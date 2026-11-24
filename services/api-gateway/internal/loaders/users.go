package loaders

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"api-gateway/graph/model"
	"api-gateway/internal/args"
	"api-gateway/internal/auth"

	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	profilepb "ouroboros/proto/generated/profile"
)

// maxConcurrentUserRPCs bounds the fan-out of per-user connection lookups that
// a single user batch performs.
const maxConcurrentUserRPCs = 16

// fetchUsers is the batch function behind the users loader. It assembles whole
// User objects from three services using one auth call, one profile call and
// (concurrently) the per-user connection counts. Users unknown to the auth
// service are omitted, which the loader surfaces as nil.
func (d Deps) fetchUsers(ctx context.Context, ids []string) (map[string]*model.User, error) {
	ids = args.UniqueIDs(ids)
	if len(ids) == 0 {
		return map[string]*model.User{}, nil
	}

	authRes, err := d.Clients.Auth.GetUsersByIds(ctx, &authpb.GetUsersByIdsRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	if authRes == nil {
		return nil, fmt.Errorf("auth service returned an empty user response")
	}

	users := make(map[string]*model.User, len(authRes.Users))
	for _, u := range authRes.Users {
		if u == nil || strings.TrimSpace(u.Id) == "" {
			continue
		}
		users[u.Id] = &model.User{ID: u.Id, Email: u.Email, Username: u.Username}
	}
	if len(users) == 0 {
		return users, nil
	}

	d.applyProfiles(ctx, users, ids)
	d.applyConnections(ctx, users)
	return users, nil
}

// applyProfiles overlays profile fields. A profile outage degrades the result
// (no
//
//	/ avatar / bio) instead of failing the whole lookup.
func (d Deps) applyProfiles(ctx context.Context, users map[string]*model.User, ids []string) {
	res, err := d.Clients.Profile.GetProfilesByUserIds(ctx, &profilepb.GetProfilesByUserIdsRequest{UserIds: ids})
	if err != nil {
		log.Printf("api-gateway: failed to batch load profiles for %d users: %v", len(ids), err)
		return
	}
	if res == nil {
		log.Printf("api-gateway: profile service returned an empty response for %d users", len(ids))
		return
	}
	for _, p := range res.Profiles {
		if p == nil {
			continue
		}
		// User.email and User.username are non-null, so never synthesize a
		// partial User for a profile that has no auth account.
		user := users[p.UserId]
		if user == nil {
			continue
		}
		if p.DisplayName != "" {
			user.DisplayName = ptr(p.DisplayName)
		}
		if p.AvatarUrl != "" {
			user.AvatarURL = ptr(p.AvatarUrl)
		}
		if p.Bio != "" {
			user.Bio = ptr(p.Bio)
		}
	}
}

// applyConnections fills follower/following counts and isFollowing. The
// connection service has no batch RPC, so lookups run concurrently (bounded).
// Failures leave the zero value.
func (d Deps) applyConnections(ctx context.Context, users map[string]*model.User) {
	viewerID := auth.UserID(ctx)
	conn := d.Clients.Connection

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentUserRPCs)
	spawn := func(fn func()) {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			fn()
		}()
	}

	// Each task writes a distinct field of a distinct user, so no locking is
	// needed; the map itself is only read here.
	for _, user := range users {
		user := user
		spawn(func() {
			res, err := conn.GetFollowersCount(ctx, &connpb.UserRequest{UserId: user.ID})
			if err != nil {
				log.Printf("api-gateway: followers count for %s: %v", user.ID, err)
				return
			}
			if res != nil {
				user.FollowersCount = res.Count
			}
		})
		spawn(func() {
			res, err := conn.GetFollowingCount(ctx, &connpb.UserRequest{UserId: user.ID})
			if err != nil {
				log.Printf("api-gateway: following count for %s: %v", user.ID, err)
				return
			}
			if res != nil {
				user.FollowingCount = res.Count
			}
		})
		if viewerID != "" && viewerID != user.ID {
			spawn(func() {
				res, err := conn.IsFollowing(ctx, &connpb.IsFollowingRequest{FollowerId: viewerID, FolloweeId: user.ID})
				if err != nil {
					log.Printf("api-gateway: isFollowing %s->%s: %v", viewerID, user.ID, err)
					return
				}
				if res != nil {
					user.IsFollowing = res.IsFollowing
				}
			})
		}
	}
	wg.Wait()
}

func ptr(s string) *string { return &s }
