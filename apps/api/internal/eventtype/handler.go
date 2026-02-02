package eventtype

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Handler handles HTTP requests for event types
type Handler struct {
	repo *Repository
}

// NewHandler creates a new event type handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateRequest represents the request body for creating an event type
type CreateRequest struct {
	Name           string          `json:"name" binding:"required"`
	Version        string          `json:"version" binding:"required"`
	Description    *string         `json:"description"`
	PayloadSchema  json.RawMessage `json:"payload_schema"`
	AllowedIntents []string        `json:"allowed_intents"`
}

// Create handles POST /event-types
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
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

	eventType, err := h.repo.Create(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event type"})
		return
	}

	c.JSON(http.StatusCreated, eventType)
}

// Get handles GET /event-types/:id
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event type ID"})
		return
	}

	eventType, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event type not found"})
		return
	}

	c.JSON(http.StatusOK, eventType)
}

// List handles GET /event-types
func (h *Handler) List(c *gin.Context) {
	eventTypes, err := h.repo.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list event types"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_types": eventTypes,
		"count":       len(eventTypes),
	})
}
