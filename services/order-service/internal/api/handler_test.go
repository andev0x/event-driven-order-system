package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andev0x/order-service/internal/order"
	"github.com/gorilla/mux"
)

// Mock service for testing
type mockOrderService struct {
	createFunc  func(ctx context.Context, req *order.CreateRequest) (*order.Order, error)
	getByIDFunc func(ctx context.Context, id string) (*order.Order, error)
	listFunc    func(ctx context.Context, limit, offset int) ([]*order.Order, error)
}

// We need to wrap the mock to satisfy the handler's dependency on order.Service
// Since the handler uses *order.Service directly, we'll test at a higher level

func TestHandler_CreateOrder_InvalidJSON(t *testing.T) {
	// Create a real service with mocks
	svc := order.NewService(
		&mockRepo{},
		&mockCache{},
		&mockPublisher{},
	)
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateOrder() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_CreateOrder_UnknownField(t *testing.T) {
	svc := order.NewService(
		&mockRepo{},
		&mockCache{},
		&mockPublisher{},
	)
	handler := NewHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{"customer_id":"cust-1","product_id":"prod-1","quantity":1,"total_amount":10,"unknown":"value"}`),
	)
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateOrder() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_CreateOrder_ValidationError(t *testing.T) {
	svc := order.NewService(
		&mockRepo{},
		&mockCache{},
		&mockPublisher{},
	)
	handler := NewHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`{"customer_id":"","product_id":"prod-1","quantity":1,"total_amount":10}`),
	)
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateOrder() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_CreateOrder_Success(t *testing.T) {
	repo := &mockRepo{
		createFunc: func(_ context.Context, _ *order.Order) error {
			return nil
		},
	}
	cache := &mockCache{
		setFunc: func(_ context.Context, _ *order.Order) error {
			return nil
		},
	}
	publisher := &mockPublisher{}

	svc := order.NewService(repo, cache, publisher)
	handler := NewHandler(svc)

	reqBody := order.CreateRequest{
		CustomerID:  "cust-123",
		ProductID:   "prod-456",
		Quantity:    2,
		TotalAmount: 99.99,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateOrder(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("CreateOrder() status = %v, want %v", w.Code, http.StatusCreated)
	}
}

func TestHandler_GetOrder_MissingID(t *testing.T) {
	svc := order.NewService(&mockRepo{}, &mockCache{}, &mockPublisher{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/orders/", nil)
	req = mux.SetURLVars(req, map[string]string{"id": ""})
	w := httptest.NewRecorder()

	handler.GetOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetOrder() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_GetOrder_Found(t *testing.T) {
	testOrder := &order.Order{
		ID:          "order-123",
		CustomerID:  "cust-123",
		ProductID:   "prod-456",
		Quantity:    2,
		TotalAmount: 99.99,
		Status:      order.StatusPending,
	}

	cache := &mockCache{
		getFunc: func(_ context.Context, _ string) (*order.Order, error) {
			return testOrder, nil
		},
	}
	svc := order.NewService(&mockRepo{}, cache, &mockPublisher{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/orders/order-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "order-123"})
	w := httptest.NewRecorder()

	handler.GetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetOrder() status = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestHandler_HealthCheck_Healthy(t *testing.T) {
	svc := order.NewService(&mockRepo{}, &mockCache{}, &mockPublisher{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HealthCheck() status = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestHandler_HealthCheck_Degraded(t *testing.T) {
	svc := order.NewService(&mockRepo{}, &mockCache{}, &mockPublisher{})
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
}

func TestHandler_ListOrders_InvalidLimit(t *testing.T) {
	svc := order.NewService(&mockRepo{}, &mockCache{}, &mockPublisher{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/orders?limit=abc", nil)
	w := httptest.NewRecorder()

	handler.ListOrders(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ListOrders() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_ListOrders_NegativeOffset(t *testing.T) {
	svc := order.NewService(&mockRepo{}, &mockCache{}, &mockPublisher{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/orders?offset=-1", nil)
	w := httptest.NewRecorder()

	handler.ListOrders(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ListOrders() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// Mock implementations for handler tests

type mockRepo struct {
	createFunc  func(ctx context.Context, o *order.Order) error
	getByIDFunc func(ctx context.Context, id string) (*order.Order, error)
	listFunc    func(ctx context.Context, limit, offset int) ([]*order.Order, error)
}

func (m *mockRepo) Create(ctx context.Context, o *order.Order) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, o)
	}
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*order.Order, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]*order.Order, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []*order.Order{}, nil
}

type mockCache struct {
	getFunc    func(ctx context.Context, id string) (*order.Order, error)
	setFunc    func(ctx context.Context, o *order.Order) error
	deleteFunc func(ctx context.Context, id string) error
}

func (m *mockCache) Get(ctx context.Context, id string) (*order.Order, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockCache) Set(ctx context.Context, o *order.Order) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, o)
	}
	return nil
}

func (m *mockCache) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

type mockPublisher struct {
	publishFunc func(ctx context.Context, o *order.Order) error
}

func (m *mockPublisher) PublishOrderCreated(ctx context.Context, o *order.Order) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, o)
	}
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}
