package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Repository handles event data access
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new event repository
func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

// Insert inserts a new event
func (r *Repository) Insert(ctx context.Context, params sqlc.InsertEventParams) (sqlc.Event, error) {
	return r.queries.InsertEvent(ctx, params)
}

// GetByID retrieves an event by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Event, error) {
	return r.queries.GetEvent(ctx, id)
}

// GetByHash retrieves an event by hash
func (r *Repository) GetByHash(ctx context.Context, hash string) (sqlc.Event, error) {
	return r.queries.GetEventByHash(ctx, hash)
}

// ListByScope retrieves events for a scope with pagination
func (r *Repository) ListByScope(ctx context.Context, params sqlc.GetEventsByScopeParams) ([]sqlc.Event, error) {
	return r.queries.GetEventsByScope(ctx, params)
}

// GetLastInScope retrieves the last event in a scope
func (r *Repository) GetLastInScope(ctx context.Context, scopeID uuid.UUID) (sqlc.Event, error) {
	return r.queries.GetLastEventInScope(ctx, scopeID)
}

// ListByActor retrieves events for an actor with pagination
func (r *Repository) ListByActor(ctx context.Context, params sqlc.GetEventsByActorParams) ([]sqlc.Event, error) {
	return r.queries.GetEventsByActor(ctx, params)
}

// GetCorrections retrieves corrections for an event
func (r *Repository) GetCorrections(ctx context.Context, eventID uuid.UUID) ([]sqlc.Event, error) {
	return r.queries.GetCorrectionsForEvent(ctx, pgtype.UUID{Bytes: eventID, Valid: true})
}

// CountInScope counts events in a scope
func (r *Repository) CountInScope(ctx context.Context, scopeID uuid.UUID) (int64, error) {
	return r.queries.CountEventsInScope(ctx, scopeID)
}
