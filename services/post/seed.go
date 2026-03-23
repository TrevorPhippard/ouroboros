package main

import (
	"log"
	"time"

	"gorm.io/gorm"
)

func seedDB(db *gorm.DB) {
	var postCount int64
	db.Model(&DBPost{}).Count(&postCount)

	// Only seed if empty
	if postCount > 0 {
		log.Println("Database already seeded, skipping...")
		return
	}

	log.Println("Seeding database...")

	posts := []DBPost{
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

	comments := []DBComment{
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