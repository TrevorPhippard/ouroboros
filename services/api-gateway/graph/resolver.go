package graph

import (
	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	feedpb "ouroboros/proto/generated/feed"
	notifpb "ouroboros/proto/generated/notification"
	postpb "ouroboros/proto/generated/post"
	profilepb "ouroboros/proto/generated/profile"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AuthClient         authpb.AuthServiceClient
	ConnectionClient   connpb.ConnectionServiceClient
	FeedClient         feedpb.FeedServiceClient
	NotificationClient notifpb.NotificationServiceClient
	PostClient         postpb.PostServiceClient
	ProfileClient      profilepb.ProfileServiceClient
}