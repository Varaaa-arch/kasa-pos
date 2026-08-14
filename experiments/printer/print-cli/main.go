package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
	printerreceipt "pos-system/internal/printer/receipt"
	"pos-system/internal/printer/transport"
)

const defaultPrinterDevice = "/dev/usb/lp0"

func main() {
	device := flag.String(
		"device",
		defaultPrinterDevice,
		"printer device path",
	)

	invoice := flag.String(
		"invoice",
		"INV-CLI-001",
		"invoice number",
	)

	cashier := flag.String(
		"cashier",
		"CLI",
		"cashier name",
	)

	flag.Parse()

	fmt.Println("KASA Printer Test CLI")
	fmt.Println("----------------------")
	fmt.Printf("Device : %s\n", *device)
	fmt.Printf("Invoice: %s\n", *invoice)
	fmt.Printf("Cashier: %s\n", *cashier)
	fmt.Println()

	fmt.Println("Connecting to printer...")

	printer := transport.NewUSBPrinter(*device)

	if err := printer.Open(); err != nil {
		log.Fatalf(
			"failed to connect to printer: %v",
			err,
		)
	}

	renderer := printerreceipt.NewRenderer()

	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-CLI-001",
			InvoiceNumber: *invoice,
			Timestamp:     time.Now(),
			Cashier:       *cashier,
		},

		Items: []domainreceipt.Item{
			{
				ProductID: "PROD-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
			},
			{
				ProductID: "PROD-002",
				SKU:       "ROTI-001",
				Name:      "Roti Bakar",
				Quantity:  1,
				UnitPrice: domainreceipt.NewMoney(12000, domainreceipt.IDR),
			},
			{
				ProductID: "PROD-003",
				SKU:       "AIR-001",
				Name:      "Air Mineral",
				Quantity:  1,
				UnitPrice: domainreceipt.NewMoney(5000, domainreceipt.IDR),
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal: domainreceipt.NewMoney(47000, domainreceipt.IDR),
			Total: domainreceipt.NewMoney(47000, domainreceipt.IDR),
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid: domainreceipt.NewMoney(50000, domainreceipt.IDR),
			Change: domainreceipt.NewMoney(3000, domainreceipt.IDR),
		},

		Footer: domainreceipt.Footer{
			Message: "CLI TEST PRINT OK",
		},
	}

	fmt.Println("Printing receipt...")

	data := renderer.Render(input)

	_, err := printer.Write(data)
	if err != nil {
		_ = printer.Close()

		if os.IsNotExist(err) {
			log.Fatalf(
				"printer device disappeared: %v",
				err,
			)
		}

		log.Fatalf(
			"failed to print receipt: %v",
			err,
		)
	}

	if err := printer.Close(); err != nil {
		log.Fatalf(
			"failed to close printer: %v",
			err,
		)
	}

	fmt.Println("Receipt printed successfully.")
}