// Package events defines shared event types used for inter-service communication.
//
// This package provides a centralized location for all event schemas that are
// published and consumed across different services in the system. By sharing
// these types, we ensure consistency and type safety across service boundaries.
//
// Example usage:
//
//	event := &events.OrderCreated{
//	    OrderID:     "order-123",
//	    CustomerID:  "cust-456",
//	    ProductID:   "prod-789",
//	    Quantity:    2,
//	    TotalAmount: 99.99,
//	    Status:      events.OrderStatusPending,
//	    CreatedAt:   time.Now(),
//	}
package events
