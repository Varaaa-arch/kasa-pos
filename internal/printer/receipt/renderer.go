package receipt

import (
	"fmt"
	"strings"

	"pos-system/internal/printer/escpos"
)

const DefaultLineWidth = 48

type Renderer struct {
	Width int
}

func NewRenderer() *Renderer {
	return &Renderer{
		Width: DefaultLineWidth,
	}
}

func (r *Renderer) Render(receipt Receipt) []byte {
	var data []byte

	data = append(data, escpos.Initialize()...)

	// Header
	data = append(data, escpos.AlignCenter()...)
	data = append(data, escpos.Bold(true)...)
	data = append(data, escpos.Text(receipt.Store.Name)...)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.Bold(false)...)

	if receipt.Store.Address != "" {
		data = append(data, escpos.Text(receipt.Store.Address)...)
		data = append(data, escpos.LF()...)
	}

	if receipt.Store.Phone != "" {
		data = append(data, escpos.Text(receipt.Store.Phone)...)
		data = append(data, escpos.LF()...)
	}

	data = append(data, escpos.Text(r.separator("="))...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)

	// Transaction
	data = append(data, escpos.AlignLeft()...)

	if receipt.Transaction.InvoiceNumber != "" {
		line := "Invoice: " + receipt.Transaction.InvoiceNumber
		data = append(data, escpos.Text(line)...)
		data = append(data, escpos.LF()...)
	}

	if !receipt.Transaction.TimeStamp.IsZero() {
		line := receipt.Transaction.TimeStamp.Format("02/01/2006 15:04:05")
		data = append(data, escpos.Text(line)...)
		data = append(data, escpos.LF()...)
	}

	if receipt.Transaction.Cashier != "" {
		line := "Kasir: " + receipt.Transaction.Cashier
		data = append(data, escpos.Text(line)...)
		data = append(data, escpos.LF()...)
	}

	data = append(data, escpos.LF()...)

	// Items
	for _, item := range receipt.Items {
		data = append(data, escpos.Text(item.Name)...)
		data = append(data, escpos.LF()...)

		itemLine := fmt.Sprintf(
			"%d x %s",
			item.Quantity,
			formatRupiah(item.UnitPrice),
		)

		priceLine := formatLeftRight(
			itemLine,
			formatRupiah(item.SubTotal),
			r.Width,
		)

		data = append(data, escpos.Text(priceLine)...)
		data = append(data, escpos.LF()...)
	}

	// Summary
	data = append(data, escpos.Text(r.separator("-"))...)
	data = append(data, escpos.LF()...)

	data = append(
		data,
		escpos.Text(
			formatLeftRight(
				"Subtotal",
				formatRupiah(receipt.Summary.SubTotal),
				r.Width,
			),
		)...,
	)
	data = append(data, escpos.LF()...)

	if receipt.Summary.Discount > 0 {
		data = append(
			data,
			escpos.Text(
				formatLeftRight(
					"Diskon",
					formatRupiah(receipt.Summary.Discount),
					r.Width,
				),
			)...,
		)
		data = append(data, escpos.LF()...)
	}

	if receipt.Summary.Tax > 0 {
		data = append(
			data,
			escpos.Text(
				formatLeftRight(
					"Pajak",
					formatRupiah(receipt.Summary.Tax),
					r.Width,
				),
			)...,
		)
		data = append(data, escpos.LF()...)
	}

	if receipt.Summary.ServiceCharge > 0 {
		data = append(
			data,
			escpos.Text(
				formatLeftRight(
					"Biaya Layanan",
					formatRupiah(receipt.Summary.ServiceCharge),
					r.Width,
				),
			)...,
		)
		data = append(data, escpos.LF()...)
	}

	data = append(data, escpos.Bold(true)...)

	data = append(
		data,
		escpos.Text(
			formatLeftRight(
				"TOTAL",
				formatRupiah(receipt.Summary.Total),
				r.Width,
			),
		)...,
	)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.Bold(false)...)

	// Payment
	data = append(data, escpos.Text(r.separator("-"))...)
	data = append(data, escpos.LF()...)

	if receipt.Payment.Method != "" {
		data = append(
			data,
			escpos.Text(
				formatLeftRight(
					"Metode",
					receipt.Payment.Method,
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
				formatRupiah(receipt.Payment.Paid),
				r.Width,
			),
		)...,
	)
	data = append(data, escpos.LF()...)

	data = append(
		data,
		escpos.Text(
			formatLeftRight(
				"Kembali",
				formatRupiah(receipt.Payment.Change),
				r.Width,
			),
		)...,
	)
	data = append(data, escpos.LF()...)

	// Footer
	data = append(data, escpos.Text(r.separator("-"))...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)

	if receipt.Footer.Message != "" {
		data = append(data, escpos.AlignCenter()...)
		data = append(data, escpos.Bold(true)...)
		data = append(data, escpos.Text(receipt.Footer.Message)...)
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

	groups = append([]string{string(digits)}, groups...)

	return sign + "Rp " + strings.Join(groups, ".")
}
