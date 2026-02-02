package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/db/sqlc"
	"github.com/lineage/api/internal/domain"
	"github.com/lineage/api/internal/kafka"
)

type EventsHandler struct {
	db       *pgxpool.Pool
	queries  *sqlc.Queries
	producer *kafka.Producer
}

func NewEventsHandler(db *pgxpool.Pool, producer *kafka.Producer) *EventsHandler {
	return &EventsHandler{
		db:       db,
		queries:  sqlc.New(db),
		producer: producer,
	}
}

// CreateEvent accepts an event and sends it to Kafka
func (h *EventsHandler) CreateEvent(c *gin.Context) {
	var input domain.EventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate intent
	if !domain.ValidateIntent(input.Intent) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intent"})
		return
	}

	// Validate correction type if present
	if input.CorrectionType != nil && !domain.ValidateCorrectionType(*input.CorrectionType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid correction_type"})
		return
	}

	// Send to Kafka
	if err := h.producer.ProduceEvent(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to produce event"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"message": "Event queued for processing",
	})
}

// GetEvent retrieves a single event by ID
func (h *EventsHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	event, err := h.queries.GetEvent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// ListEventsByScope retrieves events for a given scope
func (h *EventsHandler) ListEventsByScope(c *gin.Context) {
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

	limit := int32(100)
	offset := int32(0)

	events, err := h.queries.GetEventsByScope(c.Request.Context(), sqlc.GetEventsByScopeParams{
		ScopeID: scopeID,
		Limit:   limit,
		Offset:  offset,
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

// GetEventLineage retrieves parent and child events
func (h *EventsHandler) GetEventLineage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	parents, err := h.queries.GetParents(c.Request.Context(), id)
	if err != nil {
		parents = []sqlc.Event{}
	}

	children, err := h.queries.GetChildren(c.Request.Context(), id)
	if err != nil {
		children = []sqlc.Event{}
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id": id,
		"parents":  parents,
		"children": children,
	})
}
