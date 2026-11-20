package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
)

// Middleware authenticates HTTP requests from their bearer token. It does not
// reject anonymous requests (sign-up and sign-in are public); protected
// resolvers fail closed through Require.
func Middleware(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := Authenticate(r.Context(), BearerToken(r.Header.Get("Authorization")), secret, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authenticate attaches a bearer token's identity to ctx. The token is always
// forwarded to downstream services, which validate it themselves; the gateway
// only trusts the user ID and expiry after verifying the signature here.
func Authenticate(ctx context.Context, token, secret string, now time.Time) context.Context {
	if token == "" {
		return ctx
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	if secret == "" {
		return ctx
	}
	if userID, expiresAt, ok := Verify(token, secret, now); ok {
		ctx = WithUser(ctx, userID)
		ctx = WithExpiry(ctx, expiresAt)
	}
	return ctx
}

// BearerToken extracts the token from an "Authorization: Bearer <token>" value.
func BearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
