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

	// Initialize consumer
	application, err := app.NewConsumerApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize consumer: %v", err)
	}
	defer application.DB.Close()
	defer application.Consumer.Close()

	log.Println("Consumer initialized")
	log.Println("Connected to database")

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down consumer...")
		cancel()
	}()

	// Start consumer
	log.Printf("Starting Kafka consumer for topic: %s", application.Config.KafkaTopic)
	if err := application.Consumer.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Consumer error: %v", err)
	}

	log.Println("Consumer stopped")
}
