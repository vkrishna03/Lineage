package artifact

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Repository handles artifact data access
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new artifact repository
func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{queries: q}
}

// Create creates a new artifact
func (r *Repository) Create(ctx context.Context, params sqlc.CreateArtifactParams) (sqlc.Artifact, error) {
	artifact, err := r.queries.CreateArtifact(ctx, params)
	if err != nil {
		slog.Error("failed to create artifact", "scope_id", params.ScopeID, "content_hash", params.ContentHash, "error", err)
		return sqlc.Artifact{}, err
	}
	slog.Debug("artifact created", "id", artifact.ID, "content_hash", artifact.ContentHash)
	return artifact, nil
}

// GetByID retrieves an artifact by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Artifact, error) {
	artifact, err := r.queries.GetArtifact(ctx, id)
	if err != nil {
		slog.Debug("artifact not found", "id", id, "error", err)
		return sqlc.Artifact{}, err
	}
	return artifact, nil
}

// GetByHash retrieves an artifact by scope and content hash (for deduplication)
func (r *Repository) GetByHash(ctx context.Context, scopeID uuid.UUID, contentHash string) (sqlc.Artifact, error) {
	artifact, err := r.queries.GetArtifactByHash(ctx, sqlc.GetArtifactByHashParams{
		ScopeID:     scopeID,
		ContentHash: contentHash,
	})
	if err != nil {
		slog.Debug("artifact not found by hash", "scope_id", scopeID, "content_hash", contentHash, "error", err)
		return sqlc.Artifact{}, err
	}
	return artifact, nil
}

// ListByScope lists artifacts in a scope
func (r *Repository) ListByScope(ctx context.Context, params sqlc.ListArtifactsInScopeParams) ([]sqlc.Artifact, error) {
	artifacts, err := r.queries.ListArtifactsInScope(ctx, params)
	if err != nil {
		slog.Error("failed to list artifacts", "scope_id", params.ScopeID, "error", err)
		return nil, err
	}
	return artifacts, nil
}

// LinkToEvent links an artifact to an event
func (r *Repository) LinkToEvent(ctx context.Context, params sqlc.LinkArtifactParams) (sqlc.EventArtifact, error) {
	link, err := r.queries.LinkArtifact(ctx, params)
	if err != nil {
		slog.Error("failed to link artifact to event", "event_id", params.EventID, "artifact_id", params.ArtifactID, "error", err)
		return sqlc.EventArtifact{}, err
	}
	slog.Debug("artifact linked to event", "event_id", params.EventID, "artifact_id", params.ArtifactID, "role", params.Role)
	return link, nil
}

// GetForEvent retrieves all artifacts for an event
func (r *Repository) GetForEvent(ctx context.Context, eventID uuid.UUID) ([]sqlc.Artifact, error) {
	artifacts, err := r.queries.GetArtifactsForEvent(ctx, eventID)
	if err != nil {
		slog.Error("failed to get artifacts for event", "event_id", eventID, "error", err)
		return nil, err
	}
	return artifacts, nil
}

// GetInputsForEvent retrieves input artifacts for an event
func (r *Repository) GetInputsForEvent(ctx context.Context, eventID uuid.UUID) ([]sqlc.Artifact, error) {
	return r.queries.GetInputArtifactsForEvent(ctx, eventID)
}

// GetOutputsForEvent retrieves output artifacts for an event
func (r *Repository) GetOutputsForEvent(ctx context.Context, eventID uuid.UUID) ([]sqlc.Artifact, error) {
	return r.queries.GetOutputArtifactsForEvent(ctx, eventID)
}

// GetEventsForArtifact retrieves all events that reference an artifact
func (r *Repository) GetEventsForArtifact(ctx context.Context, artifactID uuid.UUID) ([]sqlc.Event, error) {
	return r.queries.GetEventsForArtifact(ctx, artifactID)
}
