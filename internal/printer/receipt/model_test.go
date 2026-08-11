package receipt

import (
	"testing"
	"time"
)

func TestReceiptModel(t *testing.T) {
	transactionTime := time.Date(
		2026,
		8,
		11,
		19,
		0,
		0,
		0,
		time.Local,
	)

	r := Receipt{
		Store: Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: Transaction{
			InvoiceNumber: "INV-000001",
			TimeStamp:     transactionTime,
			Cashier:       "Bizar",
		},

		Items: []Item{
			{
				Name:      "Kopi Susu",
				SKU:       "KOPI-001",
				Quantity:  2,
				UnitPrice: 15000,
				SubTotal:  30000,
			},
			{
				Name:      "Roti Bakar",
				SKU:       "ROTI-001",
				Quantity:  1,
				UnitPrice: 12000,
				SubTotal:  12000,
			},
		},

		Summary: Summary{
			SubTotal:      42000,
			Discount:      0,
			Tax:           0,
			ServiceCharge: 0,
			Total:         42000,
		},

		Payment: Payment{
			Method: "CASH",
			Paid:   50000,
			Change: 8000,
		},

		Footer: Footer{
			Message: "TERIMA KASIH",
		},
	}

	if r.Store.Name != "TOKO KASA" {
		t.Fatalf("unexpected store name: %q", r.Store.Name)
	}

	if r.Transaction.InvoiceNumber != "INV-000001" {
		t.Fatalf(
			"unexpected invoice number: %q",
			r.Transaction.InvoiceNumber,
		)
	}

	if len(r.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(r.Items))
	}

	if r.Items[0].SubTotal != 30000 {
		t.Fatalf(
			"expected first item subtotal 30000, got %d",
			r.Items[0].SubTotal,
		)
	}

	if r.Summary.Total != 42000 {
		t.Fatalf(
			"expected total 42000, got %d",
			r.Summary.Total,
		)
	}

	if r.Payment.Change != 8000 {
		t.Fatalf(
			"expected change 8000, got %d",
			r.Payment.Change,
		)
	}

	if r.Footer.Message != "TERIMA KASIH" {
		t.Fatalf(
			"unexpected footer message: %q",
			r.Footer.Message,
		)
	}
}
