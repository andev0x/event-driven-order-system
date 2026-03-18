package analytics

import (
	"context"
	"fmt"
	"log"

	"github.com/andev0x/event-driven-order-system/pkg/events"
)

// Service handles business logic for analytics.
type Service struct {
	repo  Repository
	cache Cache
}

// NewService creates a new analytics service.
func NewService(repo Repository, cache Cache) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}

// ProcessOrderEvent processes an order created event.
func (s *Service) ProcessOrderEvent(ctx context.Context, event *events.OrderCreated) error {
	// Create metric from event
	metric := NewOrderMetric(
		event.OrderID,
		event.CustomerID,
		event.ProductID,
		event.Quantity,
		event.TotalAmount,
	)

	// Save to database
	if err := s.repo.SaveOrderMetric(ctx, metric); err != nil {
		return fmt.Errorf("failed to save order metric: %w", err)
	}

	// Invalidate cache to force fresh calculation on next request
	if err := s.cache.InvalidateSummary(ctx); err != nil {
		log.Printf("Warning: failed to invalidate cache: %v", err)
	}

	log.Printf("Successfully processed order event: OrderID=%s, Amount=%.2f", event.OrderID, event.TotalAmount)
	return nil
}

// GetSummary retrieves analytics summary using cache-aside pattern.
func (s *Service) GetSummary(ctx context.Context) (*Summary, error) {
	// Try to get from cache first
	summary, err := s.cache.GetSummary(ctx)
	if err == nil {
		log.Println("Cache hit for analytics summary")
		return summary, nil
	}

	log.Println("Cache miss for analytics summary, fetching from database")

	// Cache miss, get from database
	summary, err = s.repo.GetSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	// Update cache
	if err := s.cache.SetSummary(ctx, summary); err != nil {
		log.Printf("Warning: failed to cache summary: %v", err)
	}

	return summary, nil
}
