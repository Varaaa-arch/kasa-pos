package receipt

import (
	"strings"
	"testing"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

func TestPreview(t *testing.T) {
	preview := NewPreview(32)

	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-PREVIEW-001",
			InvoiceNumber: "INV-PREVIEW-001",
			Timestamp: time.Date(
				2026,
				time.August,
				13,
				19,
				0,
				0,
				0,
				time.Local,
			),
			Cashier: "Bizar",
		},

		Items: []domainreceipt.Item{
			{
				ProductID: "PROD-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
			},
			{
				ProductID: "PROD-002",
				SKU:       "ROTI-001",
				Name:      "Roti Bakar",
				Quantity:  1,
				UnitPrice: domainreceipt.NewMoney(12000, domainreceipt.IDR),
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal: domainreceipt.NewMoney(42000, domainreceipt.IDR),
			Total: domainreceipt.NewMoney(42000, domainreceipt.IDR),
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid: domainreceipt.NewMoney(50000, domainreceipt.IDR),
			Change: domainreceipt.NewMoney(8000, domainreceipt.IDR),
		},

		Footer: domainreceipt.Footer{
			Message: "TERIMA KASIH",
		},
	}

	template := DefaultReceiptTemplate()

	preview.Render(
		input,
		template,
	)

	if len(preview.Lines) == 0 {
		t.Fatal("expected preview lines, got empty result")
	}

	for i, line := range preview.Lines {
		if len(line) > preview.Width {
			t.Fatalf(
				"line %d exceeds preview width: %d > %d: %q",
				i,
				len(line),
				preview.Width,
				line,
			)
		}
	}

	output := preview.String()

	expectedTexts := []string{
		"TOKO KASA",
		"Jl. Contoh No. 123",
		"081234567890",
		"INV-PREVIEW-001",
		"Bizar",
		"Kopi Susu",
		"Roti Bakar",
		"Rp30.000",
		"Rp12.000",
		"Rp42.000",
		"Rp50.000",
		"Rp8.000",
		"TERIMA KASIH",
	}

	for _, text := range expectedTexts {
		if !strings.Contains(output, text) {
			t.Fatalf(
				"preview missing %q:\n%s",
				text,
				output,
			)
		}
	}
}

func TestPreviewString(t *testing.T) {
	preview := NewPreview(20)

	preview.Lines = []string{
		"HELLO",
		"WORLD",
	}

	got := preview.String()

	expected := "HELLO\nWORLD"

	if got != expected {
		t.Fatalf(
			"unexpected preview string:\n got: %q\n want: %q",
			got,
			expected,
		)
	}
}

func TestPreviewRendersCompactTemplate(t *testing.T) {
	preview := NewPreview(32)

	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name: "TOKO KASA",
		},

		Footer: domainreceipt.Footer{
			Message: "TERIMA KASIH",
		},
	}

	preview.Render(
		input,
		CompactReceiptTemplate(),
	)

	output := preview.String()

	if strings.Contains(output, "TERIMA KASIH") {
		t.Fatalf(
			"compact preview should not contain footer:\n%s",
			output,
		)
	}

	if !strings.Contains(output, "TOKO KASA") {
		t.Fatalf(
			"compact preview missing header:\n%s",
			output,
		)
	}
}

func TestPreviewRenderResetsPreviousLines(t *testing.T) {
	preview := NewPreview(32)

	first := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name: "FIRST RECEIPT",
		},
	}

	second := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name: "SECOND RECEIPT",
		},
	}

	template := DefaultReceiptTemplate()

	preview.Render(first, template)

	if !strings.Contains(
		preview.String(),
		"FIRST RECEIPT",
	) {
		t.Fatal("first receipt was not rendered")
	}

	preview.Render(second, template)

	output := preview.String()

	if strings.Contains(
		output,
		"FIRST RECEIPT",
	) {
		t.Fatalf(
			"preview retained previous receipt:\n%s",
			output,
		)
	}

	if !strings.Contains(
		output,
		"SECOND RECEIPT",
	) {
		t.Fatalf(
			"preview missing second receipt:\n%s",
			output,
		)
	}
}
