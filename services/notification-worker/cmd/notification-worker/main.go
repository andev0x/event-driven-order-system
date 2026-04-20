// Package main implements the notification worker service that consumes order events from RabbitMQ
// and sends notifications to customers.
package main

import (
	"context"
	"log/slog"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "notification-worker")
	slog.SetDefault(logger)

	slog.Info("Starting notification worker")

	// Get configuration from environment
	rabbitMQURL := config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	healthPort := config.GetEnv("HEALTH_PORT", "8082")

	// Connect to RabbitMQ
	consumer, err := messaging.NewRabbitMQConsumer(rabbitMQURL, 10)
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			slog.Error("Failed to close RabbitMQ consumer", "error", err)
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
		slog.Info("Health check server listening", "port", healthPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Health check server error", "error", err)
		}
	}()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("Shutting down notification worker")
		cancel()
	}()

	// Start consuming events
	err = consumer.StartConsuming(ctx, func(ctx context.Context, event *events.OrderCreated) error {
		return notificationService.ProcessOrderCreated(ctx, event)
	})
	if err != nil {
		slog.Error("Failed to start consuming events", "error", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	<-ctx.Done()
	slog.Info("Notification worker stopped")
}
