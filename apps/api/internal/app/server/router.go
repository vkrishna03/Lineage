package server

import (
	"github.com/gin-gonic/gin"
	"github.com/lineage/api/internal/actor"
	"github.com/lineage/api/internal/event"
	"github.com/lineage/api/internal/eventtype"
	"github.com/lineage/api/internal/health"
	"github.com/lineage/api/internal/scope"
)

type Config struct {
	HealthHandler    *health.Handler
	ScopeHandler     *scope.Handler
	ActorHandler     *actor.Handler
	EventTypeHandler *eventtype.Handler
	EventHandler     *event.Handler
}

func NewRouter(cfg Config) *gin.Engine {
	r := gin.Default()

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
	}

	return r
}
