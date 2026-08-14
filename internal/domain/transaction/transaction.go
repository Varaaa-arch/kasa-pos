package transaction

import "time"

type Status string

const (
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

type Transaction struct {
	ID            string
	InvoiceNumber string

	Items []Item

	Subtotal      int64
	Discount      int64
	Tax           int64
	ServiceCharge int64
	Total         int64

	PaidAmount    int64
	Change        int64
	PaymentMethod string

	Status    Status
	CreatedAt time.Time
}

type Item struct {
	ID            string
	TransactionID string

	ProductID string
	SKU       string
	Name      string

	Quantity  int
	UnitPrice int64
	Subtotal  int64
}
