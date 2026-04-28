package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

func loggingUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {

	start := time.Now()

	p, _ := peer.FromContext(ctx)

	log.Printf("[gRPC START] method=%s peer=%v req=%+v",
		info.FullMethod,
		p.Addr,
		req,
	)

	resp, err := handler(ctx, req)

	duration := time.Since(start)

	st, _ := status.FromError(err)

	log.Printf("[gRPC END] method=%s duration=%s status=%s err=%v",
		info.FullMethod,
		duration,
		st.Code(),
		err,
	)

	return resp, err
}

// Config constants
const (
	feedMaxSize     = 1000
	defaultPageSize = 20
	grpcPort        = ":50055"
	httpPort        = ":8080"
	redisAddr       = "redis:6379"
	kafkaAddr       = "kafka:9092"
)

// --------------------
// Models & Events
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

// --------------------
// Storage Layer
// --------------------

type FeedStore struct {
	rdb *redis.Client
}

func NewFeedStore(addr string) *FeedStore {
	return &FeedStore{rdb: redis.NewClient(&redis.Options{Addr: addr})}
}

func (s *FeedStore) AddToFeed(ctx context.Context, userID string, item *pbFeed.FeedItem) error {
	data, _ := json.Marshal(item)
	key := fmt.Sprintf("feed:%s", userID)

	pipe := s.rdb.TxPipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, feedMaxSize-1)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *FeedStore) SeedFeedIfEmpty(ctx context.Context, userID string, items []*pbFeed.FeedItem) error {
	key := fmt.Sprintf("feed:%s", userID)

	length, err := s.rdb.LLen(ctx, key).Result()
	if err != nil {
		return err
	}
	if length > 0 {
		return nil
	}

	pipe := s.rdb.TxPipeline()
	for i := len(items) - 1; i >= 0; i-- {
		data, err := json.Marshal(items[i])
		if err != nil {
			return err
		}
		pipe.LPush(ctx, key, data)
	}
	pipe.LTrim(ctx, key, 0, feedMaxSize-1)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *FeedStore) GetFeed(ctx context.Context, userID string, cursor string, limit int64) ([]*pbFeed.FeedItem, string, error) {
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > 100 {
		limit = 100
	}

	var start int64
	if cursor != "" {
		parsed, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil || parsed < 0 {
			return nil, "", status.Error(codes.InvalidArgument, "cursor must be a non-negative integer")
		}
		start = parsed
	}

	end := start + limit - 1

	results, err := s.rdb.LRange(ctx, fmt.Sprintf("feed:%s", userID), start, end).Result()
	if err != nil {
		return nil, "", err
	}

	items := make([]*pbFeed.FeedItem, 0, len(results))
	for _, r := range results {
		var item pbFeed.FeedItem
		if err := json.Unmarshal([]byte(r), &item); err != nil {
			log.Printf("feed-service: skipping malformed feed entry user_id=%s: %v", userID, err)
			continue
		}
		items = append(items, &item)
	}

	nextCursor := ""
	if len(results) == int(limit) {
		nextCursor = strconv.FormatInt(end+1, 10)
	}
	return items, nextCursor, nil
}

func (s *FeedStore) CheckAndMarkProcessed(ctx context.Context, eventID string) (bool, error) {
	key := fmt.Sprintf("processed:%s", eventID)
	// Use SetNX for atomic check-and-set idempotency
	ok, err := s.rdb.SetNX(ctx, key, "1", 24*time.Hour).Result()
	return ok, err
}

// --------------------
// Service Layer (Core Logic)
// --------------------

type FeedService struct {
	pbFeed.UnimplementedFeedServiceServer
	store  *FeedStore
	social *SocialGraph
}

func (s *FeedService) GetFeed(ctx context.Context, req *pbFeed.GetFeedRequest) (*pbFeed.GetFeedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	items, next, err := s.store.GetFeed(ctx, req.UserId, req.Cursor, int64(req.Limit))
	if err != nil {
		log.Printf("feed-service: failed to fetch feed user_id=%s: %v", req.UserId, err)
		return nil, err
	}
	return &pbFeed.GetFeedResponse{Items: items, NextCursor: next}, nil
}

// HandlePostCreated performs the "Fan-out on Write" logic
func (s *FeedService) HandlePostCreated(ctx context.Context, event PostCreatedEvent) error {
	if strings.TrimSpace(event.EventID) == "" {
		return status.Error(codes.InvalidArgument, "event_id is required")
	}
	if strings.TrimSpace(event.Data.PostID) == "" || strings.TrimSpace(event.Data.AuthorID) == "" {
		return status.Error(codes.InvalidArgument, "post event must include post_id and author_id")
	}

	// Idempotency check
	isNew, err := s.store.CheckAndMarkProcessed(ctx, event.EventID)
	if err != nil || !isNew {
		if err != nil {
			return err
		}
		return nil
	}

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

	followers, err := s.social.GetFollowers(ctx, event.Data.AuthorID)
	if err != nil {
		return err
	}

	for _, userID := range followers {
		if err := s.store.AddToFeed(ctx, userID, item); err != nil {
			log.Printf("failed to update feed for user %s: %v", userID, err)
		}
	}
	return nil
}

// --------------------
// Infrastructure / Transport
// --------------------

type SocialGraph struct{}

func (g *SocialGraph) GetFollowers(ctx context.Context, userID string) ([]string, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	return []string{"user-1", "user-2", "user-3", userID}, nil
}

func runKafkaConsumer(ctx context.Context, svc *FeedService) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaAddr},
		Topic:   "posts.created",
		GroupID: "feed-service",
	})
	defer reader.Close()

	log.Println("Kafka Consumer started...")
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			} // Normal shutdown
			log.Printf("Kafka error: %v", err)
			continue
		}

		var event PostCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("feed-service: failed to decode post-created event: %v", err)
			continue
		}

		if err := svc.HandlePostCreated(ctx, event); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("feed-service: worker error: %v", err)
		}
	}
}

func seedMockFeed(ctx context.Context, store *FeedStore) error {
	items := []*pbFeed.FeedItem{
		{
			PostId: "post-3",
			Cursor: "0",
			Post: &pbFeed.Post{
				Id:        "post-3",
				AuthorId:  "user-3",
				Content:   "Observability before optimization still wins most weeks.",
				Timestamp: "2026-04-27T09:10:00Z",
			},
		},
		{
			PostId: "post-2",
			Cursor: "1",
			Post: &pbFeed.Post{
				Id:        "post-2",
				AuthorId:  "user-2",
				Content:   "GraphQL is a lot nicer when the joins are explicit and cheap.",
				Timestamp: "2026-04-27T09:05:00Z",
			},
		},
		{
			PostId: "post-1",
			Cursor: "2",
			Post: &pbFeed.Post{
				Id:        "post-1",
				AuthorId:  "user-1",
				Content:   "Shipping the first cut of the feed service today.",
				Timestamp: "2026-04-27T09:00:00Z",
			},
		},
	}

	if err := store.SeedFeedIfEmpty(ctx, "user-1", items); err != nil {
		return err
	}

	log.Println("feed-service: ensured mock Redis feed for user-1")
	return nil
}

// --------------------
// Main Execution
// --------------------

func main() {
	// turn OS signals into context cancellation
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize Dependencies
	store := NewFeedStore(redisAddr)
	if strings.EqualFold(os.Getenv("SEED_MOCK_FEED"), "true") {
		if err := seedMockFeed(ctx, store); err != nil {
			log.Printf("feed-service: failed to seed mock feed: %v", err)
		}
	}
	service := &FeedService{
		store:  store,
		social: &SocialGraph{},
	}

	// 1. Start HTTP (Metrics/Health)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	httpSrv := &http.Server{Addr: httpPort, Handler: mux}
	go httpSrv.ListenAndServe()

	// 2. Start Kafka Consumer
	go runKafkaConsumer(ctx, service)

	// 3. Start gRPC Server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor),
	)

	pbFeed.RegisterFeedServiceServer(grpcSrv, service)
	reflection.Register(grpcSrv)
	discovery.Register("consul:8500", "feed-service", 50055)

	go func() {
		log.Printf("gRPC Server on %s", grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("gRPC server failed: %v", err)
		}
	}()

	// Wait for Shutdown Signal
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	// Cleanup
	grpcSrv.GracefulStop()
	httpSrv.Shutdown(context.Background())
	log.Println("Service stopped.")
}
