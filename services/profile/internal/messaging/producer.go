package messaging

import (
	"github.com/segmentio/kafka-go"
)

func NewProfileProducer(broker string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}
