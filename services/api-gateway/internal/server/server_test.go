package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-gateway/graph"
	"api-gateway/internal/clients"
	"api-gateway/internal/config"
	"api-gateway/internal/testutil"
)

func TestOriginAllowed(t *testing.T) {
	allowed := []string{"http://localhost:3000"}
	cases := []struct {
		origin string
		want   bool
	}{
		{"", true}, {"http://localhost:3000", true}, {"HTTP://LOCALHOST:3000", true},
		{"http://evil.example", false}, {"http://localhost:3001", false},
	}
	for _, c := range cases {
		if got := originAllowed(c.origin, allowed); got != c.want {
			t.Errorf("origin %q = %v, want %v", c.origin, got, c.want)
		}
	}
	if !originAllowed("http://anything", []string{"*"}) {
		t.Error("wildcard should allow any origin")
	}
}

func TestComplexityMultipliesByLimit(t *testing.T) {
	c := complexity()
	hundred := int32(100)
	if got := c.User.Posts(5, &hundred); got != 501 {
		t.Errorf("User.posts(100) child 5 = %d, want 501", got)
	}
	if got := c.Post.Comments(5, nil); got != 51 { // default limit 10
		t.Errorf("Post.comments default = %d, want 51", got)
	}
	// A recursive query must cost far more than any sane limit.
	level := 1
	for i := 0; i < 4; i++ {
		level = c.User.Posts(level+1, &hundred)
	}
	if level < 5000 {
		t.Errorf("recursive fan-out should exceed the default limit, got %d", level)
	}
	if got := c.Query.Users(2, nil); got != 3 {
		t.Errorf("empty id list must stay non-negative/cheap, got %d", got)
	}
}

func newTestResolver() *graph.Resolver {
	return &graph.Resolver{Clients: clients.Set{
		Auth: &testutil.Auth{}, Profile: &testutil.Profile{}, Connection: &testutil.Connection{}, Post: &testutil.Post{},
	}}
}

func post(t *testing.T, h http.Handler, query string) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"query": query})
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("bad response %q: %v", rec.Body.String(), err)
	}
	return out
}

func TestComplexityLimitIsEnforcedByTheServer(t *testing.T) {
	cfg := config.Config{ComplexityLimit: 1000}
	h := newGraphQL(cfg, newTestResolver())

	out := post(t, h, `{ user(id:"u") { posts(limit:100) { author { posts(limit:100) { id } } } } }`)
	if _, ok := out["errors"]; !ok {
		t.Fatalf("expected a complexity error, got %v", out)
	}
	if out := post(t, h, `{ user(id:"u") { username } }`); out["errors"] != nil {
		t.Fatalf("a cheap query must pass, got %v", out)
	}
}

func TestHTTPSurface(t *testing.T) {
	get := func(srv *http.Server, path string) int {
		rec := httptest.NewRecorder()
		srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Code
	}

	off := New(config.Config{Port: "0"}, newTestResolver())
	if c := get(off, "/healthz"); c != http.StatusOK {
		t.Errorf("healthz = %d", c)
	}
	if c := get(off, "/"); c != http.StatusNotFound {
		t.Errorf("playground must be off by default, got %d", c)
	}
	on := New(config.Config{Port: "0", EnablePlayground: true}, newTestResolver())
	if c := get(on, "/"); c != http.StatusOK {
		t.Errorf("playground should be served when enabled, got %d", c)
	}
}
