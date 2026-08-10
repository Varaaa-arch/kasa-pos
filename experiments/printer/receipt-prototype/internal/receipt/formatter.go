package receipt

import (
	"fmt"
	"strings"
)

const DefaultLineWidth = 48

func FormatText(r Receipt) string {
	var b strings.Builder

	// Header
	b.WriteString(centerText(r.Store.Name, DefaultLineWidth) + "\n")

	if r.Store.Address != "" {
		b.WriteString(centerText(r.Store.Address, DefaultLineWidth) + "\n")
	}

	if r.Store.Phone != "" {
		b.WriteString(centerText(r.Store.Phone, DefaultLineWidth) + "\n")
	}

	b.WriteString(separator(DefaultLineWidth, "=") + "\n")
	b.WriteString("\n")

	// Transaction Info
	if r.Transaction.InvoiceNumber != "" {
		b.WriteString("Invoice: " + r.Transaction.InvoiceNumber + "\n")
	}

	if !r.Transaction.TimeStamp.IsZero() {
		b.WriteString(
			r.Transaction.TimeStamp.Format("02/01/2006 15:04:05"),
		)
		b.WriteString("\n")
	}

	if r.Transaction.Cashier != "" {
		b.WriteString("Kasir: " + r.Transaction.Cashier + "\n")
	}

	b.WriteString("\n")

	// Items
	for _, item := range r.Items {
		b.WriteString(item.Name + "\n")

		itemLine := fmt.Sprintf(
			"%d x %s",
			item.Quantity,
			formatRupiah(item.UnitPrice),
		)

		b.WriteString(
			formatLeftRight(
				itemLine,
				formatRupiah(item.SubTotal),
				DefaultLineWidth,
			),
		)

		b.WriteString("\n")
	}

	// Summary
	b.WriteString(separator(DefaultLineWidth, "-"))
	b.WriteString("\n")

	b.WriteString(
		formatLeftRight(
			"Subtotal",
			formatRupiah(r.Summary.SubTotal),
			DefaultLineWidth,
		),
	)
	b.WriteString("\n")

	if r.Summary.Discount > 0 {
		b.WriteString(
			formatLeftRight(
				"Diskon",
				formatRupiah(r.Summary.Discount),
				DefaultLineWidth,
			),
		)
		b.WriteString("\n")
	}

	if r.Summary.Tax > 0 {
		b.WriteString(
			formatLeftRight(
				"Pajak",
				formatRupiah(r.Summary.Tax),
				DefaultLineWidth,
			),
		)
		b.WriteString("\n")
	}

	if r.Summary.ServiceCharge > 0 {
		b.WriteString(
			formatLeftRight(
				"Biaya Layanan",
				formatRupiah(r.Summary.ServiceCharge),
				DefaultLineWidth,
			),
		)
		b.WriteString("\n")
	}

	b.WriteString(
		formatLeftRight(
			"TOTAL",
			formatRupiah(r.Summary.Total),
			DefaultLineWidth,
		),
	)
	b.WriteString("\n")

	// Payment
	b.WriteString(separator(DefaultLineWidth, "-"))
	b.WriteString("\n")

	b.WriteString(
		formatLeftRight(
			"Bayar",
			formatRupiah(r.Payment.Paid),
			DefaultLineWidth,
		),
	)
	b.WriteString("\n")

	b.WriteString(
		formatLeftRight(
			"Kembali",
			formatRupiah(r.Payment.Change),
			DefaultLineWidth,
		),
	)
	b.WriteString("\n")

	// Footer
	b.WriteString(separator(DefaultLineWidth, "-"))
	b.WriteString("\n")

	if r.Footer.Message != "" {
		b.WriteString(
			centerText(r.Footer.Message, DefaultLineWidth),
		)
		b.WriteString("\n")
	}

	return b.String()
}

func separator(width int, character string) string {
	return strings.Repeat(character, width)
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}

	padding := (width - len(text)) / 2

	return strings.Repeat(" ", padding) + text
}

func formatLeftRight(left, right string, width int) string {
	if len(left)+len(right) >= width {
		return left + " " + right
	}

	spaces := width - len(left) - len(right)

	return left + strings.Repeat(" ", spaces) + right
}

func formatRupiah(amount int64) string {
	if amount == 0 {
		return "Rp 0"
	}

	sign := ""

	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	digits := []rune(fmt.Sprintf("%d", amount))
	var groups []string

	for len(digits) > 3 {
		groups = append(
			[]string{string(digits[len(digits)-3:])},
			groups...,
		)

		digits = digits[:len(digits)-3]
	}

	groups = append(
		[]string{string(digits)},
		groups...,
	)

	return sign + "Rp " + strings.Join(groups, ".")
}
