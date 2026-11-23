package config

import (
	"testing"
	"time"
)

func TestLoadDefaultsAndOverrides(t *testing.T) {
	c := Load()
	if c.Port != "4000" || c.JWTSecret != "" || c.EnablePlayground || c.ComplexityLimit != 5000 ||
		c.NotificationPollInterval != 3*time.Second || c.Services.Post != "localhost:50057" {
		t.Fatalf("unexpected defaults: %+v", c)
	}

	t.Setenv("PORT", "9000")
	t.Setenv("JWT_SECRET", "  s3cret ")
	t.Setenv("ENABLE_PLAYGROUND", "true")
	t.Setenv("GRAPHQL_COMPLEXITY_LIMIT", "250")
	t.Setenv("NOTIFICATION_POLL_INTERVAL", "500ms")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example, ,https://b.example")
	c = Load()
	if c.Port != "9000" || c.JWTSecret != "s3cret" || !c.EnablePlayground || c.ComplexityLimit != 250 ||
		c.NotificationPollInterval != 500*time.Millisecond ||
		len(c.AllowedOrigins) != 2 || c.AllowedOrigins[1] != "https://b.example" {
		t.Fatalf("overrides not applied: %+v", c)
	}
}

func TestInvalidValuesFallBack(t *testing.T) {
	t.Setenv("ENABLE_PLAYGROUND", "maybe")
	t.Setenv("GRAPHQL_COMPLEXITY_LIMIT", "-3")
	t.Setenv("NOTIFICATION_POLL_INTERVAL", "soon")
	c := Load()
	if c.EnablePlayground || c.ComplexityLimit != 5000 || c.NotificationPollInterval != 3*time.Second {
		t.Fatalf("invalid values must fall back: %+v", c)
	}
}
