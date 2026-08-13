package receipt

import (
	"strings"
	"testing"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

func TestSeparator(t *testing.T) {
	layout := NewLayout(20)

	got := layout.Separator('-')
	expected := "--------------------"

	if got != expected {
		t.Fatalf(
			"unexpected separator:\n got:  %q\n want: %q",
			got,
			expected,
		)
	}
}

func TestCenter(t *testing.T) {
	layout := NewLayout(20)

	got := layout.Center("TOKO KASA")

	if len(got) != 20 {
		t.Fatalf(
			"expected line width 20, got %d",
			len(got),
		)
	}

	if !strings.Contains(got, "TOKO KASA") {
		t.Fatalf(
			"centered text missing: %q",
			got,
		)
	}
}

func TestLeftRight(t *testing.T) {
	layout := NewLayout(30)

	got := layout.LeftRight(
		"TOTAL",
		"Rp47.000",
	)

	if len(got) != 30 {
		t.Fatalf(
			"expected line width 30, got %d",
			len(got),
		)
	}

	if !strings.HasPrefix(got, "TOTAL") {
		t.Fatalf("left text missing: %q", got)
	}

	if !strings.HasSuffix(got, "Rp47.000") {
		t.Fatalf("right text missing: %q", got)
	}
}

func TestItemLayout(t *testing.T) {
	layout := NewLayout(32)

	item := domainreceipt.Item{
		Name:      "Kopi Susu",
		Quantity:  2,
		UnitPrice: 15000,
	}

	lines := layout.Item(item)

	expected := []string{
		"Kopi Susu",
		"2 x Rp15.000            Rp30.000",
	}

	if len(lines) != len(expected) {
		t.Fatalf(
			"expected %d lines, got %d: %#v",
			len(expected),
			len(lines),
			lines,
		)
	}

	for i := range expected {
		if lines[i] != expected[i] {
			t.Fatalf(
				"line %d: expected %q, got %q",
				i,
				expected[i],
				lines[i],
			)
		}

		if len(lines[i]) > 32 {
			t.Fatalf(
				"line %d exceeds printer width: %d > 32",
				i,
				len(lines[i]),
			)
		}
	}
}

func TestItemLayoutLongName(t *testing.T) {
	layout := NewLayout(32)

	item := domainreceipt.Item{
		Name:      "Kopi Susu Gula Aren Extra Large Premium",
		Quantity:  1,
		UnitPrice: 15000,
	}

	lines := layout.Item(item)

	if len(lines) < 3 {
		t.Fatalf(
			"expected long product name to wrap into multiple lines, got %d lines: %v",
			len(lines),
			lines,
		)
	}

	nameLines := lines[:len(lines)-1]

	if len(nameLines) < 2 {
		t.Fatalf(
			"expected product name to wrap, got: %v",
			nameLines,
		)
	}

	for i, line := range lines {
		if len(line) > layout.Width {
			t.Fatalf(
				"line %d exceeds printer width: %d > %d: %q",
				i,
				len(line),
				layout.Width,
				line,
			)
		}
	}

	priceLine := lines[len(lines)-1]

	if !strings.Contains(priceLine, "1 x Rp15.000") {
		t.Fatalf(
			"quantity/unit price missing: %q",
			priceLine,
		)
	}

	if !strings.Contains(priceLine, "Rp15.000") {
		t.Fatalf(
			"subtotal missing: %q",
			priceLine,
		)
	}

	joined := strings.Join(nameLines, " ")

	expectedName := "Kopi Susu Gula Aren Extra Large Premium"

	if joined != expectedName {
		t.Fatalf(
			"product name was changed during wrapping:\n got:  %q\n want: %q",
			joined,
			expectedName,
		)
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  []string
	}{
		{
			name:  "short text",
			text:  "Kopi Susu",
			width: 32,
			want:  []string{"Kopi Susu"},
		},
		{
			name:  "exact width",
			text:  "12345678901234567890123456789012",
			width: 32,
			want:  []string{"12345678901234567890123456789012"},
		},
		{
			name:  "long text",
			text:  "Ini adalah nama produk yang sangat panjang",
			width: 32,
			want: []string{
				"Ini adalah nama produk yang",
				"sangat panjang",
			},
		},
		{
			name:  "long word",
			text:  "123456789012345678901234567890123",
			width: 32,
			want: []string{
				"12345678901234567890123456789012",
				"3",
			},
		},
		{
			name:  "empty text",
			text:  "",
			width: 32,
			want:  []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.width)

			if len(got) != len(tt.want) {
				t.Fatalf(
					"expected %d lines, got %d: %#v",
					len(tt.want),
					len(got),
					got,
				)
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf(
						"line %d: expected %q, got %q",
						i,
						tt.want[i],
						got[i],
					)
				}
			}
		})
	}
}

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "thousand",
			input:    1000,
			expected: "1.000",
		},
		{
			name:     "ten thousand",
			input:    15000,
			expected: "15.000",
		},
		{
			name:     "hundred thousand",
			input:    150000,
			expected: "150.000",
		},
		{
			name:     "million",
			input:    1500000,
			expected: "1.500.000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMoney(tt.input)

			if got != tt.expected {
				t.Fatalf(
					"got %q, want %q",
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestReceiptRender(t *testing.T) {
	layout := NewLayout(32)

	r := domainreceipt.Receipt{
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
			{
				ProductID: "PROD-003",
				SKU:       "AIR-001",
				Name:      "Air Mineral",
				Quantity:  1,
				UnitPrice: 5000,
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal: 47000,
			Total:    47000,
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid:   50000,
			Change: 3000,
		},

		Footer: domainreceipt.Footer{
			Message: "TERIMA KASIH",
		},
	}

	lines := layout.Render(r)

	if len(lines) == 0 {
		t.Fatal("expected receipt lines, got empty result")
	}

	for i, line := range lines {
		if len(line) > layout.Width {
			t.Fatalf(
				"line %d exceeds receipt width: %d characters\n%q",
				i,
				len(line),
				line,
			)
		}
	}

	output := strings.Join(lines, "\n")

	expectedTexts := []string{
		"TOKO KASA",
		"Jl. Contoh No. 123",
		"081234567890",
		"INV-000001",
		"Bizar",
		"Kopi Susu",
		"Roti Bakar",
		"Air Mineral",
		"Subtotal",
		"TOTAL",
		"Rp47.000",
		"CASH",
		"Rp50.000",
		"Rp3.000",
		"TERIMA KASIH",
	}

	for _, text := range expectedTexts {
		if !strings.Contains(output, text) {
			t.Fatalf(
				"receipt missing %q:\n%s",
				text,
				output,
			)
		}
	}
}

func TestEmptyReceipt(t *testing.T) {
	layout := NewLayout(32)

	r := domainreceipt.Receipt{}

	lines := layout.Render(r)

	if len(lines) == 0 {
		t.Fatal("expected receipt layout, got empty result")
	}

	for i, line := range lines {
		if len(line) > layout.Width {
			t.Fatalf(
				"line %d exceeds width: %q",
				i,
				line,
			)
		}
	}
}

func TestItemLayoutLargePrice(t *testing.T) {
	layout := NewLayout(32)

	item := domainreceipt.Item{
		Name:      "Laptop Gaming",
		Quantity:  99,
		UnitPrice: 12500000,
	}

	lines := layout.Item(item)

	if len(lines) != 2 {
		t.Fatalf(
			"expected 2 lines, got %d: %#v",
			len(lines),
			lines,
		)
	}

	if len(lines[1]) > layout.Width {
		t.Fatalf(
			"item price line exceeds width: %d > %d",
			len(lines[1]),
			layout.Width,
		)
	}
}

func TestManyItems(t *testing.T) {
	layout := NewLayout(32)

	items := []domainreceipt.Item{
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
		{
			Name:      "Air Mineral",
			Quantity:  3,
			UnitPrice: 5000,
		},
		{
			Name:      "Nasi Goreng Spesial",
			Quantity:  2,
			UnitPrice: 25000,
		},
		{
			Name:      "Es Teh Manis",
			Quantity:  4,
			UnitPrice: 5000,
		},
		{
			Name:      "Kentang Goreng",
			Quantity:  2,
			UnitPrice: 10000,
		},
		{
			Name:      "Ayam Geprek",
			Quantity:  3,
			UnitPrice: 18000,
		},
		{
			Name:      "Mie Goreng",
			Quantity:  2,
			UnitPrice: 15000,
		},
		{
			Name:      "Jus Alpukat",
			Quantity:  1,
			UnitPrice: 12000,
		},
		{
			Name:      "Pisang Coklat Keju",
			Quantity:  2,
			UnitPrice: 13000,
		},
	}

	r := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name: "TOKO KASA",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-MANY-001",
			InvoiceNumber: "INV-MANY-001",
			Timestamp:     time.Now(),
			Cashier:       "Bizar",
		},

		Items: items,

		Summary: domainreceipt.Summary{
			Subtotal: 269000,
			Total:    269000,
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid:   300000,
			Change: 31000,
		},

		Footer: domainreceipt.Footer{
			Message: "TERIMA KASIH",
		},
	}

	lines := layout.Render(r)

	if len(lines) == 0 {
		t.Fatal("expected receipt lines, got empty result")
	}

	for i, line := range lines {
		if len(line) > layout.Width {
			t.Fatalf(
				"line %d exceeds receipt width: %d > %d: %q",
				i,
				len(line),
				layout.Width,
				line,
			)
		}
	}

	output := strings.Join(lines, "\n")

	for _, item := range items {
		if !strings.Contains(output, item.Name) {
			t.Fatalf(
				"receipt missing item %q:\n%s",
				item.Name,
				output,
			)
		}
	}

	if !strings.Contains(output, "Rp269.000") {
		t.Fatalf(
			"receipt missing subtotal/total Rp269.000:\n%s",
			output,
		)
	}
}

func TestLargePrices(t *testing.T) {
	layout := NewLayout(32)

	tests := []struct {
		name      string
		itemName  string
		quantity  int
		unitPrice int64
	}{
		{
			name:      "one million",
			itemName:  "Laptop",
			quantity:  1,
			unitPrice: 1_000_000,
		},
		{
			name:      "ten million",
			itemName:  "Smartphone",
			quantity:  2,
			unitPrice: 10_000_000,
		},
		{
			name:      "one hundred million",
			itemName:  "Server Rack",
			quantity:  1,
			unitPrice: 100_000_000,
		},
		{
			name:      "large quantity",
			itemName:  "Kursi Kantor",
			quantity:  999,
			unitPrice: 999_999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := domainreceipt.Item{
				Name:      tt.itemName,
				Quantity:  tt.quantity,
				UnitPrice: tt.unitPrice,
			}

			lines := layout.Item(item)

			if len(lines) < 2 {
				t.Fatalf(
					"expected at least 2 lines, got %d: %#v",
					len(lines),
					lines,
				)
			}

			for i, line := range lines {
				if len(line) > layout.Width {
					t.Fatalf(
						"line %d exceeds printer width: %d > %d: %q",
						i,
						len(line),
						layout.Width,
						line,
					)
				}
			}

			priceLine := lines[len(lines)-1]

			expectedUnitPrice := "Rp" + formatMoney(tt.unitPrice)

			if !strings.Contains(priceLine, expectedUnitPrice) {
				t.Fatalf(
					"unit price missing:\n got: %q\n want to contain: %q",
					priceLine,
					expectedUnitPrice,
				)
			}

			expectedTotal := "Rp" + formatMoney(itemTotal(item))

			if !strings.Contains(priceLine, expectedTotal) {
				t.Fatalf(
					"subtotal missing:\n got: %q\n want to contain: %q",
					priceLine,
					expectedTotal,
				)
			}
		})
	}
}

func TestLargeQuantities(t *testing.T) {
	layout := NewLayout(32)

	tests := []struct {
		name     string
		itemName string
		quantity int
		price    int64
	}{
		{
			name:     "quantity 99",
			itemName: "Kopi Susu",
			quantity: 99,
			price:    15000,
		},
		{
			name:     "quantity 999",
			itemName: "Air Mineral",
			quantity: 999,
			price:    5000,
		},
		{
			name:     "quantity 9999",
			itemName: "Roti Bakar",
			quantity: 9999,
			price:    12000,
		},
		{
			name:     "quantity 100000",
			itemName: "Pulpen",
			quantity: 100000,
			price:    3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := domainreceipt.Item{
				Name:      tt.itemName,
				Quantity:  tt.quantity,
				UnitPrice: tt.price,
			}

			lines := layout.Item(item)

			if len(lines) < 2 {
				t.Fatalf(
					"expected at least 2 lines, got %d: %#v",
					len(lines),
					lines,
				)
			}

			for i, line := range lines {
				if len(line) > layout.Width {
					t.Fatalf(
						"line %d exceeds printer width: %d > %d: %q",
						i,
						len(line),
						layout.Width,
						line,
					)
				}
			}

			priceLine := lines[len(lines)-1]

			expectedQuantity := strings.TrimSpace(
				strings.Join(
					[]string{
						formatQuantity(tt.quantity),
						"x Rp" + formatMoney(tt.price),
					},
					" ",
				),
			)

			if !strings.Contains(priceLine, expectedQuantity) {
				t.Fatalf(
					"quantity/price missing:\n got: %q\n want to contain: %q",
					priceLine,
					expectedQuantity,
				)
			}

			expectedTotal := "Rp" + formatMoney(itemTotal(item))

			if !strings.Contains(priceLine, expectedTotal) {
				t.Fatalf(
					"subtotal missing:\n got: %q\n want to contain: %q",
					priceLine,
					expectedTotal,
				)
			}
		})
	}
}

func TestReceiptWithoutItems(t *testing.T) {
	layout := NewLayout(32)

	r := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-EMPTY-001",
			InvoiceNumber: "INV-EMPTY-001",
			Timestamp:     time.Now(),
			Cashier:       "Bizar",
		},

		Items: []domainreceipt.Item{},

		Summary: domainreceipt.Summary{
			Subtotal: 0,
			Total:    0,
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid:   0,
			Change: 0,
		},

		Footer: domainreceipt.Footer{
			Message: "TERIMA KASIH",
		},
	}

	lines := layout.Render(r)

	if len(lines) == 0 {
		t.Fatal("expected receipt layout, got empty result")
	}

	for i, line := range lines {
		if len(line) > layout.Width {
			t.Fatalf(
				"line %d exceeds printer width: %d > %d: %q",
				i,
				len(line),
				layout.Width,
				line,
			)
		}
	}

	output := strings.Join(lines, "\n")

	expectedTexts := []string{
		"TOKO KASA",
		"INV-EMPTY-001",
		"Bizar",
		"Subtotal",
		"TOTAL",
		"CASH",
		"Bayar",
		"Kembali",
		"TERIMA KASIH",
	}

	for _, expected := range expectedTexts {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"receipt missing %q:\n%s",
				expected,
				output,
			)
		}
	}

	if strings.Contains(output, "x Rp") {
		t.Fatalf(
			"receipt unexpectedly contains item pricing:\n%s",
			output,
		)
	}
}

func formatQuantity(quantity int) string {
	return strings.TrimSpace(
		strings.TrimPrefix(
			strings.TrimSpace(
				strings.ReplaceAll(
					strings.TrimSpace(
						" "+strings.TrimSpace(
							strings.TrimLeft(
								strings.Repeat(" ", 0),
								" ",
							),
						),
					),
					"  ",
					" ",
				),
			),
			"",
		),
	)
}
