package eventtype

import (
	"context"

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
	return r.queries.CreateEventType(ctx, params)
}

// GetByID retrieves an event type by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.EventTypeRegistry, error) {
	return r.queries.GetEventType(ctx, id)
}

// GetByNameVersion retrieves an event type by name and version
func (r *Repository) GetByNameVersion(ctx context.Context, params sqlc.GetEventTypeByNameVersionParams) (sqlc.EventTypeRegistry, error) {
	return r.queries.GetEventTypeByNameVersion(ctx, params)
}

// ListActive retrieves all active event types
func (r *Repository) ListActive(ctx context.Context) ([]sqlc.EventTypeRegistry, error) {
	return r.queries.ListActiveEventTypes(ctx)
}

// Deactivate deactivates an event type
func (r *Repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeactivateEventType(ctx, id)
}
