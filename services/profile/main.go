package main

import (
	"log"
	"net"
	"os"
	"time"

	pb "ouroboros/proto/generated/profile"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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


	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	db := connectDB(dbURL)

	log.Println("Connected to post_db")

	if err := db.AutoMigrate(&Profile{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	seedDB(db)
	// Port 50058 selected to maintain uniqueness in the cluster
	lis, err := net.Listen("tcp", ":50058")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterProfileServiceServer(s, &profileServiceServer{})

	log.Println("Profile Service (gRPC) running on :50058")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}