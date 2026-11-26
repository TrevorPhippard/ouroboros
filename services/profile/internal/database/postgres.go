package database

import (
	"log"
	"profile/internal/models" // Update with your actual module name
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
	if err := db.AutoMigrate(
		&models.Profile{},
		&models.Experience{},
	); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("Database migrations completed successfully")
}

func SeedDB(db *gorm.DB) {
	var postCount int64
	db.Model(&models.Profile{}).Count(&postCount)

	// Only seed if empty
	if postCount > 0 {
		log.Println("Database already seeded, skipping...")
		return
	}

	log.Println("Seeding database...")

	Profile := []models.Profile{
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
			DisplayName: "Sarah Jenkins",
			AvatarUrl:   "https://api.dicebear.com/7.x/avataaars/svg?seed=user-2",
			Bio:         "I am Sarah! I do things with code.",
			Headline:    "Frontend Engineer | UI/UX enthusiast",
			About:       "Love creating products that people love.",
		},
		{
			ID:          "profile-2",
			UserId:      "user-2",
			DisplayName: "Alex Rivera",
			AvatarUrl:   "https://api.dicebear.com/7.x/avataaars/svg?seed=user-2",
			Bio:         "Bob here! Avid traveler and foodie.",
			Headline:    "Full Stack Developer & Coffee Addict",
			About:       "Love creating products that people love.",
		},
	}

	if err := db.Create(&Profile).Error; err != nil {
		log.Fatalf("failed to seed Profile: %v", err)
	}

	log.Println("Database seeded successfully")
}
