package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/db/sqlc"
)

type EventTypesHandler struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewEventTypesHandler(db *pgxpool.Pool) *EventTypesHandler {
	return &EventTypesHandler{
		db:      db,
		queries: sqlc.New(db),
	}
}

type CreateEventTypeRequest struct {
	Name           string          `json:"name" binding:"required"`
	Version        string          `json:"version" binding:"required"`
	Description    *string         `json:"description"`
	PayloadSchema  json.RawMessage `json:"payload_schema"`
	AllowedIntents []string        `json:"allowed_intents"`
}

func (h *EventTypesHandler) CreateEventType(c *gin.Context) {
	var req CreateEventTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := sqlc.CreateEventTypeParams{
		Name:           req.Name,
		Version:        req.Version,
		Description:    req.Description,
		PayloadSchema:  req.PayloadSchema,
		AllowedIntents: req.AllowedIntents,
	}

	eventType, err := h.queries.CreateEventType(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event type"})
		return
	}

	c.JSON(http.StatusCreated, eventType)
}

func (h *EventTypesHandler) GetEventType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event type ID"})
		return
	}

	eventType, err := h.queries.GetEventType(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event type not found"})
		return
	}

	c.JSON(http.StatusOK, eventType)
}

func (h *EventTypesHandler) ListEventTypes(c *gin.Context) {
	eventTypes, err := h.queries.ListActiveEventTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list event types"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_types": eventTypes,
		"count":       len(eventTypes),
	})
}
