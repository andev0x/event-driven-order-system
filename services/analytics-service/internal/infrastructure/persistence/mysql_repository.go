package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andev0x/analytics-service/internal/analytics"
)

// MySQLRepository implements analytics.Repository using MySQL.
type MySQLRepository struct {
	db *sql.DB
}

// NewMySQLRepository creates a new MySQL analytics repository.
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// SaveOrderMetric inserts a new order metric into the database.
func (r *MySQLRepository) SaveOrderMetric(ctx context.Context, metric *analytics.OrderMetric) error {
	query := `
		INSERT INTO order_metrics (order_id, customer_id, product_id, quantity, total_amount, processed_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		metric.OrderID,
		metric.CustomerID,
		metric.ProductID,
		metric.Quantity,
		metric.TotalAmount,
		metric.ProcessedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save order metric: %w", err)
	}

	return nil
}

// GetSummary retrieves aggregated analytics summary.
func (r *MySQLRepository) GetSummary(ctx context.Context) (*analytics.Summary, error) {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(AVG(total_amount), 0) as average_order_size
		FROM order_metrics
	`

	summary := &analytics.Summary{
		LastUpdated: time.Now(),
	}

	err := r.db.QueryRowContext(ctx, query).Scan(
		&summary.TotalOrders,
		&summary.TotalRevenue,
		&summary.AverageOrderSize,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	return summary, nil
}
