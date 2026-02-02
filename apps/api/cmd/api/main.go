package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lineage/api/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize application
	application, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.DB.Close()
	defer application.Producer.Close()

	log.Println("Application initialized")
	log.Println("Connected to database")
	log.Println("Kafka producer initialized")

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// Start server
	log.Printf("Starting API server on port %s", application.Config.Port)
	if err := application.Router.Run(":" + application.Config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
