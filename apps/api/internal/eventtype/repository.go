package eventtype

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Repository handles event type data access
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new event type repository
func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

// Create creates a new event type
func (r *Repository) Create(ctx context.Context, params sqlc.CreateEventTypeParams) (sqlc.EventTypeRegistry, error) {
	eventType, err := r.queries.CreateEventType(ctx, params)
	if err != nil {
		slog.Error("failed to create event type", "name", params.Name, "version", params.Version, "error", err)
		return sqlc.EventTypeRegistry{}, err
	}
	slog.Debug("event type created", "id", eventType.ID, "name", eventType.Name)
	return eventType, nil
}

// GetByID retrieves an event type by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.EventTypeRegistry, error) {
	eventType, err := r.queries.GetEventType(ctx, id)
	if err != nil {
		slog.Debug("event type not found", "id", id, "error", err)
		return sqlc.EventTypeRegistry{}, err
	}
	return eventType, nil
}

// GetByNameVersion retrieves an event type by name and version
func (r *Repository) GetByNameVersion(ctx context.Context, params sqlc.GetEventTypeByNameVersionParams) (sqlc.EventTypeRegistry, error) {
	eventType, err := r.queries.GetEventTypeByNameVersion(ctx, params)
	if err != nil {
		slog.Debug("event type not found", "name", params.Name, "version", params.Version, "error", err)
		return sqlc.EventTypeRegistry{}, err
	}
	return eventType, nil
}

// ListActive retrieves all active event types
func (r *Repository) ListActive(ctx context.Context) ([]sqlc.EventTypeRegistry, error) {
	eventTypes, err := r.queries.ListActiveEventTypes(ctx)
	if err != nil {
		slog.Error("failed to list event types", "error", err)
		return nil, err
	}
	return eventTypes, nil
}

// Deactivate deactivates an event type
func (r *Repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeactivateEventType(ctx, id)
	if err != nil {
		slog.Error("failed to deactivate event type", "id", id, "error", err)
		return err
	}
	slog.Debug("event type deactivated", "id", id)
	return nil
}
