package kafka

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lineage/api/internal/domain"
	"github.com/segmentio/kafka-go"
)

// Producer wraps kafka-go writer for event production
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.Hash{}, // Partition by key (scope_id)
		},
	}
}

// ProduceEvent sends an event to Kafka, partitioned by scope_id
func (p *Producer) ProduceEvent(ctx context.Context, event domain.EventInput) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Use scope_id as partition key to ensure ordering within a scope
	key := event.ScopeID[:]

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

// ProduceEventWithKey sends an event with a custom partition key
func (p *Producer) ProduceEventWithKey(ctx context.Context, key uuid.UUID, event domain.EventInput) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key[:],
		Value: value,
	})
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
