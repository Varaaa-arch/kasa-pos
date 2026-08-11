package main

import (
	"fmt"
	"log"
	"time"

	"pos-system/internal/printer/receipt"
	"pos-system/internal/printer/transport"
)

const printerDevice = "/dev/usb/lp0"

func main() {
	printer := transport.NewUSBPrinter(printerDevice)

	if err := printer.Open(); err != nil {
		log.Fatalf("failed to connect to printer: %v", err)
	}
	defer printer.Close()

	transaction := receipt.Receipt{
		Store: receipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: receipt.Transaction{
			InvoiceNumber: "INV-000001",
			TimeStamp:     time.Now(),
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

	renderer := receipt.NewRenderer()

	data := renderer.Render(transaction)

	// Debug information.
	fmt.Printf("Sending %d bytes to printer\n", len(data))
	fmt.Printf("ESC/POS HEX:\n% X\n", data)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print receipt: %v", err)
	}

	fmt.Printf(
		"Receipt printed successfully (%d bytes sent)\n",
		n,
	)
}
