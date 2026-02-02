package event

import (
	"context"
	"log/slog"

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
	event, err := r.queries.InsertEvent(ctx, params)
	if err != nil {
		slog.Error("failed to insert event", "scope_id", params.ScopeID, "error", err)
		return sqlc.Event{}, err
	}
	slog.Debug("event inserted", "id", event.ID, "scope_sequence", event.ScopeSequence)
	return event, nil
}

// GetByID retrieves an event by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Event, error) {
	event, err := r.queries.GetEvent(ctx, id)
	if err != nil {
		slog.Debug("event not found", "id", id, "error", err)
		return sqlc.Event{}, err
	}
	return event, nil
}

// GetByHash retrieves an event by hash
func (r *Repository) GetByHash(ctx context.Context, hash string) (sqlc.Event, error) {
	event, err := r.queries.GetEventByHash(ctx, hash)
	if err != nil {
		slog.Debug("event not found by hash", "hash", hash, "error", err)
		return sqlc.Event{}, err
	}
	return event, nil
}

// ListByScope retrieves events for a scope with pagination
func (r *Repository) ListByScope(ctx context.Context, params sqlc.GetEventsByScopeParams) ([]sqlc.Event, error) {
	events, err := r.queries.GetEventsByScope(ctx, params)
	if err != nil {
		slog.Error("failed to list events by scope", "scope_id", params.ScopeID, "error", err)
		return nil, err
	}
	return events, nil
}

// GetLastInScope retrieves the last event in a scope
func (r *Repository) GetLastInScope(ctx context.Context, scopeID uuid.UUID) (sqlc.Event, error) {
	event, err := r.queries.GetLastEventInScope(ctx, scopeID)
	if err != nil {
		slog.Debug("no events in scope", "scope_id", scopeID, "error", err)
		return sqlc.Event{}, err
	}
	return event, nil
}

// ListByActor retrieves events for an actor with pagination
func (r *Repository) ListByActor(ctx context.Context, params sqlc.GetEventsByActorParams) ([]sqlc.Event, error) {
	events, err := r.queries.GetEventsByActor(ctx, params)
	if err != nil {
		slog.Error("failed to list events by actor", "actor_id", params.ActorID, "error", err)
		return nil, err
	}
	return events, nil
}

// GetCorrections retrieves corrections for an event
func (r *Repository) GetCorrections(ctx context.Context, eventID uuid.UUID) ([]sqlc.Event, error) {
	events, err := r.queries.GetCorrectionsForEvent(ctx, pgtype.UUID{Bytes: eventID, Valid: true})
	if err != nil {
		slog.Error("failed to get corrections", "event_id", eventID, "error", err)
		return nil, err
	}
	return events, nil
}

// CountInScope counts events in a scope
func (r *Repository) CountInScope(ctx context.Context, scopeID uuid.UUID) (int64, error) {
	count, err := r.queries.CountEventsInScope(ctx, scopeID)
	if err != nil {
		slog.Error("failed to count events in scope", "scope_id", scopeID, "error", err)
		return 0, err
	}
	return count, nil
}
