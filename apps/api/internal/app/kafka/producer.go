package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lineage/api/internal/event"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// Producer wraps kafka-go writer for event production
type Producer struct {
	writer *kafka.Writer
}

// ProducerConfig holds configuration for creating a producer
type ProducerConfig struct {
	Brokers      []string
	Topic        string
	SASLEnabled  bool
	SASLUsername string
	SASLPassword string
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg ProducerConfig) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: &kafka.Hash{}, // Partition by key (scope_id)
	}

	// Configure SASL if enabled (for Aiven)
	if cfg.SASLEnabled {
		mechanism, err := scram.Mechanism(scram.SHA256, cfg.SASLUsername, cfg.SASLPassword)
		if err != nil {
			return nil, err
		}

		writer.Transport = &kafka.Transport{
			SASL: mechanism,
			TLS:  &tls.Config{}, // Aiven requires TLS
		}
	}

	return &Producer{writer: writer}, nil
}

// ProduceEvent sends an event to Kafka, partitioned by scope_id
func (p *Producer) ProduceEvent(ctx context.Context, input event.Input) error {
	value, err := json.Marshal(input)
	if err != nil {
		return err
	}

	// Use scope_id as partition key to ensure ordering within a scope
	key := input.ScopeID[:]

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

// ProduceEventWithKey sends an event with a custom partition key
func (p *Producer) ProduceEventWithKey(ctx context.Context, key uuid.UUID, input event.Input) error {
	value, err := json.Marshal(input)
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
