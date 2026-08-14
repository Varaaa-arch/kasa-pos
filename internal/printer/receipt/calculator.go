package receipt

import (
	domainreceipt "pos-system/internal/domain/receipt"
)

type Calculation struct {
	ItemsTotal    int64
	Subtotal      int64
	Discount      int64
	Tax           int64
	ServiceCharge int64
	Total         int64
	Paid          int64
	Change        int64
}

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Calculate(
	input domainreceipt.Receipt,
) Calculation {
	var itemsTotal int64

	for _, item := range input.Items {
		itemsTotal += int64(item.Quantity) * item.UnitPrice
	}

	subtotal := itemsTotal

	discount := input.Summary.Discount
	tax := input.Summary.Tax
	serviceCharge := input.Summary.ServiceCharge

	total := subtotal -
		discount +
		tax +
		serviceCharge

	paid := input.Payment.Paid

	change := paid - total

	if change < 0 {
		change = 0
	}

	return Calculation{
		ItemsTotal:    itemsTotal,
		Subtotal:      subtotal,
		Discount:      discount,
		Tax:           tax,
		ServiceCharge: serviceCharge,
		Total:         total,
		Paid:          paid,
		Change:        change,
	}
}
