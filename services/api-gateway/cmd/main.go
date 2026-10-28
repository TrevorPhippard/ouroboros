package main

import (
	"log"
	"net/http"
	"os"

	"api-gateway/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Importing specific generated proto packages
	authpb "ouroboros/proto/generated/auth"
	connpb "ouroboros/proto/generated/connection"
	feedpb "ouroboros/proto/generated/feed"
	notifpb "ouroboros/proto/generated/notification"
	postpb "ouroboros/proto/generated/post"
	profilepb "ouroboros/proto/generated/profile"
)

// Helper function to get env or fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// dialService encapsulates gRPC client creation and fatal error handling.
func dialService(name, envKey, defaultURL string, opts []grpc.DialOption) *grpc.ClientConn {
	url := getEnv(envKey, defaultURL)
	conn, err := grpc.NewClient(url, opts...)
	if err != nil {
		log.Fatalf("could not connect to %s service at %s: %v", name, url, err)
	}
	return conn
}

func main() {
	port := getEnv("PORT", "4000")

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Initialize gRPC service connections cleanly
	authConn := dialService("auth", "AUTH_SERVICE_URL", "localhost:50053", dialOptions)
	defer authConn.Close()

	connConn := dialService("connection", "CONNECTION_SERVICE_URL", "localhost:50054", dialOptions)
	defer connConn.Close()

	feedConn := dialService("feed", "FEED_SERVICE_URL", "localhost:50055", dialOptions)
	defer feedConn.Close()

	notifConn := dialService("notification", "NOTIFICATION_SERVICE_URL", "localhost:50056", dialOptions)
	defer notifConn.Close()

	postConn := dialService("post", "POST_SERVICE_URL", "localhost:50057", dialOptions)
	defer postConn.Close()

	profileConn := dialService("profile", "PROFILE_SERVICE_URL", "localhost:50058", dialOptions)
	defer profileConn.Close()

	// Initialize Resolver with specific clients
	resolver := &graph.Resolver{
		AuthClient:         authpb.NewAuthServiceClient(authConn),
		ConnectionClient:   connpb.NewConnectionServiceClient(connConn),
		FeedClient:         feedpb.NewFeedServiceClient(feedConn),
		NotificationClient: notifpb.NewNotificationServiceClient(notifConn),
		PostClient:         postpb.NewPostServiceClient(postConn),
		ProfileClient:      profilepb.NewProfileServiceClient(profileConn),
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://192.168.2.225:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", c.Handler(srv))

	log.Printf("Ouroboros Gateway live at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
