package receipt

import (
	"fmt"
	"strings"
)

const ReceiptWidth = 32

type Layout struct {
	Width int
}

func NewLayout(width int) Layout {
	if width <= 0 {
		width = ReceiptWidth
	}

	return Layout{
		Width: width,
	}
}

func (l Layout) Separator(char byte) string {
	return strings.Repeat(string(char), l.Width)
}

func (l Layout) Center(text string) string {
	text = truncate(text, l.Width)

	padding := l.Width - len(text)
	left := padding / 2
	right := padding - left

	return strings.Repeat(" ", left) +
		text +
		strings.Repeat(" ", right)
}

func (l Layout) LeftRight(left, right string) string {
	if l.Width <= 0 {
		return ""
	}

	// Right side alone is wider than the receipt.
	if len(right) >= l.Width {
		return truncate(right, l.Width)
	}

	// Left + right don't fit.
	if len(left)+len(right) >= l.Width {
		maxLeft := l.Width - len(right)

		if maxLeft <= 0 {
			return truncate(right, l.Width)
		}

		return truncate(left, maxLeft) + right
	}

	spaces := l.Width - len(left) - len(right)

	return left + strings.Repeat(" ", spaces) + right
}

func (l Layout) Item(item Item) []string {
	var lines []string

	// Product name.
	lines = append(
		lines,
		wrapText(item.Name, l.Width)...,
	)

	// Quantity × unit price.
	qtyPrice := fmt.Sprintf(
		"%d x Rp%s",
		item.Quantity,
		formatMoney(item.UnitPrice),
	)

	// Subtotal.
	total := "Rp" + formatMoney(item.Total())

	lines = append(
		lines,
		l.LeftRight(qtyPrice, total),
	)

	return lines
}

func (l Layout) Render(r Receipt) []string {
	var lines []string

	// ==============================
	// Header
	// ==============================

	if r.Store.Name != "" {
		lines = append(
			lines,
			l.Center(r.Store.Name),
		)
	}

	if r.Store.Address != "" {
		lines = append(
			lines,
			l.Center(r.Store.Address),
		)
	}

	if r.Store.Phone != "" {
		lines = append(
			lines,
			l.Center(r.Store.Phone),
		)
	}

	lines = append(lines, "")

	// ==============================
	// Transaction information
	// ==============================

	if r.Transaction.InvoiceNumber != "" {
		lines = append(
			lines,
			l.LeftRight(
				"Invoice : "+r.Transaction.InvoiceNumber,
				"",
			),
		)
	}

	if !r.Transaction.TimeStamp.IsZero() {
		lines = append(
			lines,
			l.LeftRight(
				"Date    : "+r.Transaction.TimeStamp.Format("02/01/2006 15:04"),
				"",
			),
		)
	}

	if r.Transaction.Cashier != "" {
		lines = append(
			lines,
			l.LeftRight(
				"Kasir   : "+r.Transaction.Cashier,
				"",
			),
		)
	}

	lines = append(lines, "")

	// ==============================
	// Items
	// ==============================

	lines = append(
		lines,
		l.Separator('-'),
	)

	for _, item := range r.Items {
		lines = append(
			lines,
			l.Item(item)...,
		)

		lines = append(lines, "")
	}

	// ==============================
	// Summary
	// ==============================

	lines = append(
		lines,
		l.Separator('-'),
	)

	subtotal := "Rp" + formatMoney(r.Subtotal())

	lines = append(
		lines,
		l.LeftRight(
			"Subtotal",
			subtotal,
		),
	)

	if r.Summary.Discount != 0 {
		lines = append(
			lines,
			l.LeftRight(
				"Discount",
				"Rp"+formatMoney(r.Summary.Discount),
			),
		)
	}

	if r.Summary.Tax != 0 {
		lines = append(
			lines,
			l.LeftRight(
				"Tax",
				"Rp"+formatMoney(r.Summary.Tax),
			),
		)
	}

	if r.Summary.ServiceCharge != 0 {
		lines = append(
			lines,
			l.LeftRight(
				"Service",
				"Rp"+formatMoney(r.Summary.ServiceCharge),
			),
		)
	}

	total := r.Summary.Total

	if total == 0 {
		total = r.Subtotal()
	}

	lines = append(
		lines,
		l.LeftRight(
			"TOTAL",
			"Rp"+formatMoney(total),
		),
	)

	// ==============================
	// Payment
	// ==============================

	lines = append(
		lines,
		l.Separator('-'),
	)

	if r.Payment.Method != "" {
		lines = append(
			lines,
			l.LeftRight(
				"Metode",
				r.Payment.Method,
			),
		)
	}

	lines = append(
		lines,
		l.LeftRight(
			"Bayar",
			"Rp"+formatMoney(r.Payment.Paid),
		),
	)

	change := r.Change()

	lines = append(
		lines,
		l.LeftRight(
			"Kembali",
			"Rp"+formatMoney(change),
		),
	)

	// ==============================
	// Footer
	// ==============================

	lines = append(
		lines,
		l.Separator('-'),
	)

	if r.Footer.Message != "" {
		lines = append(lines, "")
		lines = append(
			lines,
			l.Center(r.Footer.Message),
		)
	}

	return lines
}

func truncate(text string, width int) string {
	if width <= 0 {
		return ""
	}

	if len(text) <= width {
		return text
	}

	return text[:width]
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	text = strings.TrimSpace(text)

	if text == "" {
		return []string{""}
	}

	words := strings.Fields(text)

	var lines []string
	current := ""

	for _, word := range words {
		// Kalau satu kata lebih panjang dari lebar printer,
		// pecah menjadi beberapa bagian.
		if len(word) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}

			for len(word) > width {
				lines = append(lines, word[:width])
				word = word[width:]
			}

			if word != "" {
				current = word
			}

			continue
		}

		if current == "" {
			current = word
			continue
		}

		candidate := current + " " + word

		if len(candidate) <= width {
			current = candidate
			continue
		}

		lines = append(lines, current)
		current = word
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func formatMoney(amount int64) string {
	if amount == 0 {
		return "0"
	}

	negative := amount < 0

	if negative {
		amount = -amount
	}

	value := fmt.Sprintf("%d", amount)

	for i := len(value) - 3; i > 0; i -= 3 {
		value = value[:i] + "." + value[i:]
	}

	if negative {
		return "-" + value
	}

	return value
}
