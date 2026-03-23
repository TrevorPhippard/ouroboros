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
			UserId:      "user-1",
			DisplayName: "Alice",
			AvatarUrl:   "https://api.dicebear.com/7.x/avataaars/svg?seed=user-1",
			Bio:         "Hello, I'm Alice! I love coding and coffee.",
		},
		{
			UserId:      "user-2",
			DisplayName: "Bob",
			AvatarUrl:   "https://api.dicebear.com/7.x/avataaars/svg?seed=user-2",
			Bio:         "Bob here! Avid traveler and foodie.",

		},
	}

	if err := db.Create(&Profile).Error; err != nil {
		log.Fatalf("failed to seed Profile: %v", err)
	}

	log.Println("Database seeded successfully")
}