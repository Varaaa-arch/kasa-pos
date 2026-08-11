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

	data := escpos.Initialize()

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to initialize printer: %v", err)
	}

	fmt.Printf(
		"Printer initialized successfully (%d bytes sent)\n",
		n,
	)
}