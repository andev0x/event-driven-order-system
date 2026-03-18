package order

import (
	"time"
)

// Status constants define possible order states.
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
)

// Order represents an order in the system.
type Order struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateRequest represents the request to create an order.
type CreateRequest struct {
	CustomerID  string  `json:"customer_id" validate:"required"`
	ProductID   string  `json:"product_id" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	TotalAmount float64 `json:"total_amount" validate:"required,gt=0"`
}

// Validate validates the create request.
func (r *CreateRequest) Validate() error {
	if r.CustomerID == "" {
		return ErrMissingCustomerID
	}
	if r.ProductID == "" {
		return ErrMissingProductID
	}
	if r.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if r.TotalAmount <= 0 {
		return ErrInvalidTotalAmount
	}
	return nil
}

// NewOrder creates a new Order from a CreateRequest.
func NewOrder(id string, req *CreateRequest) *Order {
	now := time.Now()
	return &Order{
		ID:          id,
		CustomerID:  req.CustomerID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		TotalAmount: req.TotalAmount,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
