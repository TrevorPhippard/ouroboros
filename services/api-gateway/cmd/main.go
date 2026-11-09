package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func csvEnv(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	var values []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			values = append(values, item)
		}
	}
	if len(values) == 0 {
		return fallback
	}
	return values
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

	// Configure CORS through the environment instead of hard-coding one LAN IP.
	c := cors.New(cors.Options{
		AllowedOrigins:   csvEnv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})

	mux := http.NewServeMux()

	// Register both GraphQL endpoints.
	graphqlHandler := http.MaxBytesHandler(srv, 1<<20)
	mux.Handle("/graphql", graphqlHandler)
	mux.Handle("/query", graphqlHandler) // Backward compatibility

	mux.Handle("/", playground.Handler("GraphQL Playground", "/graphql"))

	// Health check.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	jwtSecret := getEnv("JWT_SECRET", "JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("WARNING: JWT_SECRET is not set; authenticated GraphQL operations will be rejected")
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           c.Handler(AuthContextMiddleware(jwtSecret, mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Ouroboros Gateway listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}
