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

	// Initialize printer.
	data = append(data, escpos.Initialize()...)

	// Center barcode.
	data = append(data, escpos.AlignCenter()...)

	// Title.
	data = append(data, escpos.Text("KASA EAN13 TEST")...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)

	// EAN-13:
	// GS k 67
	// 12 data digits; printer calculates the check digit.
	//
	// Payload:
	// 590123412345
	//
	// Expected EAN-13:
	// 5901234123457
	data = append(data,
		0x1D, 0x6B, 0x43,
		'5', '9', '0', '1', '2', '3',
		'4', '1', '2', '3', '4', '5',
	)

	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.Feed(4)...)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print EAN13 barcode: %v", err)
	}

	fmt.Printf(
		"EAN13 barcode test sent successfully (%d bytes)\n",
		n,
	)
}
