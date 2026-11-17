package graph

import (
	"context"
	"fmt"

	"api-gateway/graph/model"

	connpb "ouroboros/proto/generated/connection"
)

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

// Recommendations is the resolver for the recommendations field.
func (r *queryResolver) Recommendations(ctx context.Context) ([]*model.User, error) {
	if _, err := requireCurrentUserID(ctx); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("recommendations are unavailable: Connection service does not expose a recommendation-list RPC")
}

// FollowUser follows the specified user.
func (r *mutationResolver) FollowUser(
	ctx context.Context,
	userID string,
) (*model.ConnectResponse, error) {
	followerID, err := requireCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
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
	followerID, err := requireCurrentUserID(ctx)

	if err != nil {
		return nil, err
	}

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
