package receipt

import (
	"testing"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

func TestBuildSections(t *testing.T) {
	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-000001",
			InvoiceNumber: "INV-000001",
			Timestamp: time.Date(
				2026,
				time.August,
				13,
				18,
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

	sections := BuildSections(input)

	if len(sections) != 6 {
		t.Fatalf(
			"expected 6 sections, got %d",
			len(sections),
		)
	}

	expectedTypes := []SectionType{
		SectionHeader,
		SectionTransaction,
		SectionItems,
		SectionSummary,
		SectionPayment,
		SectionFooter,
	}

	for i, expected := range expectedTypes {
		if sections[i].Type != expected {
			t.Fatalf(
				"section %d: expected %q, got %q",
				i,
				expected,
				sections[i].Type,
			)
		}
	}
}

func TestBuildHeaderSection(t *testing.T) {
	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},
	}

	section := buildHeaderSection(input)

	if section.Type != SectionHeader {
		t.Fatalf(
			"expected HEADER, got %q",
			section.Type,
		)
	}

	if len(section.Lines) != 3 {
		t.Fatalf(
			"expected 3 lines, got %d",
			len(section.Lines),
		)
	}

	if section.Lines[0] != "TOKO KASA" {
		t.Fatalf(
			"unexpected store name: %q",
			section.Lines[0],
		)
	}
}

func TestBuildTransactionSection(t *testing.T) {
	input := domainreceipt.Receipt{
		Transaction: domainreceipt.Transaction{
			InvoiceNumber: "INV-000001",
			Timestamp: time.Date(
				2026,
				time.August,
				13,
				18,
				0,
				0,
				0,
				time.Local,
			),
			Cashier: "Bizar",
		},
	}

	section := buildTransactionSection(input)

	if section.Type != SectionTransaction {
		t.Fatalf(
			"expected TRANSACTION, got %q",
			section.Type,
		)
	}

	if len(section.Lines) != 3 {
		t.Fatalf(
			"expected 3 lines, got %d",
			len(section.Lines),
		)
	}
}

func TestBuildItemsSection(t *testing.T) {
	input := domainreceipt.Receipt{
		Items: []domainreceipt.Item{
			{
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
			},
			{
				Name:      "Roti Bakar",
				Quantity:  1,
				UnitPrice: 12000,
			},
		},
	}

	section := buildItemsSection(input)

	if section.Type != SectionItems {
		t.Fatalf(
			"expected ITEMS, got %q",
			section.Type,
		)
	}

	if len(section.Lines) == 0 {
		t.Fatal("expected item section to contain lines")
	}
}

func TestBuildSummarySection(t *testing.T) {
	input := domainreceipt.Receipt{
		Items: []domainreceipt.Item{
			{
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal: 30000,
			Total:    30000,
		},
	}

	section := buildSummarySection(input)

	if section.Type != SectionSummary {
		t.Fatalf(
			"expected SUMMARY, got %q",
			section.Type,
		)
	}

	if len(section.Lines) != 2 {
		t.Fatalf(
			"expected 2 summary lines, got %d",
			len(section.Lines),
		)
	}
}

func TestBuildPaymentSection(t *testing.T) {
	input := domainreceipt.Receipt{
		Summary: domainreceipt.Summary{
			Total: 30000,
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid:   50000,
			Change: 20000,
		},
	}

	section := buildPaymentSection(input)

	if section.Type != SectionPayment {
		t.Fatalf(
			"expected PAYMENT, got %q",
			section.Type,
		)
	}

	if len(section.Lines) != 3 {
		t.Fatalf(
			"expected 3 payment lines, got %d",
			len(section.Lines),
		)
	}
}

func TestBuildFooterSection(t *testing.T) {
	input := domainreceipt.Receipt{
		Footer: domainreceipt.Footer{
			Message: "TERIMA KASIH",
		},
	}

	section := buildFooterSection(input)

	if section.Type != SectionFooter {
		t.Fatalf(
			"expected FOOTER, got %q",
			section.Type,
		)
	}

	if len(section.Lines) != 1 {
		t.Fatalf(
			"expected 1 footer line, got %d",
			len(section.Lines),
		)
	}

	if section.Lines[0] != "TERIMA KASIH" {
		t.Fatalf(
			"unexpected footer: %q",
			section.Lines[0],
		)
	}
}
