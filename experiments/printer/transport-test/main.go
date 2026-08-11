package main

import (
	"fmt"
	"log"

	"pos-system/internal/printer/transport"
)

func main() {
	printer := transport.NewUSBPrinter("/dev/usb/lp0")

	if err := printer.Open(); err != nil {
		log.Fatal(err)
	}
	defer printer.Close()

	data := []byte("KASA TRANSPORT TEST\n")
	n, err := printer.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Wrote %d bytes\n", n)
}