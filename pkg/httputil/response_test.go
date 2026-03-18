package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		payload    interface{}
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success response",
			statusCode: http.StatusOK,
			payload:    map[string]string{"message": "success"},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"success"}`,
		},
		{
			name:       "created response",
			statusCode: http.StatusCreated,
			payload:    map[string]int{"id": 123},
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			RespondJSON(w, tt.statusCode, tt.payload)

			if w.Code != tt.wantStatus {
				t.Errorf("status code = %v, want %v", w.Code, tt.wantStatus)
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", w.Header().Get("Content-Type"))
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("body = %v, want %v", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()

	RespondError(w, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %v, want %v", w.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error != "invalid input" {
		t.Errorf("error = %v, want invalid input", resp.Error)
	}
}

func TestHealthResponse(t *testing.T) {
	t.Run("new health response is healthy", func(t *testing.T) {
		hr := NewHealthResponse("test-service")

		if hr.Status != "healthy" {
			t.Errorf("Status = %v, want healthy", hr.Status)
		}
		if hr.Service != "test-service" {
			t.Errorf("Service = %v, want test-service", hr.Service)
		}
		if !hr.IsHealthy() {
			t.Error("IsHealthy() = false, want true")
		}
	})

	t.Run("healthy check keeps status healthy", func(t *testing.T) {
		hr := NewHealthResponse("test-service")
		hr.SetCheck("database", true, "")

		if hr.Checks["database"] != "healthy" {
			t.Errorf("Checks[database] = %v, want healthy", hr.Checks["database"])
		}
		if !hr.IsHealthy() {
			t.Error("IsHealthy() = false, want true")
		}
	})

	t.Run("unhealthy check sets degraded status", func(t *testing.T) {
		hr := NewHealthResponse("test-service")
		hr.SetCheck("database", false, "connection refused")

		if hr.Checks["database"] != "unhealthy: connection refused" {
			t.Errorf("Checks[database] = %v, want unhealthy: connection refused", hr.Checks["database"])
		}
		if hr.Status != "degraded" {
			t.Errorf("Status = %v, want degraded", hr.Status)
		}
		if hr.IsHealthy() {
			t.Error("IsHealthy() = true, want false")
		}
	})
}
