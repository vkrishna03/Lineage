package actor

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Handler handles HTTP requests for actors
type Handler struct {
	repo *Repository
}

// NewHandler creates a new actor handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateRequest represents the request body for creating an actor
type CreateRequest struct {
	Type       string          `json:"type" binding:"required" example:"llm" enums:"human,llm,agent,service,tool"`
	ExternalID *string         `json:"external_id" example:"claude-opus-4"`
	Name       *string         `json:"name" example:"Claude"`
	Metadata   json.RawMessage `json:"metadata" swaggertype:"object"`
}

var validTypes = map[string]bool{
	"human":   true,
	"llm":     true,
	"agent":   true,
	"service": true,
	"tool":    true,
}

// Create handles POST /actors
// @Summary      Create an actor
// @Description  Create a new actor (human, llm, agent, service, or tool)
// @Tags         actors
// @Accept       json
// @Produce      json
// @Param        request body CreateRequest true "Actor creation request"
// @Success      201 {object} ActorResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /actors [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor type (valid: human, llm, agent, service, tool)"})
		return
	}

	params := sqlc.CreateActorParams{
		Type:       sqlc.ActorType(req.Type),
		ExternalID: req.ExternalID,
		Name:       req.Name,
		Metadata:   req.Metadata,
	}

	actor, err := h.repo.Create(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create actor"})
		return
	}

	c.JSON(http.StatusCreated, actor)
}

// Get handles GET /actors/:id
// @Summary      Get an actor
// @Description  Get an actor by ID
// @Tags         actors
// @Produce      json
// @Param        id path string true "Actor ID" format(uuid)
// @Success      200 {object} ActorResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /actors/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor ID"})
		return
	}

	actor, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "actor not found"})
		return
	}

	c.JSON(http.StatusOK, actor)
}

// ActorResponse represents the API response for an actor
// @Description Actor data returned by the API
type ActorResponse struct {
	ID           string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type         string  `json:"type" example:"llm"`
	ExternalID   *string `json:"external_id" example:"claude-opus-4"`
	Name         *string `json:"name" example:"Claude"`
	Metadata     any     `json:"metadata" swaggertype:"object"`
	RegisteredAt string  `json:"registered_at" example:"2024-01-15T10:00:00Z"`
}

// ListResponse represents the response for listing actors
type ListResponse struct {
	Actors []ActorResponse `json:"actors"`
	Count  int             `json:"count"`
}

// List handles GET /actors
// @Summary      List actors
// @Description  List all actors
// @Tags         actors
// @Produce      json
// @Success      200 {object} ListResponse
// @Failure      500 {object} map[string]string
// @Router       /actors [get]
func (h *Handler) List(c *gin.Context) {
	actors, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list actors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"actors": actors,
		"count":  len(actors),
	})
}
