package event

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
	"github.com/lineage/api/internal/eventtype"
	"github.com/lineage/api/internal/lineage"
)

// Producer defines the interface for producing events to a message queue
type Producer interface {
	ProduceEvent(ctx context.Context, input Input) error
}

// Handler handles HTTP requests for events
type Handler struct {
	repo          *Repository
	lineageRepo   *lineage.Repository
	eventtypeRepo *eventtype.Repository
	producer      Producer
	validator     *Validator
}

// NewHandler creates a new event handler
func NewHandler(repo *Repository, lineageRepo *lineage.Repository, eventtypeRepo *eventtype.Repository, producer Producer) *Handler {
	return &Handler{
		repo:          repo,
		lineageRepo:   lineageRepo,
		eventtypeRepo: eventtypeRepo,
		producer:      producer,
		validator:     NewValidator(),
	}
}

// CreateResponse represents the response for event creation
type CreateResponse struct {
	Status  string `json:"status" example:"accepted"`
	Message string `json:"message" example:"Event queued for processing"`
}

// Create handles POST /events - sends event to Kafka
// @Summary      Create an event
// @Description  Submit an event to the processing queue. Events are processed asynchronously and will be hash-chained.
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        request body Input true "Event creation request"
// @Success      202 {object} CreateResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /events [post]
func (h *Handler) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !ValidateIntent(input.Intent) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intent (valid: exploration, suggestion, assertion, decision, execution)"})
		return
	}

	if input.CorrectionType != nil && !ValidateCorrectionType(*input.CorrectionType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid correction_type (valid: supersede, amend, retract)"})
		return
	}

	// Fetch event type and validate payload against schema
	eventType, err := h.eventtypeRepo.GetByID(c.Request.Context(), input.EventTypeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_type_id"})
		return
	}

	if err := h.validator.ValidatePayload(input.Payload, eventType.PayloadSchema); err != nil {
		slog.Debug("payload validation failed", "event_type_id", input.EventTypeID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
// @Summary      Get an event
// @Description  Get an event by ID
// @Tags         events
// @Produce      json
// @Param        id path string true "Event ID" format(uuid)
// @Success      200 {object} EventResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /events/{id} [get]
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

// EventResponse represents the API response for an event
// @Description Event data returned by the API
type EventResponse struct {
	ID              string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ScopeID         string  `json:"scope_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ActorID         string  `json:"actor_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	EventTypeID     string  `json:"event_type_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	ScopeSequence   int64   `json:"scope_sequence" example:"1"`
	Intent          string  `json:"intent" example:"decision"`
	Reason          *string `json:"reason" example:"User approved the recommendation"`
	CorrectionType  *string `json:"correction_type" example:"supersede"`
	CorrectsEventID *string `json:"corrects_event_id"`
	ObservedAt      *string `json:"observed_at" example:"2024-01-15T10:30:00Z"`
	DecidedAt       *string `json:"decided_at" example:"2024-01-15T10:35:00Z"`
	IngestedAt      string  `json:"ingested_at" example:"2024-01-15T10:40:00Z"`
	PrevEventHash   *string `json:"prev_event_hash" example:"abc123..."`
	EventHash       string  `json:"event_hash" example:"def456..."`
	Payload         any     `json:"payload"`
}

// ListResponse represents the response for listing events
type ListResponse struct {
	Events []EventResponse `json:"events"`
	Count  int             `json:"count"`
}

// ListByScope handles GET /events?scope_id=...
// @Summary      List events by scope
// @Description  List all events in a scope, ordered by scope_sequence
// @Tags         events
// @Produce      json
// @Param        scope_id query string true "Scope ID" format(uuid)
// @Success      200 {object} ListResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /events [get]
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

// LineageResponse represents the response for event lineage
type LineageResponse struct {
	EventID  string          `json:"event_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Parents  []EventResponse `json:"parents"`
	Children []EventResponse `json:"children"`
}

// GetLineage handles GET /events/:id/lineage
// @Summary      Get event lineage
// @Description  Get parent and child events for a given event
// @Tags         events
// @Produce      json
// @Param        id path string true "Event ID" format(uuid)
// @Success      200 {object} LineageResponse
// @Failure      400 {object} map[string]string
// @Router       /events/{id}/lineage [get]
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
