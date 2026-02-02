package score

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lineage/api/internal/app/db/sqlc"
	"github.com/lineage/api/internal/event"
	"github.com/shopspring/decimal"
)

// Handler handles HTTP requests for scores
type Handler struct {
	repo *Repository
}

// NewHandler creates a new score handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateRequest represents the request body for creating a score
type CreateRequest struct {
	Type     string          `json:"type" binding:"required" example:"confidence" enums:"confidence,relevance,reliability,agreement"`
	Value    float64         `json:"value" binding:"required" example:"0.85"`
	ScoredBy *uuid.UUID      `json:"scored_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Reason   *string         `json:"reason,omitempty" example:"High confidence based on training data coverage"`
	Metadata json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// ScoreResponse represents the API response for a score
// @Description Score data returned by the API
type ScoreResponse struct {
	ID        string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	EventID   string   `json:"event_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Type      string   `json:"type" example:"confidence"`
	Value     float64  `json:"value" example:"0.85"`
	Category  string   `json:"category" example:"high"`
	ScoredBy  *string  `json:"scored_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	Reason    *string  `json:"reason,omitempty" example:"High confidence based on training data coverage"`
	Metadata  any      `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt string   `json:"created_at" example:"2024-01-15T10:00:00Z"`
}

// valid score types
var validScoreTypes = map[string]bool{
	"confidence":  true,
	"relevance":   true,
	"reliability": true,
	"agreement":   true,
}

// Create handles POST /events/:id/scores
// @Summary      Add score to event
// @Description  Add a new score (confidence, relevance, reliability, agreement) to an event
// @Tags         scores
// @Accept       json
// @Produce      json
// @Param        id path string true "Event ID" format(uuid)
// @Param        request body CreateRequest true "Score creation request"
// @Success      201 {object} ScoreResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /events/{id}/scores [post]
func (h *Handler) Create(c *gin.Context) {
	idStr := c.Param("id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate score type
	if !validScoreTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score type (valid: confidence, relevance, reliability, agreement)"})
		return
	}

	// Validate value range
	if req.Value < 0 || req.Value > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value must be between 0.0 and 1.0"})
		return
	}

	// Derive category from value
	category := event.DeriveScoreCategory(req.Value)

	// Convert value to pgtype.Numeric
	decValue := decimal.NewFromFloat(req.Value)
	numericValue := pgtype.Numeric{}
	if err := numericValue.Scan(decValue.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert value"})
		return
	}

	// Convert scored_by to pgtype.UUID
	var scoredBy pgtype.UUID
	if req.ScoredBy != nil {
		scoredBy = pgtype.UUID{Bytes: *req.ScoredBy, Valid: true}
	}

	params := sqlc.InsertScoreParams{
		EventID:  eventID,
		Type:     sqlc.ScoreType(req.Type),
		Value:    numericValue,
		Category: sqlc.ScoreCategory(category),
		ScoredBy: scoredBy,
		Reason:   req.Reason,
		Metadata: req.Metadata,
	}

	score, err := h.repo.Create(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create score"})
		return
	}

	c.JSON(http.StatusCreated, score)
}

// ListResponse represents the response for listing scores
type ListResponse struct {
	Scores []ScoreResponse `json:"scores"`
	Count  int             `json:"count"`
}

// GetForEvent handles GET /events/:id/scores
// @Summary      Get scores for event
// @Description  Get all scores for an event, optionally filtered by type
// @Tags         scores
// @Produce      json
// @Param        id path string true "Event ID" format(uuid)
// @Param        type query string false "Filter by score type" Enums(confidence,relevance,reliability,agreement)
// @Success      200 {object} ListResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /events/{id}/scores [get]
func (h *Handler) GetForEvent(c *gin.Context) {
	idStr := c.Param("id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	scoreType := c.Query("type")

	var scores []sqlc.EventScore
	if scoreType != "" {
		if !validScoreTypes[scoreType] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score type (valid: confidence, relevance, reliability, agreement)"})
			return
		}
		scores, err = h.repo.GetByType(c.Request.Context(), eventID, sqlc.ScoreType(scoreType))
	} else {
		scores, err = h.repo.GetForEvent(c.Request.Context(), eventID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get scores"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scores": scores,
		"count":  len(scores),
	})
}
