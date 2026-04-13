package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "ouroboros/proto/generated/auth"

	"auth/internal/consul"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type authServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *authServiceServer) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.AuthResponse, error) {
	log.Printf("SignIn attempt: %s", req.Email)

	return &pb.AuthResponse{
		Token: "mock-jwt-token",
		User: &pb.User{
			Id:       "1",
			Email:    req.Email,
			Username: "mockuser",
		},
	}, nil
}

func (s *authServiceServer) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.AuthResponse, error) {
	log.Printf("SignUp: %s", req.Email)

	return &pb.AuthResponse{
		Token: "mock-jwt-token",
		User: &pb.User{
			Id:          "new-user-id",
			Email:       req.Email,
			Username:    req.DisplayName,
			DisplayName: req.DisplayName,
		},
	}, nil
}

func (s *authServiceServer) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.SignOutResponse, error) {
	log.Println("SignOut called")

	return &pb.SignOutResponse{
		Success: true,
	}, nil
}

func (s *authServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	log.Printf("Auth Service: Fetching User ID: %s", req.Id)

	return &pb.User{
		Id:       req.Id,
		Email:    fmt.Sprintf("user_%s@example.com", req.Id),
		Username: fmt.Sprintf("user_%s", req.Id),
	}, nil
}

func (s *authServiceServer) GetUsersByIds(ctx context.Context, req *pb.GetUsersByIdsRequest) (*pb.GetUsersByIdsResponse, error) {
	log.Printf("Auth Service: Fetching %d User IDs", len(req.Ids))

	var users []*pb.User
	for _, id := range req.Ids {
		users = append(users, &pb.User{
			Id:       id,
			Email:    fmt.Sprintf("user_%s@example.com", id),
			Username: fmt.Sprintf("user_%s", id),
		})
	}

	return &pb.GetUsersByIdsResponse{Users: users}, nil
}

func main() {
	// Consul (Docker-safe address)
	addr := "consul:8500"

	agent := consul.NewAgent(&api.Config{
		Address: addr,
	})

	serviceCfg := consul.Config{
		ServiceID:   "auth-service-1",
		ServiceName: "auth-service",
		Address:     "auth-service",
		Tags:        []string{"grpc", "auth"},
		Port:        50053,

		// HTTP health check (matches new consul package)
		Check: &api.AgentServiceCheck{
			HTTP:     "http://auth-service:8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	// Register service (now returns error)
	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	// Start HTTP server (metrics + health)
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
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterAuthServiceServer(s, &authServiceServer{})

	reflection.Register(s)

	log.Println("Auth Service (gRPC) running on :50053")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}