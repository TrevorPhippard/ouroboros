package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"google.golang.org/grpc/metadata"
)

const testSecret = "test-secret"

func b64(v any) string {
	raw, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func mint(secret string, header, claims map[string]any) string {
	signing := b64(header) + "." + b64(claims)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signing))
	return signing + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func hs256(claims map[string]any) string {
	return mint(testSecret, map[string]any{"alg": "HS256", "typ": "JWT"}, claims)
}

func TestVerify(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	exp := now.Add(time.Hour).Unix()

	uid, expiresAt, ok := Verify(hs256(map[string]any{"sub": "alice", "exp": exp}), testSecret, now)
	if !ok || uid != "alice" || expiresAt.Unix() != exp {
		t.Fatalf("valid token rejected: %q %v %v", uid, expiresAt, ok)
	}

	cases := map[string]string{
		"expired":      hs256(map[string]any{"sub": "a", "exp": now.Add(-time.Second).Unix()}),
		"no exp":       hs256(map[string]any{"sub": "a"}),
		"not yet":      hs256(map[string]any{"sub": "a", "exp": exp, "nbf": now.Add(time.Minute).Unix()}),
		"no user":      hs256(map[string]any{"exp": exp}),
		"wrong secret": mint("other", map[string]any{"alg": "HS256"}, map[string]any{"sub": "a", "exp": exp}),
		"alg none":     mint(testSecret, map[string]any{"alg": "none"}, map[string]any{"sub": "a", "exp": exp}),
		"garbage":      "not.a.jwt",
	}
	for name, tok := range cases {
		if _, _, ok := Verify(tok, testSecret, now); ok {
			t.Errorf("%s token must be rejected", name)
		}
	}
	if _, _, ok := Verify(hs256(map[string]any{"sub": "a", "exp": exp}), "", now); ok {
		t.Error("an empty secret must never verify anything")
	}
}

func TestRequire(t *testing.T) {
	if _, err := Require(context.Background()); err != ErrUnauthorized {
		t.Fatalf("want ErrUnauthorized, got %v", err)
	}
	if id, err := Require(WithUser(context.Background(), " alice ")); err != nil || id != "alice" {
		t.Fatalf("got %q, %v", id, err)
	}
}

func TestWebsocketInitAuthenticatesFromPayload(t *testing.T) {
	tok := hs256(map[string]any{"sub": "alice", "exp": time.Now().Add(time.Hour).Unix()})
	init := WebsocketInit(testSecret)

	for _, key := range []string{"Authorization", "authorization"} {
		ctx, _, err := init(context.Background(), transport.InitPayload{key: "Bearer " + tok})
		if err != nil {
			t.Fatal(err)
		}
		if UserID(ctx) != "alice" {
			t.Errorf("%s: payload token should authenticate, got %q", key, UserID(ctx))
		}
		if _, ok := Expiry(ctx); !ok {
			t.Errorf("%s: expiry should be recorded", key)
		}
		md, _ := metadata.FromOutgoingContext(ctx)
		if got := md.Get("authorization"); len(got) != 1 || got[0] != "Bearer "+tok {
			t.Errorf("%s: token must be forwarded downstream, got %v", key, got)
		}
	}

	// A forged token is forwarded for the backend to judge but never trusted here.
	forged := mint("attacker", map[string]any{"alg": "HS256"}, map[string]any{"sub": "alice", "exp": time.Now().Add(time.Hour).Unix()})
	ctx, _, _ := init(context.Background(), transport.InitPayload{"Authorization": "Bearer " + forged})
	if UserID(ctx) != "" {
		t.Errorf("forged token must stay unauthenticated, got %q", UserID(ctx))
	}
}
