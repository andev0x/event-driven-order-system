package analytics

import (
	"time"
)

// OrderMetric represents aggregated order metrics stored for analytics.
type OrderMetric struct {
	ID          int       `json:"id"`
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	ProcessedAt time.Time `json:"processed_at"`
}

// Summary represents aggregated analytics data.
type Summary struct {
	TotalOrders      int       `json:"total_orders"`
	TotalRevenue     float64   `json:"total_revenue"`
	AverageOrderSize float64   `json:"average_order_size"`
	LastUpdated      time.Time `json:"last_updated"`
}

// NewOrderMetric creates a new OrderMetric from order event data.
func NewOrderMetric(orderID, customerID, productID string, quantity int, totalAmount float64) *OrderMetric {
	return &OrderMetric{
		OrderID:     orderID,
		CustomerID:  customerID,
		ProductID:   productID,
		Quantity:    quantity,
		TotalAmount: totalAmount,
		ProcessedAt: time.Now(),
	}
}
