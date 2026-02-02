package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lineage/api/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize consumer
	application, err := app.NewConsumerApp(ctx)
	if err != nil {
		slog.Error("failed to initialize consumer", "error", err)
		os.Exit(1)
	}
	defer application.DB.Close()
	defer application.Consumer.Close()

	slog.Info("consumer initialized")
	slog.Info("connected to database")

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down consumer...")
		cancel()
	}()

	// Start consumer
	slog.Info("starting kafka consumer", "topic", application.Config.KafkaTopic)
	if err := application.Consumer.Start(ctx); err != nil && err != context.Canceled {
		slog.Error("consumer error", "error", err)
		os.Exit(1)
	}

	slog.Info("consumer stopped")
}
