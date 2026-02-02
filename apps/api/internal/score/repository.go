package score

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Repository handles event score data access
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new score repository
func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{queries: q}
}

// Create creates a new score for an event
func (r *Repository) Create(ctx context.Context, params sqlc.InsertScoreParams) (sqlc.EventScore, error) {
	score, err := r.queries.InsertScore(ctx, params)
	if err != nil {
		slog.Error("failed to create score", "event_id", params.EventID, "type", params.Type, "error", err)
		return sqlc.EventScore{}, err
	}
	slog.Debug("score created", "id", score.ID, "event_id", score.EventID, "type", score.Type)
	return score, nil
}

// GetForEvent retrieves all scores for an event
func (r *Repository) GetForEvent(ctx context.Context, eventID uuid.UUID) ([]sqlc.EventScore, error) {
	scores, err := r.queries.GetScoresForEvent(ctx, eventID)
	if err != nil {
		slog.Error("failed to get scores for event", "event_id", eventID, "error", err)
		return nil, err
	}
	return scores, nil
}

// GetByType retrieves scores of a specific type for an event
func (r *Repository) GetByType(ctx context.Context, eventID uuid.UUID, scoreType sqlc.ScoreType) ([]sqlc.EventScore, error) {
	scores, err := r.queries.GetScoresByType(ctx, sqlc.GetScoresByTypeParams{
		EventID: eventID,
		Type:    scoreType,
	})
	if err != nil {
		slog.Error("failed to get scores by type", "event_id", eventID, "type", scoreType, "error", err)
		return nil, err
	}
	return scores, nil
}

// GetLatest retrieves the most recent score of a specific type for an event
func (r *Repository) GetLatest(ctx context.Context, eventID uuid.UUID, scoreType sqlc.ScoreType) (sqlc.EventScore, error) {
	score, err := r.queries.GetLatestScoreByType(ctx, sqlc.GetLatestScoreByTypeParams{
		EventID: eventID,
		Type:    scoreType,
	})
	if err != nil {
		slog.Debug("latest score not found", "event_id", eventID, "type", scoreType, "error", err)
		return sqlc.EventScore{}, err
	}
	return score, nil
}
