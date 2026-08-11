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

	data = append(data, escpos.FontSize(1, 1)...)
	data = append(data, escpos.Text("NORMAL SIZE")...)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.FontSize(2, 2)...)
	data = append(data, escpos.Text("DOUBLE SIZE")...)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.FontSize(1, 1)...)
	data = append(data, escpos.Text("NORMAL AGAIN")...)
	data = append(data, escpos.LF()...)

	data = append(data, escpos.Feed(4)...)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print font size test: %v", err)
	}

	fmt.Printf("Font size test printed successfully (%d bytes sent)\n", n)
}