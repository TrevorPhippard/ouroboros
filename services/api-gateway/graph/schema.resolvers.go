package graph

import (
	"api-gateway/graph/model"
	"context"
	"fmt"
	"log"

	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	feedpb "ouroboros/proto/generated/feed"
	notificationpb "ouroboros/proto/generated/notification"
	postpb "ouroboros/proto/generated/post"
	profilepb "ouroboros/proto/generated/profile"
)

func optionalString(s string) *string {
	if s == "" {
		return nil
	}

	value := s
	return &value
}

// SignUp is the resolver for the signUp field.
func (r *mutationResolver) SignUp(
	ctx context.Context,
	input model.SignUpInput,
) (*model.AuthPayload, error) {
	res, err := r.AuthClient.SignUp(ctx, &authpb.SignUpRequest{
		Email:       input.Email,
		Password:    input.Password,
		DisplayName: input.DisplayName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign up: %w", err)
	}

	if res == nil || res.User == nil {
		return nil, fmt.Errorf("sign-up returned no user")
	}

	usersByID, err := r.loadUsersByID(ctx, []string{res.User.Id})
	if err != nil {
		return nil, fmt.Errorf("failed to load signed-up user: %w", err)
	}

	user := usersByID[res.User.Id]
	if user == nil {
		return nil, fmt.Errorf("failed to resolve signed-up user")
	}

	return &model.AuthPayload{
		Token: res.Token,
		User:  user,
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

	if res == nil || res.User == nil {
		return nil, fmt.Errorf("sign-in returned no user")
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
	if _, err := requireCurrentUserID(ctx); err != nil {
		return nil, err
	}
	// Authentication middleware should also propagate the bearer token as gRPC metadata.
	_, err := r.AuthClient.SignOut(ctx, &authpb.SignOutRequest{})
	if err != nil {
		return &model.SignOutResponse{Success: false}, err
	}
	return &model.SignOutResponse{Success: true}, nil
}

// UpdateProfile is the resolver for the updateProfile field.
func (r *mutationResolver) UpdateProfile(
	ctx context.Context,
	userID string,
	input model.UpdateProfileInput,
) (*model.Profile, error) {
	authenticatedUserID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if authenticatedUserID != userID {
		return nil, fmt.Errorf("forbidden: cannot update another user's profile")
	}

	req := &profilepb.UpdateProfileRequest{
		UserId: userID,
	}

	if input.Headline != nil {
		req.Headline = *input.Headline
	}
	if input.About != nil {
		req.About = *input.About
	}
	if input.DisplayName != nil {
		req.DisplayName = *input.DisplayName
	}
	if input.AvatarURL != nil {
		req.AvatarUrl = *input.AvatarURL
	}
	if input.Bio != nil {
		req.Bio = *input.Bio
	}

	res, err := r.ProfileClient.UpdateProfile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	if res == nil || res.Profile == nil {
		return nil, fmt.Errorf("profile service returned no profile")
	}

	p := res.Profile

	experiences := make([]*model.Experience, 0, len(p.Experiences))
	for _, e := range p.Experiences {
		if e == nil {
			continue
		}

		experiences = append(experiences, &model.Experience{
			ID:          e.Id,
			Title:       e.Title,
			Company:     e.Company,
			StartDate:   e.StartDate,
			EndDate:     optionalString(e.EndDate),
			Description: e.Description,
		})
	}

	return &model.Profile{
		ID:          p.Id,
		UserID:      p.UserId,
		Headline:    optionalString(p.Headline),
		About:       optionalString(p.About),
		DisplayName: optionalString(p.DisplayName),
		AvatarURL:   optionalString(p.AvatarUrl),
		Bio:         optionalString(p.Bio),
		Experiences: experiences,
	}, nil
}

// CreatePost is the resolver for the createPost mutation.
func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	userID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if input.AuthorID != userID {
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

	post := &model.Post{
		ID:        res.Id,
		Content:   res.Content,
		CreatedAt: res.CreatedAt,
		AuthorID:  res.AuthorId,
		Comments:  []*model.Comment{},
	}
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
	if input.AuthorID != userID {
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
		AuthorID:  res.AuthorId,
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

	post := &model.Post{
		ID:        res.Id,
		Content:   res.Content,
		CreatedAt: res.CreatedAt,
		AuthorID:  res.AuthorId,
		Comments:  []*model.Comment{},
	}

	if err := r.attachAuthorsToPosts(ctx, []*model.Post{post}); err != nil {
		return nil, fmt.Errorf("failed to hydrate liked post author: %w", err)
	}

	return post, nil
}

// SendConnect is the resolver for the sendConnect field.
func (r *mutationResolver) SendConnect(ctx context.Context, userID string) (*model.ConnectResponse, error) {
	followerID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	res, err := r.ConnectionClient.FollowUser(ctx, &connpb.FollowUserRequest{
		FollowerId: followerID,
		FolloweeId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send connection request: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("connection service returned no response")
	}
	return &model.ConnectResponse{Success: res.Success}, nil
}

// MarkNotificationRead is the resolver for the markNotificationRead field.
func (r *mutationResolver) MarkNotificationRead(ctx context.Context, id string) (bool, error) {
	if _, err := requireCurrentUserID(ctx); err != nil {
		return false, err
	}
	res, err := r.NotificationClient.MarkAsRead(ctx, &notificationpb.MarkAsReadRequest{
		NotificationId: id,
	})
	if err != nil {
		return false, fmt.Errorf("failed to mark notification as read: %w", err)
	}
	if res == nil {
		return false, fmt.Errorf("notification service returned no response")
	}
	return res.Success, nil
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	// The HTTP auth middleware verifies the bearer token and injects the user ID.
	userID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
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
	if res == nil {
		return nil, nil
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
	if res == nil {
		return []*model.Post{}, nil
	}

	posts := make([]*model.Post, 0, len(res.Posts))
	for _, p := range res.Posts {
		if p == nil {
			continue
		}
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

// Posts is currently unsupported by the Post service: the available gateway
// client exposes GetPost and GetPostsByIds, but no list-all RPC. Return an
// explicit error rather than silently pretending that the list is empty.
func (r *queryResolver) Posts(ctx context.Context, limit *int32) ([]*model.Post, error) {
	return nil, fmt.Errorf("posts query is unavailable: Post service does not expose a list-posts RPC")
}

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

// Notifications is the resolver for the notifications field.
func (r *queryResolver) Notifications(
	ctx context.Context,
	userID string,
	limit *int32,
) ([]*model.Notification, error) {
	reqLimit := int32(20)
	if limit != nil {
		reqLimit = *limit
	}

	if reqLimit < 1 || reqLimit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100")
	}
	authenticatedUserID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if authenticatedUserID != userID {
		return nil, fmt.Errorf("forbidden: cannot read another user's notifications")
	}

	res, err := r.NotificationClient.GetNotifications(
		ctx,
		&notificationpb.GetNotificationsRequest{
			UserId: userID,
			Limit:  reqLimit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	if res == nil {
		return []*model.Notification{}, nil
	}

	notifications := make([]*model.Notification, 0, len(res.Notifications))

	for _, n := range res.Notifications {
		if n == nil {
			continue
		}

		notifications = append(notifications, &model.Notification{
			ID:        n.Id,
			UserID:    n.UserId,
			Type:      n.Type,
			ActorID:   n.ActorId,
			EntityID:  optionalString(n.EntityId),
			CreatedAt: n.CreatedAt,
			Read:      n.Read,
		})
	}

	return notifications, nil
}

// Recommendations is the resolver for the recommendations field.
func (r *queryResolver) Recommendations(ctx context.Context) ([]*model.User, error) {
	if _, err := requireCurrentUserID(ctx); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("recommendations are unavailable: Connection service does not expose a recommendation-list RPC")
}

// NotificationReceived is the resolver for the notificationReceived field.
func (r *subscriptionResolver) NotificationReceived(ctx context.Context, userID string) (<-chan *model.Notification, error) {
	authenticatedUserID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if authenticatedUserID != userID {
		return nil, fmt.Errorf("forbidden: cannot subscribe to another user's notifications")
	}
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

// FollowUser follows the specified user.
func (r *mutationResolver) FollowUser(
	ctx context.Context,
	userID string,
) (*model.ConnectResponse, error) {
	followerID := currentUserIDFromContext(ctx)
	if followerID == "" {
		return nil, fmt.Errorf("unauthorized")
	}
	if followerID == userID {
		return nil, fmt.Errorf("cannot follow yourself")
	}

	res, err := r.ConnectionClient.FollowUser(ctx, &connpb.FollowUserRequest{
		FollowerId: followerID,
		FolloweeId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to follow user: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("connection service returned no response")
	}

	return &model.ConnectResponse{Success: res.Success}, nil
}

// UnfollowUser unfollows the specified user.
func (r *mutationResolver) UnfollowUser(
	ctx context.Context,
	userID string,
) (*model.ConnectResponse, error) {
	followerID := currentUserIDFromContext(ctx)
	if followerID == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	res, err := r.ConnectionClient.UnfollowUser(ctx, &connpb.UnfollowUserRequest{
		FollowerId: followerID,
		FolloweeId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unfollow user: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("connection service returned no response")
	}

	return &model.ConnectResponse{Success: res.Success}, nil
}

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
