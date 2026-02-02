package actor

import (
	"context"

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
	return r.queries.CreateActor(ctx, params)
}

// GetByID retrieves an actor by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Actor, error) {
	return r.queries.GetActor(ctx, id)
}

// List retrieves all actors
func (r *Repository) List(ctx context.Context) ([]sqlc.Actor, error) {
	return r.queries.ListActors(ctx)
}

// GetByExternalID retrieves an actor by type and external ID
func (r *Repository) GetByExternalID(ctx context.Context, params sqlc.GetActorByExternalIDParams) (sqlc.Actor, error) {
	return r.queries.GetActorByExternalID(ctx, params)
}

// ListByType retrieves actors by type
func (r *Repository) ListByType(ctx context.Context, actorType sqlc.ActorType) ([]sqlc.Actor, error) {
	return r.queries.ListActorsByType(ctx, actorType)
}
