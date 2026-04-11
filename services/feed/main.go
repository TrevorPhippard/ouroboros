package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "ouroboros/proto/generated/feed"

	"feed/internal/consul"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type feedServiceServer struct {
	pb.UnimplementedFeedServiceServer
}

// GetFeed implements the main feed retrieval logic
func (s *feedServiceServer) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.GetFeedResponse, error) {
	log.Printf("Feed Service: Generating feed for User ID: %s", req.UserId)

	var items []*pb.FeedItem
	for i := 1; i <= 5; i++ {
		items = append(items, &pb.FeedItem{
			PostId: fmt.Sprintf("post_uuid_%d", i),
			Cursor: fmt.Sprintf("cursor_val_%d", i),
		})
	}

	return &pb.GetFeedResponse{
		Items: items,
	}, nil
}

func main() {
	// ✅ Consul (Docker-safe)
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

		// ✅ HTTP health check
		Check: &api.AgentServiceCheck{
			HTTP:     "http://feed-service:8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	// ✅ Register service (new error-returning version)
	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	// ✅ HTTP server (health + metrics)
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

	// ✅ gRPC server
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterFeedServiceServer(s, &feedServiceServer{})

	reflection.Register(s)

	log.Println("Feed Service (gRPC) running on :50055")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}