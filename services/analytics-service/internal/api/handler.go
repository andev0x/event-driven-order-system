package api

import (
	"log"
	"net/http"

	"github.com/andev0x/analytics-service/internal/analytics"
	"github.com/andev0x/event-driven-order-system/pkg/httputil"
)

// Handler handles HTTP requests for analytics.
type Handler struct {
	service     *analytics.Service
	healthCheck *HealthChecker
}

// HealthChecker provides health check functionality.
type HealthChecker struct {
	DBHealthFunc    func() error
	CacheHealthFunc func() error
	MQHealthFunc    func() error
}

// NewHandler creates a new analytics handler.
func NewHandler(service *analytics.Service) *Handler {
	return &Handler{
		service:     service,
		healthCheck: nil,
	}
}

// SetHealthChecker sets the health checker.
func (h *Handler) SetHealthChecker(hc *HealthChecker) {
	h.healthCheck = hc
}

// GetSummary handles GET /analytics/summary
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetSummary(r.Context())
	if err != nil {
		log.Printf("Error getting summary: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to get analytics summary")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, summary)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	response := httputil.NewHealthResponse("analytics-service")

	if h.healthCheck != nil {
		if h.healthCheck.DBHealthFunc != nil {
			err := h.healthCheck.DBHealthFunc()
			response.SetCheck("database", err == nil, errMsg(err))
		}

		if h.healthCheck.CacheHealthFunc != nil {
			err := h.healthCheck.CacheHealthFunc()
			response.SetCheck("cache", err == nil, errMsg(err))
		}

		if h.healthCheck.MQHealthFunc != nil {
			err := h.healthCheck.MQHealthFunc()
			response.SetCheck("mq", err == nil, errMsg(err))
		}
	}

	if response.IsHealthy() {
		httputil.RespondJSON(w, http.StatusOK, response)
	} else {
		httputil.RespondJSON(w, http.StatusServiceUnavailable, response)
	}
}

func errMsg(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
