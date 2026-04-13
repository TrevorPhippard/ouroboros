package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"

	pbFeed "ouroboros/proto/generated/feed"

	"feed/internal/consul"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type feedServiceServer struct {
	pbFeed.UnimplementedFeedServiceServer
	mu   sync.RWMutex
	feed map[string][]*pbFeed.FeedItem
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

// GetFeed now returns real (event-built) data
func (s *feedServiceServer) GetFeed(ctx context.Context, req *pbFeed.GetFeedRequest) (*pbFeed.GetFeedResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := s.feed[req.UserId]

	return &pbFeed.GetFeedResponse{
		Items: items,
	}, nil
}

func main() {
	addr := "consul:8500"

	agent := consul.NewAgent(&api.Config{
		Address: addr,
	})

	serviceCfg := consul.Config{
		ServiceID:   "feed-service-1",
		ServiceName: "feed-service",
		Address:     "feed-service",
		Tags:        []string{"grpc", "feed"},
		Port:        50055,
		Check: &api.AgentServiceCheck{
			HTTP:     "http://feed-service:8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	feedServer := &feedServiceServer{
		feed: make(map[string][]*pbFeed.FeedItem),
	}

	// 🔥 Kafka Consumer
	go func() {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"kafka:9092"},
			Topic:   "posts.created",
			GroupID: "feed-service",
		})

		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Println("kafka read error:", err)
				continue
			}

			var event PostCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Println("failed to unmarshal event")
				continue
			}

			log.Println("Received event:", event.Type)

			item := &pbFeed.FeedItem{
				PostId: event.Data.PostID,
				Cursor: event.Data.PostID,
				Post: &pbFeed.Post{
					Id:        event.Data.PostID,
					AuthorId:  event.Data.AuthorID,
					Content:   event.Data.Content,
					Timestamp: event.Timestamp,
				},
			}

			feedServer.mu.Lock()

			// For now: global feed (all users get all posts)
			for userID := range feedServer.feed {
				feedServer.feed[userID] = append([]*pbFeed.FeedItem{item}, feedServer.feed[userID]...)
			}

			// Ensure at least one user exists (for testing)
			if len(feedServer.feed) == 0 {
				feedServer.feed["test-user"] = []*pbFeed.FeedItem{item}
			}

			feedServer.mu.Unlock()
		}
	}()

	// HTTP server
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		log.Println("HTTP server running on :8080 (metrics + health)")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	// gRPC server
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pbFeed.RegisterFeedServiceServer(s, feedServer)

	reflection.Register(s)

	log.Println("Feed Service (gRPC) running on :50055")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
