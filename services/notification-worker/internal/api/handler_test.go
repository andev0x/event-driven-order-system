package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockHealthChecker is a mock implementation of HealthChecker for testing.
type mockHealthChecker struct {
	err error
}

func (m *mockHealthChecker) HealthCheck() error {
	return m.err
}

func TestHandler_Health_AllHealthy(t *testing.T) {
	handler := NewHandler(&mockHealthChecker{err: nil})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", resp.Status)
	}

	if resp.Service != "notification-worker" {
		t.Errorf("expected service 'notification-worker', got '%s'", resp.Service)
	}

	if resp.Checks["mq"] != "healthy" {
		t.Errorf("expected mq check 'healthy', got '%s'", resp.Checks["mq"])
	}
}

func TestHandler_Health_MQUnhealthy(t *testing.T) {
	handler := NewHandler(&mockHealthChecker{err: &testError{msg: "connection closed"}})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", resp.Status)
	}
}

func TestHandler_Health_NilChecker(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", resp.Status)
	}

	if resp.Checks["mq"] != "unhealthy: not configured" {
		t.Errorf("expected mq check 'unhealthy: not configured', got '%s'", resp.Checks["mq"])
	}
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
