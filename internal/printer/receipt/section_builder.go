package receipt

import (
	"fmt"

	domainreceipt "pos-system/internal/domain/receipt"
)

func BuildSections(input domainreceipt.Receipt) []Section {
	return []Section{
		buildHeaderSection(input),
		buildTransactionSection(input),
		buildItemsSection(input),
		buildSummarySection(input),
		buildPaymentSection(input),
		buildFooterSection(input),
	}
}

func buildHeaderSection(input domainreceipt.Receipt) Section {
	var lines []string

	if input.Store.Name != "" {
		lines = append(
			lines,
			input.Store.Name,
		)
	}

	if input.Store.Address != "" {
		lines = append(
			lines,
			input.Store.Address,
		)
	}

	if input.Store.Phone != "" {
		lines = append(
			lines,
			input.Store.Phone,
		)
	}

	return NewSection(
		SectionHeader,
		lines...,
	)
}

func buildTransactionSection(input domainreceipt.Receipt) Section {
	var lines []string

	if input.Transaction.InvoiceNumber != "" {
		lines = append(
			lines,
			"Invoice: "+input.Transaction.InvoiceNumber,
		)
	}

	if !input.Transaction.Timestamp.IsZero() {
		lines = append(
			lines,
			input.Transaction.Timestamp.Format(
				"02/01/2006 15:04:05",
			),
		)
	}

	if input.Transaction.Cashier != "" {
		lines = append(
			lines,
			"Kasir: "+input.Transaction.Cashier,
		)
	}

	return NewSection(
		SectionTransaction,
		lines...,
	)
}

func buildItemsSection(input domainreceipt.Receipt) Section {
	var lines []string

	layout := NewLayout(32)
	renderer := NewItemRenderer(layout)

	for _, item := range input.Items {
		lines = append(
			lines,
			renderer.Render(item)...,
		)

		lines = append(
			lines,
			"",
		)
	}

	return NewSection(
		SectionItems,
		lines...,
	)
}

func buildSummarySection(input domainreceipt.Receipt) Section {
	var lines []string

	subtotal := calculateSubtotal(input)

	lines = append(
		lines,
		formatLeftRight(
			"Subtotal",
			"Rp"+formatMoney(subtotal),
			32,
		),
	)

	if input.Summary.Discount != 0 {
		lines = append(
			lines,
			formatLeftRight(
				"Diskon",
				"Rp"+formatMoney(input.Summary.Discount),
				32,
			),
		)
	}

	if input.Summary.Tax != 0 {
		lines = append(
			lines,
			formatLeftRight(
				"Pajak",
				"Rp"+formatMoney(input.Summary.Tax),
				32,
			),
		)
	}

	if input.Summary.ServiceCharge != 0 {
		lines = append(
			lines,
			formatLeftRight(
				"Biaya Layanan",
				"Rp"+formatMoney(input.Summary.ServiceCharge),
				32,
			),
		)
	}

	total := input.Summary.Total

	if total == 0 {
		total =
			subtotal -
				input.Summary.Discount +
				input.Summary.Tax +
				input.Summary.ServiceCharge
	}

	lines = append(
		lines,
		formatLeftRight(
			"TOTAL",
			"Rp"+formatMoney(total),
			32,
		),
	)

	return NewSection(
		SectionSummary,
		lines...,
	)
}

func buildPaymentSection(input domainreceipt.Receipt) Section {
	var lines []string

	if input.Payment.Method != "" {
		lines = append(
			lines,
			formatLeftRight(
				"Metode",
				input.Payment.Method,
				32,
			),
		)
	}

	lines = append(
		lines,
		formatLeftRight(
			"Bayar",
			"Rp"+formatMoney(input.Payment.Paid),
			32,
		),
	)

	change := input.Payment.Change

	if change == 0 {
		subtotal := calculateSubtotal(input)

		total := input.Summary.Total

		if total == 0 {
			total =
				subtotal -
					input.Summary.Discount +
					input.Summary.Tax +
					input.Summary.ServiceCharge
		}

		change = input.Payment.Paid - total
	}

	lines = append(
		lines,
		formatLeftRight(
			"Kembali",
			"Rp"+formatMoney(change),
			32,
		),
	)

	return NewSection(
		SectionPayment,
		lines...,
	)
}

func buildFooterSection(input domainreceipt.Receipt) Section {
	var lines []string

	if input.Footer.Message != "" {
		lines = append(
			lines,
			input.Footer.Message,
		)
	}

	return NewSection(
		SectionFooter,
		lines...,
	)
}

func formatSectionLine(label, value string) string {
	if value == "" {
		return label
	}

	return fmt.Sprintf(
		"%s: %s",
		label,
		value,
	)
}
