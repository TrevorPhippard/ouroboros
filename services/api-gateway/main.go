package main

import (
	"api-gateway/graph"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors" // Import the CORS package
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "ouroboros/proto"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "4000" }

	// Connect to User Service
	userConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to user-service: %v", err)
	}
	defer userConn.Close()


		// Connect to Auth Service
	authConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to auth-service: %v", err)
	}
	defer authConn.Close()

		// Connect to Connection Service
	connConn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to connection-service: %v", err)
	}
	defer connConn.Close()

		// Connect to Feed Service
	feedConn, err := grpc.Dial("localhost:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to feed-service: %v", err)
	}
	defer feedConn.Close()

		// Connect to Notification Service
	notConn, err := grpc.Dial("localhost:50055", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to notification-service: %v", err)
	}
	defer notConn.Close()

		// Connect to Post Service
	postConn, err := grpc.Dial("localhost:50056", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to post-service: %v", err)
	}
	defer postConn.Close()

		// Connect to Profile Service
	profileConn, err := grpc.Dial("localhost:50057", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to profile-service: %v", err)
	}
	defer profileConn.Close()


	// Initialize Gateway with both clients
	resolver := &graph.Resolver{
		AuthServiceClient: pb.NewService2Client(authConn),
		ConnectionServiceClient: pb.NewService2Client(connConn),
		FeedServiceClient: pb.NewService2Client(feedConn),
		NotificationServiceClient: pb.NewService2Client(notConn),
		PostServiceClient: pb.NewService2Client(postConn),
		ProfileServiceClient: pb.NewService2Client(profileConn),
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Allow your Next.js app
		AllowCredentials: true,                              // Required if you are sending cookies/sessions
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		Debug:            false, // Set to true if you need to debug CORS headers in your Go console
	})

// Wrap the GraphQL server handler with the CORS middleware
graphqlHandler := c.Handler(srv)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", graphqlHandler)

	log.Printf("Gateway live at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}