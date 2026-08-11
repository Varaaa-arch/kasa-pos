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

	data = append(data, escpos.AlignLeft()...)
	data = append(data, escpos.Text("NORMAL TEXT")...)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.Bold(true)...)
	data = append(data, escpos.Text("BOLD TEXT")...)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.Bold(false)...)
	data = append(data, escpos.Text("NORMAL AGAIN")...)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.Feed(3)...)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print bold test: %v", err)
	}

	fmt.Printf("Bold test printed successfully (%d bytes sent)\n", n)
}