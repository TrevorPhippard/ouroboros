package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"auth/internal/config"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	logger *slog.Logger
}

func NewProducer(cfg *config.Config, logger *slog.Logger) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.KafkaAddr),
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
		},
		logger: logger,
	}
}

func (p *Producer) Publish(
	ctx context.Context,
	topic string,
	value any,
) error {
	if p == nil || p.writer == nil {
		return errors.New("kafka producer is not initialized")
	}
	if topic == "" {
		return errors.New("kafka topic is required")
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: payload,
	})
	if err != nil {
		if p.logger != nil {
			p.logger.Error("Kafka publish failed",
				"topic", topic,
				"error", err,
			)
		}
		return err
	}

	return nil
}

func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
