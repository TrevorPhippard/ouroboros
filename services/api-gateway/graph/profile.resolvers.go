package graph

import (
	"context"
	"fmt"

	"api-gateway/graph/model"

	profilepb "ouroboros/proto/generated/profile"
)

func (r *queryResolver) Profiles(
	ctx context.Context,
	userIDs []string,
) ([]*model.Profile, error) {

	userIDs = uniqueIDs(userIDs)

	if len(userIDs) == 0 {
		return []*model.Profile{}, nil
	}

	res, err := r.ProfileClient.GetProfilesByUserIds(
		ctx,
		&profilepb.GetProfilesByUserIdsRequest{
			UserIds: userIDs,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch profiles: %w", err)
	}

	if res == nil {
		return []*model.Profile{}, nil
	}

	profiles := make([]*model.Profile, 0, len(res.Profiles))

	for _, p := range res.Profiles {
		if p == nil {
			continue
		}

		profiles = append(profiles, mapProfile(p))
	}

	return profiles, nil
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
