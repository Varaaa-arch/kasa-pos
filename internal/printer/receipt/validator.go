package receipt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

var (
	ErrInvalidReceipt       = errors.New("invalid receipt")
	ErrInvoiceRequired      = errors.New("invoice number is required")
	ErrTimestampInvalid     = errors.New("timestamp is invalid")
	ErrItemsRequired        = errors.New("receipt must contain at least one item")
	ErrInvalidItem          = errors.New("invalid receipt item")
	ErrQuantityInvalid      = errors.New("quantity must be greater than zero")
	ErrPriceInvalid         = errors.New("unit price must be greater than or equal to zero")
	ErrTotalInvalid         = errors.New("receipt total is invalid")
	ErrPaymentInvalid       = errors.New("payment is invalid")
	ErrPaymentAmountInvalid = errors.New("payment amount is invalid")
	ErrChangeInvalid        = errors.New("change is invalid")
)

type ValidationError struct {
	Field string
	Err   error
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Err.Error()
	}

	return fmt.Sprintf(
		"%s: %v",
		e.Field,
		e.Err,
	)
}

func (e ValidationError) Unwrap() error {
	return e.Err
}

type Validator struct {
	RequireItems bool
}

func NewValidator() *Validator {
	return &Validator{
		RequireItems: true,
	}
}

func (v *Validator) Validate(
	input domainreceipt.Receipt,
) error {
	if strings.TrimSpace(
		input.Transaction.InvoiceNumber,
	) == "" {
		return ValidationError{
			Field: "transaction.invoice_number",
			Err:   ErrInvoiceRequired,
		}
	}

	if input.Transaction.Timestamp.IsZero() {
		return ValidationError{
			Field: "transaction.timestamp",
			Err:   ErrTimestampInvalid,
		}
	}

	if input.Transaction.Timestamp.After(
		time.Now().Add(time.Minute),
	) {
		return ValidationError{
			Field: "transaction.timestamp",
			Err:   ErrTimestampInvalid,
		}
	}

	if v.RequireItems && len(input.Items) == 0 {
		return ValidationError{
			Field: "items",
			Err:   ErrItemsRequired,
		}
	}

	for index, item := range input.Items {
		if strings.TrimSpace(item.Name) == "" {
			return ValidationError{
				Field: fmt.Sprintf(
					"items[%d].name",
					index,
				),
				Err: ErrInvalidItem,
			}
		}

		if item.Quantity <= 0 {
			return ValidationError{
				Field: fmt.Sprintf(
					"items[%d].quantity",
					index,
				),
				Err: ErrQuantityInvalid,
			}
		}

		if item.UnitPrice.IsNegative() {
			return ValidationError{
				Field: fmt.Sprintf(
					"items[%d].unit_price",
					index,
				),
				Err: ErrPriceInvalid,
			}
		}
	}

	calculator := NewCalculator()

	calculation := calculator.Calculate(input)

	if !input.Summary.Subtotal.IsZero() &&
		!input.Summary.Subtotal.Equal(calculation.Subtotal) {
		return ValidationError{
			Field: "summary.subtotal",
			Err:   ErrTotalInvalid,
		}
	}

	if calculation.Total.IsNegative() {
		return ValidationError{
			Field: "summary.total",
			Err:   ErrTotalInvalid,
		}
	}

	if !input.Summary.Total.IsZero() &&
		!input.Summary.Total.Equal(calculation.Total) {
		return ValidationError{
			Field: "summary.total",
			Err:   ErrTotalInvalid,
		}
	}

	if strings.TrimSpace(input.Payment.Method) == "" &&
		!input.Payment.Paid.IsZero() {
		return ValidationError{
			Field: "payment.method",
			Err:   ErrPaymentInvalid,
		}
	}

	if input.Payment.Paid.IsNegative() {
		return ValidationError{
			Field: "payment.paid",
			Err:   ErrPaymentAmountInvalid,
		}
	}

	if calculation.Paid.LessThan(calculation.Total) {
		return ValidationError{
			Field: "payment.change",
			Err:   ErrChangeInvalid,
		}
	}

	if !input.Payment.Change.IsZero() &&
		!input.Payment.Change.Equal(calculation.Change) {
		return ValidationError{
			Field: "payment.change",
			Err:   ErrChangeInvalid,
		}
	}

	return nil
}
