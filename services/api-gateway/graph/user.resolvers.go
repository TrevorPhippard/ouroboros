package graph

import (
	"context"
	"fmt"

	"api-gateway/graph/model"
)

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
