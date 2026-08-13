package receipt

import (
	"fmt"
	"strings"

	domainreceipt "pos-system/internal/domain/receipt"
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

	if len(right) >= l.Width {
		return truncate(right, l.Width)
	}

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

func (l Layout) Item(item domainreceipt.Item) []string {
	var lines []string

	lines = append(
		lines,
		wrapText(item.Name, l.Width)...,
	)

	qtyPrice := fmt.Sprintf(
		"%d x Rp%s",
		item.Quantity,
		formatMoney(item.UnitPrice),
	)

	total := "Rp" + formatMoney(itemTotal(item))

	lines = append(
		lines,
		l.LeftRight(qtyPrice, total),
	)

	return lines
}

func (l Layout) Render(r domainreceipt.Receipt) []string {
	var lines []string

	// Header
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

	// Transaction
	if r.Transaction.InvoiceNumber != "" {
		lines = append(
			lines,
			l.LeftRight(
				"Invoice : "+r.Transaction.InvoiceNumber,
				"",
			),
		)
	}

	if !r.Transaction.Timestamp.IsZero() {
		lines = append(
			lines,
			l.LeftRight(
				"Date    : "+r.Transaction.Timestamp.Format(
					"02/01/2006 15:04",
				),
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

	// Items
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

	// Summary
	lines = append(
		lines,
		l.Separator('-'),
	)

	subtotal := calculateSubtotal(r)

	lines = append(
		lines,
		l.LeftRight(
			"Subtotal",
			"Rp"+formatMoney(subtotal),
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
		total = subtotal -
			r.Summary.Discount +
			r.Summary.Tax +
			r.Summary.ServiceCharge
	}

	lines = append(
		lines,
		l.LeftRight(
			"TOTAL",
			"Rp"+formatMoney(total),
		),
	)

	// Payment
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

	change := r.Payment.Change

	if change == 0 {
		change = r.Payment.Paid - total
	}

	lines = append(
		lines,
		l.LeftRight(
			"Kembali",
			"Rp"+formatMoney(change),
		),
	)

	// Footer
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

func itemTotal(item domainreceipt.Item) int64 {
	if item.Quantity <= 0 || item.UnitPrice <= 0 {
		return 0
	}

	return int64(item.Quantity) * item.UnitPrice
}

func calculateSubtotal(r domainreceipt.Receipt) int64 {
	var total int64

	for _, item := range r.Items {
		total += itemTotal(item)
	}

	return total
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
		if len(word) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}

			for len(word) > width {
				lines = append(
					lines,
					word[:width],
				)

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
