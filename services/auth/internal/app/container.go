package app

import (
	"log/slog"

	"auth/internal/config"
	"auth/internal/database"
	messagingkafka "auth/internal/messaging/kafka"
	service "auth/internal/service"

	"gorm.io/gorm"
)

type Container struct {
	Config *config.Config
	Logger *slog.Logger

	DB *gorm.DB

	authService *service.Service
	Producer    *messagingkafka.Producer
	// Consumer    *messagingkafka.Consumer
}

func BuildContainer(cfg *config.Config, logger *slog.Logger) *Container {
	db := database.Connect(cfg.PostgresDSN)
	database.Migrate(db)
	database.SeedDB(db)

	producer := messagingkafka.NewProducer(cfg, logger)
	authService := service.NewService(db, producer, cfg.JwtSecret, cfg.KafkaTopic, logger)
	// consumer := messagingkafka.NewConsumer(cfg, logger)

	return &Container{
		Config:      cfg,
		Logger:      logger,
		DB:          db,
		authService: authService,
		Producer:    producer,
		// Consumer:    consumer,
	}
}
