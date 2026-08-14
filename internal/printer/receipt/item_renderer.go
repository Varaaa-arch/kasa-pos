package receipt

import (
	"fmt"

	domainreceipt "pos-system/internal/domain/receipt"
)

type ItemRenderer struct {
	Layout     Layout
	Calculator *Calculator
}

func NewItemRenderer(layout Layout) *ItemRenderer {
	calculator := layout.Calculator

	if calculator == nil {
		calculator = NewCalculator()
	}

	return &ItemRenderer{
		Layout:     layout,
		Calculator: calculator,
	}
}

func (r *ItemRenderer) Render(
	item domainreceipt.Item,
) []string {
	var lines []string

	// Product name.
	nameLines := wrapText(
		item.Name,
		r.Layout.ContentWidth(),
	)

	lines = append(
		lines,
		nameLines...,
	)

	// Quantity x unit price.
	qtyPrice := fmt.Sprintf(
		"%d x Rp%s",
		item.Quantity,
		formatMoney(item.UnitPrice),
	)

	// Item subtotal.
	calculation := r.Calculator.Calculate(
		domainreceipt.Receipt{
			Items: []domainreceipt.Item{
				item,
			},
		},
	)

	totalText := "Rp" + formatMoney(
		calculation.ItemsTotal,
	)

	// Need at least one character between
	// the left and right columns.
	requiredWidth :=
		len(qtyPrice) +
			1 +
			len(totalText)

	if requiredWidth <= r.Layout.ContentWidth() {
		lines = append(
			lines,
			r.Layout.LeftRight(
				qtyPrice,
				totalText,
			),
		)

		return lines
	}

	// Large values that cannot fit side-by-side.
	lines = append(
		lines,
		r.Layout.Left(qtyPrice),
	)

	lines = append(
		lines,
		r.Layout.Right(totalText),
	)

	return lines
}
