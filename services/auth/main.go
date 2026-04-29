package main

import (
	"log"
	"os"

	"auth/internal/auth"
	"auth/internal/database"
	"auth/internal/discovery"
	"auth/internal/messaging"
	"auth/internal/transport"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("super-secret")
	}

	db := database.Connect(dbURL)
	database.Migrate(db)
	database.SeedDB(db)

	discovery.Register("consul:8500", "auth-service", 50053)

	writer := messaging.NewPostProducer("kafka:9092", "user.signed_up")
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("failed to close kafka writer: %v", err)
		}
	}()

	authServer := auth.NewService(db, writer, jwtSecret)

	go transport.StartHTTPServer(":8080")
	transport.StartGRPCServer(":50053", authServer)
}
