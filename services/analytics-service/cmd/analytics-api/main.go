// Package main provides the entry point for the analytics service API server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andev0x/analytics-service/internal/analytics"
	"github.com/andev0x/analytics-service/internal/api"
	"github.com/andev0x/analytics-service/internal/infrastructure/cache"
	"github.com/andev0x/analytics-service/internal/infrastructure/messaging"
	"github.com/andev0x/analytics-service/internal/infrastructure/persistence"
	"github.com/andev0x/event-driven-order-system/pkg/config"
	"github.com/andev0x/event-driven-order-system/pkg/database"
	"github.com/andev0x/event-driven-order-system/pkg/events"
	pkgredis "github.com/andev0x/event-driven-order-system/pkg/redis"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting Analytics Service...")

	// Load configuration
	cfg := loadConfig()

	// Initialize database
	log.Println("Connecting to database...")
	dbCfg := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
	}
	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()
	log.Println("Database connected successfully")

	// Initialize Redis
	log.Println("Connecting to Redis...")
	redisCfg := pkgredis.Config{
		Host: cfg.RedisHost,
		Port: cfg.RedisPort,
	}
	redisClient, err := pkgredis.Connect(redisCfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}()
	log.Println("Redis connected successfully")

	// Create infrastructure implementations
	analyticsRepo := persistence.NewMySQLRepository(db)
	analyticsCache := cache.NewRedisCache(redisClient)

	// Create domain service
	analyticsService := analytics.NewService(analyticsRepo, analyticsCache)

	// Create API handler
	analyticsHandler := api.NewHandler(analyticsService)

	// Initialize RabbitMQ consumer
	log.Println("Connecting to RabbitMQ...")
	consumer, err := messaging.NewRabbitMQConsumer(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing consumer: %v", err)
		}
	}()
	log.Println("RabbitMQ connected successfully")

	// Setup health checker
	healthChecker := &api.HealthChecker{
		DBHealthFunc: func() error {
			return database.HealthCheck(db)
		},
		CacheHealthFunc: func() error {
			return pkgredis.HealthCheck(redisClient)
		},
		MQHealthFunc: func() error {
			return consumer.HealthCheck()
		},
	}
	analyticsHandler.SetHealthChecker(healthChecker)

	// Setup router
	router := mux.NewRouter()

	// Health and metrics endpoints
	router.HandleFunc("/health", analyticsHandler.HealthCheck).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Analytics endpoints
	router.HandleFunc("/analytics/summary", analyticsHandler.GetSummary).Methods("GET")

	// Setup server
	srv := &http.Server{
		Addr:         ":" + cfg.ServicePort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start consuming events
	ctx, cancel := context.WithCancel(context.Background())

	err = consumer.StartConsuming(ctx, func(event *events.OrderCreated) error {
		return analyticsService.ProcessOrderEvent(context.Background(), event)
	})
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Analytics Service listening on port %s", cfg.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	cancel() // Stop the consumer

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// Config holds application configuration.
type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	RedisHost   string
	RedisPort   string
	RabbitMQURL string
	ServicePort string
}

// loadConfig loads configuration from environment variables.
func loadConfig() Config {
	return Config{
		DBHost:      config.GetEnv("DB_HOST", "localhost"),
		DBPort:      config.GetEnv("DB_PORT", "3306"),
		DBUser:      config.GetEnv("DB_USER", "analyticsuser"),
		DBPassword:  config.GetEnv("DB_PASSWORD", "analyticspass"),
		DBName:      config.GetEnv("DB_NAME", "analytics_db"),
		RedisHost:   config.GetEnv("REDIS_HOST", "localhost"),
		RedisPort:   config.GetEnv("REDIS_PORT", "6379"),
		RabbitMQURL: config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		ServicePort: config.GetEnv("SERVICE_PORT", "8081"),
	}
}
