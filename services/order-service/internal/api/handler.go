package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/andev0x/event-driven-order-system/pkg/httputil"
	"github.com/andev0x/order-service/internal/order"
	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for orders.
type Handler struct {
	service     *order.Service
	healthCheck *HealthChecker
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
		service:     service,
		healthCheck: nil,
	}
}

// SetHealthChecker sets the health checker.
func (h *Handler) SetHealthChecker(hc *HealthChecker) {
	h.healthCheck = hc
}

// CreateOrder handles POST /orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req order.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	o, err := h.service.Create(r.Context(), &req)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, o)
}

// GetOrder handles GET /orders/{id}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	o, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting order: %v", err)
		httputil.RespondError(w, http.StatusNotFound, "Order not found")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, o)
}

// ListOrders handles GET /orders
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	orders, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		log.Printf("Error listing orders: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to list orders")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, orders)
}

// HealthCheck handles GET /health
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
