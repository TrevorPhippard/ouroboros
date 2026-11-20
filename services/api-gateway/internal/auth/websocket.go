package auth

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
)

// WebsocketInit authenticates a WebSocket connection. Browsers cannot set an
// Authorization header on a WebSocket, so graphql-ws clients send the bearer
// token in the connection_init payload ("Authorization" or "authorization").
// A token in the upgrade request's header, already handled by Middleware, keeps
// working when the payload has none.
func WebsocketInit(secret string) transport.WebsocketInitFunc {
	return func(ctx context.Context, payload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
		header := payload.Authorization()
		if header == "" {
			if v, ok := payload["authorization"].(string); ok {
				header = v
			}
		}
		if token := BearerToken(header); token != "" {
			ctx = Authenticate(ctx, token, secret, time.Now())
		}
		return ctx, &payload, nil
	}
}
