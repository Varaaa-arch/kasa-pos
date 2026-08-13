package receipt

import (
	"fmt"
	"strings"

	domainreceipt "pos-system/internal/domain/receipt"
	"pos-system/internal/printer/escpos"
)

const DefaultLineWidth = 32

type Renderer struct {
	Width int
}

func NewRenderer() *Renderer {
	return &Renderer{
		Width: DefaultLineWidth,
	}
}

func (r *Renderer) Render(input domainreceipt.Receipt) []byte {
	var data []byte

	data = append(data, escpos.Initialize()...)

	// Header
	data = append(data, escpos.AlignCenter()...)
	data = append(data, escpos.Bold(true)...)
	data = append(data, escpos.Text(input.Store.Name)...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.Bold(false)...)

	if input.Store.Address != "" {
		data = append(data, escpos.Text(input.Store.Address)...)
		data = append(data, escpos.LF()...)
	}

	if input.Store.Phone != "" {
		data = append(data, escpos.Text(input.Store.Phone)...)
		data = append(data, escpos.LF()...)
	}

	data = append(data, escpos.Text(r.separator("="))...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)

	// Transaction
	data = append(data, escpos.AlignLeft()...)

	if input.Transaction.InvoiceNumber != "" {
		line := "Invoice: " + input.Transaction.InvoiceNumber

		data = append(data, escpos.Text(line)...)
		data = append(data, escpos.LF()...)
	}

	if !input.Transaction.Timestamp.IsZero() {
		line := input.Transaction.Timestamp.Format(
			"02/01/2006 15:04:05",
		)

		data = append(data, escpos.Text(line)...)
		data = append(data, escpos.LF()...)
	}

	if input.Transaction.Cashier != "" {
		line := "Kasir: " + input.Transaction.Cashier

		data = append(data, escpos.Text(line)...)
		data = append(data, escpos.LF()...)
	}

	data = append(data, escpos.LF()...)

	// Items
	for _, item := range input.Items {
		data = append(data, escpos.Text(item.Name)...)
		data = append(data, escpos.LF()...)

		itemLine := fmt.Sprintf(
			"%d x %s",
			item.Quantity,
			formatRupiah(item.UnitPrice),
		)

		total := itemTotal(item)

		priceLine := formatLeftRight(
			itemLine,
			formatRupiah(total),
			r.Width,
		)

		data = append(data, escpos.Text(priceLine)...)
		data = append(data, escpos.LF()...)
	}

	// Summary
	data = append(data, escpos.Text(r.separator("-"))...)
	data = append(data, escpos.LF()...)

	subtotal := calculateSubtotal(input)

	data = append(
		data,
		escpos.Text(
			formatLeftRight(
				"Subtotal",
				formatRupiah(subtotal),
				r.Width,
			),
		)...,
	)
	data = append(data, escpos.LF()...)

	if input.Summary.Discount > 0 {
		data = append(
			data,
			escpos.Text(
				formatLeftRight(
					"Diskon",
					formatRupiah(input.Summary.Discount),
					r.Width,
				),
			)...,
		)

		data = append(data, escpos.LF()...)
	}

	if input.Summary.Tax > 0 {
		data = append(
			data,
			escpos.Text(
				formatLeftRight(
					"Pajak",
					formatRupiah(input.Summary.Tax),
					r.Width,
				),
			)...,
		)

		data = append(data, escpos.LF()...)
	}

	if input.Summary.ServiceCharge > 0 {
		data = append(
			data,
			escpos.Text(
				formatLeftRight(
					"Biaya Layanan",
					formatRupiah(input.Summary.ServiceCharge),
					r.Width,
				),
			)...,
		)

		data = append(data, escpos.LF()...)
	}

	total := input.Summary.Total

	if total == 0 {
		total = subtotal -
			input.Summary.Discount +
			input.Summary.Tax +
			input.Summary.ServiceCharge
	}

	data = append(data, escpos.Bold(true)...)

	data = append(
		data,
		escpos.Text(
			formatLeftRight(
				"TOTAL",
				formatRupiah(total),
				r.Width,
			),
		)...,
	)

	data = append(data, escpos.LF()...)
	data = append(data, escpos.Bold(false)...)

	// Payment
	data = append(data, escpos.Text(r.separator("-"))...)
	data = append(data, escpos.LF()...)

	if input.Payment.Method != "" {
		data = append(
			data,
			escpos.Text(
				formatLeftRight(
					"Metode",
					input.Payment.Method,
					r.Width,
				),
			)...,
		)

		data = append(data, escpos.LF()...)
	}

	data = append(
		data,
		escpos.Text(
			formatLeftRight(
				"Bayar",
				formatRupiah(input.Payment.Paid),
				r.Width,
			),
		)...,
	)

	data = append(data, escpos.LF()...)

	change := input.Payment.Change

	if change == 0 {
		change = input.Payment.Paid - total
	}

	data = append(
		data,
		escpos.Text(
			formatLeftRight(
				"Kembali",
				formatRupiah(change),
				r.Width,
			),
		)...,
	)

	data = append(data, escpos.LF()...)

	// Footer
	data = append(data, escpos.Text(r.separator("-"))...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)

	if input.Footer.Message != "" {
		data = append(data, escpos.AlignCenter()...)
		data = append(data, escpos.Bold(true)...)
		data = append(
			data,
			escpos.Text(input.Footer.Message)...,
		)
		data = append(data, escpos.LF()...)
		data = append(data, escpos.Bold(false)...)
	}

	data = append(data, escpos.Feed(3)...)

	return data
}

func (r *Renderer) separator(character string) string {
	return strings.Repeat(character, r.Width)
}

func formatLeftRight(left, right string, width int) string {
	if width <= 0 {
		return ""
	}

	if len(right) >= width {
		return right[:width]
	}

	if len(left)+len(right) >= width {
		maxLeft := width - len(right)

		if maxLeft <= 0 {
			return right
		}

		return truncate(left, maxLeft) + right
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
			[]string{
				string(digits[len(digits)-3:]),
			},
			groups...,
		)

		digits = digits[:len(digits)-3]
	}

	groups = append(
		[]string{
			string(digits),
		},
		groups...,
	)

	return sign + "Rp " + strings.Join(groups, ".")
}
