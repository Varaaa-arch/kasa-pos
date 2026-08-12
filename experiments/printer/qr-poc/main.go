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

	// Initialize
	data = append(data, escpos.Initialize()...)

	// Center
	data = append(data, escpos.AlignCenter()...)

	// Header
	data = append(data, escpos.Text("KASA QR TEST")...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)

	payload := []byte("https://kasa.local/test")

	// --------------------------------------------------
	// 1. Select QR Code model 2
	// GS ( k 4 0 49 65 50 0
	// --------------------------------------------------
	data = append(data,
		0x1D, 0x28, 0x6B,
		0x04, 0x00,
		0x31, 0x41,
		0x32, 0x00,
	)

	// --------------------------------------------------
	// 2. Set QR module size = 4
	// GS ( k 3 0 49 67 04
	// --------------------------------------------------
	data = append(data,
		0x1D, 0x28, 0x6B,
		0x03, 0x00,
		0x31, 0x43,
		0x04,
	)

	// --------------------------------------------------
	// 3. Set error correction level M
	// GS ( k 3 0 49 69 49
	// --------------------------------------------------
	data = append(data,
		0x1D, 0x28, 0x6B,
		0x03, 0x00,
		0x31, 0x45,
		0x31,
	)

	// --------------------------------------------------
	// 4. Store QR data
	//
	// GS ( k pL pH 49 80 48 data
	// pL/pH = length of 49 80 48 + payload
	// --------------------------------------------------
	storeLength := len(payload) + 3

	pL := byte(storeLength & 0xFF)
	pH := byte((storeLength >> 8) & 0xFF)

	data = append(data,
		0x1D, 0x28, 0x6B,
		pL, pH,
		0x31, 0x50, 0x30,
	)

	data = append(data, payload...)

	// --------------------------------------------------
	// 5. Print QR
	// GS ( k 3 0 49 81 48
	// --------------------------------------------------
	data = append(data,
		0x1D, 0x28, 0x6B,
		0x03, 0x00,
		0x31, 0x51,
		0x30,
	)

	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.Feed(4)...)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print QR code: %v", err)
	}

	fmt.Printf(
		"QR test sent successfully (%d bytes)\n",
		n,
	)
}
