package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andev0x/analytics-service/internal/analytics"
)

func TestHandler_GetSummary(t *testing.T) {
	testSummary := &analytics.Summary{
		TotalOrders:      100,
		TotalRevenue:     9999.99,
		AverageOrderSize: 99.99,
		LastUpdated:      time.Now(),
	}

	t.Run("returns summary successfully", func(t *testing.T) {
		repo := &mockRepo{
			getSummaryFunc: func(_ context.Context) (*analytics.Summary, error) {
				return testSummary, nil
			},
		}
		cache := &mockCache{
			getSummaryFunc: func(_ context.Context) (*analytics.Summary, error) {
				return testSummary, nil
			},
		}

		svc := analytics.NewService(repo, cache)
		handler := NewHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
		w := httptest.NewRecorder()

		handler.GetSummary(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetSummary() status = %v, want %v", w.Code, http.StatusOK)
		}
	})
}

func TestHandler_HealthCheck(t *testing.T) {
	t.Run("healthy when no checks configured", func(t *testing.T) {
		svc := analytics.NewService(&mockRepo{}, &mockCache{})
		handler := NewHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HealthCheck() status = %v, want %v", w.Code, http.StatusOK)
		}
	})

	t.Run("degraded when db unhealthy", func(t *testing.T) {
		svc := analytics.NewService(&mockRepo{}, &mockCache{})
		handler := NewHandler(svc)
		handler.SetHealthChecker(&HealthChecker{
			DBHealthFunc: func() error {
				return errors.New("db down")
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("HealthCheck() status = %v, want %v", w.Code, http.StatusServiceUnavailable)
		}
	})
}

// Mock implementations

type mockRepo struct {
	saveOrderMetricFunc func(ctx context.Context, metric *analytics.OrderMetric) error
	getSummaryFunc      func(ctx context.Context) (*analytics.Summary, error)
}

func (m *mockRepo) SaveOrderMetric(ctx context.Context, metric *analytics.OrderMetric) error {
	if m.saveOrderMetricFunc != nil {
		return m.saveOrderMetricFunc(ctx, metric)
	}
	return nil
}

func (m *mockRepo) GetSummary(ctx context.Context) (*analytics.Summary, error) {
	if m.getSummaryFunc != nil {
		return m.getSummaryFunc(ctx)
	}
	return &analytics.Summary{}, nil
}

type mockCache struct {
	getSummaryFunc        func(ctx context.Context) (*analytics.Summary, error)
	setSummaryFunc        func(ctx context.Context, summary *analytics.Summary) error
	invalidateSummaryFunc func(ctx context.Context) error
}

func (m *mockCache) GetSummary(ctx context.Context) (*analytics.Summary, error) {
	if m.getSummaryFunc != nil {
		return m.getSummaryFunc(ctx)
	}
	return nil, errors.New("not found")
}

func (m *mockCache) SetSummary(ctx context.Context, summary *analytics.Summary) error {
	if m.setSummaryFunc != nil {
		return m.setSummaryFunc(ctx, summary)
	}
	return nil
}

func (m *mockCache) InvalidateSummary(ctx context.Context) error {
	if m.invalidateSummaryFunc != nil {
		return m.invalidateSummaryFunc(ctx)
	}
	return nil
}
