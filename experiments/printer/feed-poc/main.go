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
	data = append(data, escpos.Text("PAPER FEED TEST")...)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.Feed(5)...)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print feed test: %v", err)
	}

	fmt.Printf("Paper feed test printed successfully (%d bytes sent)\n", n)
}