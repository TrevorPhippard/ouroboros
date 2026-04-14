package main

import (
	"log"
	"net"
	"net/http"
	"os"

	pb "ouroboros/proto/generated/post"
	"post/internal/database"
	"post/internal/discovery"
	"post/internal/messaging"
	"post/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Setup Infrastructure
	dbURL := os.Getenv("DB_URL")
	db := database.Connect(dbURL)
	database.Migrate(db)
	database.SeedDB(db)

	writer := messaging.NewPostProducer("kafka:9092", "posts.created")
	discovery.Register("consul:8500", "post-service", 50057)

	// 2. Start HTTP Server (Metrics + Health)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		log.Println("HTTP server running on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	// 3. Start gRPC Server
	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	postService := &service.PostServiceServer{DB: db, Writer: writer}
	pb.RegisterPostServiceServer(s, postService)
	reflection.Register(s)

	log.Println("Post Service running on :50057")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
