package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "ouroboros/proto/generated/post"

	"google.golang.org/grpc"
)

type postServiceServer struct {
	pb.UnimplementedPostServiceServer
}

func (s *postServiceServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.Post, error) {
	log.Printf("Post Service: Creating post for Author: %s", req.AuthorId)

	return &pb.Post{
		Id:        "new-post-uuid",
		AuthorId:  req.AuthorId,
		Content:   req.Content,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *postServiceServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.Post, error) {
	log.Printf("Post Service: Fetching Post ID: %s", req.Id)

	return &pb.Post{
		Id:        req.Id,
		AuthorId:  "mock-author-id",
		Content:   "This is a mock post content",
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *postServiceServer) GetPostsByIds(ctx context.Context, req *pb.GetPostsByIdsRequest) (*pb.GetPostsByIdsResponse, error) {
	log.Printf("Post Service: Batch fetching %d posts", len(req.Ids))

	var posts []*pb.Post
	for _, id := range req.Ids {
		posts = append(posts, &pb.Post{
			Id:        id,
			AuthorId:  "mock-author-id",
			Content:   fmt.Sprintf("Content for post %s", id),
			CreatedAt: time.Now().Format(time.RFC3339),
		})
	}

	return &pb.GetPostsByIdsResponse{Posts: posts}, nil
}

func (s *postServiceServer) GetCommentsByPostId(ctx context.Context, req *pb.GetCommentsRequest) (*pb.GetCommentsResponse, error) {
	log.Printf("Post Service: Fetching comments for Post: %s", req.PostId)

	var comments []*pb.Comment
	comments = append(comments, &pb.Comment{
		Id:        "comment-1",
		PostId:    req.PostId,
		AuthorId:  "commenter-id",
		Content:   "Great post!",
		CreatedAt: time.Now().Format(time.RFC3339),
	})

	return &pb.GetCommentsResponse{Comments: comments}, nil
}

func main() {
	// Port 50057 to keep the Ouroboros service map unique
	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPostServiceServer(s, &postServiceServer{})

	log.Println("Post Service (gRPC) running on :50057")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}