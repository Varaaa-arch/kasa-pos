package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
	printerreceipt "pos-system/internal/printer/receipt"
)

func main() {
	width := flag.Int(
		"width",
		32,
		"receipt preview width",
	)

	invoice := flag.String(
		"invoice",
		"INV-PREVIEW-001",
		"invoice number",
	)

	cashier := flag.String(
		"cashier",
		"Bizar",
		"cashier name",
	)

	templateName := flag.String(
		"template",
		"default",
		"receipt template: default, compact, detailed",
	)

	flag.Parse()

	template, err := getTemplate(*templateName)
	if err != nil {
		log.Fatal(err)
	}

	receipt := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-PREVIEW-001",
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
			Message: "TERIMA KASIH",
		},
	}

	preview := printerreceipt.NewPreview(*width)

	preview.Render(
		receipt,
		template,
	)

	fmt.Println("KASA RECEIPT PREVIEW")
	fmt.Println("====================")
	fmt.Printf("Template : %s\n", template.Name)
	fmt.Printf("Width    : %d\n", preview.Width)
	fmt.Println("====================")
	fmt.Println()

	fmt.Println(preview.String())
}

func getTemplate(name string) (
	printerreceipt.Template,
	error,
) {
	switch name {
	case "default":
		return printerreceipt.DefaultReceiptTemplate(), nil

	case "compact":
		return printerreceipt.CompactReceiptTemplate(), nil

	case "detailed":
		return printerreceipt.DetailedReceiptTemplate(), nil

	default:
		return printerreceipt.Template{}, fmt.Errorf(
			"unknown template %q: use default, compact, or detailed",
			name,
		)
	}
}
