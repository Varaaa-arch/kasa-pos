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
	UnitPrice Money
}

type Summary struct {
	Subtotal      Money
	Discount      Money
	Tax           Money
	ServiceCharge Money
	Total         Money
}

type Payment struct {
	Method string
	Paid   Money
	Change Money
}

type Footer struct {
	Message string
}
