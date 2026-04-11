package main

import (
	"context"
	"fmt"
	"log"
	"net"

	// Ensure this path matches your generated output directory
	pb "ouroboros/proto/generated/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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


// GetUser implements the single user lookup
func (s *authServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	log.Printf("Auth Service: Fetching User ID: %s", req.Id)

	return &pb.User{
		Id:       req.Id,
		Email:    fmt.Sprintf("user_%s@example.com", req.Id),
		Username: fmt.Sprintf("user_%s", req.Id),
	}, nil
}

// GetUsersByIds implements the batch lookup
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
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Updated registration function to match the service name in auth.proto
	pb.RegisterAuthServiceServer(s, &authServiceServer{})

	reflection.Register(s)

	log.Println("Auth Service (gRPC) running on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}