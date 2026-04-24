package order

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andev0x/event-driven-order-system/pkg/observability"
	"github.com/google/uuid"
)

// Service handles business logic for orders.
type Service struct {
	repo      Repository
	cache     Cache
	publisher EventPublisher
}

// NewService creates a new order service.
func NewService(repo Repository, cache Cache, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
	}
}

// Create creates a new order.
func (s *Service) Create(ctx context.Context, req *CreateRequest) (*Order, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Create order entity
	order := NewOrder(uuid.New().String(), req)

	// Persist to database
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOrderCreationFailed, err)
	}

	// Cache the order (non-blocking)
	if err := s.cache.Set(ctx, order); err != nil {
		slog.Warn("Failed to cache order", "order_id", order.ID, "error", err)
	}

	// Publish event asynchronously
	publishCtx := observability.DetachContext(ctx)
	go func() {
		if err := s.publisher.PublishOrderCreated(publishCtx, order); err != nil {
			slog.Error("Failed to publish order created event", "order_id", order.ID, "error", err)
		}
	}()

	slog.Info("Order created successfully", "order_id", order.ID)
	return order, nil
}

// GetByID retrieves an order by ID using cache-aside pattern.
func (s *Service) GetByID(ctx context.Context, id string) (*Order, error) {
	// Try cache first
	order, err := s.cache.Get(ctx, id)
	if err == nil {
		slog.Info("Order cache hit", "order_id", id)
		return order, nil
	}

	slog.Info("Order cache miss, fetching from database", "order_id", id)

	// Cache miss, get from database
	order, err = s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOrderNotFound, err)
	}

	// Update cache
	if err := s.cache.Set(ctx, order); err != nil {
		slog.Warn("Failed to cache order", "order_id", id, "error", err)
	}

	return order, nil
}

// List retrieves a list of orders with pagination.
func (s *Service) List(ctx context.Context, limit, offset int) ([]*Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, nil
}
