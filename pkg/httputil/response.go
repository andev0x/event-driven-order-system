package httputil

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents a standard error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondJSON writes a JSON response with the given status code.
func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, writeErr := w.Write([]byte(`{"error":"Internal server error"}`)); writeErr != nil {
			log.Printf("Error writing error response: %v", writeErr)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(response); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// RespondError writes a JSON error response with the given status code.
func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, ErrorResponse{Error: message})
}

// HealthResponse represents a standard health check response.
type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks,omitempty"`
}

// NewHealthResponse creates a new health response for a service.
func NewHealthResponse(serviceName string) *HealthResponse {
	return &HealthResponse{
		Status:  "healthy",
		Service: serviceName,
		Checks:  make(map[string]string),
	}
}

// SetCheck sets the health status for a specific dependency.
func (h *HealthResponse) SetCheck(name string, healthy bool, errMsg string) {
	if healthy {
		h.Checks[name] = "healthy"
	} else {
		h.Checks[name] = "unhealthy: " + errMsg
		h.Status = "degraded"
	}
}

// IsHealthy returns true if overall status is healthy.
func (h *HealthResponse) IsHealthy() bool {
	return h.Status == "healthy"
}
