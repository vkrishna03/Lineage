package artifact

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lineage/api/internal/app/db/sqlc"
)

// Handler handles HTTP requests for artifacts
type Handler struct {
	repo *Repository
}

// NewHandler creates a new artifact handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateRequest represents the request body for creating an artifact
type CreateRequest struct {
	ScopeID     uuid.UUID       `json:"scope_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	ContentHash string          `json:"content_hash" binding:"required" example:"sha256:abc123..."`
	ContentType string          `json:"content_type" binding:"required" example:"application/json"`
	URI         *string         `json:"uri" example:"s3://bucket/path/file.json"`
	Metadata    json.RawMessage `json:"metadata" swaggertype:"object"`
}

// ArtifactResponse represents the API response for an artifact
// @Description Artifact data returned by the API
type ArtifactResponse struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ScopeID     string  `json:"scope_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ContentHash string  `json:"content_hash" example:"sha256:abc123..."`
	ContentType string  `json:"content_type" example:"application/json"`
	URI         *string `json:"uri" example:"s3://bucket/path/file.json"`
	Metadata    any     `json:"metadata" swaggertype:"object"`
	CreatedAt   string  `json:"created_at" example:"2024-01-15T10:00:00Z"`
}

// Create handles POST /artifacts
// @Summary      Create an artifact
// @Description  Create a new artifact with content hash for deduplication
// @Tags         artifacts
// @Accept       json
// @Produce      json
// @Param        request body CreateRequest true "Artifact creation request"
// @Success      201 {object} ArtifactResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /artifacts [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := sqlc.CreateArtifactParams{
		ScopeID:     req.ScopeID,
		ContentHash: req.ContentHash,
		ContentType: req.ContentType,
		Uri:         req.URI,
		Metadata:    req.Metadata,
	}

	artifact, err := h.repo.Create(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create artifact"})
		return
	}

	c.JSON(http.StatusCreated, artifact)
}

// Get handles GET /artifacts/:id
// @Summary      Get an artifact
// @Description  Get an artifact by ID
// @Tags         artifacts
// @Produce      json
// @Param        id path string true "Artifact ID" format(uuid)
// @Success      200 {object} ArtifactResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /artifacts/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid artifact ID"})
		return
	}

	artifact, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "artifact not found"})
		return
	}

	c.JSON(http.StatusOK, artifact)
}

// ListResponse represents the response for listing artifacts
type ListResponse struct {
	Artifacts []ArtifactResponse `json:"artifacts"`
	Count     int                `json:"count"`
}

// List handles GET /artifacts
// @Summary      List or find artifacts
// @Description  List artifacts by scope, or find a specific artifact by scope and content hash (for deduplication)
// @Tags         artifacts
// @Produce      json
// @Param        scope_id query string true "Scope ID" format(uuid)
// @Param        content_hash query string false "Content hash (if provided, returns single artifact)"
// @Success      200 {object} ListResponse "When listing artifacts"
// @Success      200 {object} ArtifactResponse "When content_hash is provided"
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /artifacts [get]
func (h *Handler) List(c *gin.Context) {
	scopeIDStr := c.Query("scope_id")
	contentHash := c.Query("content_hash")

	if scopeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scope_id is required"})
		return
	}

	scopeID, err := uuid.Parse(scopeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope_id"})
		return
	}

	// If content_hash provided, return single artifact (dedup check)
	if contentHash != "" {
		artifact, err := h.repo.GetByHash(c.Request.Context(), scopeID, contentHash)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "artifact not found"})
			return
		}
		c.JSON(http.StatusOK, artifact)
		return
	}

	// Otherwise list all artifacts in scope
	artifacts, err := h.repo.ListByScope(c.Request.Context(), sqlc.ListArtifactsInScopeParams{
		ScopeID: scopeID,
		Limit:   100,
		Offset:  0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list artifacts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"artifacts": artifacts,
		"count":     len(artifacts),
	})
}

// LinkRequest represents the request to link an artifact to an event
type LinkRequest struct {
	ArtifactID uuid.UUID `json:"artifact_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role       string    `json:"role" binding:"required" example:"input" enums:"input,output"`
}

// GetForEvent handles GET /events/:id/artifacts
// @Summary      Get artifacts for event
// @Description  Get all artifacts linked to an event
// @Tags         artifacts
// @Produce      json
// @Param        id path string true "Event ID" format(uuid)
// @Success      200 {object} ListResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /events/{id}/artifacts [get]
func (h *Handler) GetForEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	artifacts, err := h.repo.GetForEvent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get artifacts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"artifacts": artifacts,
		"count":     len(artifacts),
	})
}

// LinkToEvent handles POST /events/:id/artifacts
// @Summary      Link artifact to event
// @Description  Link an existing artifact to an event as input or output
// @Tags         artifacts
// @Accept       json
// @Produce      json
// @Param        id path string true "Event ID" format(uuid)
// @Param        request body LinkRequest true "Link request"
// @Success      201 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /events/{id}/artifacts [post]
func (h *Handler) LinkToEvent(c *gin.Context) {
	idStr := c.Param("id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	var req LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role != "input" && req.Role != "output" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role (valid: input, output)"})
		return
	}

	_, err = h.repo.LinkToEvent(c.Request.Context(), sqlc.LinkArtifactParams{
		EventID:    eventID,
		ArtifactID: req.ArtifactID,
		Role:       sqlc.ArtifactRole(req.Role),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link artifact"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "linked",
		"message": "Artifact linked to event",
	})
}
