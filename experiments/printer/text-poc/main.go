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

	data := append(
		escpos.Initialize(),
		escpos.Text("HELLO WORLD")...,
	)

	data = append(data, escpos.LF()...)
	data = append(data, escpos.LF()...)
	data = append(data, escpos.Feed(3)...)

	n, err := printer.Write(data)
	if err != nil {
		log.Fatalf("failed to print text: %v", err)
	}

	fmt.Printf(
		"Text printed successfully (%d bytes sent)\n",
		n,
	)
}