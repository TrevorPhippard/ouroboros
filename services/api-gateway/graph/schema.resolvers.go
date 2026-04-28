package graph

import (
	"api-gateway/graph/model"
	"context"
	"fmt"
	"log"
	"time"

	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	feedpb "ouroboros/proto/generated/feed"
	notificationpb "ouroboros/proto/generated/notification"
	postpb "ouroboros/proto/generated/post"
	profilepb "ouroboros/proto/generated/profile"
)

// SignUp is the resolver for the signUp field.
func (r *mutationResolver) SignUp(ctx context.Context, input model.SignUpInput) (*model.User, error) {
	res, err := r.AuthClient.SignUp(ctx, &authpb.SignUpRequest{
		Email:       input.Email,
		Password:    input.Password,
		DisplayName: input.DisplayName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign up: %w", err)
	}

	usersByID, err := r.loadUsersByID(ctx, []string{res.User.Id})
	if err != nil {
		return nil, fmt.Errorf("failed to load signed-up user: %w", err)
	}

	user := usersByID[res.User.Id]
	if user == nil {
		return nil, fmt.Errorf("failed to resolve signed-up user")
	}

	return user, nil
}

// SignIn is the resolver for the signIn field.
func (r *mutationResolver) SignIn(ctx context.Context, input model.SignInInput) (*model.AuthPayload, error) {
	res, err := r.AuthClient.SignIn(ctx, &authpb.SignInRequest{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	usersByID, err := r.loadUsersByID(ctx, []string{res.User.Id})
	if err != nil {
		return nil, fmt.Errorf("failed to load signed-in user: %w", err)
	}

	user := usersByID[res.User.Id]
	if user == nil {
		return nil, fmt.Errorf("failed to resolve signed-in user")
	}

	return &model.AuthPayload{
		Token: res.Token,
		User:  user,
	}, nil
}

// SignOut is the resolver for the signOut field.
func (r *mutationResolver) SignOut(ctx context.Context) (*model.SignOutResponse, error) {
	// Usually requires passing the token via gRPC metadata extracted from ctx
	_, err := r.AuthClient.SignOut(ctx, &authpb.SignOutRequest{})
	if err != nil {
		return &model.SignOutResponse{Success: false}, err
	}
	return &model.SignOutResponse{Success: true}, nil
}

// UpdateProfile is the resolver for the updateProfile field.
func (r *mutationResolver) UpdateProfile(ctx context.Context, userID string, input model.UpdateProfileInput) (*model.Profile, error) {
	// Convert pointers safely
	var headline, about string
	if input.Headline != nil {
		headline = *input.Headline
	}
	if input.About != nil {
		about = *input.About
	}

	res, err := r.ProfileClient.UpdateProfile(ctx, &profilepb.UpdateProfileRequest{
		UserId:   userID,
		Headline: headline,
		About:    about,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &model.Profile{
		ID:       res.Profile.Id,
		Headline: &res.Profile.Headline,
		About:    &res.Profile.About,
		// Assuming experiences need mapping too
	}, nil
}

// CreatePost is the resolver for the createPost mutation.
func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	res, err := r.PostClient.CreatePost(ctx, &postpb.CreatePostRequest{
		AuthorId: input.AuthorID,
		Content:  input.Content,
	})
	if err != nil {
		return nil, err
	}

	return &model.Post{
		ID:        res.Id,
		Content:   res.Content,
		CreatedAt: res.CreatedAt,
		AuthorID:  res.AuthorId,
	}, nil
}

// CreateComment is the resolver for the createComment field.
func (r *mutationResolver) CreateComment(ctx context.Context, input model.CreateCommentInput) (*model.Comment, error) {
	if _, err := r.PostClient.GetPost(ctx, &postpb.GetPostRequest{Id: input.PostID}); err != nil {
		return nil, fmt.Errorf("failed to find post for comment: %w", err)
	}

	usersByID, err := r.loadUsersByID(ctx, []string{input.AuthorID})
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate comment author: %w", err)
	}

	return &model.Comment{
		ID:        fmt.Sprintf("comment-%d", time.Now().UnixNano()),
		PostID:    input.PostID,
		AuthorID:  input.AuthorID,
		Content:   input.Content,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Author:    usersByID[input.AuthorID],
	}, nil
}

// LikePost is the resolver for the likePost field.
func (r *mutationResolver) LikePost(ctx context.Context, postID string) (*model.Post, error) {
	res, err := r.PostClient.GetPost(ctx, &postpb.GetPostRequest{Id: postID})
	if err != nil {
		return nil, fmt.Errorf("failed to like post: %w", err)
	}

	post := &model.Post{
		ID:        res.Id,
		Content:   res.Content,
		CreatedAt: res.CreatedAt,
		AuthorID:  res.AuthorId,
	}

	if err := r.attachAuthorsToPosts(ctx, []*model.Post{post}); err != nil {
		return nil, err
	}

	return post, nil
}

// SendConnect is the resolver for the sendConnect field.
func (r *mutationResolver) SendConnect(ctx context.Context, userID string) (*model.ConnectResponse, error) {
	res, err := r.ConnectionClient.FollowUser(ctx, &connpb.FollowUserRequest{
		FollowerId: currentUserIDFromContext(ctx),
		FolloweeId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send connection request: %w", err)
	}
	return &model.ConnectResponse{Success: res.Success}, nil
}

// MarkNotificationRead is the resolver for the markNotificationRead field.
func (r *mutationResolver) MarkNotificationRead(ctx context.Context, id string) (bool, error) {
	res, err := r.NotificationClient.MarkAsRead(ctx, &notificationpb.MarkAsReadRequest{
		NotificationId: id,
	})
	if err != nil {
		return false, fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return res.Success, nil
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	// Requires HTTP middleware to parse JWT and inject into context
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	return r.User(ctx, userID)
}

// User is the resolver for the user query.
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	usersByID, err := r.loadUsersByID(ctx, []string{id})
	if err != nil {
		return nil, fmt.Errorf("failed to get user auth record: %w", err)
	}

	user := usersByID[id]
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, ids []string) ([]*model.User, error) {
	usersByID, err := r.loadUsersByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	users := make([]*model.User, 0, len(ids))
	for _, id := range ids {
		if user := usersByID[id]; user != nil {
			users = append(users, user)
		}
	}
	return users, nil
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	res, err := r.PostClient.GetPost(ctx, &postpb.GetPostRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}

	post := &model.Post{
		ID:        res.Id,
		Content:   res.Content,
		CreatedAt: res.CreatedAt,
		AuthorID:  res.AuthorId,
		Comments:  []*model.Comment{},
	}
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

	var posts []*model.Post
	for _, p := range res.Posts {
		posts = append(posts, &model.Post{
			ID:        p.Id,
			Content:   p.Content,
			CreatedAt: p.CreatedAt,
			AuthorID:  p.AuthorId,
			Comments:  []*model.Comment{},
		})
	}
	if err := r.attachAuthorsToPosts(ctx, posts); err != nil {
		return nil, fmt.Errorf("failed to hydrate posts: %w", err)
	}
	return posts, nil
}

// Feed is the resolver for the feed query.
func (r *queryResolver) Feed(ctx context.Context, userID string, limit *int32, cursor *string) (*model.FeedResponse, error) {
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
	feedRes, err := r.FeedClient.GetFeed(ctx, &feedpb.GetFeedRequest{
		UserId: userID,
		Limit:  reqLimit,
		Cursor: reqCursor,
	})
	if err != nil {
		return nil, err
	}

	// 2. Extract all the Post IDs to fetch them in a single batch
	var postIDs []string
	for _, item := range feedRes.Items {
		if item.PostId != "" {
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
		} else {
			posts := make([]*model.Post, 0, len(postsRes.Posts))
			// Create a map for quick lookup by ID
			for _, p := range postsRes.Posts {
				post := &model.Post{
					ID:        p.Id,
					Content:   p.Content,
					CreatedAt: p.CreatedAt,
					AuthorID:  p.AuthorId,
					Comments:  []*model.Comment{},
				}
				posts = append(posts, post)
				postMap[p.Id] = post
			}
			if err := r.attachAuthorsToPosts(ctx, posts); err != nil {
				return nil, fmt.Errorf("failed to hydrate feed author data: %w", err)
			}
		}
	}

	// 4. Stitch the Feed items and the Posts together
	var items []*model.FeedItem
	for _, item := range feedRes.Items {
		items = append(items, &model.FeedItem{
			PostID: item.PostId,
			Cursor: item.Cursor,
			Post:   postMap[item.PostId], // If the post wasn't found, this naturally defaults to nil
		})
	}

	return &model.FeedResponse{Items: items}, nil
}

// Notifications is the resolver for the notifications field.
func (r *queryResolver) Notifications(ctx context.Context, userID string, limit *int32) ([]*model.Notification, error) {
	reqLimit := int32(20)
	if limit != nil {
		reqLimit = *limit
	}

	res, err := r.NotificationClient.GetNotifications(ctx, &notificationpb.GetNotificationsRequest{
		UserId: userID,
		Limit:  reqLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	var notifications []*model.Notification
	for _, n := range res.Notifications {
		notifications = append(notifications, &model.Notification{
			ID:        n.Id,
			Type:      n.Type,
			ActorID:   n.ActorId,
			CreatedAt: n.CreatedAt,
			Read:      n.Read,
		})
	}
	return notifications, nil
}

// Recommendations is the resolver for the recommendations field.
func (r *queryResolver) Recommendations(ctx context.Context) ([]*model.User, error) {
	currentUserID := currentUserIDFromContext(ctx)
	candidateIDs := []string{"user-1", "user-2", "user-3"}

	usersByID, err := r.loadUsersByID(ctx, candidateIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load recommendation candidates: %w", err)
	}

	recommendations := make([]*model.User, 0, len(candidateIDs))
	for _, candidateID := range candidateIDs {
		if candidateID == currentUserID {
			continue
		}
		user := usersByID[candidateID]
		if user == nil {
			continue
		}

		isFollowing, err := r.ConnectionClient.IsFollowing(ctx, &connpb.IsFollowingRequest{
			FollowerId: currentUserID,
			FolloweeId: candidateID,
		})
		if err == nil && isFollowing.IsFollowing {
			continue
		}

		recommendations = append(recommendations, user)
	}

	return recommendations, nil
}

// NotificationReceived is the resolver for the notificationReceived field.
func (r *subscriptionResolver) NotificationReceived(ctx context.Context, userID string) (<-chan *model.Notification, error) {
	return nil, fmt.Errorf("subscriptions are not enabled")
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	func (r *mutationResolver) FollowUser(ctx context.Context, followerID string, followeeID string) (bool, error) {
	res, err := r.ConnectionClient.FollowUser(ctx, &connpb.FollowUserRequest{
		FollowerId: followerID,
		FolloweeId: followeeID,
	})
	if err != nil {
		return false, err
	}

	return res.Success, nil
}
func (r *mutationResolver) UnfollowUser(ctx context.Context, followerID string, followeeID string) (bool, error) {
	res, err := r.ConnectionClient.UnfollowUser(ctx, &connpb.UnfollowUserRequest{
		FollowerId: followerID,
		FolloweeId: followeeID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to unfollow user: %w", err)
	}

	return res.Success, nil
}
*/
