package main

import (
	"log"

	"gorm.io/gorm"
)

func seedDB(db *gorm.DB) {
	var postCount int64
	db.Model(&Profile{}).Count(&postCount)

	// Only seed if empty
	if postCount > 0 {
		log.Println("Database already seeded, skipping...")
		return
	}

	log.Println("Seeding database...")

	Profile := []Profile{
		{
			ID:          "profile-1",
			UserId:      "user-1",
			DisplayName: "Alice",
			AvatarUrl:   "https://api.dicebear.com/7.x/avataaars/svg?seed=user-1",
			Bio:         "Hello, I'm Alice! I love coding and coffee.",
			Headline:    "Software Engineer",
			About:       "Passionate about building scalable systems.",
		},
		{
			ID:          "profile-2",
			UserId:      "user-2",
			DisplayName: "Bob",
			AvatarUrl:   "https://api.dicebear.com/7.x/avataaars/svg?seed=user-2",
			Bio:         "Bob here! Avid traveler and foodie.",
			Headline:    "Product Manager",
			About:       "Love creating products that people love.",
		},
	}

	if err := db.Create(&Profile).Error; err != nil {
		log.Fatalf("failed to seed Profile: %v", err)
	}

	log.Println("Database seeded successfully")
}