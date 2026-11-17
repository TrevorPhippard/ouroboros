package graph

import (
	"context"
	"fmt"

	"api-gateway/graph/model"

	authpb "ouroboros/proto/generated/auth"
)

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
