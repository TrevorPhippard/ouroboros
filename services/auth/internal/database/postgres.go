package database

import (
	"auth/internal/models" // Update with your actual module name
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dbURL string) *gorm.DB {
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

func Migrate(db *gorm.DB) {
	if err := db.AutoMigrate(&models.DBUser{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

// SeedDB placeholder logic
func SeedDB(db *gorm.DB) {
	var userCount int64
	db.Model(&models.DBUser{}).Count(&userCount)

	// Only seed if empty
	if userCount > 0 {
		log.Println("Database already seeded, skipping...")
		return
	}

	log.Println("Seeding database...")

	users := []models.DBUser{
		{
			ID:           "user-1",
			Email:        "user1@example.com",
			PasswordHash: "user1",
			DisplayName:  "User One",
		},
		{
			ID:           "user-2",
			Email:        "user2@example.com",
			PasswordHash: "user2",
			DisplayName:  "User Two",
		},
	}

	if err := db.Create(&users).Error; err != nil {
		log.Fatalf("failed to seed users: %v", err)
	}

	log.Println("Database seeded successfully")
}
