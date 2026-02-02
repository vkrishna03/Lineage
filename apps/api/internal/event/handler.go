package event

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
	"github.com/lineage/api/internal/lineage"
)

// Producer defines the interface for producing events to a message queue
type Producer interface {
	ProduceEvent(ctx context.Context, input Input) error
}

// Handler handles HTTP requests for events
type Handler struct {
	repo        *Repository
	lineageRepo *lineage.Repository
	producer    Producer
}

// NewHandler creates a new event handler
func NewHandler(repo *Repository, lineageRepo *lineage.Repository, producer Producer) *Handler {
	return &Handler{
		repo:        repo,
		lineageRepo: lineageRepo,
		producer:    producer,
	}
}

// Create handles POST /events - sends event to Kafka
func (h *Handler) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !ValidateIntent(input.Intent) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intent"})
		return
	}

	if input.CorrectionType != nil && !ValidateCorrectionType(*input.CorrectionType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid correction_type"})
		return
	}

	if err := h.producer.ProduceEvent(c.Request.Context(), input); err != nil {
		slog.Error("kafka produce failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to produce event: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"message": "Event queued for processing",
	})
}

// Get handles GET /events/:id
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	event, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// ListByScope handles GET /events?scope_id=...
func (h *Handler) ListByScope(c *gin.Context) {
	scopeIDStr := c.Query("scope_id")
	if scopeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope_id is required"})
		return
	}

	scopeID, err := uuid.Parse(scopeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope_id"})
		return
	}

	events, err := h.repo.ListByScope(c.Request.Context(), sqlc.GetEventsByScopeParams{
		ScopeID: scopeID,
		Limit:   100,
		Offset:  0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// GetLineage handles GET /events/:id/lineage
func (h *Handler) GetLineage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	parents, err := h.lineageRepo.GetParents(c.Request.Context(), id)
	if err != nil {
		parents = []sqlc.Event{}
	}

	children, err := h.lineageRepo.GetChildren(c.Request.Context(), id)
	if err != nil {
		children = []sqlc.Event{}
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id": id,
		"parents":  parents,
		"children": children,
	})
}
