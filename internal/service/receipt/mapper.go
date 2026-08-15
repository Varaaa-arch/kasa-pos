package receipt

import (
	domainreceipt "pos-system/internal/domain/receipt"
	domaintransaction "pos-system/internal/domain/transaction"
)

func FromTransaction(
	tx domaintransaction.Transaction,
) domainreceipt.Receipt {
	items := make([]domainreceipt.Item, 0, len(tx.Items))

	for _, item := range tx.Items {
		items = append(
			items,
			domainreceipt.Item{
				ProductID: item.ProductID,
				SKU:       item.SKU,
				Name:      item.Name,
				Quantity:  item.Quantity,
				UnitPrice: domainreceipt.NewMoney(
					item.UnitPrice,
					domainreceipt.IDR,
				),
			},
		)
	}

	return domainreceipt.Receipt{
		Transaction: domainreceipt.Transaction{
			ID:            tx.ID,
			InvoiceNumber: tx.InvoiceNumber,
			Timestamp:     tx.CreatedAt,
			Cashier:       "",
		},

		Items: items,

		Summary: domainreceipt.Summary{
			Subtotal: domainreceipt.NewMoney(
				tx.Subtotal,
				domainreceipt.IDR,
			),
			Discount: domainreceipt.NewMoney(
				tx.Discount,
				domainreceipt.IDR,
			),
			Tax: domainreceipt.NewMoney(
				tx.Tax,
				domainreceipt.IDR,
			),
			ServiceCharge: domainreceipt.NewMoney(
				tx.ServiceCharge,
				domainreceipt.IDR,
			),
			Total: domainreceipt.NewMoney(
				tx.Total,
				domainreceipt.IDR,
			),
		},

		Payment: domainreceipt.Payment{
			Method: tx.PaymentMethod,
			Paid: domainreceipt.NewMoney(
				tx.PaidAmount,
				domainreceipt.IDR,
			),
			Change: domainreceipt.NewMoney(
				tx.Change,
				domainreceipt.IDR,
			),
		},
	}
}
