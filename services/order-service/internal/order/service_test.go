package order

import (
	"context"
	"errors"
	"testing"
)

// Mock implementations for testing

type mockRepository struct {
	CreateFunc  func(ctx context.Context, order *Order) error
	GetByIDFunc func(ctx context.Context, id string) (*Order, error)
	ListFunc    func(ctx context.Context, limit, offset int) ([]*Order, error)
}

func (m *mockRepository) Create(ctx context.Context, order *Order) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order)
	}
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Order, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]*Order, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return nil, errors.New("not implemented")
}

type mockCache struct {
	GetFunc    func(ctx context.Context, id string) (*Order, error)
	SetFunc    func(ctx context.Context, order *Order) error
	DeleteFunc func(ctx context.Context, id string) error
}

func (m *mockCache) Get(ctx context.Context, id string) (*Order, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockCache) Set(ctx context.Context, order *Order) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, order)
	}
	return nil
}

func (m *mockCache) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

type mockPublisher struct {
	PublishOrderCreatedFunc func(ctx context.Context, order *Order) error
}

func (m *mockPublisher) PublishOrderCreated(ctx context.Context, order *Order) error {
	if m.PublishOrderCreatedFunc != nil {
		return m.PublishOrderCreatedFunc(ctx, order)
	}
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name        string
		request     *CreateRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid order",
			request: &CreateRequest{
				CustomerID:  "customer-123",
				ProductID:   "product-456",
				Quantity:    2,
				TotalAmount: 99.99,
			},
			wantErr: false,
		},
		{
			name: "missing customer_id",
			request: &CreateRequest{
				CustomerID:  "",
				ProductID:   "product-456",
				Quantity:    2,
				TotalAmount: 99.99,
			},
			wantErr:     true,
			errContains: "customer_id",
		},
		{
			name: "missing product_id",
			request: &CreateRequest{
				CustomerID:  "customer-123",
				ProductID:   "",
				Quantity:    2,
				TotalAmount: 99.99,
			},
			wantErr:     true,
			errContains: "product_id",
		},
		{
			name: "invalid quantity",
			request: &CreateRequest{
				CustomerID:  "customer-123",
				ProductID:   "product-456",
				Quantity:    0,
				TotalAmount: 99.99,
			},
			wantErr:     true,
			errContains: "quantity",
		},
		{
			name: "invalid total_amount",
			request: &CreateRequest{
				CustomerID:  "customer-123",
				ProductID:   "product-456",
				Quantity:    2,
				TotalAmount: 0,
			},
			wantErr:     true,
			errContains: "total_amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				CreateFunc: func(_ context.Context, _ *Order) error {
					return nil
				},
			}
			cache := &mockCache{
				SetFunc: func(_ context.Context, _ *Order) error {
					return nil
				},
			}
			publisher := &mockPublisher{
				PublishOrderCreatedFunc: func(_ context.Context, _ *Order) error {
					return nil
				},
			}

			svc := NewService(repo, cache, publisher)
			order, err := svc.Create(context.Background(), tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error but got none")
				}
				if tt.errContains != "" && err != nil {
					errStr := err.Error()
					found := false
					for i := 0; i <= len(errStr)-len(tt.errContains); i++ {
						if errStr[i:i+len(tt.errContains)] == tt.errContains {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Create() error = %v, want error containing %v", err, tt.errContains)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Create() unexpected error = %v", err)
				}
				if order == nil {
					t.Errorf("Create() returned nil order")
				}
				if order != nil && order.Status != StatusPending {
					t.Errorf("Create() status = %v, want %v", order.Status, StatusPending)
				}
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	testOrder := &Order{
		ID:          "order-123",
		CustomerID:  "customer-123",
		ProductID:   "product-456",
		Quantity:    2,
		TotalAmount: 99.99,
		Status:      StatusPending,
	}

	t.Run("cache hit", func(t *testing.T) {
		repo := &mockRepository{}
		cache := &mockCache{
			GetFunc: func(_ context.Context, _ string) (*Order, error) {
				return testOrder, nil
			},
		}
		publisher := &mockPublisher{}

		svc := NewService(repo, cache, publisher)
		order, err := svc.GetByID(context.Background(), "order-123")

		if err != nil {
			t.Errorf("GetByID() unexpected error = %v", err)
		}
		if order.ID != testOrder.ID {
			t.Errorf("GetByID() returned wrong order")
		}
	})

	t.Run("cache miss, db hit", func(t *testing.T) {
		repo := &mockRepository{
			GetByIDFunc: func(_ context.Context, _ string) (*Order, error) {
				return testOrder, nil
			},
		}
		cache := &mockCache{
			GetFunc: func(_ context.Context, _ string) (*Order, error) {
				return nil, errors.New("not found")
			},
			SetFunc: func(_ context.Context, _ *Order) error {
				return nil
			},
		}
		publisher := &mockPublisher{}

		svc := NewService(repo, cache, publisher)
		order, err := svc.GetByID(context.Background(), "order-123")

		if err != nil {
			t.Errorf("GetByID() unexpected error = %v", err)
		}
		if order.ID != testOrder.ID {
			t.Errorf("GetByID() returned wrong order")
		}
	})
}

func TestService_List(t *testing.T) {
	t.Run("applies default limit", func(t *testing.T) {
		var capturedLimit int
		repo := &mockRepository{
			ListFunc: func(_ context.Context, limit, _ int) ([]*Order, error) {
				capturedLimit = limit
				return []*Order{}, nil
			},
		}
		cache := &mockCache{}
		publisher := &mockPublisher{}

		svc := NewService(repo, cache, publisher)
		_, err := svc.List(context.Background(), 0, 0)

		if err != nil {
			t.Errorf("List() unexpected error = %v", err)
		}
		if capturedLimit != 10 {
			t.Errorf("List() limit = %v, want 10 (default)", capturedLimit)
		}
	})

	t.Run("caps maximum limit", func(t *testing.T) {
		var capturedLimit int
		repo := &mockRepository{
			ListFunc: func(_ context.Context, limit, _ int) ([]*Order, error) {
				capturedLimit = limit
				return []*Order{}, nil
			},
		}
		cache := &mockCache{}
		publisher := &mockPublisher{}

		svc := NewService(repo, cache, publisher)
		_, err := svc.List(context.Background(), 1000, 0)

		if err != nil {
			t.Errorf("List() unexpected error = %v", err)
		}
		if capturedLimit != 100 {
			t.Errorf("List() limit = %v, want 100 (max)", capturedLimit)
		}
	})
}

func TestCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateRequest
		wantErr error
	}{
		{
			name: "valid request",
			req: &CreateRequest{
				CustomerID:  "cust-1",
				ProductID:   "prod-1",
				Quantity:    1,
				TotalAmount: 10.0,
			},
			wantErr: nil,
		},
		{
			name: "missing customer_id",
			req: &CreateRequest{
				CustomerID:  "",
				ProductID:   "prod-1",
				Quantity:    1,
				TotalAmount: 10.0,
			},
			wantErr: ErrMissingCustomerID,
		},
		{
			name: "missing product_id",
			req: &CreateRequest{
				CustomerID:  "cust-1",
				ProductID:   "",
				Quantity:    1,
				TotalAmount: 10.0,
			},
			wantErr: ErrMissingProductID,
		},
		{
			name: "zero quantity",
			req: &CreateRequest{
				CustomerID:  "cust-1",
				ProductID:   "prod-1",
				Quantity:    0,
				TotalAmount: 10.0,
			},
			wantErr: ErrInvalidQuantity,
		},
		{
			name: "negative quantity",
			req: &CreateRequest{
				CustomerID:  "cust-1",
				ProductID:   "prod-1",
				Quantity:    -1,
				TotalAmount: 10.0,
			},
			wantErr: ErrInvalidQuantity,
		},
		{
			name: "zero total_amount",
			req: &CreateRequest{
				CustomerID:  "cust-1",
				ProductID:   "prod-1",
				Quantity:    1,
				TotalAmount: 0,
			},
			wantErr: ErrInvalidTotalAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
