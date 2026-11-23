// Command api-gateway is the Ouroboros GraphQL gateway: a single GraphQL API in
// front of the auth, connection, feed, notification, post and profile services.
package main

import (
	"log"

	"api-gateway/graph"
	"api-gateway/internal/catalog"
	"api-gateway/internal/clients"
	"api-gateway/internal/config"
	"api-gateway/internal/server"
)

func main() {
	cfg := config.Load()
	cfg.Warn()
	backend, closeBackend, err := clients.Dial(cfg.Services)
	if err != nil {
		log.Fatal(err)
	}
	defer closeBackend()

	// Optional capabilities that need Post/Connection RPCs beyond today's API;
	// both are nil unless built with -tags postext (docs/proto-extensions.md).
	posts, recommender := catalog.FromClients(backend)
	if posts == nil {
		log.Printf("post catalog not configured: Query.posts is unavailable and User.posts / Post.comments are empty (build with -tags postext once the Post service exposes the list RPCs)")
	}
	if recommender == nil {
		log.Printf("recommender not configured: Query.recommendations is unavailable (build with -tags postext once the Connection service exposes GetRecommendations)")
	}

	srv := server.New(cfg, &graph.Resolver{
		Clients:                  backend,
		Catalog:                  posts,
		Recommender:              recommender,
		NotificationPollInterval: cfg.NotificationPollInterval,
	})

	log.Printf("Ouroboros Gateway listening on :%s", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
