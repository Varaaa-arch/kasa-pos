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
	data = append(data, escpos.Text("KASA CODE128 TEST")...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)

	// GS k 73:
	// CODE128 with explicit data length.
	//
	// 0x49 = ASCII 'I' = CODE128 on many ESC/POS implementations.
	//
	// Data starts with {B to select CODE128 Code Set B.
	// KASA-001 is the barcode payload.
	//
	// 1D 6B 49 n d1...dn
	barcodeData := []byte("{BKASA-001")
	command := []byte{
		0x1D,
		0x6B,
		0x49,
		byte(len(barcodeData)),
	}

	data = append(data, command...)
	data = append(data, barcodeData...)

	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.Feed(4)...)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print CODE128 barcode: %v", err)
	}

	fmt.Printf(
		"CODE128 barcode test sent successfully (%d bytes)\n",
		n,
	)
}
