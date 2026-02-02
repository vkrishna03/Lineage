package lineage

import (
	"context"

	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Repository handles lineage data access
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new lineage repository
func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

// Create creates a new lineage edge
func (r *Repository) Create(ctx context.Context, params sqlc.CreateLineageParams) (sqlc.EventLineage, error) {
	return r.queries.CreateLineage(ctx, params)
}

// GetParents retrieves parent events for an event
func (r *Repository) GetParents(ctx context.Context, childID uuid.UUID) ([]sqlc.Event, error) {
	return r.queries.GetParents(ctx, childID)
}

// GetChildren retrieves child events for an event
func (r *Repository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]sqlc.Event, error) {
	return r.queries.GetChildren(ctx, parentID)
}

// GetEdge retrieves a specific lineage edge
func (r *Repository) GetEdge(ctx context.Context, params sqlc.GetLineageEdgeParams) (sqlc.EventLineage, error) {
	return r.queries.GetLineageEdge(ctx, params)
}

// GetDirectParentIDs retrieves direct parent event IDs
func (r *Repository) GetDirectParentIDs(ctx context.Context, childID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.GetDirectParentIDs(ctx, childID)
}

// GetDirectChildIDs retrieves direct child event IDs
func (r *Repository) GetDirectChildIDs(ctx context.Context, parentID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.GetDirectChildIDs(ctx, parentID)
}
