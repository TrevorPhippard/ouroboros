package main

import (
	"context"
	"log"
	"net"
	"net/http"

	pb "ouroboros/proto/generated/connection"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"connection/internal/consul"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type connServiceServer struct {
	pb.UnimplementedConnectionServiceServer
}

func (s *connServiceServer) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.FollowUserResponse, error) {
	log.Printf("Connection Service: %s is following %s", req.FollowerId, req.FolloweeId)
	return &pb.FollowUserResponse{Success: true}, nil
}

func (s *connServiceServer) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.UnfollowUserResponse, error) {
	log.Printf("Connection Service: %s unfollowed %s", req.FollowerId, req.FolloweeId)
	return &pb.UnfollowUserResponse{Success: true}, nil
}

func (s *connServiceServer) GetFollowersCount(ctx context.Context, req *pb.UserRequest) (*pb.CountResponse, error) {
	log.Printf("Connection Service: Getting followers count for %s", req.UserId)
	return &pb.CountResponse{Count: 42}, nil
}

func (s *connServiceServer) GetFollowingCount(ctx context.Context, req *pb.UserRequest) (*pb.CountResponse, error) {
	log.Printf("Connection Service: Getting following count for %s", req.UserId)
	return &pb.CountResponse{Count: 10}, nil
}

func (s *connServiceServer) IsFollowing(ctx context.Context, req *pb.IsFollowingRequest) (*pb.IsFollowingResponse, error) {
	log.Printf("Connection Service: Checking if %s follows %s", req.FollowerId, req.FolloweeId)
	return &pb.IsFollowingResponse{IsFollowing: true}, nil
}

func main() {
	// Correct Consul address for Docker
	addr := "consul:8500"

	agent := consul.NewAgent(&api.Config{
		Address: addr,
	})

	serviceCfg := consul.Config{
		ServiceID:   "connection-service-1",
		ServiceName: "connection-service",
		Address:     "connection-service",
		Tags:        []string{"grpc", "connection"},
		Port:        50054,

		// Proper HTTP health check (NO TTL)
		Check: &api.AgentServiceCheck{
			HTTP:     "http://connection-service:8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}
	// Register service with health check
	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	// Start HTTP server for health + metrics
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
	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterConnectionServiceServer(s, &connServiceServer{})

	reflection.Register(s)

	log.Println("Connection Service (gRPC) running on :50054")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}