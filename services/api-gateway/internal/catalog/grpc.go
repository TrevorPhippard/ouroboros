//go:build postext

package catalog

import (
	"context"
	"fmt"

	"api-gateway/graph/model"
	"api-gateway/internal/clients"
	"api-gateway/internal/mapping"

	"google.golang.org/grpc"

	connpb "ouroboros/proto/generated/connection"
	postpb "ouroboros/proto/generated/post"
)

// Compiled only with `-tags postext`, once the Post and Connection services
// expose the RPCs described in docs/proto-extensions.md and their Go code has
// been regenerated. The narrow client interfaces below are satisfied by the
// generated PostServiceClient / ConnectionServiceClient.

type postListClient interface {
	ListPosts(ctx context.Context, in *postpb.ListPostsRequest, opts ...grpc.CallOption) (*postpb.ListPostsResponse, error)
	GetPostsByAuthors(ctx context.Context, in *postpb.GetPostsByAuthorsRequest, opts ...grpc.CallOption) (*postpb.GetPostsByAuthorsResponse, error)
	GetCommentsByPostIds(ctx context.Context, in *postpb.GetCommentsByPostIdsRequest, opts ...grpc.CallOption) (*postpb.GetCommentsByPostIdsResponse, error)
}

type recommendationClient interface {
	GetRecommendations(ctx context.Context, in *connpb.GetRecommendationsRequest, opts ...grpc.CallOption) (*connpb.GetRecommendationsResponse, error)
}

// FromClients wires the gRPC-backed capabilities.
func FromClients(c clients.Set) (PostCatalog, Recommender) {
	return NewGRPCCatalog(c.Post), NewGRPCRecommender(c.Connection)
}

// NewGRPCCatalog adapts the Post service to PostCatalog.
func NewGRPCCatalog(client postListClient) PostCatalog { return grpcCatalog{client: client} }

// NewGRPCRecommender adapts the Connection service to Recommender.
func NewGRPCRecommender(client recommendationClient) Recommender {
	return grpcRecommender{client: client}
}

type grpcCatalog struct{ client postListClient }

func (c grpcCatalog) ListPosts(ctx context.Context, limit int32) ([]*model.Post, error) {
	res, err := c.client.ListPosts(ctx, &postpb.ListPostsRequest{Limit: limit})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	posts := make([]*model.Post, 0, len(res.Posts))
	for _, p := range res.Posts {
		if p != nil {
			posts = append(posts, mapping.Post(p))
		}
	}
	return posts, nil
}

func (c grpcCatalog) PostsByAuthors(ctx context.Context, authorIDs []string, limit int32) (map[string][]*model.Post, error) {
	res, err := c.client.GetPostsByAuthors(ctx, &postpb.GetPostsByAuthorsRequest{
		AuthorIds:      authorIDs,
		LimitPerAuthor: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string][]*model.Post, len(authorIDs))
	if res == nil {
		return out, nil
	}
	for _, group := range res.Authors {
		if group == nil {
			continue
		}
		for _, p := range group.Posts {
			if p != nil {
				out[group.AuthorId] = append(out[group.AuthorId], mapping.Post(p))
			}
		}
	}
	return out, nil
}

func (c grpcCatalog) CommentsByPosts(ctx context.Context, postIDs []string, limit int32) (map[string][]*model.Comment, error) {
	res, err := c.client.GetCommentsByPostIds(ctx, &postpb.GetCommentsByPostIdsRequest{
		PostIds:      postIDs,
		LimitPerPost: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string][]*model.Comment, len(postIDs))
	if res == nil {
		return out, nil
	}
	for _, group := range res.Posts {
		if group == nil {
			continue
		}
		for _, cm := range group.Comments {
			if cm != nil {
				out[group.PostId] = append(out[group.PostId], mapping.Comment(cm))
			}
		}
	}
	return out, nil
}

type grpcRecommender struct{ client recommendationClient }

func (g grpcRecommender) RecommendedUserIDs(ctx context.Context, userID string, limit int32) ([]string, error) {
	res, err := g.client.GetRecommendations(ctx, &connpb.GetRecommendationsRequest{UserId: userID, Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("get recommendations: %w", err)
	}
	if res == nil {
		return nil, nil
	}
	return res.UserIds, nil
}
