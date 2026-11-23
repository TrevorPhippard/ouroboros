package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Verify validates an HS256 token's signature and time claims and
// returns the user ID and expiry it carries.
func Verify(token, secret string, now time.Time) (userID string, expiresAt time.Time, ok bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || secret == "" {
		return "", time.Time{}, false
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", time.Time{}, false
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Algorithm != "HS256" {
		return "", time.Time{}, false
	}

	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", time.Time{}, false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(providedSignature, mac.Sum(nil)) {
		return "", time.Time{}, false
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", time.Time{}, false
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return "", time.Time{}, false
	}

	exp, exists := claims["exp"]
	if !exists {
		return "", time.Time{}, false
	}
	expiration, ok := numericClaim(exp)
	if !ok || now.Unix() >= expiration {
		return "", time.Time{}, false
	}
	if nbf, exists := claims["nbf"]; exists {
		timestamp, ok := numericClaim(nbf)
		if !ok || now.Unix() < timestamp {
			return "", time.Time{}, false
		}
	}

	for _, key := range []string{"sub", "user_id", "userId", "userID", "id"} {
		if value, ok := claims[key].(string); ok && strings.TrimSpace(value) != "" {
			userID = strings.TrimSpace(value)
			break
		}
	}
	if userID == "" {
		return "", time.Time{}, false
	}
	return userID, time.Unix(expiration, 0), true
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
