package events

import (
	"testing"
	"time"
)

func TestNewOrderCreated(t *testing.T) {
	tests := []struct {
		name        string
		orderID     string
		customerID  string
		productID   string
		quantity    int
		totalAmount float64
		status      string
		createdAt   time.Time
	}{
		{
			name:        "creates valid event",
			orderID:     "order-123",
			customerID:  "customer-456",
			productID:   "product-789",
			quantity:    2,
			totalAmount: 99.99,
			status:      OrderStatusPending,
			createdAt:   time.Now(),
		},
		{
			name:        "handles zero quantity",
			orderID:     "order-000",
			customerID:  "customer-001",
			productID:   "product-002",
			quantity:    0,
			totalAmount: 0,
			status:      OrderStatusCancelled,
			createdAt:   time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewOrderCreated(
				tt.orderID,
				tt.customerID,
				tt.productID,
				tt.quantity,
				tt.totalAmount,
				tt.status,
				tt.createdAt,
			)

			if event.OrderID != tt.orderID {
				t.Errorf("OrderID = %v, want %v", event.OrderID, tt.orderID)
			}
			if event.CustomerID != tt.customerID {
				t.Errorf("CustomerID = %v, want %v", event.CustomerID, tt.customerID)
			}
			if event.ProductID != tt.productID {
				t.Errorf("ProductID = %v, want %v", event.ProductID, tt.productID)
			}
			if event.Quantity != tt.quantity {
				t.Errorf("Quantity = %v, want %v", event.Quantity, tt.quantity)
			}
			if event.TotalAmount != tt.totalAmount {
				t.Errorf("TotalAmount = %v, want %v", event.TotalAmount, tt.totalAmount)
			}
			if event.Status != tt.status {
				t.Errorf("Status = %v, want %v", event.Status, tt.status)
			}
			if !event.CreatedAt.Equal(tt.createdAt) {
				t.Errorf("CreatedAt = %v, want %v", event.CreatedAt, tt.createdAt)
			}
			if event.EventType != EventTypeOrderCreated {
				t.Errorf("EventType = %v, want %v", event.EventType, EventTypeOrderCreated)
			}
		})
	}
}

func TestOrderStatusConstants(t *testing.T) {
	// Verify status constants are correctly defined
	if OrderStatusPending != "pending" {
		t.Errorf("OrderStatusPending = %v, want pending", OrderStatusPending)
	}
	if OrderStatusConfirmed != "confirmed" {
		t.Errorf("OrderStatusConfirmed = %v, want confirmed", OrderStatusConfirmed)
	}
	if OrderStatusCancelled != "cancelled" {
		t.Errorf("OrderStatusCancelled = %v, want cancelled", OrderStatusCancelled)
	}
}
