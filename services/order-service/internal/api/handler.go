package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/andev0x/event-driven-order-system/pkg/httputil"
	"github.com/andev0x/order-service/internal/order"
	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for orders.
type Handler struct {
	service             *order.Service
	healthCheck         *HealthChecker
	internalAuthHandler *httputil.InternalAuthHandler
}

// HealthChecker provides health check functionality.
type HealthChecker struct {
	DBHealthFunc    func() error
	CacheHealthFunc func() error
	MQHealthFunc    func() error
}

// NewHandler creates a new order handler.
func NewHandler(service *order.Service) *Handler {
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

// CreateOrder handles POST /orders
// @Summary Create order
// @Description Create a new order and publish an OrderCreated event.
// @Tags orders
// @Accept json
// @Produce json
// @Param request body order.CreateRequest true "Order payload"
// @Success 201 {object} order.Order
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security BearerAuth
// @Router /orders [post]
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req order.CreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		httputil.RespondError(w, http.StatusBadRequest, "request body must contain only one JSON object")
		return
	}

	o, err := h.service.Create(r.Context(), &req)
	if err != nil {
		if order.IsValidationError(err) {
			httputil.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		slog.Error("Failed to create order", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, o)
}

// GetOrder handles GET /orders/{id}
// @Summary Get order by ID
// @Description Retrieve a single order by its unique identifier.
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} order.Order
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Security BearerAuth
// @Router /orders/{id} [get]
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	o, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get order", "order_id", id, "error", err)
		httputil.RespondError(w, http.StatusNotFound, "Order not found")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, o)
}

// ListOrders handles GET /orders
// @Summary List orders
// @Description List orders using pagination.
// @Tags orders
// @Produce json
// @Param limit query int false "Max number of items to return" default(10)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {array} order.Order
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Security BearerAuth
// @Router /orders [get]
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "limit must be a valid integer")
			return
		}

		if l <= 0 {
			httputil.RespondError(w, http.StatusBadRequest, "limit must be greater than 0")
			return
		}

		limit = l
	}

	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "offset must be a valid integer")
			return
		}

		if o < 0 {
			httputil.RespondError(w, http.StatusBadRequest, "offset must be greater than or equal to 0")
			return
		}

		offset = o
	}

	orders, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		slog.Error("Failed to list orders", "limit", limit, "offset", offset, "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to list orders")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, orders)
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
	response := httputil.NewHealthResponse("order-service")

	// Check dependencies if health checker is configured
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
