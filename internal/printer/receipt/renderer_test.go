package receipt

import (
	"bytes"
	"testing"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
	"pos-system/internal/printer/escpos"
)

func TestRenderer(t *testing.T) {
	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-000001",
			InvoiceNumber: "INV-000001",
			Timestamp:     time.Date(2026, 8, 11, 20, 0, 0, 0, time.Local),
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

	renderer := NewRenderer()
	output := renderer.Render(input)

	if len(output) == 0 {
		t.Fatal("renderer returned empty output")
	}

	if !bytes.HasPrefix(output, escpos.Initialize()) {
		t.Fatal(
			"receipt does not start with ESC/POS initialize command",
		)
	}

	expectedTexts := []string{
		"TOKO KASA",
		"Jl. Contoh No. 123",
		"081234567890",
		"INV-000001",
		"11/08/2026 20:00:00",
		"Kasir: Bizar",
		"Kopi Susu",
		"Roti Bakar",
		"Rp30.000",
		"Rp12.000",
		"Rp42.000",
		"Rp50.000",
		"Rp8.000",
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

func TestRendererEmptyReceipt(t *testing.T) {
	renderer := NewRenderer()

	output := renderer.Render(
		domainreceipt.Receipt{},
	)

	if len(output) == 0 {
		t.Fatal("renderer returned empty output for empty receipt")
	}

	if !bytes.HasPrefix(output, escpos.Initialize()) {
		t.Fatal(
			"empty receipt does not start with ESC/POS initialize command",
		)
	}
}
