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
	calculator := NewCalculator()
	calculation := calculator.Calculate(input)

	var lines []string

	lines = append(
		lines,
		formatLeftRight(
			"Subtotal",
			calculation.Subtotal.String(),
			32,
		),
	)

	if !input.Summary.Discount.IsZero() {
		lines = append(
			lines,
			formatLeftRight(
				"Diskon",
				input.Summary.Discount.String(),
				32,
			),
		)
	}

	if !input.Summary.Tax.IsZero() {
		lines = append(
			lines,
			formatLeftRight(
				"Pajak",
				input.Summary.Tax.String(),
				32,
			),
		)
	}

	if !input.Summary.ServiceCharge.IsZero() {
		lines = append(
			lines,
			formatLeftRight(
				"Biaya Layanan",
				input.Summary.ServiceCharge.String(),
				32,
			),
		)
	}

	lines = append(
		lines,
		formatLeftRight(
			"TOTAL",
			calculation.Total.String(),
			32,
		),
	)

	return NewSection(
		SectionSummary,
		lines...,
	)
}

func buildPaymentSection(input domainreceipt.Receipt) Section {
	calculator := NewCalculator()
	calculation := calculator.Calculate(input)

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
			calculation.Paid.String(),
			32,
		),
	)

	lines = append(
		lines,
		formatLeftRight(
			"Kembali",
			calculation.Change.String(),
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
