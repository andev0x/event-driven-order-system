// Package main provides the entry point for the analytics service API server.
// @title Analytics Service API
// @version 1.0
// @description REST API for analytics summary and service observability.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andev0x/analytics-service/internal/analytics"
	"github.com/andev0x/analytics-service/internal/api"
	_ "github.com/andev0x/analytics-service/internal/api/docs"
	"github.com/andev0x/analytics-service/internal/infrastructure/cache"
	"github.com/andev0x/analytics-service/internal/infrastructure/messaging"
	"github.com/andev0x/analytics-service/internal/infrastructure/persistence"
	"github.com/andev0x/event-driven-order-system/pkg/config"
	"github.com/andev0x/event-driven-order-system/pkg/database"
	"github.com/andev0x/event-driven-order-system/pkg/events"
	"github.com/andev0x/event-driven-order-system/pkg/httputil"
	"github.com/andev0x/event-driven-order-system/pkg/observability"
	pkgredis "github.com/andev0x/event-driven-order-system/pkg/redis"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "analytics-service")
	slog.SetDefault(logger)

	slog.Info("Starting analytics service")

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

	shutdownTracing, err := observability.InitTracing(
		context.Background(),
		"analytics-service",
		cfg.OTLPEndpoint,
		cfg.OTLPInsecure,
	)
	if err != nil {
		slog.Error("Failed to initialize OpenTelemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if shutdownErr := shutdownTracing(shutdownCtx); shutdownErr != nil {
			slog.Error("Failed to shutdown OpenTelemetry", "error", shutdownErr)
		}
	}()

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
			slog.Error("Failed to close database", "error", err)
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
		os.Exit(1)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("Failed to close Redis", "error", err)
		}
	}()
	slog.Info("Redis connected successfully")

	// Create infrastructure implementations
	analyticsRepo := persistence.NewMySQLRepository(db)
	analyticsCache := cache.NewRedisCache(redisClient)

	// Create domain service
	analyticsService := analytics.NewService(analyticsRepo, analyticsCache)

	// Create API handler
	analyticsHandler := api.NewHandler(analyticsService)
	authHandler := httputil.NewInternalAuthHandler(
		cfg.JWTSecret,
		cfg.InternalAuthKey,
		cfg.InternalAuthIssuer,
		cfg.InternalAuthTokenTTL,
	)
	analyticsHandler.SetInternalAuthHandler(authHandler)

	// Initialize RabbitMQ consumer
	slog.Info("Connecting to RabbitMQ")
	consumer, err := messaging.NewRabbitMQConsumer(cfg.RabbitMQURL)
	if err != nil {
		slog.Error("Failed to initialize RabbitMQ consumer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			slog.Error("Failed to close RabbitMQ consumer", "error", err)
		}
	}()
	slog.Info("RabbitMQ connected successfully")

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
	router.Use(otelmux.Middleware("analytics-service"))
	jwtMiddleware := httputil.JWTMiddleware(cfg.JWTSecret)
	protectedRouter := router.NewRoute().Subrouter()
	protectedRouter.Use(jwtMiddleware)

	// Health and metrics endpoints
	router.HandleFunc("/health", analyticsHandler.HealthCheck).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	router.HandleFunc("/internal/auth/token", analyticsHandler.IssueToken).Methods("POST")
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Analytics endpoints
	protectedRouter.HandleFunc("/analytics/summary", analyticsHandler.GetSummary).Methods("GET")

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

	err = consumer.StartConsuming(ctx, func(msgCtx context.Context, event *events.OrderCreated) error {
		return analyticsService.ProcessOrderEvent(msgCtx, event)
	})
	if err != nil {
		slog.Error("Failed to start consuming events", "error", err)
		os.Exit(1)
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Analytics service listening", "port", cfg.ServicePort)
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
	cancel() // Stop the consumer

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
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
	OTLPEndpoint         string
	OTLPInsecure         bool
	ServicePort          string
}

// loadConfig loads configuration from environment variables.
func loadConfig() Config {
	return Config{
		DBHost:               config.GetEnv("DB_HOST", "localhost"),
		DBPort:               config.GetEnv("DB_PORT", "3306"),
		DBUser:               config.GetEnv("DB_USER", "analyticsuser"),
		DBPassword:           config.GetEnv("DB_PASSWORD", "analyticspass"),
		DBName:               config.GetEnv("DB_NAME", "analytics_db"),
		RedisHost:            config.GetEnv("REDIS_HOST", "localhost"),
		RedisPort:            config.GetEnv("REDIS_PORT", "6379"),
		RabbitMQURL:          config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		JWTSecret:            config.GetEnv("JWT_SECRET", ""),
		InternalAuthKey:      config.GetEnv("INTERNAL_AUTH_KEY", ""),
		InternalAuthIssuer:   config.GetEnv("INTERNAL_AUTH_ISSUER", "analytics-service"),
		InternalAuthTokenTTL: config.GetEnvDuration("INTERNAL_AUTH_TOKEN_TTL", time.Hour),
		OTLPEndpoint:         config.GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		OTLPInsecure:         config.GetEnvBool("OTEL_EXPORTER_OTLP_INSECURE", true),
		ServicePort:          config.GetEnv("SERVICE_PORT", "8081"),
	}
}
