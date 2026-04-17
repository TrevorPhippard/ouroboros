package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"feed/internal/discovery"
	pbFeed "ouroboros/proto/generated/feed"
	pbPost "ouroboros/proto/generated/post"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	feedMaxItems = 100
	redisAddr    = "redis:6379"
	kafkaAddr    = "kafka:9092"
	postSvcAddr  = "post-service:50056"
)

type feedServiceServer struct {
	pbFeed.UnimplementedFeedServiceServer
	rdb        *redis.Client
	postClient pbPost.PostServiceClient
}

type PostCreatedEvent struct {
	EventID   string `json:"eventId"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		PostID   string `json:"postId"`
		AuthorID string `json:"authorId"`
		Content  string `json:"content"`
	} `json:"data"`
}

// --- gRPC HANDLERS ---

func (s *feedServiceServer) GetFeed(ctx context.Context, req *pbFeed.GetFeedRequest) (*pbFeed.GetFeedResponse, error) {
	feedKey := fmt.Sprintf("feed:%s", req.UserId)

	// 1. Fetch the "pointers"
	postIDs, err := s.rdb.LRange(ctx, feedKey, 0, feedMaxItems-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}

	if len(postIDs) == 0 {
		return &pbFeed.GetFeedResponse{Items: []*pbFeed.FeedItem{}}, nil
	}

	// 2. Hydrate the IDs into full Post objects
	items := s.hydrateFeed(ctx, postIDs)

	return &pbFeed.GetFeedResponse{Items: items}, nil
}

// --- BUSINESS LOGIC (ABSTRACTIONS) ---

func (s *feedServiceServer) hydrateFeed(ctx context.Context, postIDs []string) []*pbFeed.FeedItem {
	entityKeys := make([]string, len(postIDs))
	for i, id := range postIDs {
		entityKeys[i] = fmt.Sprintf("p:%s", id)
	}

	contents, err := s.rdb.MGet(ctx, entityKeys...).Result()
	if err != nil {
		log.Printf("MGet error: %v", err)
		return nil
	}

	var items []*pbFeed.FeedItem
	for i, c := range contents {
		postID := postIDs[i]

		if c != nil {
			if item, err := unmarshalFeedItem(c.(string)); err == nil {
				items = append(items, item)
				continue
			}
		}

		// Cache Miss: Self-Heal
		if item := s.fetchAndCachePost(ctx, postID); item != nil {
			items = append(items, item)
		}
	}
	return items
}

func (s *feedServiceServer) fetchAndCachePost(ctx context.Context, postID string) *pbFeed.FeedItem {
	resp, err := s.postClient.GetPost(ctx, &pbPost.GetPostRequest{Id: postID})
	if err != nil {
		log.Printf("Fallback failed for %s: %v", postID, err)
		return nil
	}

	item := &pbFeed.FeedItem{
		PostId: resp.Post.Id,
		Post: &pbFeed.Post{
			Id:        resp.Post.Id,
			AuthorId:  resp.Post.AuthorId,
			Content:   resp.Post.Content,
			Timestamp: resp.Post.Timestamp,
		},
	}

	// Async update to Redis so we don't block the current request
	go func() {
		data, _ := protojson.Marshal(item)
		s.rdb.Set(context.Background(), fmt.Sprintf("p:%s", postID), data, 0)
	}()

	return item
}

func (s *feedServiceServer) handlePostCreated(ctx context.Context, event PostCreatedEvent) {
	// 1. Update Entity Cache
	item := &pbFeed.FeedItem{
		PostId: event.Data.PostID,
		Post: &pbFeed.Post{
			Id:        event.Data.PostID,
			AuthorId:  event.Data.AuthorID,
			Content:   event.Data.Content,
			Timestamp: event.Timestamp,
		},
	}
	itemJSON, _ := protojson.Marshal(item)
	s.rdb.Set(ctx, fmt.Sprintf("p:%s", event.Data.PostID), itemJSON, 0)

	// 2. Fan-out IDs to followers (Simulated list)
	followers := []string{"test-user", "another-user"}
	for _, userID := range followers {
		key := fmt.Sprintf("feed:%s", userID)
		pipe := s.rdb.Pipeline()
		pipe.LPush(ctx, key, event.Data.PostID)
		pipe.LTrim(ctx, key, 0, feedMaxItems-1)
		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("Fan-out failed for %s: %v", userID, err)
		}
	}
}

func runKafkaConsumer(s *feedServiceServer) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaAddr},
		Topic:   "posts.created",
		GroupID: "feed-service",
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Kafka error: %v", err)
			continue
		}

		var event PostCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err == nil {
			s.handlePostCreated(context.Background(), event)
		}
	}
}

// --- HELPERS ---

func unmarshalFeedItem(data string) (*pbFeed.FeedItem, error) {
	item := &pbFeed.FeedItem{}
	err := protojson.Unmarshal([]byte(data), item)
	return item, err
}

func mustDialGRPC(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", addr, err)
	}
	return conn
}

func startHTTPServer() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	log.Println("Metrics/Health on :8080")
	http.ListenAndServe(":8080", nil)
}

func startGRPCServer(server *feedServiceServer) {
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pbFeed.RegisterFeedServiceServer(s, server)
	reflection.Register(s)
	log.Println("Feed Service gRPC on :50055")
	s.Serve(lis)
}

func main() {
	discovery.Register("consul:8500", "feed-service", 50055)

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	postConn := mustDialGRPC(postSvcAddr)
	defer postConn.Close()

	server := &feedServiceServer{
		rdb:        rdb,
		postClient: pbPost.NewPostServiceClient(postConn),
	}

	// Start Background Tasks
	go runKafkaConsumer(server)
	go startHTTPServer()

	// Start Primary gRPC Service
	startGRPCServer(server)
}
