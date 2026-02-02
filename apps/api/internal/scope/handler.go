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
	Project     string  `json:"project" binding:"required"`
	Domain      *string `json:"domain"`
	Environment *string `json:"environment"`
}

// Create handles POST /scopes
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

// List handles GET /scopes
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
