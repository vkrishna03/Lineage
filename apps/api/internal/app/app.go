package app

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lineage/api/internal/actor"
	"github.com/lineage/api/internal/app/config"
	"github.com/lineage/api/internal/app/db/sqlc"
	"github.com/lineage/api/internal/app/kafka"
	"github.com/lineage/api/internal/app/server"
	"github.com/lineage/api/internal/event"
	"github.com/lineage/api/internal/eventtype"
	"github.com/lineage/api/internal/health"
	"github.com/lineage/api/internal/lineage"
	"github.com/lineage/api/internal/scope"
)

// App holds all application dependencies for the API server
type App struct {
	Config   *config.Config
	DB       *pgxpool.Pool
	Producer *kafka.Producer
	Router   *gin.Engine
}

// ConsumerApp holds all dependencies for the Kafka consumer
type ConsumerApp struct {
	Config   *config.Config
	DB       *pgxpool.Pool
	Consumer *kafka.Consumer
}

// NewApp creates and wires up all dependencies for the API server
func NewApp(ctx context.Context) (*App, error) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Connect to database
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create SQLC queries
	queries := sqlc.New(db)

	// Create Kafka producer
	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:      cfg.KafkaBrokers,
		Topic:        cfg.KafkaTopic,
		SASLEnabled:  cfg.KafkaSASLEnabled,
		SASLUsername: cfg.KafkaSASLUsername,
		SASLPassword: cfg.KafkaSASLPassword,
		CAPath:       cfg.KafkaCAPath,
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Create repositories
	scopeRepo := scope.NewRepository(queries)
	actorRepo := actor.NewRepository(queries)
	eventTypeRepo := eventtype.NewRepository(queries)
	eventRepo := event.NewRepository(queries)
	lineageRepo := lineage.NewRepository(queries)

	// Create handlers
	healthHandler := health.NewHandler(db, cfg)
	scopeHandler := scope.NewHandler(scopeRepo)
	actorHandler := actor.NewHandler(actorRepo)
	eventTypeHandler := eventtype.NewHandler(eventTypeRepo)
	eventHandler := event.NewHandler(eventRepo, lineageRepo, eventTypeRepo, producer)

	// Create router
	router := server.NewRouter(server.Config{
		HealthHandler:    healthHandler,
		ScopeHandler:     scopeHandler,
		ActorHandler:     actorHandler,
		EventTypeHandler: eventTypeHandler,
		EventHandler:     eventHandler,
	})

	return &App{
		Config:   cfg,
		DB:       db,
		Producer: producer,
		Router:   router,
	}, nil
}

// NewConsumerApp creates and wires up all dependencies for the Kafka consumer
func NewConsumerApp(ctx context.Context) (*ConsumerApp, error) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to database
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:      cfg.KafkaBrokers,
		Topic:        cfg.KafkaTopic,
		GroupID:      cfg.KafkaGroupID,
		SASLEnabled:  cfg.KafkaSASLEnabled,
		SASLUsername: cfg.KafkaSASLUsername,
		SASLPassword: cfg.KafkaSASLPassword,
		CAPath:       cfg.KafkaCAPath,
	}, db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	return &ConsumerApp{
		Config:   cfg,
		DB:       db,
		Consumer: consumer,
	}, nil
}
