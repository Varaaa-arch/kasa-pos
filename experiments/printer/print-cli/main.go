package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"pos-system/internal/printer/receipt"
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

	renderer := receipt.NewRenderer()

	r := receipt.Receipt{
		Store: receipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: receipt.Transaction{
			InvoiceNumber: *invoice,
			TimeStamp:     time.Now(),
			Cashier:       *cashier,
		},

		Items: []receipt.Item{
			{
				Name:      "Kopi Susu",
				SKU:       "KOPI-001",
				Quantity:  2,
				UnitPrice: 15000,
			},
			{
				Name:      "Roti Bakar",
				SKU:       "ROTI-001",
				Quantity:  1,
				UnitPrice: 12000,
			},
			{
				Name:      "Air Mineral",
				SKU:       "AIR-001",
				Quantity:  1,
				UnitPrice: 5000,
			},
		},

		Summary: receipt.Summary{
			SubTotal: 47000,
			Total:    47000,
		},

		Payment: receipt.Payment{
			Method: "CASH",
			Paid:   50000,
			Change: 3000,
		},

		Footer: receipt.Footer{
			Message: "CLI TEST PRINT OK",
		},
	}

	fmt.Println("Printing receipt...")

	data := renderer.Render(r)

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
