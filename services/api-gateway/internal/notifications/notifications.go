// Package notifications reads a user's notifications and turns the Notification
// service's request/response API into a live stream.
package notifications

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"api-gateway/graph/model"
	"api-gateway/internal/auth"
	"api-gateway/internal/mapping"

	notifpb "ouroboros/proto/generated/notification"
)

const (
	// DefaultPollInterval is used when Service.PollInterval is unset.
	DefaultPollInterval = 3 * time.Second
	// pageSize is how many of the user's newest notifications each poll
	// inspects. It only needs to exceed the number that can arrive within one
	// poll interval.
	pageSize = 50
	// buffer lets a poll deliver a burst without blocking on a slow client;
	// beyond it, delivery applies backpressure to the poller.
	buffer = 16
)

// Service reads notifications from the Notification backend.
type Service struct {
	Client       notifpb.NotificationServiceClient
	PollInterval time.Duration
}

// Latest returns a user's newest notifications.
func (s Service) Latest(ctx context.Context, userID string, limit int32) ([]*model.Notification, error) {
	res, err := s.Client.GetNotifications(ctx, &notifpb.GetNotificationsRequest{UserId: userID, Limit: limit})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return []*model.Notification{}, nil
	}
	out := make([]*model.Notification, 0, len(res.Notifications))
	for _, n := range res.Notifications {
		if n != nil {
			out = append(out, mapping.Notification(n))
		}
	}
	return out, nil
}

// Subscribe streams notifications the first time each is seen, oldest first.
//
// The backend has no push RPC, so this polls. Notifications that already exist
// when the subscription starts are not replayed (query them instead). The
// stream ends, and the channel is closed, when ctx is done or the bearer token
// the caller authenticated with expires.
func (s Service) Subscribe(ctx context.Context, userID string) (<-chan *model.Notification, error) {
	interval := s.PollInterval
	if interval <= 0 {
		interval = DefaultPollInterval
	}

	// Prime synchronously so a broken backend fails the subscription up front
	// instead of producing a silent, empty stream.
	initial, err := s.Latest(ctx, userID, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to start notification subscription: %w", err)
	}

	events := make(chan *model.Notification, buffer)
	go func() {
		defer close(events)
		s.stream(ctx, userID, interval, idSet(initial), events)
	}()
	return events, nil
}

func (s Service) stream(ctx context.Context, userID string, interval time.Duration, seen map[string]struct{}, events chan<- *model.Notification) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if expiry, ok := auth.Expiry(ctx); ok && !time.Now().Before(expiry) {
			return // the token this subscription was authorized with has expired
		}

		latest, err := s.Latest(ctx, userID, pageSize)
		if err != nil {
			// Transient backend failures must not kill the subscription.
			log.Printf("api-gateway: notification poll failed user_id=%s: %v", userID, err)
			continue
		}

		fresh := make([]*model.Notification, 0, len(latest))
		for _, n := range latest {
			if _, already := seen[n.ID]; !already {
				fresh = append(fresh, n)
			}
		}
		// Only IDs from the latest window can recur, so replacing the set keeps
		// memory bounded for arbitrarily long-lived subscriptions.
		seen = idSet(latest)

		// Deliver oldest first. CreatedAt is RFC 3339, so it sorts lexically.
		sort.SliceStable(fresh, func(i, j int) bool { return fresh[i].CreatedAt < fresh[j].CreatedAt })
		for _, n := range fresh {
			select {
			case events <- n:
			case <-ctx.Done():
				return
			}
		}
	}
}

func idSet(ns []*model.Notification) map[string]struct{} {
	set := make(map[string]struct{}, len(ns))
	for _, n := range ns {
		set[n.ID] = struct{}{}
	}
	return set
}
