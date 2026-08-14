package receipt

import (
	domainreceipt "pos-system/internal/domain/receipt"
)

type Calculation struct {
	ItemsTotal    domainreceipt.Money
	Subtotal      domainreceipt.Money
	Discount      domainreceipt.Money
	Tax           domainreceipt.Money
	ServiceCharge domainreceipt.Money
	Total         domainreceipt.Money
	Paid          domainreceipt.Money
	Change        domainreceipt.Money
}

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Calculate(
	input domainreceipt.Receipt,
) Calculation {
	itemsTotal := domainreceipt.ZeroMoney(domainreceipt.IDR)

	for _, item := range input.Items {
		itemsTotal = itemsTotal.Add(
			item.UnitPrice.Mul(
				int64(item.Quantity),
			),
		)
	}

	subtotal := itemsTotal

	discount := moneyOrZero(input.Summary.Discount)
	tax := moneyOrZero(input.Summary.Tax)
	serviceCharge := moneyOrZero(input.Summary.ServiceCharge)

	total := subtotal.
		Sub(discount).
		Add(tax).
		Add(serviceCharge)

	paid := moneyOrZero(input.Payment.Paid)

	change := paid.Sub(total)

	if change.IsNegative() {
		change = domainreceipt.ZeroMoney(
			domainreceipt.IDR,
		)
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

// moneyOrZero returns m if it has a valid currency,
// otherwise returns ZeroMoney(IDR).
func moneyOrZero(m domainreceipt.Money) domainreceipt.Money {
	if m.Currency == "" {
		return domainreceipt.ZeroMoney(domainreceipt.IDR)
	}
	return m
}
