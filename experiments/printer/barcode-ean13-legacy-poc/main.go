package main

import (
	"fmt"
	"log"

	"pos-system/internal/printer/escpos"
	"pos-system/internal/printer/transport"
)

const printerDevice = "/dev/usb/lp0"

func main() {
	printer := transport.NewUSBPrinter(printerDevice)

	if err := printer.Open(); err != nil {
		log.Fatalf("failed to connect to printer: %v", err)
	}
	defer printer.Close()

	var data []byte

	data = append(data, escpos.Initialize()...)
	data = append(data, escpos.AlignCenter()...)

	data = append(data, escpos.Text("KASA EAN13 LEGACY TEST")...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)

	// GS k 2
	// EAN-13
	// 12 data digits
	// Printer generates/checks the 13th digit.
	barcodeData := []byte("590123412345")

	data = append(data,
		0x1D, 0x6B, 0x02,
	)

	data = append(data, barcodeData...)
	data = append(data, 0x00)

	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.Feed(4)...)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print EAN13 barcode: %v", err)
	}

	fmt.Printf(
		"EAN13 legacy test sent successfully (%d bytes)\n",
		n,
	)
}
