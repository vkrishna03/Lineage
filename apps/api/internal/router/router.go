package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/handler"
	"github.com/lineage/api/internal/kafka"
)

type Config struct {
	DB           *pgxpool.Pool
	Producer     *kafka.Producer
	KafkaBrokers []string
}

func New(cfg Config) *gin.Engine {
	r := gin.Default()

	// Health handler
	healthHandler := handler.NewHealthHandler(cfg.DB, cfg.KafkaBrokers)
	r.GET("/health", healthHandler.Health)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Scopes
		scopesHandler := handler.NewScopesHandler(cfg.DB)
		v1.POST("/scopes", scopesHandler.CreateScope)
		v1.GET("/scopes/:id", scopesHandler.GetScope)
		v1.GET("/scopes", scopesHandler.ListScopes)

		// Actors
		actorsHandler := handler.NewActorsHandler(cfg.DB)
		v1.POST("/actors", actorsHandler.CreateActor)
		v1.GET("/actors/:id", actorsHandler.GetActor)
		v1.GET("/actors", actorsHandler.ListActors)

		// Event Types
		eventTypesHandler := handler.NewEventTypesHandler(cfg.DB)
		v1.POST("/event-types", eventTypesHandler.CreateEventType)
		v1.GET("/event-types/:id", eventTypesHandler.GetEventType)
		v1.GET("/event-types", eventTypesHandler.ListEventTypes)

		// Events
		eventsHandler := handler.NewEventsHandler(cfg.DB, cfg.Producer)
		v1.POST("/events", eventsHandler.CreateEvent)
		v1.GET("/events/:id", eventsHandler.GetEvent)
		v1.GET("/events/:id/lineage", eventsHandler.GetEventLineage)
		v1.GET("/events", eventsHandler.ListEventsByScope)
	}

	return r
}
