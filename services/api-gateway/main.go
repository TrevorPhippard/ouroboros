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
	if port == "" { port = "8080" }

	// Connect to User Service
	userConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect to user-service: %v", err)
	}
	defer userConn.Close()

	// Connect to Todo Service
	todoConn, err := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("could not connect to todo-service: %v", err)
	}
	defer todoConn.Close()

	// Initialize Gateway with both clients
	resolver := &graph.Resolver{
		UserServiceClient: pb.NewService1Client(userConn),
		TodoServiceClient: pb.NewService2Client(todoConn),
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