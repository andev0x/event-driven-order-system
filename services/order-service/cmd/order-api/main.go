// Package main provides the entry point for the order service API.
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
	"github.com/andev0x/event-driven-order-system/pkg/database"
	"github.com/andev0x/event-driven-order-system/pkg/httputil"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "order-service")
	slog.SetDefault(logger)

	slog.Info("Starting order service")

	// Load configuration
	cfg := loadConfig()
	if cfg.JWTSecret == "" {
		slog.Error("Missing required configuration", "key", "JWT_SECRET")
		os.Exit(1)
	}
	if cfg.InternalAuthKey == "" {
		slog.Error("Missing required configuration", "key", "INTERNAL_AUTH_KEY")
		os.Exit(1)
	}

	// Initialize database
	slog.Info("Connecting to database")
	dbCfg := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
	}
	db, err := database.Connect(dbCfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()
	slog.Info("Database connected successfully")

	// Initialize Redis
	slog.Info("Connecting to Redis")
	redisCfg := pkgredis.Config{
		Host: cfg.RedisHost,
		Port: cfg.RedisPort,
	}
	redisClient, err := pkgredis.Connect(redisCfg)
	if err != nil {
		slog.Error("Failed to initialize Redis", "error", err)
		return
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("Failed to close Redis connection", "error", err)
		}
	}()
	slog.Info("Redis connected successfully")

	// Initialize RabbitMQ publisher
	slog.Info("Connecting to RabbitMQ")
	publisher, err := messaging.NewRabbitMQPublisher(cfg.RabbitMQURL)
	if err != nil {
		slog.Error("Failed to initialize RabbitMQ publisher", "error", err)
		return
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			slog.Error("Failed to close RabbitMQ publisher", "error", err)
		}
	}()
	slog.Info("RabbitMQ connected successfully")

	// Create infrastructure implementations
	orderRepo := persistence.NewMySQLRepository(db)
	orderCache := cache.NewRedisCache(redisClient)

	// Create domain service
	orderService := order.NewService(orderRepo, orderCache, publisher)

	// Create API handler
	orderHandler := api.NewHandler(orderService)
	authHandler := httputil.NewInternalAuthHandler(
		cfg.JWTSecret,
		cfg.InternalAuthKey,
		cfg.InternalAuthIssuer,
		cfg.InternalAuthTokenTTL,
	)

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
	jwtMiddleware := httputil.JWTMiddleware(cfg.JWTSecret)
	protectedRouter := router.NewRoute().Subrouter()
	protectedRouter.Use(jwtMiddleware)

	// Health and metrics endpoints
	router.HandleFunc("/health", orderHandler.HealthCheck).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	router.HandleFunc("/internal/auth/token", authHandler.IssueToken).Methods("POST")

	// Order endpoints
	protectedRouter.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	protectedRouter.HandleFunc("/orders/{id}", orderHandler.GetOrder).Methods("GET")
	protectedRouter.HandleFunc("/orders", orderHandler.ListOrders).Methods("GET")

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
		slog.Info("Order service listening", "port", cfg.ServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited gracefully")
}

// Config holds application configuration.
type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	RedisHost            string
	RedisPort            string
	RabbitMQURL          string
	JWTSecret            string
	InternalAuthKey      string
	InternalAuthIssuer   string
	InternalAuthTokenTTL time.Duration
	ServicePort          string
}

// loadConfig loads configuration from environment variables.
func loadConfig() Config {
	return Config{
		DBHost:               config.GetEnv("DB_HOST", "localhost"),
		DBPort:               config.GetEnv("DB_PORT", "3306"),
		DBUser:               config.GetEnv("DB_USER", "orderuser"),
		DBPassword:           config.GetEnv("DB_PASSWORD", "orderpass"),
		DBName:               config.GetEnv("DB_NAME", "order_db"),
		RedisHost:            config.GetEnv("REDIS_HOST", "localhost"),
		RedisPort:            config.GetEnv("REDIS_PORT", "6379"),
		RabbitMQURL:          config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		JWTSecret:            config.GetEnv("JWT_SECRET", ""),
		InternalAuthKey:      config.GetEnv("INTERNAL_AUTH_KEY", ""),
		InternalAuthIssuer:   config.GetEnv("INTERNAL_AUTH_ISSUER", "order-service"),
		InternalAuthTokenTTL: config.GetEnvDuration("INTERNAL_AUTH_TOKEN_TTL", time.Hour),
		ServicePort:          config.GetEnv("SERVICE_PORT", "8080"),
	}
}
