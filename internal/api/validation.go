package api

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError wraps a validation failure with a human-readable message.
// Implements the error interface so it can be used with errors.As.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// validateCheckoutRequest checks the request before touching the database.
// Returns a *ValidationError if any rule is violated, nil otherwise.
//
// Rules:
//   - items must not be empty
//   - each item.product_id must not be blank
//   - each item.quantity must be > 0
//   - paid_amount must be > 0
func validateCheckoutRequest(req CheckoutRequest) error {
	if len(req.Items) == 0 {
		return &ValidationError{Message: "cart is empty"}
	}

	for i, item := range req.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("items[%d].product_id", i),
				Message: "product_id is required",
			}
		}

		if item.Quantity <= 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("items[%d].quantity", i),
				Message: "quantity must be greater than 0",
			}
		}
	}

	if req.PaidAmount <= 0 {
		return &ValidationError{
			Field:   "paid_amount",
			Message: "paid_amount must be greater than 0",
		}
	}

	return nil
}

// isValidationError reports whether err is a *ValidationError.
func isValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
