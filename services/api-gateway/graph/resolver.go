package graph

// This file will not be regenerated automatically.
//
// The resolvers in this package are deliberately thin: they validate arguments,
// authorize, and delegate. Domain logic lives in ../internal. Keep helpers out
// of *.resolvers.go files: gqlgen comments out anything in them that is not a
// resolver method.

import (
	"context"
	"time"

	"api-gateway/internal/catalog"
	"api-gateway/internal/clients"
	"api-gateway/internal/loaders"
	"api-gateway/internal/notifications"
)

// recommendationCount is how many suggestions `recommendations` returns.
const recommendationCount = 10

// Resolver is the dependency container shared by all resolvers.
type Resolver struct {
	Clients clients.Set

	// Catalog backs Query.posts, User.posts and Post.comments. Optional: when
	// nil, Query.posts reports catalog.ErrUnavailable and the nested lists are
	// empty.
	Catalog catalog.PostCatalog
	// Recommender backs Query.recommendations. Optional: when nil the query
	// reports catalog.ErrUnavailable.
	Recommender catalog.Recommender

	// NotificationPollInterval is how often a notificationReceived subscription
	// polls the Notification service. Defaults to notifications.DefaultPollInterval.
	NotificationPollInterval time.Duration
}

// LoaderDeps returns what the per-operation data loaders need. The server
// passes it to loaders.Extension.
func (r *Resolver) LoaderDeps() loaders.Deps {
	return loaders.Deps{Clients: r.Clients, Catalog: r.Catalog}
}

// loaders returns the loaders for this operation.
func (r *Resolver) loaders(ctx context.Context) *loaders.Loaders {
	return r.LoaderDeps().For(ctx)
}

func (r *Resolver) notifications() notifications.Service {
	return notifications.Service{Client: r.Clients.Notification, PollInterval: r.NotificationPollInterval}
}
