package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"feed/internal/discovery"
	pbFeed "ouroboros/proto/generated/feed"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// --------------------
// Config
// --------------------

const (
	feedMaxSize     = 1000
	defaultPageSize = 20
	grpcPort        = ":50055"
	httpPort        = ":8080"
	redisAddr       = "redis:6379"
	kafkaAddr       = "kafka:9092"

	batchSize       = 100
	maxConcurrency  = 8  // per job
	globalWorkers   = 16 // global worker pool
	globalQueueSize = 1000
	rateLimitEvery  = 2 * time.Millisecond // ~500 ops/sec
)

// --------------------
// Global Backpressure + Rate Limiting
// --------------------

var (
	fanoutQueue = make(chan func(), globalQueueSize)
	rateLimiter = time.Tick(rateLimitEvery)
)

func startGlobalWorkers() {
	for i := 0; i < globalWorkers; i++ {
		go func() {
			for job := range fanoutQueue {
				job()
			}
		}()
	}
}

// --------------------
// Interceptors
// --------------------

func loggingUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {

	start := time.Now()
	p, _ := peer.FromContext(ctx)

	log.Printf("[gRPC START] method=%s peer=%v", info.FullMethod, p.Addr)

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	st, _ := status.FromError(err)

	log.Printf("[gRPC END] method=%s duration=%s status=%s err=%v",
		info.FullMethod, duration, st.Code(), err,
	)

	return resp, err
}

// --------------------
// Models
// --------------------

type PostCreatedEvent struct {
	EventID   string `json:"eventId"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		PostID   string `json:"postId"`
		AuthorID string `json:"authorId"`
		Content  string `json:"content"`
	} `json:"data"`
}

type FanoutJob struct {
	EventID   string   `json:"event_id"`
	PostID    string   `json:"post_id"`
	AuthorID  string   `json:"author_id"`
	Followers []string `json:"followers"`

	BatchSize int          `json:"batch_size"`
	Completed map[int]bool `json:"completed"`
	Cursor    int          `json:"cursor"`
}

// --------------------
// Storage Layer
// --------------------

type FeedStore struct {
	rdb *redis.Client
}

func NewFeedStore(addr string) *FeedStore {
	return &FeedStore{
		rdb: redis.NewClient(&redis.Options{
			Addr:         addr,
			PoolSize:     50,
			MinIdleConns: 10,
		}),
	}
}

func feedKey(userID string) string {
	return fmt.Sprintf("feed:%s", userID)
}

func jobKey(eventID string) string {
	return "fanout:job:" + eventID
}

func parseTimestamp(ts string) (float64, error) {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return 0, err
	}
	return float64(t.UnixNano()), nil
}

// --------------------
// Redis Operations
// --------------------

func (s *FeedStore) FanoutBatch(
	ctx context.Context,
	userIDs []string,
	item *pbFeed.FeedItem,
) error {

	score, err := parseTimestamp(item.Post.Timestamp)
	if err != nil {
		return err
	}

	pipe := s.rdb.Pipeline()

	for _, userID := range userIDs {
		key := feedKey(userID)

		pipe.ZAdd(ctx, key, &redis.Z{
			Score:  score,
			Member: item.PostId, // ✅ idempotent
		})

		pipe.ZRemRangeByRank(ctx, key, 0, -(feedMaxSize + 1))
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (s *FeedStore) SaveJob(ctx context.Context, job *FanoutJob) error {
	data, _ := json.Marshal(job)
	return s.rdb.Set(ctx, jobKey(job.EventID), data, 24*time.Hour).Err()
}

func (s *FeedStore) LoadJob(ctx context.Context, eventID string) (*FanoutJob, error) {
	data, err := s.rdb.Get(ctx, jobKey(eventID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var job FanoutJob
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// --------------------
// Read Path
// --------------------

func (s *FeedStore) GetFeed(
	ctx context.Context,
	userID string,
	cursor string,
	limit int64,
) ([]*pbFeed.FeedItem, string, error) {

	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > 100 {
		limit = 100
	}

	key := feedKey(userID)

	maxScore := "+inf"
	if cursor != "" {
		maxScore = cursor
	}

	// 1. FIXED: Use WithScores to get the actual timestamp data alongside the PostID
	results, err := s.rdb.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Max:   maxScore,
		Min:   "-inf",
		Count: limit,
	}).Result()

	if err != nil {
		return nil, "", err
	}

	items := make([]*pbFeed.FeedItem, 0, len(results))

	var lastScore float64

	for _, r := range results {
		// 2. FIXED: Type assert the member to a string safely
		postID, ok := r.Member.(string)
		if !ok {
			continue // Handle unexpected data types gracefully
		}

		items = append(items, &pbFeed.FeedItem{
			PostId: postID,
		})

		// 3. FIXED: Capture the actual score of the item from Redis
		lastScore = r.Score
	}

	nextCursor := ""
	if len(items) == int(limit) {
		// 4. FIXED: Use "%.0f" to prevent large UnixNano floats from
		// degrading into scientific notation (e.g., 1.682e+18), which Redis rejects.
		nextCursor = fmt.Sprintf("%.0f", lastScore-1)
	}

	return items, nextCursor, nil
}

// --------------------
// Service Layer
// --------------------

type FeedService struct {
	pbFeed.UnimplementedFeedServiceServer
	store  *FeedStore
	social *SocialGraph
}

// --------------------
// Parallel Fanout Engine
// --------------------

func (s *FeedService) ProcessFanoutParallel(
	ctx context.Context,
	job *FanoutJob,
	item *pbFeed.FeedItem,
) error {

	totalBatches := (len(job.Followers) + job.BatchSize - 1) / job.BatchSize

	sem := make(chan struct{}, maxConcurrency)
	errCh := make(chan error, totalBatches)

	var mu sync.Mutex

	for batchIdx := 0; batchIdx < totalBatches; batchIdx++ {

		if job.Completed[batchIdx] {
			continue
		}

		sem <- struct{}{}

		start := batchIdx * job.BatchSize
		end := start + job.BatchSize
		if end > len(job.Followers) {
			end = len(job.Followers)
		}

		batch := job.Followers[start:end]

		fanoutQueue <- func() {

			defer func() { <-sem }()

			<-rateLimiter

			err := s.store.FanoutBatch(ctx, batch, item)
			if err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			job.Completed[batchIdx] = true

			for job.Completed[job.Cursor] {
				job.Cursor++
			}

			_ = s.store.SaveJob(ctx, job)
			mu.Unlock()
		}
	}

	// wait for all
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// --------------------
// Event Handler
// --------------------

func (s *FeedService) HandlePostCreated(ctx context.Context, event PostCreatedEvent) error {

	item := &pbFeed.FeedItem{
		PostId: event.Data.PostID,
		Cursor: event.Timestamp,
		Post: &pbFeed.Post{
			Id:        event.Data.PostID,
			AuthorId:  event.Data.AuthorID,
			Content:   event.Data.Content,
			Timestamp: event.Timestamp,
		},
	}

	job, err := s.store.LoadJob(ctx, event.EventID)
	if err != nil {
		return err
	}

	if job == nil {
		followers, err := s.social.GetFollowers(ctx, event.Data.AuthorID)
		if err != nil {
			return err
		}

		job = &FanoutJob{
			EventID:   event.EventID,
			PostID:    event.Data.PostID,
			AuthorID:  event.Data.AuthorID,
			Followers: followers,
			BatchSize: batchSize,
			Completed: make(map[int]bool),
			Cursor:    0,
		}
	}

	return s.ProcessFanoutParallel(ctx, job, item)
}

// --------------------
// gRPC
// --------------------

func (s *FeedService) GetFeed(ctx context.Context, req *pbFeed.GetFeedRequest) (*pbFeed.GetFeedResponse, error) {
	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	items, next, err := s.store.GetFeed(ctx, req.UserId, req.Cursor, int64(req.Limit))
	if err != nil {
		return nil, err
	}

	return &pbFeed.GetFeedResponse{
		Items:      items,
		NextCursor: next,
	}, nil
}

// --------------------
// Infra
// --------------------

type SocialGraph struct{}

func (g *SocialGraph) GetFollowers(ctx context.Context, userID string) ([]string, error) {
	return []string{"user-1", "user-2", "user-3", userID}, nil
}

// --------------------
// Kafka Consumer (SAFE)
// --------------------

func runKafkaConsumer(ctx context.Context, svc *FeedService) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaAddr},
		Topic:          "posts.created",
		GroupID:        "feed-service",
		CommitInterval: 0,
	})
	defer reader.Close()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		var event PostCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}

		err = svc.HandlePostCreated(ctx, event)
		if err != nil {
			log.Println("fanout failed:", err)
			continue // no commit
		}

		_ = reader.CommitMessages(ctx, msg)
	}
}

// --------------------
// Main
// --------------------

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startGlobalWorkers()

	store := NewFeedStore(redisAddr)

	service := &FeedService{
		store:  store,
		social: &SocialGraph{},
	}

	// HTTP
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	httpSrv := &http.Server{Addr: httpPort, Handler: mux}
	go httpSrv.ListenAndServe()

	// Kafka
	go runKafkaConsumer(ctx, service)

	// gRPC
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor),
	)

	pbFeed.RegisterFeedServiceServer(grpcSrv, service)
	reflection.Register(grpcSrv)
	discovery.Register("consul:8500", "feed-service", 50055)

	go grpcSrv.Serve(lis)

	<-ctx.Done()

	grpcSrv.GracefulStop()
	httpSrv.Shutdown(context.Background())
}
