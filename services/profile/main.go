package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "ouroboros/proto/generated/profile"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"profile/internal/database"
	"profile/internal/discovery"
	"profile/internal/messaging"
	"profile/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func connectDB(dbURL string) *gorm.DB {
	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err == nil {
			return db
		}

		log.Println("Waiting for DB...")
		time.Sleep(2 * time.Second)
	}

	log.Fatal("Failed to connect to DB:", err)
	return nil
}

func main() {
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

	log.Println("Profile Service (gRPC) running on :50058")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
