// Package server assembles the gateway's HTTP surface: the GraphQL endpoint
// (HTTP and WebSocket), CORS, authentication, the optional playground, and the
// health check.
package server

import (
	"net/http"
	"time"

	"api-gateway/graph"
	"api-gateway/internal/auth"
	"api-gateway/internal/config"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"
)

// maxRequestBody caps GraphQL request bodies at 1 MiB.
const maxRequestBody = 1 << 20

// New returns the configured HTTP server, ready for ListenAndServe.
func New(cfg config.Config, resolver *graph.Resolver) *http.Server {
	mux := http.NewServeMux()

	gql := http.MaxBytesHandler(newGraphQL(cfg, resolver), maxRequestBody)
	mux.Handle("/graphql", gql)
	mux.Handle("/query", gql) // backward compatibility

	// The playground is a development tool: expose it only when asked to.
	if cfg.EnablePlayground {
		mux.Handle("/", playground.Handler("GraphQL Playground", "/graphql"))
	}

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	corsPolicy := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           corsPolicy.Handler(auth.Middleware(cfg.JWTSecret, mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
