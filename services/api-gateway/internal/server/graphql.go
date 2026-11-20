package server

import (
	"net/http"
	"time"

	"api-gateway/graph"
	"api-gateway/graph/generated"
	"api-gateway/internal/auth"
	"api-gateway/internal/config"
	"api-gateway/internal/loaders"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"
)

// newGraphQL builds the GraphQL handler. It mirrors gqlgen's NewDefaultServer
// (query cache, introspection, persisted queries) and adds what a public
// gateway needs: per-operation data loaders, a query-complexity limit,
// authenticated WebSocket subscriptions with an origin check, and no multipart
// upload transport (the schema has no uploads).
func newGraphQL(cfg config.Config, resolver *graph.Resolver) *handler.Server {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers:  resolver,
		Complexity: complexity(),
	}))

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		InitFunc:              auth.WebsocketInit(cfg.JWTSecret),
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return originAllowed(r.Header.Get("Origin"), cfg.AllowedOrigins)
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})
	srv.Use(extension.FixedComplexityLimit(cfg.ComplexityLimit))
	srv.Use(loaders.Extension(resolver.LoaderDeps()))
	return srv
}
