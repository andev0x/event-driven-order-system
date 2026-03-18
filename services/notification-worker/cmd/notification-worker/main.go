// Package main implements the notification worker service that consumes order events from RabbitMQ
// and sends notifications to customers.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andev0x/event-driven-order-system/pkg/config"
	"github.com/andev0x/event-driven-order-system/pkg/events"
	"github.com/andev0x/notification-worker/internal/api"
	"github.com/andev0x/notification-worker/internal/infrastructure/messaging"
	"github.com/andev0x/notification-worker/internal/notification"
)

func main() {
	log.Println("Starting Notification Worker...")

	// Get configuration from environment
	rabbitMQURL := config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	healthPort := config.GetEnv("HEALTH_PORT", "8082")

	// Connect to RabbitMQ
	consumer, err := messaging.NewRabbitMQConsumer(rabbitMQURL, 10)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing RabbitMQ consumer: %v", err)
		}
	}()

	// Create notification service with console sender
	sender := notification.NewConsoleSender()
	notificationService := notification.NewService(sender)

	// Setup HTTP server for health checks
	handler := api.NewHandler(consumer)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + healthPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Health check server listening on port %s", healthPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down notification worker...")
		cancel()
	}()

	// Start consuming events
	err = consumer.StartConsuming(ctx, func(ctx context.Context, event *events.OrderCreated) error {
		return notificationService.ProcessOrderCreated(ctx, event)
	})
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Notification worker stopped")
}
