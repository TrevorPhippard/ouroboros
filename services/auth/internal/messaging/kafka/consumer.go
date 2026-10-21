package kafka

import (
	"auth/internal/config"
	"auth/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type UserService interface {
	CreateUser(ctx context.Context, post *domain.UserRecord) (*domain.UserRecord, error)
}

type Consumer struct {
	reader  *kafka.Reader
	handler UserService
	logger  *slog.Logger
}

type UserCreatedEvent struct {
	ID           string    `gorm:"primaryKey"`
	Email        string    `gorm:"unique;not null"`
	DisplayName  string    `gorm:"not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func NewConsumer(cfg *config.Config, h UserService, l *slog.Logger) *Consumer {

	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{cfg.KafkaAddr},
			Topic:   cfg.KafkaTopic,
			GroupID: cfg.KafkaGroup,
		}),
		handler: h,
		logger:  l,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			time.Sleep(time.Second)
			continue
		}

		var evt UserCreatedEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			continue
		}

		c.logger.Info("Received event", "event", evt)

		user := &domain.UserRecord{
			ID:           uuid.NewString(),
			Email:        evt.Email,
			DisplayName:  evt.DisplayName,
			PasswordHash: evt.PasswordHash,
			CreatedAt:    evt.CreatedAt,
		}

		if _, err := c.handler.CreateUser(ctx, user); err != nil {
			c.logger.Error("Failed to create user", "error", err)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("Failed to commit message", "error", err)
		}

		_ = c.reader.CommitMessages(ctx, msg)
	}
}
