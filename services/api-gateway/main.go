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

	// Importing the specific generated proto packages
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	// Helper function to create gRPC connections
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// 1. Auth Service (:50053)
	authConn, err := grpc.NewClient(getEnv("AUTH_SERVICE_URL", "localhost:50053"), dialOptions...)
	if err != nil {
		log.Fatalf("could not connect to auth-service: %v", err)
	}
	defer authConn.Close()

	// 2. Connection Service (:50054)
	connConn, err := grpc.NewClient(getEnv("CONNECTION_SERVICE_URL", "localhost:50054"), dialOptions...)
	if err != nil {
		log.Fatalf("could not connect to connection-service: %v", err)
	}
	defer connConn.Close()

	// 3. Feed Service (:50055)
	feedConn, err := grpc.NewClient(getEnv("FEED_SERVICE_URL", "localhost:50055"), dialOptions...)
	if err != nil {
		log.Fatalf("could not connect to feed-service: %v", err)
	}
	defer feedConn.Close()

	// 4. Notification Service (:50056)
	notifConn, err := grpc.NewClient(getEnv("NOTIFICATION_SERVICE_URL", "localhost:50056"), dialOptions...)
	if err != nil {
		log.Fatalf("could not connect to notification-service: %v", err)
	}
	defer notifConn.Close()

	// 5. Post Service (:50057)
	postConn, err := grpc.NewClient(getEnv("POST_SERVICE_URL", "localhost:50057"), dialOptions...)
	if err != nil {
		log.Fatalf("could not connect to post-service: %v", err)
	}
	defer postConn.Close()

	// 6. Profile Service (:50058)
	profileConn, err := grpc.NewClient(getEnv("PROFILE_SERVICE_URL", "localhost:50058"), dialOptions...)
	if err != nil {
		log.Fatalf("could not connect to profile-service: %v", err)
	}
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
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", c.Handler(srv))

	log.Printf("Ouroboros Gateway live at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}