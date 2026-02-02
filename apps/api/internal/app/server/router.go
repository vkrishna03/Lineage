package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/lineage/api/internal/actor"
	"github.com/lineage/api/internal/artifact"
	"github.com/lineage/api/internal/event"
	"github.com/lineage/api/internal/eventtype"
	"github.com/lineage/api/internal/health"
	"github.com/lineage/api/internal/scope"
	"github.com/lineage/api/internal/score"
)

type Config struct {
	HealthHandler    *health.Handler
	ScopeHandler     *scope.Handler
	ActorHandler     *actor.Handler
	EventTypeHandler *eventtype.Handler
	EventHandler     *event.Handler
	ArtifactHandler  *artifact.Handler
	ScoreHandler     *score.Handler
}

func NewRouter(cfg Config) *gin.Engine {
	r := gin.Default()

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health endpoint
	r.GET("/health", cfg.HealthHandler.Check)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Scopes
		v1.POST("/scopes", cfg.ScopeHandler.Create)
		v1.GET("/scopes/:id", cfg.ScopeHandler.Get)
		v1.GET("/scopes", cfg.ScopeHandler.List)

		// Actors
		v1.POST("/actors", cfg.ActorHandler.Create)
		v1.GET("/actors/:id", cfg.ActorHandler.Get)
		v1.GET("/actors", cfg.ActorHandler.List)

		// Event Types
		v1.POST("/event-types", cfg.EventTypeHandler.Create)
		v1.GET("/event-types/:id", cfg.EventTypeHandler.Get)
		v1.GET("/event-types", cfg.EventTypeHandler.List)

		// Events
		v1.POST("/events", cfg.EventHandler.Create)
		v1.GET("/events/:id", cfg.EventHandler.Get)
		v1.GET("/events/:id/lineage", cfg.EventHandler.GetLineage)
		v1.GET("/events", cfg.EventHandler.ListByScope)

		// Event Scores
		v1.POST("/events/:id/scores", cfg.ScoreHandler.Create)
		v1.GET("/events/:id/scores", cfg.ScoreHandler.GetForEvent)

		// Event Artifacts (linked to events)
		v1.GET("/events/:id/artifacts", cfg.ArtifactHandler.GetForEvent)
		v1.POST("/events/:id/artifacts", cfg.ArtifactHandler.LinkToEvent)

		// Artifacts
		v1.POST("/artifacts", cfg.ArtifactHandler.Create)
		v1.GET("/artifacts/:id", cfg.ArtifactHandler.Get)
		v1.GET("/artifacts", cfg.ArtifactHandler.List)
	}

	return r
}
