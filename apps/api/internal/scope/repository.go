package scope

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Repository handles scope data access
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new scope repository
func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

// Create creates a new scope
func (r *Repository) Create(ctx context.Context, params sqlc.CreateScopeParams) (sqlc.Scope, error) {
	scope, err := r.queries.CreateScope(ctx, params)
	if err != nil {
		slog.Error("failed to create scope", "project", params.Project, "error", err)
		return sqlc.Scope{}, err
	}
	slog.Debug("scope created", "id", scope.ID, "project", scope.Project)
	return scope, nil
}

// GetByID retrieves a scope by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Scope, error) {
	scope, err := r.queries.GetScope(ctx, id)
	if err != nil {
		slog.Debug("scope not found", "id", id, "error", err)
		return sqlc.Scope{}, err
	}
	return scope, nil
}

// List retrieves all scopes
func (r *Repository) List(ctx context.Context) ([]sqlc.Scope, error) {
	scopes, err := r.queries.ListScopes(ctx)
	if err != nil {
		slog.Error("failed to list scopes", "error", err)
		return nil, err
	}
	return scopes, nil
}

// GetByProject retrieves a scope by project, domain, and environment
func (r *Repository) GetByProject(ctx context.Context, params sqlc.GetScopeByProjectParams) (sqlc.Scope, error) {
	scope, err := r.queries.GetScopeByProject(ctx, params)
	if err != nil {
		slog.Debug("scope not found by project", "project", params.Project, "error", err)
		return sqlc.Scope{}, err
	}
	return scope, nil
}
