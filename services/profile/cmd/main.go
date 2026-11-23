package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "ouroboros/proto/generated/profile"

	"profile/internal/database"
	"profile/internal/discovery"
	"profile/internal/messaging"
	"profile/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Setup Infrastructure
	dbURL := os.Getenv("DB_URL")
	db := database.Connect(dbURL)
	database.Migrate(db)
	database.SeedDB(db)

	writer := messaging.NewProfileProducer("kafka:9092", "posts.created")
	discovery.Register("consul:8500", "profile-service", 50058)

	// HTTP server (metrics + health)
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		log.Println("HTTP server running on :8080 (metrics + health)")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	// gRPC server
	lis, err := net.Listen("tcp", ":50058")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	profileService := &service.ProfileServiceServer{DB: db, Writer: writer}
	pb.RegisterProfileServiceServer(s, profileService)
	reflection.Register(s)

	go messaging.RunUserSignedUpConsumer(ctx, profileService)

	log.Println("Profile Service (gRPC) running on :50058")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
