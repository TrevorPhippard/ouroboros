package service

import (
	pb "ouroboros/proto/generated/post"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	DB     *gorm.DB
	Writer *kafka.Writer
}

// Implement your gRPC methods here (e.g., CreatePost, etc.)
