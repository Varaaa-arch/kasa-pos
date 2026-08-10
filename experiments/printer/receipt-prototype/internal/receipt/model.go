package receipt

import "time"

type Receipt struct {
	Store       Store
	Transaction Transaction
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
	InvoiceNumber string
	TimeStamp     time.Time
	Cashier       string
}

type Item struct {
	Name      string
	SKU       string
	Quantity  int
	UnitPrice int64
	Discount  int64
	SubTotal  int64
}

type Summary struct {
	SubTotal      int64
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
	QRCode  string
}
