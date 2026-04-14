package database

import (
	"log"
	"post/internal/models" // Update with your actual module name
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
	if err := db.AutoMigrate(&models.DBPost{}, &models.DBComment{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

// SeedDB placeholder logic
func SeedDB(db *gorm.DB) {
	var postCount int64
	db.Model(&models.DBPost{}).Count(&postCount)

	// Only seed if empty
	if postCount > 0 {
		log.Println("Database already seeded, skipping...")
		return
	}

	log.Println("Seeding database...")

	posts := []models.DBPost{
		{
			ID:        "post-1",
			AuthorID:  "user-1",
			Content:   "Hello world!",
			CreatedAt: time.Now(),
		},
		{
			ID:        "post-2",
			AuthorID:  "user-2",
			Content:   "Second post 🚀",
			CreatedAt: time.Now(),
		},
	}

	comments := []models.DBComment{
		{
			ID:        "comment-1",
			PostID:    "post-1",
			AuthorID:  "user-2",
			Content:   "Nice post!",
			CreatedAt: time.Now(),
		},
		{
			ID:        "comment-2",
			PostID:    "post-1",
			AuthorID:  "user-3",
			Content:   "Agreed 👍",
			CreatedAt: time.Now(),
		},
	}

	if err := db.Create(&posts).Error; err != nil {
		log.Fatalf("failed to seed posts: %v", err)
	}

	if err := db.Create(&comments).Error; err != nil {
		log.Fatalf("failed to seed comments: %v", err)
	}

	log.Println("Database seeded successfully")
}
