package scope

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Handler handles HTTP requests for scopes
type Handler struct {
	repo *Repository
}

// NewHandler creates a new scope handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateRequest represents the request body for creating a scope
type CreateRequest struct {
	Project     string  `json:"project" binding:"required" example:"my-project"`
	Domain      *string `json:"domain" example:"payments"`
	Environment *string `json:"environment" example:"production"`
}

// Create handles POST /scopes
// @Summary      Create a scope
// @Description  Create a new scope for grouping events
// @Tags         scopes
// @Accept       json
// @Produce      json
// @Param        request body CreateRequest true "Scope creation request"
// @Success      201 {object} ScopeResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /scopes [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := sqlc.CreateScopeParams{
		Project:     req.Project,
		Domain:      req.Domain,
		Environment: req.Environment,
	}

	scope, err := h.repo.Create(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create scope"})
		return
	}

	c.JSON(http.StatusCreated, scope)
}

// Get handles GET /scopes/:id
// @Summary      Get a scope
// @Description  Get a scope by ID
// @Tags         scopes
// @Produce      json
// @Param        id path string true "Scope ID" format(uuid)
// @Success      200 {object} ScopeResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /scopes/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope ID"})
		return
	}

	scope, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scope not found"})
		return
	}

	c.JSON(http.StatusOK, scope)
}

// ScopeResponse represents the API response for a scope
// @Description Scope data returned by the API
type ScopeResponse struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Project     string  `json:"project" example:"my-project"`
	Domain      *string `json:"domain" example:"payments"`
	Environment *string `json:"environment" example:"production"`
	CreatedAt   string  `json:"created_at" example:"2024-01-15T10:00:00Z"`
}

// ListResponse represents the response for listing scopes
type ListResponse struct {
	Scopes []ScopeResponse `json:"scopes"`
	Count  int             `json:"count"`
}

// List handles GET /scopes
// @Summary      List scopes
// @Description  List all scopes
// @Tags         scopes
// @Produce      json
// @Success      200 {object} ListResponse
// @Failure      500 {object} map[string]string
// @Router       /scopes [get]
func (h *Handler) List(c *gin.Context) {
	scopes, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list scopes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scopes": scopes,
		"count":  len(scopes),
	})
}
