package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// HealthChecker is an interface for components that can report their health.
type HealthChecker interface {
	HealthCheck() error
}

// Handler handles HTTP requests for the notification worker.
type Handler struct {
	mqChecker HealthChecker
}

// NewHandler creates a new Handler with the given health checker.
func NewHandler(mqChecker HealthChecker) *Handler {
	return &Handler{
		mqChecker: mqChecker,
	}
}

// HealthResponse represents the health check response structure.
type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks"`
}

// Health handles the health check endpoint.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "healthy",
		Service: "notification-worker",
		Checks:  make(map[string]string),
	}

	overallHealthy := true

	// Check RabbitMQ connection
	if h.mqChecker != nil {
		if err := h.mqChecker.HealthCheck(); err != nil {
			response.Checks["mq"] = "unhealthy: " + err.Error()
			overallHealthy = false
		} else {
			response.Checks["mq"] = "healthy"
		}
	} else {
		response.Checks["mq"] = "unhealthy: not configured"
		overallHealthy = false
	}

	if !overallHealthy {
		response.Status = "degraded"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("Failed to encode degraded health response", "error", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode health response", "error", err)
	}
}

// RegisterRoutes registers the handler routes to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
}
