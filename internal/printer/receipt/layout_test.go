package receipt

import (
	"strings"
	"testing"
	"time"
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

	item := Item{
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
	layout := NewLayout(20)

	item := Item{
		Name:      "Kopi Susu Gula Aren Extra Large",
		Quantity:  1,
		UnitPrice: 15000,
	}

	lines := layout.Item(item)

	if len(lines) < 3 {
		t.Fatalf(
			"expected long name to wrap, got %d lines: %v",
			len(lines),
			lines,
		)
	}

	for i, line := range lines {
		if len(line) > 20 {
			t.Fatalf(
				"line %d exceeds width: %d: %q",
				i,
				len(line),
				line,
			)
		}
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
	layout := NewLayout(48)

	r := Receipt{
		Store: Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: Transaction{
			InvoiceNumber: "INV-000001",
			TimeStamp:     time.Date(2026, 8, 11, 20, 0, 0, 0, time.Local),
			Cashier:       "Bizar",
		},

		Items: []Item{
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
				Quantity:  1,
				UnitPrice: 5000,
			},
		},

		Payment: Payment{
			Method: "CASH",
			Paid:   50000,
		},

		Footer: Footer{
			Message: "TERIMA KASIH",
		},
	}

	lines := layout.Render(r)

	if len(lines) == 0 {
		t.Fatal("expected receipt lines, got empty result")
	}

	for i, line := range lines {
		if len(line) > 48 {
			t.Fatalf(
				"line %d exceeds receipt width: %d characters\n%q",
				i,
				len(line),
				line,
			)
		}
	}

	receipt := strings.Join(lines, "\n")

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
		if !strings.Contains(receipt, text) {
			t.Fatalf(
				"receipt missing %q:\n%s",
				text,
				receipt,
			)
		}
	}
}

func TestEmptyReceipt(t *testing.T) {
	layout := NewLayout(48)

	r := Receipt{}

	lines := layout.Render(r)

	if len(lines) == 0 {
		t.Fatal("expected receipt layout, got empty result")
	}

	for i, line := range lines {
		if len(line) > 48 {
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

	item := Item{
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

	if len(lines[1]) > 32 {
		t.Fatalf(
			"item price line exceeds width: %d > 32",
			len(lines[1]),
		)
	}

}
