package main

import (
	"context"
	"log"
	"net"

	pb "ouroboros/proto/generated/connection"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type connServiceServer struct {
	pb.UnimplementedConnectionServiceServer
}

func (s *connServiceServer) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.FollowUserResponse, error) {
	log.Printf("Connection Service: %s is following %s", req.FollowerId, req.FolloweeId)
	// Logic for Neo4j or database insertion would go here
	return &pb.FollowUserResponse{Success: true}, nil
}

func (s *connServiceServer) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.UnfollowUserResponse, error) {
	log.Printf("Connection Service: %s unfollowed %s", req.FollowerId, req.FolloweeId)
	return &pb.UnfollowUserResponse{Success: true}, nil
}

func (s *connServiceServer) GetFollowersCount(ctx context.Context, req *pb.UserRequest) (*pb.CountResponse, error) {
	log.Printf("Connection Service: Getting followers count for %s", req.UserId)
	return &pb.CountResponse{Count: 42}, nil // Mock count
}

func (s *connServiceServer) GetFollowingCount(ctx context.Context, req *pb.UserRequest) (*pb.CountResponse, error) {
	log.Printf("Connection Service: Getting following count for %s", req.UserId)
	return &pb.CountResponse{Count: 10}, nil // Mock count
}

func (s *connServiceServer) IsFollowing(ctx context.Context, req *pb.IsFollowingRequest) (*pb.IsFollowingResponse, error) {
	log.Printf("Connection Service: Checking if %s follows %s", req.FollowerId, req.FolloweeId)
	return &pb.IsFollowingResponse{IsFollowing: true}, nil
}

func main() {
	// Note: Ensure this port doesn't conflict with your Auth service (:50053)
	// Usually, Connection Service might sit on :50054 or similar
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