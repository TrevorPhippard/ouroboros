// Package testutil holds in-memory fakes of the backend services, shared by the
// tests of the packages that talk to them. Not for production use.
package testutil

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"api-gateway/graph/model"

	"google.golang.org/grpc"

	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	notifpb "ouroboros/proto/generated/notification"
	postpb "ouroboros/proto/generated/post"
	profilepb "ouroboros/proto/generated/profile"
)

// The fakes embed the generated client interfaces so they keep compiling when
// the services gain RPCs; only the methods the tests need are implemented.

// Auth fakes the auth service: every ID exists except "ghost".
type Auth struct {
	authpb.AuthServiceClient
	Calls atomic.Int32
}

func (f *Auth) GetUsersByIds(_ context.Context, in *authpb.GetUsersByIdsRequest, _ ...grpc.CallOption) (*authpb.GetUsersByIdsResponse, error) {
	f.Calls.Add(1)
	res := &authpb.GetUsersByIdsResponse{}
	for _, id := range in.Ids {
		if id != "ghost" {
			res.Users = append(res.Users, &authpb.User{Id: id, Email: id + "@x.io", Username: id})
		}
	}
	return res, nil
}

// Profile fakes the profile service: every user has the display name "Name <id>".
type Profile struct {
	profilepb.ProfileServiceClient
	Calls atomic.Int32
}

func (f *Profile) GetProfilesByUserIds(_ context.Context, in *profilepb.GetProfilesByUserIdsRequest, _ ...grpc.CallOption) (*profilepb.GetProfilesByUserIdsResponse, error) {
	f.Calls.Add(1)
	res := &profilepb.GetProfilesByUserIdsResponse{}
	for _, id := range in.UserIds {
		res.Profiles = append(res.Profiles, &profilepb.Profile{Id: "p-" + id, UserId: id, DisplayName: "Name " + id})
	}
	return res, nil
}

// Connection fakes the connection service: 7 followers, 3 following, always followed.
type Connection struct {
	connpb.ConnectionServiceClient
	Calls atomic.Int32
}

func (f *Connection) GetFollowersCount(context.Context, *connpb.UserRequest, ...grpc.CallOption) (*connpb.CountResponse, error) {
	f.Calls.Add(1)
	return &connpb.CountResponse{Count: 7}, nil
}
func (f *Connection) GetFollowingCount(context.Context, *connpb.UserRequest, ...grpc.CallOption) (*connpb.CountResponse, error) {
	f.Calls.Add(1)
	return &connpb.CountResponse{Count: 3}, nil
}
func (f *Connection) IsFollowing(context.Context, *connpb.IsFollowingRequest, ...grpc.CallOption) (*connpb.IsFollowingResponse, error) {
	f.Calls.Add(1)
	return &connpb.IsFollowingResponse{IsFollowing: true}, nil
}

// Post fakes the post service: every post exists except "missing"; the author
// of post X is "author-X".
type Post struct {
	postpb.PostServiceClient
	ByIDsCalls atomic.Int32
}

func (f *Post) GetPostsByIds(_ context.Context, in *postpb.GetPostsByIdsRequest, _ ...grpc.CallOption) (*postpb.GetPostsByIdsResponse, error) {
	f.ByIDsCalls.Add(1)
	res := &postpb.GetPostsByIdsResponse{}
	for _, id := range in.Ids {
		if id != "missing" {
			res.Posts = append(res.Posts, &postpb.Post{Id: id, Content: "c", AuthorId: "author-" + id})
		}
	}
	return res, nil
}

// Notification fakes the notification service with scripted pages: call i
// returns Pages[i], and the last page repeats.
type Notification struct {
	notifpb.NotificationServiceClient
	mu    sync.Mutex
	calls int
	Pages [][]*notifpb.Notification
}

func (f *Notification) GetNotifications(context.Context, *notifpb.GetNotificationsRequest, ...grpc.CallOption) (*notifpb.GetNotificationsResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	i := f.calls
	if i >= len(f.Pages) {
		i = len(f.Pages) - 1
	}
	f.calls++
	return &notifpb.GetNotificationsResponse{Notifications: f.Pages[i]}, nil
}

// Notif builds a notification for user "alice" from actor "bob".
func Notif(id, createdAt string) *notifpb.Notification {
	return &notifpb.Notification{Id: id, UserId: "alice", Type: "like", ActorId: "bob", CreatedAt: createdAt}
}

// Catalog fakes the optional PostCatalog and records batch-call counts. It
// deliberately over-returns so tests can prove the gateway enforces limits.
type Catalog struct {
	mu                       sync.Mutex
	AuthorCalls, CommentCall int
}

func (f *Catalog) ListPosts(_ context.Context, limit int32) ([]*model.Post, error) {
	out := make([]*model.Post, 0, limit+5)
	for i := 0; i < int(limit)+5; i++ {
		out = append(out, &model.Post{ID: fmt.Sprint(i)})
	}
	return out, nil
}

func (f *Catalog) PostsByAuthors(_ context.Context, ids []string, _ int32) (map[string][]*model.Post, error) {
	f.mu.Lock()
	f.AuthorCalls++
	f.mu.Unlock()
	out := map[string][]*model.Post{}
	for _, id := range ids {
		out[id] = []*model.Post{{ID: "p1-" + id}, {ID: "p2-" + id}, {ID: "p3-" + id}}
	}
	return out, nil
}

func (f *Catalog) CommentsByPosts(_ context.Context, ids []string, _ int32) (map[string][]*model.Comment, error) {
	f.mu.Lock()
	f.CommentCall++
	f.mu.Unlock()
	out := map[string][]*model.Comment{}
	for _, id := range ids {
		out[id] = []*model.Comment{{ID: "c-" + id, PostID: id, UserID: "u1"}}
	}
	return out, nil
}

// Recommender fakes the optional Recommender with a fixed ranking.
type Recommender struct{ IDs []string }

func (f Recommender) RecommendedUserIDs(context.Context, string, int32) ([]string, error) {
	return f.IDs, nil
}
