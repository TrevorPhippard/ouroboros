package services

import (
	service "auth/internal/service"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	pb "ouroboros/proto/generated/auth"

	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPasswordHashAndCheck(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == password || hash == "" {
		t.Fatalf("HashPassword() returned an invalid hash: %q", hash)
	}
	if err := service.CheckPassword(password, hash); err != nil {
		t.Errorf("CheckPassword() rejected the correct password: %v", err)
	}
	if err := service.CheckPassword("wrong password", hash); err == nil {
		t.Error("CheckPassword() accepted an incorrect password")
	}
}

func TestGenerateToken(t *testing.T) {
	const userID = "user-123"
	secret := []byte("test-secret")
	token, err := service.GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d segments; want 3", len(parts))
	}
	unsigned := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(unsigned))
	if got := base64.RawURLEncoding.EncodeToString(mac.Sum(nil)); got != parts[2] {
		t.Fatal("token signature did not match")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode claims: %v", err)
	}
	var claims struct {
		Subject   string `json:"sub"`
		ExpiresAt int64  `json:"exp"`
	}
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		t.Fatalf("unmarshal claims: %v", err)
	}
	if claims.Subject != userID {
		t.Errorf("subject = %q; want %q", claims.Subject, userID)
	}
	if !time.Unix(claims.ExpiresAt, 0).After(time.Now()) {
		t.Errorf("token expiration %v is not in the future", time.Unix(claims.ExpiresAt, 0))
	}
}

func TestGenerateTokenRequiresUserAndSecret(t *testing.T) {
	for _, tc := range []struct {
		name   string
		userID string
		secret []byte
	}{
		{name: "missing user ID", secret: []byte("secret")},
		{name: "missing secret", userID: "user-123"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if token, err := service.GenerateToken(tc.userID, tc.secret); err == nil || token != "" {
				t.Errorf("GenerateToken() = (%q, %v); want empty token and error", token, err)
			}
		})
	}
}

func TestServiceRequestValidation(t *testing.T) {
	svc := service.NewService(nil, nil, []byte("secret"), "", slog.Default())
	ctx := t.Context()

	tests := []struct {
		name string
		call func() error
	}{
		{"signup nil request", func() error { _, err := svc.SignUp(ctx, nil); return err }},
		{"signup missing email", func() error { _, err := svc.SignUp(ctx, &pb.SignUpRequest{Password: "pw"}); return err }},
		{"signup missing password", func() error { _, err := svc.SignUp(ctx, &pb.SignUpRequest{Email: "a@example.com"}); return err }},
		{"signin nil request", func() error { _, err := svc.SignIn(ctx, nil); return err }},
		{"signin missing email", func() error { _, err := svc.SignIn(ctx, &pb.SignInRequest{Password: "pw"}); return err }},
		{"signin missing password", func() error { _, err := svc.SignIn(ctx, &pb.SignInRequest{Email: "a@example.com"}); return err }},
		{"get user nil request", func() error { _, err := svc.GetUser(ctx, nil); return err }},
		{"get user missing ID", func() error { _, err := svc.GetUser(ctx, &pb.GetUserRequest{}); return err }},
		{"get users nil request", func() error { _, err := svc.GetUsersByIds(ctx, nil); return err }},
		{"get users empty IDs", func() error { _, err := svc.GetUsersByIds(ctx, &pb.GetUsersByIdsRequest{}); return err }},
		{"get users only empty IDs", func() error {
			_, err := svc.GetUsersByIds(ctx, &pb.GetUsersByIdsRequest{Ids: []string{"", ""}})
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); status.Code(err) != codes.InvalidArgument {
				t.Errorf("status.Code(error) = %v; want %v (error: %v)", status.Code(err), codes.InvalidArgument, err)
			}
		})
	}
}
