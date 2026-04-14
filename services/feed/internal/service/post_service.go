package service

import (
	pb "ouroboros/proto/generated/feed"

	"github.com/segmentio/kafka-go"
)

type FeedServiceServer struct {
	pb.UnimplementedFeedServiceServer
	Writer *kafka.Writer
}

// Implement your gRPC methods here (e.g., CreatePost, etc.)
