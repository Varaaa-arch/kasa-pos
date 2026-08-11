package receipt

import (
	"bytes"
	"testing"
	"time"

	"pos-system/internal/printer/escpos"
)

func TestRenderer(t *testing.T) {
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

	input := Receipt{
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

	renderer := NewRenderer()
	output := renderer.Render(input)

	if len(output) == 0 {
		t.Fatal("renderer returned empty output")
	}

	if !bytes.HasPrefix(output, escpos.Initialize()) {
		t.Fatal("receipt does not start with ESC/POS initialize command")
	}

	expectedTexts := []string{
		"TOKO KASA",
		"Jl. Contoh No. 123",
		"081234567890",
		"INV-000001",
		"Kasir: Bizar",
		"Kopi Susu",
		"Roti Bakar",
		"Rp 30.000",
		"Rp 12.000",
		"Rp 42.000",
		"Rp 50.000",
		"Rp 8.000",
		"TERIMA KASIH",
	}

	for _, expected := range expectedTexts {
		if !bytes.Contains(output, []byte(expected)) {
			t.Errorf(
				"rendered receipt does not contain %q",
				expected,
			)
		}
	}
}
