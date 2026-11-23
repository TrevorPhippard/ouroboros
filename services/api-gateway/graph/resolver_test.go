package graph

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"api-gateway/graph/model"
	"api-gateway/internal/auth"
	"api-gateway/internal/clients"
	"api-gateway/internal/dataloader"
	"api-gateway/internal/loaders"
	"api-gateway/internal/testutil"
)

type fixture struct {
	r    *Resolver
	auth *testutil.Auth
	prof *testutil.Profile
	conn *testutil.Connection
	post *testutil.Post
}

func newFixture() *fixture {
	f := &fixture{auth: &testutil.Auth{}, prof: &testutil.Profile{}, conn: &testutil.Connection{}, post: &testutil.Post{}}
	f.r = &Resolver{Clients: clients.Set{
		Auth: f.auth, Profile: f.prof, Connection: f.conn, Post: f.post,
	}}
	return f
}

// opCtx mimics what loaders.Extension and the auth middleware give a real operation.
func (f *fixture) opCtx(viewer string) context.Context {
	ctx := context.Background()
	if viewer != "" {
		ctx = auth.WithUser(ctx, viewer)
	}
	return loaders.With(ctx, loaders.New(ctx, f.r.LoaderDeps(), dataloader.Config{Wait: 20 * time.Millisecond}))
}

func parallel(n int, fn func(i int)) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); fn(i) }(i)
	}
	wg.Wait()
}

func TestPostAuthorsAreBatchedIntoOneUserLookup(t *testing.T) {
	f := newFixture()
	ctx := f.opCtx("viewer")
	pr := &postResolver{f.r}

	// 20 posts by 5 distinct authors.
	parallel(20, func(i int) {
		u, err := pr.Author(ctx, &model.Post{ID: fmt.Sprint(i), UserID: fmt.Sprintf("u%d", i%5)})
		if err != nil || u == nil {
			t.Errorf("author %d: %v, %v", i, u, err)
			return
		}
		if u.DisplayName == nil || *u.DisplayName != "Name "+u.ID {
			t.Errorf("profile not merged: %+v", u)
		}
		if u.FollowersCount != 7 || u.FollowingCount != 3 || !u.IsFollowing {
			t.Errorf("connection data not merged: %+v", u)
		}
	})

	if n := f.auth.Calls.Load(); n != 1 {
		t.Errorf("expected 1 auth call, got %d", n)
	}
	if n := f.prof.Calls.Load(); n != 1 {
		t.Errorf("expected 1 profile call, got %d", n)
	}
	if n := f.conn.Calls.Load(); n != 15 { // 5 users x (followers + following + isFollowing)
		t.Errorf("expected 15 connection calls (5 users, deduped), got %d", n)
	}
}

func TestMissingAuthorIsAnError(t *testing.T) {
	f := newFixture()
	_, err := (&postResolver{f.r}).Author(f.opCtx("v"), &model.Post{UserID: "ghost"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestNotificationActorMissingIsNull(t *testing.T) {
	f := newFixture()
	u, err := (&notificationResolver{f.r}).Actor(f.opCtx("v"), &model.Notification{ActorID: "ghost"})
	if err != nil || u != nil {
		t.Fatalf("expected null actor, got %v, %v", u, err)
	}
}

func TestFeedItemPostsAreBatched(t *testing.T) {
	f := newFixture()
	ctx := f.opCtx("v")
	fr := &feedItemResolver{f.r}

	parallel(10, func(i int) {
		id := fmt.Sprint(i)
		if i == 9 {
			id = "missing"
		}
		p, err := fr.Post(ctx, &model.FeedItem{PostID: id})
		if err != nil {
			t.Errorf("post %s: %v", id, err)
		}
		if (id == "missing") != (p == nil) {
			t.Errorf("post %s: unexpected result %v", id, p)
		}
	})
	if n := f.post.ByIDsCalls.Load(); n != 1 {
		t.Errorf("expected 1 GetPostsByIds call, got %d", n)
	}
}

func TestPostUserIDIsTheAuthor(t *testing.T) {
	f := newFixture()
	posts, err := (&queryResolver{f.r}).PostsByIds(f.opCtx("v"), []string{"a"})
	if err != nil || len(posts) != 1 || posts[0].UserID != "author-a" {
		t.Fatalf("Post.UserID must be the author, got %v, %v", posts, err)
	}
}

func TestNestedListsEmptyWithoutCatalog(t *testing.T) {
	f := newFixture()
	ctx := f.opCtx("v")
	posts, err := (&userResolver{f.r}).Posts(ctx, &model.User{ID: "u"}, nil)
	if err != nil || posts == nil || len(posts) != 0 {
		t.Fatalf("want empty non-nil list, got %v, %v", posts, err)
	}
	cm, err := (&postResolver{f.r}).Comments(ctx, &model.Post{ID: "p"}, nil)
	if err != nil || cm == nil || len(cm) != 0 {
		t.Fatalf("want empty non-nil list, got %v, %v", cm, err)
	}
	if _, err := (&queryResolver{f.r}).Posts(ctx, nil); err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("Query.posts must report unavailable, got %v", err)
	}
}

func TestNestedListsUseCatalogInOneBatchAndHonourLimit(t *testing.T) {
	f := newFixture()
	cat := &testutil.Catalog{}
	f.r.Catalog = cat
	ctx := f.opCtx("v")
	ur, pr := &userResolver{f.r}, &postResolver{f.r}
	two := int32(2)

	parallel(8, func(i int) {
		posts, err := ur.Posts(ctx, &model.User{ID: fmt.Sprintf("u%d", i)}, &two)
		if err != nil || len(posts) != 2 {
			t.Errorf("want 2 posts (limit enforced), got %d, %v", len(posts), err)
		}
		cms, err := pr.Comments(ctx, &model.Post{ID: fmt.Sprintf("p%d", i)}, nil)
		if err != nil || len(cms) != 1 {
			t.Errorf("want 1 comment, got %d, %v", len(cms), err)
		}
	})
	if cat.AuthorCalls != 1 || cat.CommentCall != 1 {
		t.Errorf("expected one batched call each, got posts=%d comments=%d", cat.AuthorCalls, cat.CommentCall)
	}

	for _, bad := range []int32{0, -1, 101} {
		bad := bad
		if _, err := ur.Posts(ctx, &model.User{ID: "u"}, &bad); err == nil {
			t.Errorf("limit %d should be rejected", bad)
		}
	}

	posts, err := (&queryResolver{f.r}).Posts(ctx, &two)
	if err != nil || len(posts) != 2 {
		t.Errorf("Query.posts must cap an over-returning catalog, got %d, %v", len(posts), err)
	}
}

func TestRecommendationsFilterSelfAndDuplicatesPreservingOrder(t *testing.T) {
	f := newFixture()
	f.r.Recommender = testutil.Recommender{IDs: []string{"b", "viewer", "a", "b", "ghost", "c"}}

	users, err := (&queryResolver{f.r}).Recommendations(f.opCtx("viewer"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, u := range users {
		got = append(got, u.ID)
	}
	if strings.Join(got, ",") != "b,a,c" {
		t.Fatalf("want b,a,c got %v", got)
	}

	f.r.Recommender = nil
	if _, err := (&queryResolver{f.r}).Recommendations(f.opCtx("viewer")); err == nil {
		t.Fatal("expected unavailable error")
	}
	if _, err := (&queryResolver{f.r}).Recommendations(f.opCtx("")); err != auth.ErrUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestProfileQuery(t *testing.T) {
	f := newFixture()
	p, err := (&queryResolver{f.r}).Profile(context.Background(), "alice")
	if err != nil || p == nil || p.UserID != "alice" {
		t.Fatalf("got %+v, %v", p, err)
	}
}

func TestSubscriptionRequiresMatchingIdentity(t *testing.T) {
	sr := &subscriptionResolver{&Resolver{}}
	if _, err := sr.NotificationReceived(context.Background(), "alice"); err != auth.ErrUnauthorized {
		t.Fatalf("want unauthorized, got %v", err)
	}
	if _, err := sr.NotificationReceived(auth.WithUser(context.Background(), "mallory"), "alice"); err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("want forbidden, got %v", err)
	}
}
