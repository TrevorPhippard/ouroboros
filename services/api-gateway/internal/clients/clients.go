// Package clients owns the gateway's connections to the backend gRPC services.
package clients

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	feedpb "ouroboros/proto/generated/feed"
	notifpb "ouroboros/proto/generated/notification"
	postpb "ouroboros/proto/generated/post"
	profilepb "ouroboros/proto/generated/profile"
)

// Set is one typed client per backend service.
type Set struct {
	Auth         authpb.AuthServiceClient
	Connection   connpb.ConnectionServiceClient
	Feed         feedpb.FeedServiceClient
	Notification notifpb.NotificationServiceClient
	Post         postpb.PostServiceClient
	Profile      profilepb.ProfileServiceClient
}

// Addrs holds the host:port of each backend service.
type Addrs struct {
	Auth, Connection, Feed, Notification, Post, Profile string
}

// Dial creates a client for every service. gRPC connects lazily, so this only
// fails on malformed addresses. Call the returned function to close everything.
func Dial(a Addrs) (Set, func(), error) {
	var conns []*grpc.ClientConn
	closeAll := func() {
		for _, c := range conns {
			_ = c.Close()
		}
	}

	var firstErr error
	dial := func(name, addr string) *grpc.ClientConn {
		if firstErr != nil {
			return nil
		}
		// Service-to-service traffic inside the cluster does not use TLS.
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			firstErr = fmt.Errorf("could not connect to %s service at %s: %w", name, addr, err)
			return nil
		}
		conns = append(conns, conn)
		return conn
	}

	authConn := dial("auth", a.Auth)
	connConn := dial("connection", a.Connection)
	feedConn := dial("feed", a.Feed)
	notifConn := dial("notification", a.Notification)
	postConn := dial("post", a.Post)
	profileConn := dial("profile", a.Profile)
	if firstErr != nil {
		closeAll()
		return Set{}, nil, errors.Join(firstErr)
	}

	return Set{
		Auth:         authpb.NewAuthServiceClient(authConn),
		Connection:   connpb.NewConnectionServiceClient(connConn),
		Feed:         feedpb.NewFeedServiceClient(feedConn),
		Notification: notifpb.NewNotificationServiceClient(notifConn),
		Post:         postpb.NewPostServiceClient(postConn),
		Profile:      profilepb.NewProfileServiceClient(profileConn),
	}, closeAll, nil
}
