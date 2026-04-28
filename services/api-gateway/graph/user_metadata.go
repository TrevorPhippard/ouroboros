package graph

import (
	"context"
	"log"
	"strings"

	"api-gateway/graph/model"

	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	profilepb "ouroboros/proto/generated/profile"
)

func (r *Resolver) loadUsersByID(ctx context.Context, ids []string) (map[string]*model.User, error) {
	uniqueIDs := uniqueIDs(ids)
	if len(uniqueIDs) == 0 {
		return map[string]*model.User{}, nil
	}

	authRes, err := r.AuthClient.GetUsersByIds(ctx, &authpb.GetUsersByIdsRequest{Ids: uniqueIDs})
	if err != nil {
		return nil, err
	}

	users := make(map[string]*model.User, len(authRes.Users))
	for _, user := range authRes.Users {
		users[user.Id] = &model.User{
			ID:          user.Id,
			Email:       user.Email,
			Username:    user.Username,
			Posts:       []*model.Post{},
			DisplayName: nil,
			AvatarURL:   nil,
			Bio:         nil,
		}
	}

	profileRes, err := r.ProfileClient.GetProfilesByUserIds(ctx, &profilepb.GetProfilesByUserIdsRequest{UserIds: uniqueIDs})
	if err != nil {
		log.Printf("api-gateway: failed to batch load profiles for %d users: %v", len(uniqueIDs), err)
		return users, nil
	}

	for _, profile := range profileRes.Profiles {
		user := users[profile.UserId]
		if user == nil {
			user = &model.User{
				ID:    profile.UserId,
				Posts: []*model.Post{},
			}
			users[profile.UserId] = user
		}

		if profile.DisplayName != "" {
			user.DisplayName = stringPtr(profile.DisplayName)
		}
		if profile.AvatarUrl != "" {
			user.AvatarURL = stringPtr(profile.AvatarUrl)
		}
		if profile.Bio != "" {
			user.Bio = stringPtr(profile.Bio)
		}
	}

	currentUserID := currentUserIDFromContext(ctx)
	for _, user := range users {
		if user == nil {
			continue
		}

		followersCount, err := r.ConnectionClient.GetFollowersCount(ctx, &connpb.UserRequest{UserId: user.ID})
		if err == nil {
			user.FollowersCount = followersCount.Count
		}

		followingCount, err := r.ConnectionClient.GetFollowingCount(ctx, &connpb.UserRequest{UserId: user.ID})
		if err == nil {
			user.FollowingCount = followingCount.Count
		}

		if currentUserID != "" && currentUserID != user.ID {
			isFollowing, err := r.ConnectionClient.IsFollowing(ctx, &connpb.IsFollowingRequest{
				FollowerId: currentUserID,
				FolloweeId: user.ID,
			})
			if err == nil {
				user.IsFollowing = isFollowing.IsFollowing
			}
		}
	}

	return users, nil
}

func (r *Resolver) attachAuthorsToPosts(ctx context.Context, posts []*model.Post) error {
	authorIDs := make([]string, 0, len(posts))
	for _, post := range posts {
		if post == nil {
			continue
		}
		post.Comments = nonNilComments(post.Comments)
		if strings.TrimSpace(post.AuthorID) != "" {
			authorIDs = append(authorIDs, post.AuthorID)
		}
	}

	usersByID, err := r.loadUsersByID(ctx, authorIDs)
	if err != nil {
		return err
	}

	for _, post := range posts {
		if post == nil {
			continue
		}
		post.Author = usersByID[post.AuthorID]
	}

	return nil
}

func uniqueIDs(values []string) []string {
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

func stringPtr(value string) *string {
	v := value
	return &v
}

func nonNilComments(comments []*model.Comment) []*model.Comment {
	if comments == nil {
		return []*model.Comment{}
	}
	return comments
}

func currentUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("userID").(string); ok && strings.TrimSpace(userID) != "" {
		return userID
	}
	return "user-1"
}
