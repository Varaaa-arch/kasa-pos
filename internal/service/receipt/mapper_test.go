package receipt

import (
	"testing"
	"time"

	domaintransaction "pos-system/internal/domain/transaction"
)

func TestFromTransactionDoesNotRecalculate(t *testing.T) {
	tx := domaintransaction.Transaction{
		ID:            "trx-001",
		InvoiceNumber: "INV-0001",
		CreatedAt:     time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC),

		Subtotal:      50000,
		Discount:      5000,
		Tax:           4500,
		ServiceCharge: 1000,
		Total:         50500,

		PaidAmount:    60000,
		Change:        9500,
		PaymentMethod: "CASH",

		Items: []domaintransaction.Item{
			{
				ID:            "item-001",
				TransactionID: "trx-001",
				ProductID:     "prod-001",
				SKU:           "KOPI-001",
				Name:          "Kopi Susu",
				Quantity:      2,
				UnitPrice:     25000,
				Subtotal:      50000,
			},
		},
	}

	receipt := FromTransaction(tx)

	if receipt.Transaction.InvoiceNumber != tx.InvoiceNumber {
		t.Fatalf("invoice mismatch")
	}

	if receipt.Summary.Subtotal.Amount != 50000 {
		t.Fatalf(
			"subtotal = %d, want 50000",
			receipt.Summary.Subtotal.Amount,
		)
	}

	if receipt.Summary.Total.Amount != 50500 {
		t.Fatalf(
			"total = %d, want 50500",
			receipt.Summary.Total.Amount,
		)
	}

	if receipt.Payment.Paid.Amount != 60000 {
		t.Fatalf(
			"paid = %d, want 60000",
			receipt.Payment.Paid.Amount,
		)
	}

	if receipt.Payment.Change.Amount != 9500 {
		t.Fatalf(
			"change = %d, want 9500",
			receipt.Payment.Change.Amount,
		)
	}
}
