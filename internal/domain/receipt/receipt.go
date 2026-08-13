package receipt

import "time"

type Receipt struct {
	Store       Store
	Transaction Transaction
	Customer    Customer
	Items       []Item
	Summary     Summary
	Payment     Payment
	Footer      Footer
}

type Store struct {
	Name    string
	Address string
	Phone   string
}

type Transaction struct {
	ID            string
	InvoiceNumber string
	Timestamp     time.Time
	Cashier       string
}

type Customer struct {
	ID    string
	Name  string
	Phone string
}

type Item struct {
	ProductID string
	SKU       string
	Name      string
	Quantity  int
	UnitPrice int64
}

type Summary struct {
	Subtotal      int64
	Discount      int64
	Tax           int64
	ServiceCharge int64
	Total         int64
}

type Payment struct {
	Method string
	Paid   int64
	Change int64
}

type Footer struct {
	Message string
}
