package agent

import (
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

type printRequest struct {
	Store       printStore       `json:"store"`
	Transaction printTransaction `json:"transaction"`
	Items       []printItem      `json:"items"`
	Summary     printSummary     `json:"summary"`
	Payment     printPayment     `json:"payment"`
	Footer      printFooter      `json:"footer"`
}

type printStore struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type printTransaction struct {
	ID            string    `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	Timestamp     time.Time `json:"timestamp"`
	Cashier       string    `json:"cashier"`
}

type printItem struct {
	ProductID string `json:"product_id"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

type printSummary struct {
	Subtotal      int64 `json:"subtotal"`
	Discount      int64 `json:"discount"`
	Tax           int64 `json:"tax"`
	ServiceCharge int64 `json:"service_charge"`
	Total         int64 `json:"total"`
}

type printPayment struct {
	Method string `json:"method"`
	Paid   int64  `json:"paid"`
	Change int64  `json:"change"`
}

type printFooter struct {
	Message string `json:"message"`
}

func toPrintRequest(receipt domainreceipt.Receipt) printRequest {
	items := make([]printItem, len(receipt.Items))
	for i, item := range receipt.Items {
		items[i] = printItem{
			ProductID: item.ProductID,
			SKU:       item.SKU,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice.Amount,
		}
	}

	return printRequest{
		Store: printStore{
			Name:    receipt.Store.Name,
			Address: receipt.Store.Address,
			Phone:   receipt.Store.Phone,
		},
		Transaction: printTransaction{
			ID:            receipt.Transaction.ID,
			InvoiceNumber: receipt.Transaction.InvoiceNumber,
			Timestamp:     receipt.Transaction.Timestamp,
			Cashier:       receipt.Transaction.Cashier,
		},
		Items: items,
		Summary: printSummary{
			Subtotal:      receipt.Summary.Subtotal.Amount,
			Discount:      receipt.Summary.Discount.Amount,
			Tax:           receipt.Summary.Tax.Amount,
			ServiceCharge: receipt.Summary.ServiceCharge.Amount,
			Total:         receipt.Summary.Total.Amount,
		},
		Payment: printPayment{
			Method: receipt.Payment.Method,
			Paid:   receipt.Payment.Paid.Amount,
			Change: receipt.Payment.Change.Amount,
		},
		Footer: printFooter{
			Message: receipt.Footer.Message,
		},
	}
}
