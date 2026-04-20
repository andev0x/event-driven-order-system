package api

import (
	"log/slog"
	"net/http"

	"github.com/andev0x/analytics-service/internal/analytics"
	"github.com/andev0x/event-driven-order-system/pkg/httputil"
)

// Handler handles HTTP requests for analytics.
type Handler struct {
	service             *analytics.Service
	healthCheck         *HealthChecker
	internalAuthHandler *httputil.InternalAuthHandler
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
		service:             service,
		healthCheck:         nil,
		internalAuthHandler: nil,
	}
}

// SetHealthChecker sets the health checker.
func (h *Handler) SetHealthChecker(hc *HealthChecker) {
	h.healthCheck = hc
}

// SetInternalAuthHandler sets the internal auth handler.
func (h *Handler) SetInternalAuthHandler(authHandler *httputil.InternalAuthHandler) {
	h.internalAuthHandler = authHandler
}

// IssueToken handles POST /internal/auth/token.
// @Summary Issue internal auth token
// @Description Issue a JWT token for service-to-service or local automation access.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Internal-Auth-Key header string true "Internal authentication key"
// @Param request body httputil.TokenRequest false "Optional token request payload"
// @Success 200 {object} httputil.TokenResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /internal/auth/token [post]
func (h *Handler) IssueToken(w http.ResponseWriter, r *http.Request) {
	if h.internalAuthHandler == nil {
		httputil.RespondError(w, http.StatusInternalServerError, "internal auth handler not configured")
		return
	}

	h.internalAuthHandler.IssueToken(w, r)
}

// GetSummary handles GET /analytics/summary
// @Summary Get analytics summary
// @Description Retrieve aggregated order analytics metrics.
// @Tags analytics
// @Produce json
// @Success 200 {object} analytics.Summary
// @Failure 500 {object} httputil.ErrorResponse
// @Security BearerAuth
// @Router /analytics/summary [get]
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetSummary(r.Context())
	if err != nil {
		slog.Error("Failed to get analytics summary", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to get analytics summary")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, summary)
}

// HealthCheck handles GET /health
// @Summary Health check
// @Description Returns service health and dependency checks.
// @Tags health
// @Produce json
// @Success 200 {object} httputil.HealthResponse
// @Failure 503 {object} httputil.HealthResponse
// @Router /health [get]
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
