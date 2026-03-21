package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "ouroboros/proto/generated/feed"

	"google.golang.org/grpc"
)

type feedServiceServer struct {
	pb.UnimplementedFeedServiceServer
}

// GetFeed implements the main feed retrieval logic
func (s *feedServiceServer) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.GetFeedResponse, error) {
	log.Printf("Feed Service: Generating feed for User ID: %s", req.UserId)

	// Mocking some feed items.
	// In a real scenario, this would involve calling the Connection service
	// to get "following" IDs and then the Post service to get content.
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
	// Shifted to :50055 to avoid conflicts with Auth (:50053) and Connection (:50054)
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterFeedServiceServer(s, &feedServiceServer{})

	log.Println("Feed Service (gRPC) running on :50055")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}