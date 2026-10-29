package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api-gateway/graph"

	"google.golang.org/grpc/metadata"
)

// AuthContextMiddleware verifies HS256 bearer tokens when JWT_SECRET is set,
// adds the verified user ID to the GraphQL context, and forwards the bearer
// token to downstream gRPC services as authorization metadata. It intentionally
// does not reject every unauthenticated request because sign-up/sign-in are
// public operations; protected resolvers fail closed when no user ID is set.
func AuthContextMiddleware(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if token := bearerToken(r.Header.Get("Authorization")); token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
			if secret != "" {
				if userID, ok := verifyHS256JWT(token, secret, time.Now()); ok {
					ctx = graph.WithUserID(ctx, userID)
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func verifyHS256JWT(token, secret string, now time.Time) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || secret == "" {
		return "", false
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Algorithm != "HS256" {
		return "", false
	}

	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(providedSignature, mac.Sum(nil)) {
		return "", false
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return "", false
	}

	exp, exists := claims["exp"]
	if !exists {
		return "", false
	}
	expiration, ok := numericClaim(exp)
	if !ok || now.Unix() >= expiration {
		return "", false
	}
	if nbf, exists := claims["nbf"]; exists {
		timestamp, ok := numericClaim(nbf)
		if !ok || now.Unix() < timestamp {
			return "", false
		}
	}

	var userID string
	for _, key := range []string{"sub", "user_id", "userId", "userID", "id"} {
		if value, ok := claims[key].(string); ok && strings.TrimSpace(value) != "" {
			userID = strings.TrimSpace(value)
			break
		}
	}
	return userID, userID != ""
}

func numericClaim(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), v == float64(int64(v))
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}
