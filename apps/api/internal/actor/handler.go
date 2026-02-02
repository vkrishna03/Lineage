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
	Type       string          `json:"type" binding:"required"`
	ExternalID *string         `json:"external_id"`
	Name       *string         `json:"name"`
	Metadata   json.RawMessage `json:"metadata"`
}

var validTypes = map[string]bool{
	"human":   true,
	"llm":     true,
	"agent":   true,
	"service": true,
	"tool":    true,
}

// Create handles POST /actors
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor type"})
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

// List handles GET /actors
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
