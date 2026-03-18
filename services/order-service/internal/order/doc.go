// Package order contains the core domain logic for order management.
//
// This package implements the Order domain following domain-driven design principles.
// It contains the Order entity, repository interfaces, and business logic services.
//
// The order domain is responsible for:
//   - Creating new orders with validation
//   - Retrieving orders by ID or listing orders
//   - Publishing order events for other services to consume
//
// Example usage:
//
//	svc := order.NewService(repo, cache, publisher)
//	order, err := svc.Create(ctx, &order.CreateRequest{
//	    CustomerID:  "cust-123",
//	    ProductID:   "prod-456",
//	    Quantity:    2,
//	    TotalAmount: 99.99,
//	})
package order
