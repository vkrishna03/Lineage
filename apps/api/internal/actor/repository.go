package actor

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Repository handles actor data access
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new actor repository
func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

// Create creates a new actor
func (r *Repository) Create(ctx context.Context, params sqlc.CreateActorParams) (sqlc.Actor, error) {
	actor, err := r.queries.CreateActor(ctx, params)
	if err != nil {
		slog.Error("failed to create actor", "type", params.Type, "error", err)
		return sqlc.Actor{}, err
	}
	slog.Debug("actor created", "id", actor.ID, "type", actor.Type)
	return actor, nil
}

// GetByID retrieves an actor by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Actor, error) {
	actor, err := r.queries.GetActor(ctx, id)
	if err != nil {
		slog.Debug("actor not found", "id", id, "error", err)
		return sqlc.Actor{}, err
	}
	return actor, nil
}

// List retrieves all actors
func (r *Repository) List(ctx context.Context) ([]sqlc.Actor, error) {
	actors, err := r.queries.ListActors(ctx)
	if err != nil {
		slog.Error("failed to list actors", "error", err)
		return nil, err
	}
	return actors, nil
}

// GetByExternalID retrieves an actor by type and external ID
func (r *Repository) GetByExternalID(ctx context.Context, params sqlc.GetActorByExternalIDParams) (sqlc.Actor, error) {
	actor, err := r.queries.GetActorByExternalID(ctx, params)
	if err != nil {
		slog.Debug("actor not found by external id", "type", params.Type, "error", err)
		return sqlc.Actor{}, err
	}
	return actor, nil
}

// ListByType retrieves actors by type
func (r *Repository) ListByType(ctx context.Context, actorType sqlc.ActorType) ([]sqlc.Actor, error) {
	actors, err := r.queries.ListActorsByType(ctx, actorType)
	if err != nil {
		slog.Error("failed to list actors by type", "type", actorType, "error", err)
		return nil, err
	}
	return actors, nil
}
