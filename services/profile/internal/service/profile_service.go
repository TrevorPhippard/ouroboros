package service

import (
	pb "ouroboros/proto/generated/profile"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type ProfileServiceServer struct {
	pb.UnimplementedProfileServiceServer
	DB     *gorm.DB
	Writer *kafka.Writer
}

// Implement your gRPC methods here (e.g., CreatePost, etc.)
