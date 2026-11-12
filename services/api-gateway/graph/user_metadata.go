package graph

import (
	"context"
	"fmt"
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
	if authRes == nil {
		return nil, fmt.Errorf("auth service returned an empty user response")
	}

	users := make(map[string]*model.User, len(authRes.Users))
	for _, user := range authRes.Users {
		if user == nil || strings.TrimSpace(user.Id) == "" {
			continue
		}
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
	if profileRes == nil {
		log.Printf("api-gateway: profile service returned an empty response for %d users", len(uniqueIDs))
		return users, nil
	}

	for _, profile := range profileRes.Profiles {
		if profile == nil || strings.TrimSpace(profile.UserId) == "" {
			continue
		}
		// User.email and User.username are non-null fields in the GraphQL
		// schema, so do not synthesize a partial User when the auth service
		// has no corresponding account record.
		user := users[profile.UserId]
		if user == nil {
			continue
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

	currentUserID, err := requireCurrentUserID(ctx)
	for _, user := range users {
		if user == nil {
			continue
		}

		followersCount, err := r.ConnectionClient.GetFollowersCount(ctx, &connpb.UserRequest{UserId: user.ID})
		if err == nil && followersCount != nil {
			user.FollowersCount = followersCount.Count
		}

		followingCount, err := r.ConnectionClient.GetFollowingCount(ctx, &connpb.UserRequest{UserId: user.ID})
		if err == nil && followingCount != nil {
			user.FollowingCount = followingCount.Count
		}

		if currentUserID != "" && currentUserID != user.ID {
			isFollowing, err := r.ConnectionClient.IsFollowing(ctx, &connpb.IsFollowingRequest{
				FollowerId: currentUserID,
				FolloweeId: user.ID,
			})
			if err == nil && isFollowing != nil {
				user.IsFollowing = isFollowing.IsFollowing
			}
		}
	}

	return users, nil
}

func (r *Resolver) attachAuthorsToPosts(ctx context.Context, posts []*model.Post) error {
	UserIds := make([]string, 0, len(posts))
	for _, post := range posts {
		if post == nil {
			continue
		}
		post.Comments = nonNilComments(post.Comments)
		if strings.TrimSpace(post.UserId) != "" {
			UserIds = append(UserIds, post.UserId)
		}
	}

	usersByID, err := r.loadUsersByID(ctx, UserIds)
	if err != nil {
		return err
	}

	for _, post := range posts {
		if post == nil {
			continue
		}
		author := usersByID[post.UserId]
		if author == nil {
			return fmt.Errorf("author %q was not found", post.UserId)
		}
		post.Author = author
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

type userIDContextKey struct{}

// WithUserID adds an identity that has already been authenticated by trusted
// middleware. Callers must not populate it directly from an arbitrary request
// header or query parameter.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, strings.TrimSpace(userID))
}

func mapProfile(p *profilepb.Profile) *model.Profile {
	if p == nil {
		return nil
	}

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
	}
}
