package main

import (
	"fmt"
	"time"

	"pos-system/experiments/printer/receipt-prototype/internal/receipt"
)

func main() {
	transaction := receipt.Receipt{
		Store: receipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: receipt.Transaction{
			InvoiceNumber: "INV-000001",
			TimeStamp:     time.Date(2026, 8, 10, 19, 0, 0, 0, time.Local),
			Cashier:       "Bizar",
		},

		Items: []receipt.Item{
			{
				Name:      "Kopi Susu",
				SKU:       "KOPI-001",
				Quantity:  2,
				UnitPrice: 15000,
				SubTotal:  30000,
			},
			{
				Name:      "Roti Bakar",
				SKU:       "ROTI-001",
				Quantity:  1,
				UnitPrice: 12000,
				SubTotal:  12000,
			},
			{
				Name:      "Air Mineral",
				SKU:       "AIR-001",
				Quantity:  1,
				UnitPrice: 5000,
				SubTotal:  5000,
			},
		},

		Summary: receipt.Summary{
			SubTotal:      47000,
			Discount:      0,
			Tax:           0,
			ServiceCharge: 0,
			Total:         47000,
		},

		Payment: receipt.Payment{
			Method: "CASH",
			Paid:   50000,
			Change: 3000,
		},

		Footer: receipt.Footer{
			Message: "TERIMA KASIH",
		},
	}

	output := receipt.FormatText(transaction)

	fmt.Println(output)
}
