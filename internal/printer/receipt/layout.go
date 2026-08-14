package receipt

import (
	"fmt"
	"strings"

	domainreceipt "pos-system/internal/domain/receipt"
)

const (
	DefaultReceiptWidth = 32
	DefaultLeftMargin   = 0
	DefaultRightMargin  = 0
)

type Layout struct {
	Width       int
	LeftMargin  int
	RightMargin int
	Calculator  *Calculator
}

func NewLayout(width int) Layout {
	if width <= 0 {
		width = DefaultReceiptWidth
	}

	return Layout{
		Width:       width,
		LeftMargin:  DefaultLeftMargin,
		RightMargin: DefaultRightMargin,
		Calculator:  NewCalculator(),
	}
}

func NewLayoutWithMargins(
	width int,
	leftMargin int,
	rightMargin int,
) Layout {
	if width <= 0 {
		width = DefaultReceiptWidth
	}

	if leftMargin < 0 {
		leftMargin = 0
	}

	if rightMargin < 0 {
		rightMargin = 0
	}

	if leftMargin+rightMargin >= width {
		leftMargin = 0
		rightMargin = 0
	}

	return Layout{
		Width:       width,
		LeftMargin:  leftMargin,
		RightMargin: rightMargin,
		Calculator:  NewCalculator(),
	}
}

func (l Layout) ContentWidth() int {
	width := l.Width -
		l.LeftMargin -
		l.RightMargin

	if width < 0 {
		return 0
	}

	return width
}
func (l Layout) Separator(char byte) string {
	if l.Width <= 0 {
		return ""
	}

	return strings.Repeat(string(char), l.Width)
}

func (l Layout) withMargins(content string) string {
	width := l.ContentWidth()

	content = truncate(content, width)

	return strings.Repeat(" ", l.LeftMargin) +
		content +
		strings.Repeat(
			" ",
			l.RightMargin+(width-len(content)),
		)
}

func (l Layout) Center(text string) string {
	width := l.ContentWidth()

	text = truncate(text, width)

	padding := width - len(text)

	left := padding / 2
	right := padding - left

	content :=
		strings.Repeat(" ", left) +
			text +
			strings.Repeat(" ", right)

	return l.withMargins(content)
}

func (l Layout) Left(text string) string {
	return l.withMargins(text)
}

func (l Layout) Right(text string) string {
	width := l.ContentWidth()

	text = truncate(text, width)

	return l.withMargins(
		padLeft(text, width),
	)
}

func (l Layout) LeftRight(left, right string) string {
	width := l.ContentWidth()

	if width <= 0 {
		return ""
	}

	left = truncate(left, width)
	right = truncate(right, width)

	if len(right) >= width {
		return l.withMargins(right)
	}

	if len(left)+len(right) >= width {
		maxLeft := width - len(right)

		if maxLeft <= 0 {
			return l.withMargins(right)
		}

		left = truncate(left, maxLeft)

		return l.withMargins(left + right)
	}

	spaces := width - len(left) - len(right)

	content :=
		left +
			strings.Repeat(" ", spaces) +
			right

	return l.withMargins(content)
}

func (l Layout) Item(item domainreceipt.Item) []string {
	renderer := NewItemRenderer(l)

	return renderer.Render(item)
}
func (l Layout) Render(input domainreceipt.Receipt) []string {
	var lines []string

	calculation := l.Calculator.Calculate(input)

	// =========================
	// HEADER
	// =========================

	if input.Store.Name != "" {
		lines = append(
			lines,
			l.Center(input.Store.Name),
		)
	}

	if input.Store.Address != "" {
		for _, line := range wrapText(
			input.Store.Address,
			l.ContentWidth(),
		) {
			lines = append(
				lines,
				l.Center(line),
			)
		}
	}

	if input.Store.Phone != "" {
		lines = append(
			lines,
			l.Center(input.Store.Phone),
		)
	}

	lines = append(
		lines,
		l.BlankLine(),
		l.SeparatorLine('='),
	)

	// =========================
	// TRANSACTION
	// =========================

	if input.Transaction.InvoiceNumber != "" {
		lines = append(
			lines,
			l.Left(
				"Invoice: "+input.Transaction.InvoiceNumber,
			),
		)
	}

	if !input.Transaction.Timestamp.IsZero() {
		lines = append(
			lines,
			l.Left(
				input.Transaction.Timestamp.Format(
					"02/01/2006 15:04:05",
				),
			),
		)
	}

	if input.Transaction.Cashier != "" {
		lines = append(
			lines,
			l.Left(
				"Kasir: "+input.Transaction.Cashier,
			),
		)
	}

	lines = append(
		lines,
		l.SeparatorLine('-'),
	)

	// =========================
	// ITEMS
	// =========================

	for _, item := range input.Items {
		lines = append(
			lines,
			l.Item(item)...,
		)

		lines = append(
			lines,
			l.BlankLine(),
		)
	}

	// =========================
	// SUMMARY
	// =========================

	lines = append(
		lines,
		l.SeparatorLine('-'),
	)

	lines = append(
		lines,
		l.LeftRight(
			"Subtotal",
			"Rp"+formatMoney(calculation.Subtotal),
		),
	)

	if input.Summary.Discount != 0 {
		lines = append(
			lines,
			l.LeftRight(
				"Diskon",
				"Rp"+formatMoney(input.Summary.Discount),
			),
		)
	}

	if input.Summary.Tax != 0 {
		lines = append(
			lines,
			l.LeftRight(
				"Pajak",
				"Rp"+formatMoney(input.Summary.Tax),
			),
		)
	}

	if input.Summary.ServiceCharge != 0 {
		lines = append(
			lines,
			l.LeftRight(
				"Biaya Layanan",
				"Rp"+formatMoney(
					input.Summary.ServiceCharge,
				),
			),
		)
	}

	total := calculation.Total

	lines = append(
		lines,
		l.LeftRight(
			"TOTAL",
			"Rp"+formatMoney(total),
		),
	)

	// =========================
	// PAYMENT
	// =========================

	lines = append(
		lines,
		l.SeparatorLine('-'),
	)

	if input.Payment.Method != "" {
		lines = append(
			lines,
			l.LeftRight(
				"Metode",
				input.Payment.Method,
			),
		)
	}

	lines = append(
		lines,
		l.LeftRight(
			"Bayar",
			"Rp"+formatMoney(calculation.Paid),
		),
	)

	change := calculation.Change

	lines = append(
		lines,
		l.LeftRight(
			"Kembali",
			"Rp"+formatMoney(change),
		),
	)

	// =========================
	// FOOTER
	// =========================

	lines = append(
		lines,
		l.SeparatorLine('-'),
		l.BlankLine(),
	)

	if input.Footer.Message != "" {
		for _, line := range wrapText(
			input.Footer.Message,
			l.ContentWidth(),
		) {
			lines = append(
				lines,
				l.Center(line),
			)
		}
	}

	return lines
}

func (l Layout) SeparatorLine(char byte) string {
	width := l.ContentWidth()

	line := strings.Repeat(
		string(char),
		width,
	)

	return l.withMargins(line)
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

func padLeft(text string, width int) string {
	if width <= 0 {
		return ""
	}

	if len(text) >= width {
		return truncate(text, width)
	}

	return strings.Repeat(
		" ",
		width-len(text),
	) + text
}

func padRight(text string, width int) string {
	if width <= 0 {
		return ""
	}

	if len(text) >= width {
		return truncate(text, width)
	}

	return text +
		strings.Repeat(
			" ",
			width-len(text),
		)
}

func (l Layout) Padding(left, right int, content string) string {
	if left < 0 {
		left = 0
	}

	if right < 0 {
		right = 0
	}

	width := l.ContentWidth()

	if left+right >= width {
		return l.withMargins(
			truncate(content, width),
		)
	}

	available := width - left - right
	content = truncate(content, available)

	return l.withMargins(
		strings.Repeat(" ", left) +
			content +
			strings.Repeat(" ", right),
	)
}

func (l Layout) BlankLine() string {
	return l.withMargins("")
}
