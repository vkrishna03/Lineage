package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/db/sqlc"
)

type ScopesHandler struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewScopesHandler(db *pgxpool.Pool) *ScopesHandler {
	return &ScopesHandler{
		db:      db,
		queries: sqlc.New(db),
	}
}

type CreateScopeRequest struct {
	Project     string  `json:"project" binding:"required"`
	Domain      *string `json:"domain"`
	Environment *string `json:"environment"`
}

func (h *ScopesHandler) CreateScope(c *gin.Context) {
	var req CreateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := sqlc.CreateScopeParams{
		Project:     req.Project,
		Domain:      req.Domain,
		Environment: req.Environment,
	}

	scope, err := h.queries.CreateScope(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create scope"})
		return
	}

	c.JSON(http.StatusCreated, scope)
}

func (h *ScopesHandler) GetScope(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope ID"})
		return
	}

	scope, err := h.queries.GetScope(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scope not found"})
		return
	}

	c.JSON(http.StatusOK, scope)
}

func (h *ScopesHandler) ListScopes(c *gin.Context) {
	scopes, err := h.queries.ListScopes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list scopes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scopes": scopes,
		"count":  len(scopes),
	})
}
