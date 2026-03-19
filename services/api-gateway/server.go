package main

import (
	"api-gateway/graph"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
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
	todoConn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("Gateway live at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}