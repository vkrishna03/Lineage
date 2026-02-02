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
	Name           string          `json:"name" binding:"required" example:"decision"`
	Version        string          `json:"version" binding:"required" example:"1.0"`
	Description    *string         `json:"description" example:"A decision event type"`
	PayloadSchema  json.RawMessage `json:"payload_schema" swaggertype:"object"`
	AllowedIntents []string        `json:"allowed_intents" example:"exploration,suggestion,assertion,decision,execution"`
}

// Create handles POST /event-types
// @Summary      Create an event type
// @Description  Create a new event type with optional JSON schema for payload validation
// @Tags         event-types
// @Accept       json
// @Produce      json
// @Param        request body CreateRequest true "Event type creation request"
// @Success      201 {object} EventTypeResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /event-types [post]
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
// @Summary      Get an event type
// @Description  Get an event type by ID
// @Tags         event-types
// @Produce      json
// @Param        id path string true "Event Type ID" format(uuid)
// @Success      200 {object} EventTypeResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /event-types/{id} [get]
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

// EventTypeResponse represents the API response for an event type
// @Description Event type data returned by the API
type EventTypeResponse struct {
	ID             string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name           string   `json:"name" example:"decision"`
	Version        string   `json:"version" example:"1.0"`
	Description    *string  `json:"description" example:"A decision event type"`
	PayloadSchema  any      `json:"payload_schema" swaggertype:"object"`
	AllowedIntents []string `json:"allowed_intents" example:"exploration,suggestion,assertion,decision,execution"`
	IsActive       bool     `json:"is_active" example:"true"`
	CreatedAt      string   `json:"created_at" example:"2024-01-15T10:00:00Z"`
}

// ListResponse represents the response for listing event types
type ListResponse struct {
	EventTypes []EventTypeResponse `json:"event_types"`
	Count      int                 `json:"count"`
}

// List handles GET /event-types
// @Summary      List event types
// @Description  List all active event types
// @Tags         event-types
// @Produce      json
// @Success      200 {object} ListResponse
// @Failure      500 {object} map[string]string
// @Router       /event-types [get]
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
