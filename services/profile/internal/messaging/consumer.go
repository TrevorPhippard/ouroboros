package messaging

import (
	"context"
	"encoding/json"
	"log"
	"profile/internal/service"

	"github.com/segmentio/kafka-go"
)

type userSignedUpEvent struct {
	EventID   string `json:"eventId"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		UserID      string `json:"userId"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	} `json:"data"`
}

func RunUserSignedUpConsumer(ctx context.Context, profileService *service.ProfileServiceServer) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"kafka:9092"},
		Topic:          "user.signed_up",
		GroupID:        "profile-service",
		CommitInterval: 0,
	})
	defer reader.Close()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("profile-service: failed to fetch user.signed_up message: %v", err)
			continue
		}

		var event userSignedUpEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("profile-service: failed to decode user.signed_up message: %v", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		if err := profileService.EnsureProfileForUser(ctx, event.Data.UserID, event.Data.DisplayName); err != nil {
			log.Printf("profile-service: failed processing user.signed_up user_id=%s: %v", event.Data.UserID, err)
			continue
		}

		_ = reader.CommitMessages(ctx, msg)
	}
}
