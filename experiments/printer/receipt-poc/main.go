package main

import (
	"fmt"
	"log"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
	printerreceipt "pos-system/internal/printer/receipt"
	"pos-system/internal/printer/transport"
)

const printerDevice = "/dev/usb/lp0"

func main() {
	printer := transport.NewUSBPrinter(printerDevice)

	if err := printer.Open(); err != nil {
		log.Fatalf("failed to connect to printer: %v", err)
	}
	defer printer.Close()

	transaction := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-000001",
			InvoiceNumber: "INV-000001",
			Timestamp:     time.Now(),
			Cashier:       "Bizar",
		},

		Items: []domainreceipt.Item{
			{
				ProductID: "PROD-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
			},
			{
				ProductID: "PROD-002",
				SKU:       "ROTI-001",
				Name:      "Roti Bakar",
				Quantity:  1,
				UnitPrice: 12000,
			},
			{
				ProductID: "PROD-003",
				SKU:       "AIR-001",
				Name:      "Air Mineral",
				Quantity:  1,
				UnitPrice: 5000,
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal: 47000,
			Total:    47000,
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid:   50000,
			Change: 3000,
		},

		Footer: domainreceipt.Footer{
			Message: "DAY 3 DOMAIN RECEIPT",
		},
	}

	renderer := printerreceipt.NewRenderer()

	data := renderer.Render(transaction)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print receipt: %v", err)
	}

	fmt.Printf(
		"Receipt printed successfully (%d bytes sent)\n",
		n,
	)
}
