package messaging

import (
	"context"
	"encoding/json"
	"log"

	"feed/internal/config"
	"feed/internal/feed"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	svc    *feed.Service
}

func New(cfg config.Config, svc *feed.Service) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{cfg.KafkaAddr},
			Topic:   cfg.KafkaTopic,
			GroupID: cfg.KafkaGroup,
		}),
		svc: svc,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	defer c.reader.Close()

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return
		}

		var event feed.PostCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}

		if err := c.svc.HandlePostCreated(ctx, event); err != nil {
			log.Println("fanout failed:", err)
			continue
		}

		_ = c.reader.CommitMessages(ctx, msg)
	}
}
