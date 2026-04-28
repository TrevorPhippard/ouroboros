package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	pb "ouroboros/proto/generated/post"
	"post/internal/models"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	DB     *gorm.DB
	Writer *kafka.Writer
}

type postCreatedEvent struct {
	EventID   string `json:"eventId"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		PostID   string `json:"postId"`
		AuthorID string `json:"authorId"`
		Content  string `json:"content"`
	} `json:"data"`
}

func (s *PostServiceServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if strings.TrimSpace(req.AuthorId) == "" {
		return nil, status.Error(codes.InvalidArgument, "author_id is required")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	post := &models.DBPost{
		ID:        uuid.NewString(),
		AuthorID:  req.AuthorId,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.DB.WithContext(ctx).Create(post).Error; err != nil {
		log.Printf("post-service: failed to create post for author=%s: %v", req.AuthorId, err)
		return nil, status.Error(codes.Internal, "failed to create post")
	}

	if err := s.publishPostCreated(ctx, post, uuid.NewString()); err != nil {
		log.Printf("post-service: failed to publish post-created event post_id=%s: %v", post.ID, err)
	}

	return toProtoPost(post), nil
}

func (s *PostServiceServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.Post, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	var post models.DBPost
	if err := s.DB.WithContext(ctx).First(&post, "id = ?", req.Id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "post not found")
		}
		log.Printf("post-service: failed to fetch post id=%s: %v", req.Id, err)
		return nil, status.Error(codes.Internal, "failed to fetch post")
	}

	return toProtoPost(&post), nil
}

func (s *PostServiceServer) GetPostsByIds(ctx context.Context, req *pb.GetPostsByIdsRequest) (*pb.GetPostsByIdsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	ids := uniqueNonEmpty(req.Ids)
	if len(ids) == 0 {
		return &pb.GetPostsByIdsResponse{Posts: []*pb.Post{}}, nil
	}

	var posts []models.DBPost
	if err := s.DB.WithContext(ctx).Where("id IN ?", ids).Find(&posts).Error; err != nil {
		log.Printf("post-service: failed to fetch %d posts: %v", len(ids), err)
		return nil, status.Error(codes.Internal, "failed to fetch posts")
	}

	postByID := make(map[string]*models.DBPost, len(posts))
	for i := range posts {
		postByID[posts[i].ID] = &posts[i]
	}

	result := make([]*pb.Post, 0, len(ids))
	for _, id := range ids {
		if post := postByID[id]; post != nil {
			result = append(result, toProtoPost(post))
		}
	}

	return &pb.GetPostsByIdsResponse{Posts: result}, nil
}

func (s *PostServiceServer) GetCommentsByPostId(ctx context.Context, req *pb.GetCommentsRequest) (*pb.GetCommentsResponse, error) {
	if req == nil || strings.TrimSpace(req.PostId) == "" {
		return nil, status.Error(codes.InvalidArgument, "post_id is required")
	}

	var comments []models.DBComment
	if err := s.DB.WithContext(ctx).
		Where("post_id = ?", req.PostId).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		log.Printf("post-service: failed to fetch comments for post_id=%s: %v", req.PostId, err)
		return nil, status.Error(codes.Internal, "failed to fetch comments")
	}

	result := make([]*pb.Comment, 0, len(comments))
	for i := range comments {
		result = append(result, &pb.Comment{
			Id:        comments[i].ID,
			PostId:    comments[i].PostID,
			AuthorId:  comments[i].AuthorID,
			Content:   comments[i].Content,
			CreatedAt: comments[i].CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &pb.GetCommentsResponse{Comments: result}, nil
}

func BootstrapFeedEvents(ctx context.Context, db *gorm.DB, writer *kafka.Writer) error {
	if db == nil || writer == nil {
		return nil
	}

	var posts []models.DBPost
	if err := db.WithContext(ctx).Order("created_at DESC").Find(&posts).Error; err != nil {
		return err
	}

	svc := &PostServiceServer{DB: db, Writer: writer}
	for i := range posts {
		if err := svc.publishPostCreated(ctx, &posts[i], "bootstrap-"+posts[i].ID); err != nil {
			return err
		}
	}

	log.Printf("post-service: published %d bootstrap post events", len(posts))
	return nil
}

func (s *PostServiceServer) publishPostCreated(ctx context.Context, post *models.DBPost, eventID string) error {
	if s.Writer == nil || post == nil {
		return nil
	}

	event := postCreatedEvent{
		EventID:   eventID,
		Type:      "PostCreated",
		Timestamp: post.CreatedAt.UTC().Format(time.RFC3339),
	}
	event.Data.PostID = post.ID
	event.Data.AuthorID = post.AuthorID
	event.Data.Content = post.Content

	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(post.AuthorID),
		Value: value,
	})
}

func toProtoPost(post *models.DBPost) *pb.Post {
	return &pb.Post{
		Id:        post.ID,
		AuthorId:  post.AuthorID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
