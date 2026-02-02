package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/app/db/sqlc"
	"github.com/lineage/api/internal/event"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/segmentio/kafka-go"
)

// Consumer processes events from Kafka and writes to Postgres
type Consumer struct {
	reader  *kafka.Reader
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic, groupID string, db *pgxpool.Pool) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10e6, // 10MB
		}),
		db:      db,
		queries: sqlc.New(db),
	}
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Starting Kafka consumer for topic: %s", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Error fetching message: %v", err)
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Error processing message: %v", err)
				// Don't commit on error - message will be retried
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

// processMessage handles a single Kafka message
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var input event.Input
	if err := json.Unmarshal(msg.Value, &input); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Validate intent
	if !event.ValidateIntent(input.Intent) {
		return fmt.Errorf("invalid intent: %s", input.Intent)
	}

	// Validate correction type if present
	if input.CorrectionType != nil && !event.ValidateCorrectionType(*input.CorrectionType) {
		return fmt.Errorf("invalid correction type: %s", *input.CorrectionType)
	}

	// Get event type for schema validation
	eventType, err := c.queries.GetEventType(ctx, input.EventTypeID)
	if err != nil {
		return fmt.Errorf("failed to get event type: %w", err)
	}

	// Validate payload against JSON Schema (if schema exists)
	if len(eventType.PayloadSchema) > 0 {
		if err := validatePayload(input.Payload, eventType.PayloadSchema); err != nil {
			return fmt.Errorf("payload validation failed: %w", err)
		}
	}

	// Get the last event in scope to determine scope_sequence and prev_hash
	lastEvent, err := c.queries.GetLastEventInScope(ctx, input.ScopeID)
	var scopeSequence int64
	var prevEventHash *string

	if err != nil {
		// No previous events - this is the genesis event
		scopeSequence = 0
		prevEventHash = nil
	} else {
		scopeSequence = lastEvent.ScopeSequence + 1
		prevEventHash = &lastEvent.EventHash
	}

	// Compute event hash using RFC 8785 canonical JSON
	hashable := event.Hashable{
		ScopeID:         input.ScopeID,
		ActorID:         input.ActorID,
		EventTypeID:     input.EventTypeID,
		Intent:          input.Intent,
		ScopeSequence:   scopeSequence,
		CorrectionType:  input.CorrectionType,
		CorrectsEventID: input.CorrectsEventID,
		ObservedAt:      input.ObservedAt,
		DecidedAt:       input.DecidedAt,
		Reason:          input.Reason,
		Payload:         input.Payload,
	}

	eventHash, err := event.ComputeHash(hashable)
	if err != nil {
		return fmt.Errorf("failed to compute event hash: %w", err)
	}

	// Insert event into database
	insertParams := sqlc.InsertEventParams{
		ScopeID:       input.ScopeID,
		ActorID:       input.ActorID,
		EventTypeID:   input.EventTypeID,
		ScopeSequence: scopeSequence,
		Intent:        sqlc.EventIntent(input.Intent),
		Payload:       input.Payload,
		EventHash:     eventHash,
		Reason:        input.Reason,
		PrevEventHash: prevEventHash,
	}

	// Handle optional fields with pgtype
	if input.CorrectionType != nil {
		insertParams.CorrectionType = sqlc.NullCorrectionType{
			CorrectionType: sqlc.CorrectionType(*input.CorrectionType),
			Valid:          true,
		}
	}
	if input.CorrectsEventID != nil {
		insertParams.CorrectsEventID = pgtype.UUID{Bytes: *input.CorrectsEventID, Valid: true}
	}
	if input.ObservedAt != nil {
		insertParams.ObservedAt = pgtype.Timestamptz{Time: *input.ObservedAt, Valid: true}
	}
	if input.DecidedAt != nil {
		insertParams.DecidedAt = pgtype.Timestamptz{Time: *input.DecidedAt, Valid: true}
	}

	event, err := c.queries.InsertEvent(ctx, insertParams)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	// Create lineage edges if parent events specified
	for _, parentID := range input.ParentEventIDs {
		_, err := c.queries.CreateLineage(ctx, sqlc.CreateLineageParams{
			ParentEventID: parentID,
			ChildEventID:  event.ID,
			Relationship:  "derived_from",
		})
		if err != nil {
			log.Printf("Warning: failed to create lineage edge: %v", err)
		}
	}

	log.Printf("Processed event: id=%s scope_sequence=%d hash=%s", event.ID, event.ScopeSequence, event.EventHash)
	return nil
}

// validatePayload validates JSON payload against a JSON Schema
func validatePayload(payload json.RawMessage, schemaBytes []byte) error {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", bytes.NewReader(schemaBytes)); err != nil {
		return err
	}

	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return err
	}

	var payloadData interface{}
	if err := json.Unmarshal(payload, &payloadData); err != nil {
		return err
	}

	return schema.Validate(payloadData)
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
