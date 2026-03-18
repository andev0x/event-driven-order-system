// Package main provides the entry point for the order service API.
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
	"github.com/andev0x/event-driven-order-system/pkg/database"
	pkgredis "github.com/andev0x/event-driven-order-system/pkg/redis"
	"github.com/andev0x/order-service/internal/api"
	"github.com/andev0x/order-service/internal/infrastructure/cache"
	"github.com/andev0x/order-service/internal/infrastructure/messaging"
	"github.com/andev0x/order-service/internal/infrastructure/persistence"
	"github.com/andev0x/order-service/internal/order"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting Order Service...")

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
			log.Printf("Error closing database connection: %v", err)
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
		log.Printf("Failed to initialize Redis: %v", err)
		return
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()
	log.Println("Redis connected successfully")

	// Initialize RabbitMQ publisher
	log.Println("Connecting to RabbitMQ...")
	publisher, err := messaging.NewRabbitMQPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("Failed to initialize RabbitMQ publisher: %v", err)
		return
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Error closing RabbitMQ publisher: %v", err)
		}
	}()
	log.Println("RabbitMQ connected successfully")

	// Create infrastructure implementations
	orderRepo := persistence.NewMySQLRepository(db)
	orderCache := cache.NewRedisCache(redisClient)

	// Create domain service
	orderService := order.NewService(orderRepo, orderCache, publisher)

	// Create API handler
	orderHandler := api.NewHandler(orderService)

	// Setup health checker
	healthChecker := &api.HealthChecker{
		DBHealthFunc: func() error {
			return database.HealthCheck(db)
		},
		CacheHealthFunc: func() error {
			return pkgredis.HealthCheck(redisClient)
		},
		MQHealthFunc: func() error {
			return publisher.HealthCheck()
		},
	}
	orderHandler.SetHealthChecker(healthChecker)

	// Setup router
	router := mux.NewRouter()

	// Health and metrics endpoints
	router.HandleFunc("/health", orderHandler.HealthCheck).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Order endpoints
	router.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	router.HandleFunc("/orders/{id}", orderHandler.GetOrder).Methods("GET")
	router.HandleFunc("/orders", orderHandler.ListOrders).Methods("GET")

	// Setup server
	srv := &http.Server{
		Addr:         ":" + cfg.ServicePort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Order Service listening on port %s", cfg.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
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
		DBUser:      config.GetEnv("DB_USER", "orderuser"),
		DBPassword:  config.GetEnv("DB_PASSWORD", "orderpass"),
		DBName:      config.GetEnv("DB_NAME", "order_db"),
		RedisHost:   config.GetEnv("REDIS_HOST", "localhost"),
		RedisPort:   config.GetEnv("REDIS_PORT", "6379"),
		RabbitMQURL: config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		ServicePort: config.GetEnv("SERVICE_PORT", "8080"),
	}
}
