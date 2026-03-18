package analytics

import "context"

// Repository defines the interface for analytics persistence.
type Repository interface {
	// SaveOrderMetric persists an order metric to the database.
	SaveOrderMetric(ctx context.Context, metric *OrderMetric) error

	// GetSummary retrieves aggregated analytics summary.
	GetSummary(ctx context.Context) (*Summary, error)
}

// Cache defines the interface for analytics caching.
type Cache interface {
	// GetSummary retrieves analytics summary from cache.
	GetSummary(ctx context.Context) (*Summary, error)

	// SetSummary stores analytics summary in cache.
	SetSummary(ctx context.Context, summary *Summary) error

	// InvalidateSummary removes analytics summary from cache.
	InvalidateSummary(ctx context.Context) error
}
