package main

import (
	"context"
	pb "ouroboros/proto/generated/post"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostServiceServer implements the PostService gRPC interface
type PostServiceServer struct {
	pb.UnimplementedPostServiceServer
	DB *gorm.DB
}



func (s *PostServiceServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	post := DBPost{
		ID:        uuid.NewString(),
		AuthorID:  req.AuthorId,
		Content:   req.Content,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.DB.Create(&post).Error; err != nil {
		return nil, err
	}

	return &pb.Post{
		Id:        post.ID,
		AuthorId:  post.AuthorID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *PostServiceServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.Post, error) {
	var post DBPost

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

// GetPosts pulls real data from the isolated post_db
func (s *PostServiceServer) GetPosts(ctx context.Context, req *pb.GetPostsByIdsRequest) (*pb.GetPostsByIdsResponse, error) {
	var dbPosts []DBPost

	// Use GORM to find posts by the requested AuthorID
	result := s.DB.Where("author_id = ?", req.Ids).Find(&dbPosts)
	if result.Error != nil {
		return nil, result.Error
	}

	// Map GORM models to Protobuf messages
	pbPosts := make([]*pb.Post, 0, len(dbPosts))

	for _, p := range dbPosts {
		pbPosts = append(pbPosts, &pb.Post{
			Id:       p.ID,
			AuthorId: p.AuthorID,
			Content:  p.Content,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetPostsByIdsResponse{Posts: pbPosts}, nil
}



// func (s *PostServiceServer) GetCommentsByPostId(ctx context.Context, req *pb.GetCommentsRequest) (*pb.GetCommentsResponse, error) {
// 	var comments []DBComment

// 	if err := s.DB.Where("post_id = ?", req.PostId).Find(&comments).Error; err != nil {
// 		return nil, err
// 	}

// 	pbComments := make([]*pb.Comment, 0, len(comments))
// 	for _, c := range comments {
// 		pbComments = append(pbComments, &pb.Comment{
// 			Id:        c.ID,
// 			PostId:    c.PostID,
// 			AuthorId:  c.AuthorID,
// 			Content:   c.Content,
// 			CreatedAt: c.CreatedAt.Format(time.RFC3339),
// 		})
// 	}

// 	return &pb.GetCommentsResponse{Comments: pbComments}, nil
// }
