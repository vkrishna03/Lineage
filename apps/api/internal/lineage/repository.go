package lineage

import (
	"context"
	"log/slog"

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
	edge, err := r.queries.CreateLineage(ctx, params)
	if err != nil {
		slog.Error("failed to create lineage edge", "parent_id", params.ParentEventID, "child_id", params.ChildEventID, "error", err)
		return sqlc.EventLineage{}, err
	}
	slog.Debug("lineage edge created", "parent_id", params.ParentEventID, "child_id", params.ChildEventID)
	return edge, nil
}

// GetParents retrieves parent events for an event
func (r *Repository) GetParents(ctx context.Context, childID uuid.UUID) ([]sqlc.Event, error) {
	events, err := r.queries.GetParents(ctx, childID)
	if err != nil {
		slog.Error("failed to get parent events", "child_id", childID, "error", err)
		return nil, err
	}
	return events, nil
}

// GetChildren retrieves child events for an event
func (r *Repository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]sqlc.Event, error) {
	events, err := r.queries.GetChildren(ctx, parentID)
	if err != nil {
		slog.Error("failed to get child events", "parent_id", parentID, "error", err)
		return nil, err
	}
	return events, nil
}

// GetEdge retrieves a specific lineage edge
func (r *Repository) GetEdge(ctx context.Context, params sqlc.GetLineageEdgeParams) (sqlc.EventLineage, error) {
	edge, err := r.queries.GetLineageEdge(ctx, params)
	if err != nil {
		slog.Debug("lineage edge not found", "parent_id", params.ParentEventID, "child_id", params.ChildEventID, "error", err)
		return sqlc.EventLineage{}, err
	}
	return edge, nil
}

// GetDirectParentIDs retrieves direct parent event IDs
func (r *Repository) GetDirectParentIDs(ctx context.Context, childID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := r.queries.GetDirectParentIDs(ctx, childID)
	if err != nil {
		slog.Error("failed to get parent ids", "child_id", childID, "error", err)
		return nil, err
	}
	return ids, nil
}

// GetDirectChildIDs retrieves direct child event IDs
func (r *Repository) GetDirectChildIDs(ctx context.Context, parentID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := r.queries.GetDirectChildIDs(ctx, parentID)
	if err != nil {
		slog.Error("failed to get child ids", "parent_id", parentID, "error", err)
		return nil, err
	}
	return ids, nil
}
