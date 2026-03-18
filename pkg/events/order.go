package events

import "time"

// EventType constants define the types of events in the system.
const (
	EventTypeOrderCreated   = "OrderCreated"
	EventTypeOrderConfirmed = "OrderConfirmed"
	EventTypeOrderCancelled = "OrderCancelled"
)

// OrderStatus constants define possible order states.
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusCancelled = "cancelled"
)

// OrderCreated represents the event published when a new order is created.
// This event is consumed by analytics and notification services.
type OrderCreated struct {
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	EventType   string    `json:"event_type"`
}

// NewOrderCreated creates a new OrderCreated event with the EventType pre-filled.
func NewOrderCreated(orderID, customerID, productID string, quantity int, totalAmount float64, status string, createdAt time.Time) *OrderCreated {
	return &OrderCreated{
		OrderID:     orderID,
		CustomerID:  customerID,
		ProductID:   productID,
		Quantity:    quantity,
		TotalAmount: totalAmount,
		Status:      status,
		CreatedAt:   createdAt,
		EventType:   EventTypeOrderCreated,
	}
}
