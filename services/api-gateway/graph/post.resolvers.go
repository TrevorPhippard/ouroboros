package graph

import (
	"context"
	"fmt"

	"api-gateway/graph/model"

	postpb "ouroboros/proto/generated/post"
)

// CreatePost is the resolver for the createPost mutation.
func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	userID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if input.UserId != userID {
		return nil, fmt.Errorf("forbidden: post author must be the authenticated user")
	}

	res, err := r.PostClient.CreatePost(ctx, &postpb.CreatePostRequest{
		AuthorId: userID,
		Content:  input.Content,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("post service returned no post")
	}

	post := mapPost(res)
	if err := r.attachAuthorsToPosts(ctx, []*model.Post{post}); err != nil {
		return nil, fmt.Errorf("post was created but its author could not be loaded: %w", err)
	}
	return post, nil
}

// CreateComment is the resolver for the createComment field.
func (r *mutationResolver) CreateComment(
	ctx context.Context,
	input model.CreateCommentInput,
) (*model.Comment, error) {
	userID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if input.UserId != userID {
		return nil, fmt.Errorf("forbidden: comment author must be the authenticated user")
	}

	// Validate the post before attempting to create a comment.
	postRes, err := r.PostClient.GetPost(ctx, &postpb.GetPostRequest{
		Id: input.PostID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find post for comment: %w", err)
	}
	if postRes == nil {
		return nil, fmt.Errorf("post service returned no post")
	}

	// Verify that the author can be resolved.
	usersByID, err := r.loadUsersByID(ctx, []string{userID})
	if err != nil {
		return nil, fmt.Errorf("failed to load comment author: %w", err)
	}

	author := usersByID[userID]
	if author == nil {
		return nil, fmt.Errorf("comment author not found")
	}

	// Persist the comment through the post service.
	res, err := r.PostClient.CreateComment(ctx, &postpb.CreateCommentRequest{
		PostId:   input.PostID,
		AuthorId: userID,
		Content:  input.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("post service returned no comment")
	}

	return &model.Comment{
		ID:        res.Id,
		PostID:    res.PostId,
		UserId:    res.AuthorId,
		Content:   res.Content,
		CreatedAt: res.CreatedAt,
		Author:    author,
	}, nil
}

// LikePost is the resolver for the likePost field.
func (r *mutationResolver) LikePost(
	ctx context.Context,
	postID string,
) (*model.Post, error) {
	userID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	likeRes, err := r.PostClient.LikePost(ctx, &postpb.LikePostRequest{
		PostId: postID,
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to like post: %w", err)
	}
	if likeRes == nil || !likeRes.Success {
		return nil, fmt.Errorf("post could not be liked")
	}

	res, err := r.PostClient.GetPost(ctx, &postpb.GetPostRequest{
		Id: postID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch liked post: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("post service returned no post")
	}

	post := mapPost(res)

	if err := r.attachAuthorsToPosts(ctx, []*model.Post{post}); err != nil {
		return nil, fmt.Errorf("failed to hydrate liked post author: %w", err)
	}

	return post, nil
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	res, err := r.PostClient.GetPost(ctx, &postpb.GetPostRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	if res == nil {
		return nil, nil
	}

	post := mapPost(res)

	if err := r.attachAuthorsToPosts(ctx, []*model.Post{post}); err != nil {
		return nil, fmt.Errorf("failed to hydrate post author: %w", err)
	}
	return post, nil
}

// PostsByIds is the resolver for the postsByIds field.
func (r *queryResolver) PostsByIds(ctx context.Context, ids []string) ([]*model.Post, error) {
	res, err := r.PostClient.GetPostsByIds(ctx, &postpb.GetPostsByIdsRequest{Ids: ids})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	if res == nil {
		return []*model.Post{}, nil
	}

	posts := make([]*model.Post, 0, len(res.Posts))
	for _, p := range res.Posts {
		if p == nil {
			continue
		}
		posts = append(posts, mapPost(p))

	}
	if err := r.attachAuthorsToPosts(ctx, posts); err != nil {
		return nil, fmt.Errorf("failed to hydrate posts: %w", err)
	}
	return posts, nil
}

// Posts is currently unsupported by the Post service: the available gateway
// client exposes GetPost and GetPostsByIds, but no list-all RPC. Return an
// explicit error rather than silently pretending that the list is empty.
func (r *queryResolver) Posts(ctx context.Context, limit *int32) ([]*model.Post, error) {
	return nil, fmt.Errorf("posts query is unavailable: Post service does not expose a list-posts RPC")
}

func mapPost(p *postpb.Post) *model.Post {
	if p == nil {
		return nil
	}

	return &model.Post{
		ID:        p.Id,
		Content:   p.Content,
		CreatedAt: p.CreatedAt,
		UserId:    p.Id,
		Comments:  []*model.Comment{},
	}
}
