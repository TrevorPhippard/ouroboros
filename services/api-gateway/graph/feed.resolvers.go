package graph

import (
	"context"
	"fmt"
	"log"

	"api-gateway/graph/model"

	feedpb "ouroboros/proto/generated/feed"
	postpb "ouroboros/proto/generated/post"
)

// Feed is the resolver for the feed query.
func (r *queryResolver) Feed(ctx context.Context, userID string, limit *int32, cursor *string) (*model.FeedResponse, error) {
	authenticatedUserID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if authenticatedUserID != userID {
		return nil, fmt.Errorf("forbidden: cannot read another user's feed")
	}

	reqLimit := int32(20)
	if limit != nil {
		reqLimit = *limit
	}

	reqCursor := ""
	if cursor != nil {
		reqCursor = *cursor
	}

	// 1. Get the feed (which just contains Post IDs and Cursors)
	// Give me the next 20 Post IDs for User A
	if reqLimit < 1 || reqLimit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100")
	}
	feedRes, err := r.FeedClient.GetFeed(ctx, &feedpb.GetFeedRequest{
		UserId: userID,
		Limit:  reqLimit,
		Cursor: reqCursor,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	if feedRes == nil {
		return &model.FeedResponse{Items: []*model.FeedItem{}}, nil
	}

	// 2. Extract all the Post IDs to fetch them in a single batch
	var postIDs []string
	for _, item := range feedRes.Items {
		if item != nil && item.PostId != "" {
			postIDs = append(postIDs, item.PostId)
		}
	}

	// 3. Fetch the actual posts from the Post service
	postMap := make(map[string]*model.Post)
	if len(postIDs) > 0 {
		postsRes, err := r.PostClient.GetPostsByIds(ctx, &postpb.GetPostsByIdsRequest{Ids: postIDs})
		if err != nil {
			// You can choose to return an error here, or just log it and return partial data (null posts)
			log.Printf("api-gateway: failed to fetch posts for feed user_id=%s: %v", userID, err)
		} else if postsRes != nil {
			posts := make([]*model.Post, 0, len(postsRes.Posts))
			// Create a map for quick lookup by ID
			for _, p := range postsRes.Posts {
				if p == nil {
					continue
				}
				post := mapPost(p)
				posts = append(posts, post)
				postMap[p.Id] = post
			}
			if err := r.attachAuthorsToPosts(ctx, posts); err != nil {
				return nil, fmt.Errorf("failed to hydrate feed author data: %w", err)
			}
		}
	}

	// 4. Stitch the Feed items and the Posts together
	items := make([]*model.FeedItem, 0, len(feedRes.Items))
	for _, item := range feedRes.Items {
		if item == nil {
			continue
		}
		items = append(items, &model.FeedItem{
			PostID: item.PostId,
			Cursor: item.Cursor,
			Post:   postMap[item.PostId], // If the post wasn't found, this naturally defaults to nil
		})
	}

	return &model.FeedResponse{
		Items:      items,
		NextCursor: optionalString(feedRes.NextCursor),
	}, nil

}
