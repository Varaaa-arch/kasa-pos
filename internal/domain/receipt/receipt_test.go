package receipt

import (
	"testing"
	"time"
)

func TestReceiptDomainModel(t *testing.T) {
	now := time.Date(
		2026,
		time.August,
		13,
		16,
		30,
		0,
		0,
		time.UTC,
	)

	r := Receipt{
		Store: Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: Transaction{
			ID:            "TXN-000001",
			InvoiceNumber: "INV-000001",
			Timestamp:     now,
			Cashier:       "Bizar",
		},

		Customer: Customer{
			ID:    "CUS-000001",
			Name:  "Customer Test",
			Phone: "081234567890",
		},

		Items: []Item{
			{
				ProductID: "PROD-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
			},
		},

		Summary: Summary{
			Subtotal:      30000,
			Discount:      0,
			Tax:           0,
			ServiceCharge: 0,
			Total:         30000,
		},

		Payment: Payment{
			Method: "CASH",
			Paid:   50000,
			Change: 20000,
		},

		Footer: Footer{
			Message: "TERIMA KASIH",
		},
	}

	if r.Store.Name != "TOKO KASA" {
		t.Fatalf(
			"unexpected store name: %q",
			r.Store.Name,
		)
	}

	if r.Transaction.ID != "TXN-000001" {
		t.Fatalf(
			"unexpected transaction ID: %q",
			r.Transaction.ID,
		)
	}

	if r.Transaction.InvoiceNumber != "INV-000001" {
		t.Fatalf(
			"unexpected invoice number: %q",
			r.Transaction.InvoiceNumber,
		)
	}

	if r.Customer.Name != "Customer Test" {
		t.Fatalf(
			"unexpected customer name: %q",
			r.Customer.Name,
		)
	}

	if len(r.Items) != 1 {
		t.Fatalf(
			"expected 1 item, got %d",
			len(r.Items),
		)
	}

	if r.Items[0].Quantity != 2 {
		t.Fatalf(
			"expected quantity 2, got %d",
			r.Items[0].Quantity,
		)
	}

	if r.Items[0].UnitPrice != 15000 {
		t.Fatalf(
			"expected unit price 15000, got %d",
			r.Items[0].UnitPrice,
		)
	}

	if r.Summary.Total != 30000 {
		t.Fatalf(
			"expected total 30000, got %d",
			r.Summary.Total,
		)
	}

	if r.Payment.Paid != 50000 {
		t.Fatalf(
			"expected paid amount 50000, got %d",
			r.Payment.Paid,
		)
	}

	if r.Footer.Message != "TERIMA KASIH" {
		t.Fatalf(
			"unexpected footer message: %q",
			r.Footer.Message,
		)
	}
}
