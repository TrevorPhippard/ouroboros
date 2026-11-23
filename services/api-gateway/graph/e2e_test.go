package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-gateway/graph/generated"
	"api-gateway/internal/auth"
	"api-gateway/internal/loaders"
	"api-gateway/internal/testutil"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

func gql(t *testing.T, h http.Handler, ctx context.Context, query string) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"query": query})
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(string(body))).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("bad response %q: %v", rec.Body.String(), err)
	}
	return out
}

// Through the real generated schema: N posts selecting `author` must cost one
// auth lookup, not N, and the result must be correct end to end.
func TestEndToEndLoadersEngagedByExtension(t *testing.T) {
	f := newFixture()
	f.r.Catalog = &testutil.Catalog{}

	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: f.r}))
	srv.AddTransport(transport.POST{})
	srv.Use(loaders.Extension(f.r.LoaderDeps()))

	ctx := auth.WithUser(context.Background(), "viewer")
	out := gql(t, srv, ctx, `{ postsByIds(ids: ["a","b","c","d"]) {
		id UserId author { id username displayName followersCount }
		comments(limit: 1) { id author { username } } } }`)

	if errs, ok := out["errors"]; ok {
		t.Fatalf("unexpected errors: %v", errs)
	}
	posts := out["data"].(map[string]any)["postsByIds"].([]any)
	if len(posts) != 4 {
		t.Fatalf("want 4 posts, got %d", len(posts))
	}
	first := posts[0].(map[string]any)
	if first["UserId"] != "author-a" || first["author"].(map[string]any)["id"] != "author-a" {
		t.Fatalf("author wiring wrong: %v", first)
	}
	if first["author"].(map[string]any)["displayName"] != "Name author-a" {
		t.Fatalf("profile not merged: %v", first)
	}
	// comments are authored by u1: that author is loaded once for all 4 posts,
	// in the same loader as the post authors' next level.
	if len(first["comments"].([]any)) != 1 {
		t.Fatalf("comments missing: %v", first)
	}

	// Level 1: 4 post authors in one lookup. Level 2: u1 (comment author) in one more.
	if n := f.auth.Calls.Load(); n != 2 {
		t.Errorf("expected 2 batched auth lookups (one per query level), got %d", n)
	}
	if n := f.post.ByIDsCalls.Load(); n != 1 {
		t.Errorf("expected 1 GetPostsByIds, got %d", n)
	}
}
