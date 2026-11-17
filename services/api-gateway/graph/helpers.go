package graph

import (
	"context"
	"fmt"
	"strings"
)

func optionalString(s string) *string {
	if s == "" {
		return nil
	}

	value := s
	return &value
}

// currentUserIDFromContext returns only an identity established by trusted
// authentication middleware. Never synthesize a fallback identity: doing so
// lets unauthenticated requests act as a real user.
func currentUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userID, ok := ctx.Value(userIDContextKey{}).(string); ok {
		return strings.TrimSpace(userID)
	}
	return ""
}

func requireCurrentUserID(ctx context.Context) (string, error) {
	userID := currentUserIDFromContext(ctx)
	if userID == "" {
		return "", fmt.Errorf("unauthorized")
	}
	return userID, nil
}
