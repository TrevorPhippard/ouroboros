package graph

import (
	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	feedpb "ouroboros/proto/generated/feed"
	notificationpb "ouroboros/proto/generated/notification"
	postpb "ouroboros/proto/generated/post"
	profilepb "ouroboros/proto/generated/profile"
)

// Resolver holds all dependencies required by GraphQL resolvers.
// It acts as the dependency injection container for the GraphQL API.
type Resolver struct {
	AuthClient         authpb.AuthServiceClient
	ConnectionClient   connpb.ConnectionServiceClient
	FeedClient         feedpb.FeedServiceClient
	NotificationClient notificationpb.NotificationServiceClient
	PostClient         postpb.PostServiceClient
	ProfileClient      profilepb.ProfileServiceClient
}

// Resolver implementations.
// These are shared across all resolver files.
type mutationResolver struct {
	*Resolver
}

type queryResolver struct {
	*Resolver
}

type subscriptionResolver struct {
	*Resolver
}

// Mutation returns the MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query returns the QueryResolver implementation.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Subscription returns the SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}
