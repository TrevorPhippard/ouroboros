// Package auth establishes and reads the caller's identity: JWT verification,
// the HTTP and WebSocket entry points, and the request-context accessors that
// resolvers use. Only identities set here are ever trusted.
package auth

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrUnauthorized is returned when an operation needs a verified identity.
var ErrUnauthorized = errors.New("unauthorized")

type (
	userKey   struct{}
	expiryKey struct{}
)

// WithUser stores an identity that has ALREADY been verified. Never call it
// with a value taken straight from a request header or argument.
func WithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userKey{}, strings.TrimSpace(userID))
}

// UserID returns the verified caller, or "" when unauthenticated.
func UserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(userKey{}).(string)
	return strings.TrimSpace(id)
}

// Require returns the verified caller or ErrUnauthorized. There is deliberately
// no fallback identity: one would let anonymous requests act as a real user.
func Require(ctx context.Context) (string, error) {
	if id := UserID(ctx); id != "" {
		return id, nil
	}
	return "", ErrUnauthorized
}

// WithExpiry records when the caller's token expires so long-lived operations
// (subscriptions) can end when it does.
func WithExpiry(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, expiryKey{}, t)
}

// Expiry returns the token expiry, if one was recorded.
func Expiry(ctx context.Context) (time.Time, bool) {
	t, ok := ctx.Value(expiryKey{}).(time.Time)
	return t, ok && !t.IsZero()
}
