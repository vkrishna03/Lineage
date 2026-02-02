package scope

import (
	"context"

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
	return r.queries.CreateScope(ctx, params)
}

// GetByID retrieves a scope by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Scope, error) {
	return r.queries.GetScope(ctx, id)
}

// List retrieves all scopes
func (r *Repository) List(ctx context.Context) ([]sqlc.Scope, error) {
	return r.queries.ListScopes(ctx)
}

// GetByProject retrieves a scope by project, domain, and environment
func (r *Repository) GetByProject(ctx context.Context, params sqlc.GetScopeByProjectParams) (sqlc.Scope, error) {
	return r.queries.GetScopeByProject(ctx, params)
}
