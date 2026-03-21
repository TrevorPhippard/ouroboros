package graph

import (
	authPb "ouroboros/proto/generated/auth"
	connectionPb "ouroboros/proto/generated/connection"
	feedPb "ouroboros/proto/generated/feed"
	notificationPb "ouroboros/proto/generated/notification"
	postPb "ouroboros/proto/generated/post"
	profilePb "ouroboros/proto/generated/profile"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	ConnectionServiceClient connectionPb.ConnectionServiceClient
	AuthServiceClient       authPb.AuthServiceClient
	FeedServiceClient       feedPb.FeedServiceClient
	NotificationServiceClient notificationPb.NotificationServiceClient
	PostServiceClient       postPb.PostServiceClient
	ProfileServiceClient    profilePb.ProfileServiceClient
}