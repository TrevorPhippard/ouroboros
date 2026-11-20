// Package config reads the gateway's configuration from the environment.
package config

import (
	"log"
	"time"

	"api-gateway/internal/clients"
)

// Config is the gateway's complete runtime configuration.
type Config struct {
	Port string

	// JWTSecret verifies HS256 bearer tokens. There is deliberately no default:
	// a guessable secret would let anyone mint valid tokens. When empty,
	// protected operations fail closed.
	JWTSecret string

	// AllowedOrigins drives both CORS and the WebSocket origin check.
	AllowedOrigins []string

	EnablePlayground         bool
	ComplexityLimit          int
	NotificationPollInterval time.Duration

	Services clients.Addrs
}

// Load reads the environment, falling back to development defaults. Invalid
// values are logged and ignored rather than silently misapplied.
func Load() Config {
	return Config{
		Port:      str("PORT", "4000"),
		JWTSecret: str("JWT_SECRET", ""),
		AllowedOrigins: csv("CORS_ALLOWED_ORIGINS",
			[]string{"http://localhost:3000", "http://127.0.0.1:3000"}),
		EnablePlayground:         boolean("ENABLE_PLAYGROUND", false),
		ComplexityLimit:          integer("GRAPHQL_COMPLEXITY_LIMIT", 5000),
		NotificationPollInterval: duration("NOTIFICATION_POLL_INTERVAL", 3*time.Second),
		Services: clients.Addrs{
			Auth:         str("AUTH_SERVICE_URL", "localhost:50053"),
			Connection:   str("CONNECTION_SERVICE_URL", "localhost:50054"),
			Feed:         str("FEED_SERVICE_URL", "localhost:50055"),
			Notification: str("NOTIFICATION_SERVICE_URL", "localhost:50056"),
			Post:         str("POST_SERVICE_URL", "localhost:50057"),
			Profile:      str("PROFILE_SERVICE_URL", "localhost:50058"),
		},
	}
}

// Warn logs configuration problems that are not fatal but matter.
func (c Config) Warn() {
	if c.JWTSecret == "" {
		log.Printf("WARNING: JWT_SECRET is not set; authenticated GraphQL operations will be rejected")
	}
}
