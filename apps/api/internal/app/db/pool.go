package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/app/config"
)

// NewPool creates a new database connection pool
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, cfg.DatabaseURL)
}
