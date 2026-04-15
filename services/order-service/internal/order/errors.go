package order

import "errors"

// Domain errors for order operations.
var (
	ErrMissingCustomerID   = errors.New("customer_id is required")
	ErrMissingProductID    = errors.New("product_id is required")
	ErrInvalidQuantity     = errors.New("quantity must be greater than 0")
	ErrInvalidTotalAmount  = errors.New("total_amount must be greater than 0")
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderCreationFailed = errors.New("failed to create order")
)

// IsValidationError reports whether an error is caused by invalid input.
func IsValidationError(err error) bool {
	return errors.Is(err, ErrMissingCustomerID) ||
		errors.Is(err, ErrMissingProductID) ||
		errors.Is(err, ErrInvalidQuantity) ||
		errors.Is(err, ErrInvalidTotalAmount)
}
