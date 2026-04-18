package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/andev0x/order-service/internal/order"
)

// MySQLRepository implements order.Repository using MySQL.
type MySQLRepository struct {
	db *sql.DB
}

// NewMySQLRepository creates a new MySQL order repository.
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// Create inserts a new order into the database.
func (r *MySQLRepository) Create(ctx context.Context, o *order.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, product_id, quantity, total_amount, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		o.ID,
		o.CustomerID,
		o.ProductID,
		o.Quantity,
		o.TotalAmount,
		o.Status,
		o.CreatedAt,
		o.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID retrieves an order by its ID.
func (r *MySQLRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	query := `
		SELECT id, customer_id, product_id, quantity, total_amount, status, created_at, updated_at
		FROM orders
		WHERE id = ?
	`

	o := &order.Order{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID,
		&o.CustomerID,
		&o.ProductID,
		&o.Quantity,
		&o.TotalAmount,
		&o.Status,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, order.ErrOrderNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return o, nil
}

// List retrieves a list of orders with pagination.
func (r *MySQLRepository) List(ctx context.Context, limit, offset int) ([]*order.Order, error) {
	query := `
		SELECT id, customer_id, product_id, quantity, total_amount, status, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Failed to close rows", "error", err)
		}
	}()

	var orders []*order.Order
	for rows.Next() {
		o := &order.Order{}
		err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.ProductID,
			&o.Quantity,
			&o.TotalAmount,
			&o.Status,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}
