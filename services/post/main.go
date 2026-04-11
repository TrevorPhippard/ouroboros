package main

import (
	"log"
	"net"
	"os"
	"time"

	pb "ouroboros/proto/generated/post"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB models — renamed so they do not collide with pb.Post / pb.Comment

func (DBPost) TableName() string {
	return "posts"
}

func (DBComment) TableName() string {
	return "comments"
}

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

	if err := db.AutoMigrate(&DBPost{}, &DBComment{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	seedDB(db)

	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterPostServiceServer(server, &PostServiceServer{DB: db})

	log.Println("Post Service running on :50057")

	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}