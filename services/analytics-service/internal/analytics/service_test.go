package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andev0x/event-driven-order-system/pkg/events"
)

// Mock implementations

type mockRepository struct {
	saveOrderMetricFunc func(ctx context.Context, metric *OrderMetric) error
	getSummaryFunc      func(ctx context.Context) (*Summary, error)
}

func (m *mockRepository) SaveOrderMetric(ctx context.Context, metric *OrderMetric) error {
	if m.saveOrderMetricFunc != nil {
		return m.saveOrderMetricFunc(ctx, metric)
	}
	return nil
}

func (m *mockRepository) GetSummary(ctx context.Context) (*Summary, error) {
	if m.getSummaryFunc != nil {
		return m.getSummaryFunc(ctx)
	}
	return &Summary{}, nil
}

type mockCache struct {
	getSummaryFunc        func(ctx context.Context) (*Summary, error)
	setSummaryFunc        func(ctx context.Context, summary *Summary) error
	invalidateSummaryFunc func(ctx context.Context) error
}

func (m *mockCache) GetSummary(ctx context.Context) (*Summary, error) {
	if m.getSummaryFunc != nil {
		return m.getSummaryFunc(ctx)
	}
	return nil, errors.New("not found")
}

func (m *mockCache) SetSummary(ctx context.Context, summary *Summary) error {
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

func TestService_ProcessOrderEvent(t *testing.T) {
	t.Run("successfully processes event", func(t *testing.T) {
		var savedMetric *OrderMetric
		repo := &mockRepository{
			saveOrderMetricFunc: func(_ context.Context, metric *OrderMetric) error {
				savedMetric = metric
				return nil
			},
		}
		cache := &mockCache{
			invalidateSummaryFunc: func(_ context.Context) error {
				return nil
			},
		}

		svc := NewService(repo, cache)

		event := &events.OrderCreated{
			OrderID:     "order-123",
			CustomerID:  "cust-456",
			ProductID:   "prod-789",
			Quantity:    2,
			TotalAmount: 99.99,
		}

		err := svc.ProcessOrderEvent(context.Background(), event)
		if err != nil {
			t.Errorf("ProcessOrderEvent() error = %v", err)
		}

		if savedMetric == nil {
			t.Error("ProcessOrderEvent() did not save metric")
		}
		if savedMetric != nil && savedMetric.OrderID != event.OrderID {
			t.Errorf("OrderID = %v, want %v", savedMetric.OrderID, event.OrderID)
		}
	})

	t.Run("returns error when repo fails", func(t *testing.T) {
		repo := &mockRepository{
			saveOrderMetricFunc: func(_ context.Context, _ *OrderMetric) error {
				return errors.New("db error")
			},
		}
		cache := &mockCache{}

		svc := NewService(repo, cache)

		event := &events.OrderCreated{
			OrderID: "order-123",
		}

		err := svc.ProcessOrderEvent(context.Background(), event)
		if err == nil {
			t.Error("ProcessOrderEvent() expected error")
		}
	})
}

func TestService_GetSummary(t *testing.T) {
	testSummary := &Summary{
		TotalOrders:      100,
		TotalRevenue:     9999.99,
		AverageOrderSize: 99.99,
		LastUpdated:      time.Now(),
	}

	t.Run("cache hit", func(t *testing.T) {
		repo := &mockRepository{}
		cache := &mockCache{
			getSummaryFunc: func(_ context.Context) (*Summary, error) {
				return testSummary, nil
			},
		}

		svc := NewService(repo, cache)
		summary, err := svc.GetSummary(context.Background())

		if err != nil {
			t.Errorf("GetSummary() error = %v", err)
		}
		if summary.TotalOrders != testSummary.TotalOrders {
			t.Errorf("TotalOrders = %v, want %v", summary.TotalOrders, testSummary.TotalOrders)
		}
	})

	t.Run("cache miss, db hit", func(t *testing.T) {
		repo := &mockRepository{
			getSummaryFunc: func(_ context.Context) (*Summary, error) {
				return testSummary, nil
			},
		}
		cache := &mockCache{
			getSummaryFunc: func(_ context.Context) (*Summary, error) {
				return nil, errors.New("not found")
			},
			setSummaryFunc: func(_ context.Context, _ *Summary) error {
				return nil
			},
		}

		svc := NewService(repo, cache)
		summary, err := svc.GetSummary(context.Background())

		if err != nil {
			t.Errorf("GetSummary() error = %v", err)
		}
		if summary.TotalOrders != testSummary.TotalOrders {
			t.Errorf("TotalOrders = %v, want %v", summary.TotalOrders, testSummary.TotalOrders)
		}
	})
}

func TestNewOrderMetric(t *testing.T) {
	metric := NewOrderMetric("order-1", "cust-1", "prod-1", 5, 49.99)

	if metric.OrderID != "order-1" {
		t.Errorf("OrderID = %v, want order-1", metric.OrderID)
	}
	if metric.CustomerID != "cust-1" {
		t.Errorf("CustomerID = %v, want cust-1", metric.CustomerID)
	}
	if metric.ProductID != "prod-1" {
		t.Errorf("ProductID = %v, want prod-1", metric.ProductID)
	}
	if metric.Quantity != 5 {
		t.Errorf("Quantity = %v, want 5", metric.Quantity)
	}
	if metric.TotalAmount != 49.99 {
		t.Errorf("TotalAmount = %v, want 49.99", metric.TotalAmount)
	}
	if metric.ProcessedAt.IsZero() {
		t.Error("ProcessedAt should not be zero")
	}
}
