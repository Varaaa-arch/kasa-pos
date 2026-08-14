package receipt

import (
	"errors"
	"testing"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

func validReceipt() domainreceipt.Receipt {
	return domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-VALID-001",
			InvoiceNumber: "INV-VALID-001",
			Timestamp:     time.Now(),
			Cashier:       "Bizar",
		},

		Items: []domainreceipt.Item{
			{
				ProductID: "PROD-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
			},
			{
				ProductID: "PROD-002",
				SKU:       "ROTI-001",
				Name:      "Roti Bakar",
				Quantity:  1,
				UnitPrice: 12000,
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal: 42000,
			Total:    42000,
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid:   50000,
			Change: 8000,
		},

		Footer: domainreceipt.Footer{
			Message: "TERIMA KASIH",
		},
	}
}

func TestValidatorValidReceipt(t *testing.T) {
	validator := NewValidator()

	err := validator.Validate(validReceipt())

	if err != nil {
		t.Fatalf(
			"expected valid receipt, got error: %v",
			err,
		)
	}
}

func TestValidatorRequiresInvoice(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Transaction.InvoiceNumber = ""

	err := validator.Validate(input)

	if !errors.Is(err, ErrInvoiceRequired) {
		t.Fatalf(
			"expected ErrInvoiceRequired, got %v",
			err,
		)
	}

	var validationErr ValidationError

	if !errors.As(err, &validationErr) {
		t.Fatalf(
			"expected ValidationError, got %T",
			err,
		)
	}

	if validationErr.Field != "transaction.invoice_number" {
		t.Fatalf(
			"unexpected field: %q",
			validationErr.Field,
		)
	}
}

func TestValidatorRequiresTimestamp(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Transaction.Timestamp = time.Time{}

	err := validator.Validate(input)

	if !errors.Is(err, ErrTimestampInvalid) {
		t.Fatalf(
			"expected ErrTimestampInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRejectsFutureTimestamp(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Transaction.Timestamp = time.Now().Add(2 * time.Hour)

	err := validator.Validate(input)

	if !errors.Is(err, ErrTimestampInvalid) {
		t.Fatalf(
			"expected ErrTimestampInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRequiresItems(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Items = nil
	input.Summary.Subtotal = 0
	input.Summary.Total = 0
	input.Payment.Paid = 0
	input.Payment.Change = 0

	err := validator.Validate(input)

	if !errors.Is(err, ErrItemsRequired) {
		t.Fatalf(
			"expected ErrItemsRequired, got %v",
			err,
		)
	}
}

func TestValidatorRejectsEmptyItemName(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Items[0].Name = ""

	err := validator.Validate(input)

	if !errors.Is(err, ErrInvalidItem) {
		t.Fatalf(
			"expected ErrInvalidItem, got %v",
			err,
		)
	}
}

func TestValidatorRejectsZeroQuantity(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Items[0].Quantity = 0

	err := validator.Validate(input)

	if !errors.Is(err, ErrQuantityInvalid) {
		t.Fatalf(
			"expected ErrQuantityInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRejectsNegativeQuantity(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Items[0].Quantity = -1

	err := validator.Validate(input)

	if !errors.Is(err, ErrQuantityInvalid) {
		t.Fatalf(
			"expected ErrQuantityInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRejectsNegativeUnitPrice(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Items[0].UnitPrice = -1

	err := validator.Validate(input)

	if !errors.Is(err, ErrPriceInvalid) {
		t.Fatalf(
			"expected ErrPriceInvalid, got %v",
			err,
		)
	}
}

func TestValidatorAcceptsZeroUnitPrice(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Items[0].UnitPrice = 0
	input.Summary.Subtotal = 12000
	input.Summary.Total = 12000
	input.Payment.Paid = 12000
	input.Payment.Change = 0

	err := validator.Validate(input)

	if err != nil {
		t.Fatalf(
			"expected zero unit price to be accepted, got %v",
			err,
		)
	}
}

func TestValidatorRejectsIncorrectSubtotal(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Summary.Subtotal = 99999

	err := validator.Validate(input)

	if !errors.Is(err, ErrTotalInvalid) {
		t.Fatalf(
			"expected ErrTotalInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRejectsIncorrectTotal(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Summary.Total = 99999

	err := validator.Validate(input)

	if !errors.Is(err, ErrTotalInvalid) {
		t.Fatalf(
			"expected ErrTotalInvalid, got %v",
			err,
		)
	}
}

func TestValidatorCalculatesTotalWithDiscount(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()

	input.Summary.Subtotal = 42000
	input.Summary.Discount = 5000
	input.Summary.Tax = 0
	input.Summary.ServiceCharge = 0
	input.Summary.Total = 37000

	input.Payment.Paid = 50000
	input.Payment.Change = 13000

	err := validator.Validate(input)

	if err != nil {
		t.Fatalf(
			"expected valid discounted receipt, got %v",
			err,
		)
	}
}

func TestValidatorCalculatesTotalWithTax(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()

	input.Summary.Subtotal = 42000
	input.Summary.Discount = 0
	input.Summary.Tax = 4200
	input.Summary.ServiceCharge = 0
	input.Summary.Total = 46200

	input.Payment.Paid = 50000
	input.Payment.Change = 3800

	err := validator.Validate(input)

	if err != nil {
		t.Fatalf(
			"expected valid taxed receipt, got %v",
			err,
		)
	}
}

func TestValidatorRejectsNegativeTotal(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()

	input.Summary.Subtotal = 42000
	input.Summary.Discount = 50000
	input.Summary.Tax = 0
	input.Summary.ServiceCharge = 0
	input.Summary.Total = -8000

	err := validator.Validate(input)

	if !errors.Is(err, ErrTotalInvalid) {
		t.Fatalf(
			"expected ErrTotalInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRejectsMissingPaymentMethod(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Payment.Method = ""

	err := validator.Validate(input)

	if !errors.Is(err, ErrPaymentInvalid) {
		t.Fatalf(
			"expected ErrPaymentInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRejectsNegativePaidAmount(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Payment.Paid = -1

	err := validator.Validate(input)

	if !errors.Is(err, ErrPaymentAmountInvalid) {
		t.Fatalf(
			"expected ErrPaymentAmountInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRejectsInsufficientPayment(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Payment.Paid = 30000
	input.Payment.Change = 0

	err := validator.Validate(input)

	if !errors.Is(err, ErrChangeInvalid) {
		t.Fatalf(
			"expected ErrChangeInvalid, got %v",
			err,
		)
	}
}

func TestValidatorRejectsIncorrectChange(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Payment.Paid = 50000
	input.Payment.Change = 9999

	err := validator.Validate(input)

	if !errors.Is(err, ErrChangeInvalid) {
		t.Fatalf(
			"expected ErrChangeInvalid, got %v",
			err,
		)
	}
}

func TestValidatorAllowsCalculatedChange(t *testing.T) {
	validator := NewValidator()

	input := validReceipt()
	input.Payment.Change = 0

	err := validator.Validate(input)

	if err != nil {
		t.Fatalf(
			"expected change to be calculated automatically, got %v",
			err,
		)
	}
}

func TestValidatorItemCases(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		mutate  func(*domainreceipt.Receipt)
		wantErr error
	}{
		{
			name: "valid",
			mutate: func(r *domainreceipt.Receipt) {
			},
			wantErr: nil,
		},
		{
			name: "empty item name",
			mutate: func(r *domainreceipt.Receipt) {
				r.Items[0].Name = "   "
			},
			wantErr: ErrInvalidItem,
		},
		{
			name: "zero quantity",
			mutate: func(r *domainreceipt.Receipt) {
				r.Items[0].Quantity = 0
			},
			wantErr: ErrQuantityInvalid,
		},
		{
			name: "negative quantity",
			mutate: func(r *domainreceipt.Receipt) {
				r.Items[0].Quantity = -5
			},
			wantErr: ErrQuantityInvalid,
		},
		{
			name: "negative price",
			mutate: func(r *domainreceipt.Receipt) {
				r.Items[0].UnitPrice = -100
			},
			wantErr: ErrPriceInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validReceipt()

			tt.mutate(&input)

			err := validator.Validate(input)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf(
						"expected no error, got %v",
						err,
					)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected %v, got %v",
					tt.wantErr,
					err,
				)
			}
		})
	}
}
