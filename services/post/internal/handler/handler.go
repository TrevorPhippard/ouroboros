package handler

import (
	"context"
	"encoding/json"
	pb "ouroboros/proto/generated/post"
	"post/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

// PostServiceServer implements the PostService gRPC interface
type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	DB     *gorm.DB
	Writer *kafka.Writer
}

type PostCreatedEvent struct {
	EventID   string    `json:"eventId"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		PostID   string `json:"postId"`
		AuthorID string `json:"authorId"`
		Content  string `json:"content"`
	} `json:"data"`
}

func (s *PostServiceServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	post := &models.DBPost{
		ID:        uuid.NewString(),
		AuthorID:  req.AuthorId,
		Content:   req.Content,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.DB.Create(&post).Error; err != nil {
		return nil, err
	}

	// Create event
	event := PostCreatedEvent{
		EventID:   uuid.NewString(),
		Type:      "PostCreated",
		Timestamp: time.Now().UTC(),
	}
	event.Data.PostID = post.ID
	event.Data.AuthorID = post.AuthorID
	event.Data.Content = post.Content

	value, _ := json.Marshal(event)

	// Publish to Kafka
	err := s.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(post.AuthorID),
		Value: value,
	})

	if err != nil {
		println("failed to publish event:", err.Error())
	}

	return &pb.Post{
		Id:        post.ID,
		AuthorId:  post.AuthorID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *PostServiceServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.Post, error) {
	var post *models.DBPost

	if err := s.DB.First(&post, "id = ?", req.Id).Error; err != nil {
		return nil, err
	}

	return &pb.Post{
		Id:        post.ID,
		AuthorId:  post.AuthorID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *PostServiceServer) GetPosts(ctx context.Context, req *pb.GetPostsByIdsRequest) (*pb.GetPostsByIdsResponse, error) {
	var dbPosts []*models.DBPost

	result := s.DB.Where("author_id = ?", req.Ids).Find(&dbPosts)
	if result.Error != nil {
		return nil, result.Error
	}

	pbPosts := make([]*pb.Post, 0, len(dbPosts))

	for _, p := range dbPosts {
		pbPosts = append(pbPosts, &pb.Post{
			Id:        p.ID,
			AuthorId:  p.AuthorID,
			Content:   p.Content,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetPostsByIdsResponse{Posts: pbPosts}, nil
}
