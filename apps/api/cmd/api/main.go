package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lineage/api/docs" // Swagger docs
	"github.com/lineage/api/internal/app"
)

// @title           Lineage API
// @version         1.0
// @description     Epistemic Transparency & Event Lineage System API
// @description     An append-only event store with hash-chaining for AI decision tracking

// @contact.name   Lineage Support
// @contact.url    https://github.com/vkrishna03/Lineage

// @license.name  Elastic License 2.0
// @license.url   https://www.elastic.co/licensing/elastic-license

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize application
	application, err := app.NewApp(ctx)
	if err != nil {
		slog.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}
	defer application.DB.Close()
	defer application.Producer.Close()

	slog.Info("application initialized")
	slog.Info("connected to database")
	slog.Info("kafka producer initialized")

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down...")
		cancel()
	}()

	// Start server
	slog.Info("starting api server", "port", application.Config.Port)
	if err := application.Router.Run(":" + application.Config.Port); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
