package graph

import (
	"api-gateway/graph/model"
	"context"
	"fmt"

	authpb "ouroboros/proto/generated/auth"
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

	// Assuming SignUp returns the user details. If it only returns an ID,
	// you'd call r.User(ctx, res.UserId) here.
	return &model.User{
		ID:       res.User.Id,
		Email:    res.User.Email,
		Username: res.User.Username,
	}, nil
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

	return &model.AuthPayload{
		Token: res.Token,
		User: &model.User{
			ID:       res.User.Id,
			Email:    res.User.Email,
			Username: res.User.Username,
		},
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
	if input.Headline != nil { headline = *input.Headline }
	if input.About != nil { about = *input.About }

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
		panic(fmt.Errorf("not implemented: CreateComment - CreateComment"))

	// res, err := r.PostClient.CreateComment(ctx, &postpb.CreateCommentRequest{
	// 	PostId:   input.PostID,
	// 	AuthorId: input.AuthorID,
	// 	Content:  input.Content,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create comment: %w", err)
	// }

	// return &model.Comment{
	// 	ID:        res.Comment.Id,
	// 	PostID:    res.Comment.PostId,
	// 	AuthorID:  res.Comment.AuthorId,
	// 	Content:   res.Comment.Content,
	// 	CreatedAt: res.Comment.CreatedAt,
	// }, nil
}

// LikePost is the resolver for the likePost field.
func (r *mutationResolver) LikePost(ctx context.Context, postID string) (*model.Post, error) {
	panic(fmt.Errorf("not implemented: LikePost - LikePost"))
}

// SendConnect is the resolver for the sendConnect field.
func (r *mutationResolver) SendConnect(ctx context.Context, userID string) (*model.ConnectResponse, error) {
	panic(fmt.Errorf("not implemented: SendConnect - SendConnect"))
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
	authRes, err := r.AuthClient.GetUser(ctx, &authpb.GetUserRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get user auth record: %w", err)
	}

	profRes, err := r.ProfileClient.GetProfile(ctx, &profilepb.GetProfileRequest{UserId: id})
	if err != nil {
		// Decide if partial degradation is acceptable here
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &model.User{
		ID:          authRes.Id,
		Email:       authRes.Email,
		Username:    authRes.Username,
		DisplayName: &profRes.DisplayName,
		AvatarURL:   &profRes.AvatarUrl,
		Bio:         &profRes.Bio,
	}, nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, ids []string) ([]*model.User, error) {
	// Optimized approach: Use a batch endpoint if your gRPC service supports it.
	// Doing this in a loop creates an N+1 scaling nightmare across the network boundary.
	var users []*model.User

	// Assuming a batch implementation exists in your gRPC service:
	res, err := r.AuthClient.GetUsersByIds(ctx, &authpb.GetUsersByIdsRequest{Ids: ids})
	if err != nil {
		return nil, err
	}

	for _, u := range res.Users {
		users = append(users, &model.User{
			ID:       u.Id,
			Email:    u.Email,
			Username: u.Username,
		})
	}

	// Note: You would also want to batch-fetch profiles here and stitch them.
	return users, nil
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	res, err := r.PostClient.GetPost(ctx, &postpb.GetPostRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}

	return &model.Post{
		ID:        res.Id,
		Content:   res.Content,
		CreatedAt: res.CreatedAt,
		AuthorID:  res.AuthorId,
	}, nil
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
		})
	}
	return posts, nil
}

// Feed is the resolver for the feed query.
func (r *queryResolver) Feed(ctx context.Context, userID string, limit *int32, cursor *string) (*model.FeedResponse, error) {
	// reqLimit := int32(20)
	// if limit != nil { reqLimit = *limit }

	// reqCursor := ""
	// if cursor != nil { reqCursor = *cursor }

	feedRes, err := r.FeedClient.GetFeed(ctx, &feedpb.GetFeedRequest{
		UserId: userID,
		// Limit:  reqLimit,
		// Cursor: reqCursor,
	})
	if err != nil {
		return nil, err
	}

	var items []*model.FeedItem
	for _, item := range feedRes.Items {
		items = append(items, &model.FeedItem{
			PostID: item.PostId,
			Cursor: item.Cursor,
		})
	}

	return &model.FeedResponse{Items: items}, nil
}

// Notifications is the resolver for the notifications field.
func (r *queryResolver) Notifications(ctx context.Context, userID string, limit *int32) ([]*model.Notification, error) {
	reqLimit := int32(20)
	if limit != nil { reqLimit = *limit }

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
	panic(fmt.Errorf("not implemented: Recommendations - Recommendations"))
}

// NotificationReceived is the resolver for the notificationReceived field.
func (r *subscriptionResolver) NotificationReceived(ctx context.Context, userID string) (<-chan *model.Notification, error) {
	panic(fmt.Errorf("not implemented: NotificationReceived - NotificationReceived"))
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
