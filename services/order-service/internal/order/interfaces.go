package order

import "context"

// Repository defines the interface for order persistence.
type Repository interface {
	// Create persists a new order to the database.
	Create(ctx context.Context, order *Order) error

	// GetByID retrieves an order by its ID.
	GetByID(ctx context.Context, id string) (*Order, error)

	// List retrieves orders with pagination.
	List(ctx context.Context, limit, offset int) ([]*Order, error)
}

// Cache defines the interface for order caching.
type Cache interface {
	// Get retrieves an order from cache.
	Get(ctx context.Context, id string) (*Order, error)

	// Set stores an order in cache.
	Set(ctx context.Context, order *Order) error

	// Delete removes an order from cache.
	Delete(ctx context.Context, id string) error
}

// EventPublisher defines the interface for publishing order events.
type EventPublisher interface {
	// PublishOrderCreated publishes an order created event.
	PublishOrderCreated(ctx context.Context, order *Order) error

	// Close closes the publisher connection.
	Close() error
}
