package receipt

import (
	"strings"
	"testing"

	domainreceipt "pos-system/internal/domain/receipt"
)

func TestItemRenderer(t *testing.T) {
	layout := NewLayout(32)
	renderer := NewItemRenderer(layout)

	item := domainreceipt.Item{
		Name:      "Kopi Susu",
		Quantity:  2,
		UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
	}

	lines := renderer.Render(item)

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
				"line %d:\n got:  %q\n want: %q",
				i,
				lines[i],
				expected[i],
			)
		}
	}
}

func TestItemRendererLongName(t *testing.T) {
	layout := NewLayout(32)
	renderer := NewItemRenderer(layout)

	item := domainreceipt.Item{
		Name:      "Kopi Susu Gula Aren Extra Large Premium",
		Quantity:  1,
		UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
	}

	lines := renderer.Render(item)

	if len(lines) < 3 {
		t.Fatalf(
			"expected long name to wrap, got %d lines: %#v",
			len(lines),
			lines,
		)
	}

	for i, line := range lines {
		if len(line) > layout.Width {
			t.Fatalf(
				"line %d exceeds width: %d > %d: %q",
				i,
				len(line),
				layout.Width,
				line,
			)
		}
	}

	productName := strings.Join(
		lines[:len(lines)-1],
		" ",
	)

	expectedName :=
		"Kopi Susu Gula Aren Extra Large Premium"

	if productName != expectedName {
		t.Fatalf(
			"product name changed during rendering:\n got:  %q\n want: %q",
			productName,
			expectedName,
		)
	}
}

func TestItemRendererLargePrice(t *testing.T) {
	layout := NewLayout(32)
	renderer := NewItemRenderer(layout)

	item := domainreceipt.Item{
		Name:      "Laptop Gaming",
		Quantity:  99,
		UnitPrice: domainreceipt.NewMoney(12500000, domainreceipt.IDR),
	}

	lines := renderer.Render(item)

	if len(lines) != 3 {
		t.Fatalf(
			"expected 3 lines, got %d: %#v",
			len(lines),
			lines,
		)
	}

	if !strings.Contains(
		lines[1],
		"99 x Rp12.500.000",
	) {
		t.Fatalf(
			"unit price missing: %q",
			lines[1],
		)
	}

	if !strings.Contains(
		lines[2],
		"Rp1.237.500.000",
	) {
		t.Fatalf(
			"subtotal missing: %q",
			lines[2],
		)
	}
}

func TestItemRendererLargeQuantity(t *testing.T) {
	layout := NewLayout(32)
	renderer := NewItemRenderer(layout)

	item := domainreceipt.Item{
		Name:      "Kopi Susu",
		Quantity:  100000,
		UnitPrice: domainreceipt.NewMoney(3000, domainreceipt.IDR),
	}

	lines := renderer.Render(item)

	if len(lines) != 2 {
		t.Fatalf(
			"expected 2 lines, got %d: %#v",
			len(lines),
			lines,
		)
	}

	if !strings.Contains(
		lines[1],
		"100000 x Rp3.000",
	) {
		t.Fatalf(
			"large quantity missing: %q",
			lines[1],
		)
	}

	if !strings.Contains(
		lines[1],
		"Rp300.000.000",
	) {
		t.Fatalf(
			"subtotal missing: %q",
			lines[1],
		)
	}

	if len(lines[1]) > layout.Width {
		t.Fatalf(
			"line exceeds width: %d > %d",
			len(lines[1]),
			layout.Width,
		)
	}
}

func TestItemRendererZeroQuantity(t *testing.T) {
	layout := NewLayout(32)
	renderer := NewItemRenderer(layout)

	item := domainreceipt.Item{
		Name:      "Kopi Susu",
		Quantity:  0,
		UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
	}

	lines := renderer.Render(item)

	if len(lines) < 2 {
		t.Fatalf(
			"expected at least 2 lines, got %d: %#v",
			len(lines),
			lines,
		)
	}

	if !strings.Contains(
		lines[len(lines)-1],
		"Rp0",
	) {
		t.Fatalf(
			"expected zero subtotal, got: %q",
			lines[len(lines)-1],
		)
	}
}

func TestItemRendererZeroPrice(t *testing.T) {
	layout := NewLayout(32)
	renderer := NewItemRenderer(layout)

	item := domainreceipt.Item{
		Name:      "Free Sample",
		Quantity:  1,
		UnitPrice: domainreceipt.NewMoney(0, domainreceipt.IDR),
	}

	lines := renderer.Render(item)

	if len(lines) < 2 {
		t.Fatalf(
			"expected at least 2 lines, got %d: %#v",
			len(lines),
			lines,
		)
	}

	if !strings.Contains(
		lines[len(lines)-1],
		"Rp0",
	) {
		t.Fatalf(
			"expected zero subtotal, got: %q",
			lines[len(lines)-1],
		)
	}
}
