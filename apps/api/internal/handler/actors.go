package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/db/sqlc"
)

type ActorsHandler struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewActorsHandler(db *pgxpool.Pool) *ActorsHandler {
	return &ActorsHandler{
		db:      db,
		queries: sqlc.New(db),
	}
}

type CreateActorRequest struct {
	Type       string          `json:"type" binding:"required"`
	ExternalID *string         `json:"external_id"`
	Name       *string         `json:"name"`
	Metadata   json.RawMessage `json:"metadata"`
}

var validActorTypes = map[string]bool{
	"human":   true,
	"llm":     true,
	"agent":   true,
	"service": true,
	"tool":    true,
}

func (h *ActorsHandler) CreateActor(c *gin.Context) {
	var req CreateActorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validActorTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor type"})
		return
	}

	params := sqlc.CreateActorParams{
		Type:       sqlc.ActorType(req.Type),
		ExternalID: req.ExternalID,
		Name:       req.Name,
		Metadata:   req.Metadata,
	}

	actor, err := h.queries.CreateActor(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create actor"})
		return
	}

	c.JSON(http.StatusCreated, actor)
}

func (h *ActorsHandler) GetActor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor ID"})
		return
	}

	actor, err := h.queries.GetActor(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "actor not found"})
		return
	}

	c.JSON(http.StatusOK, actor)
}

func (h *ActorsHandler) ListActors(c *gin.Context) {
	actors, err := h.queries.ListActors(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list actors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"actors": actors,
		"count":  len(actors),
	})
}
