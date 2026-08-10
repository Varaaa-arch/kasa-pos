package main

import (
	"fmt"
	"os"
	"time"

	"pos-system/experiments/printer/receipt-prototype/pkg/receipt"
)

const printerDevice = "/dev/usb/lp0"

func main() {
	printer, err := os.OpenFile(
		printerDevice,
		os.O_WRONLY,
		0,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to open printer device: %v", err))
	}
	defer printer.Close()

	escpos := receipt.NewESCPosPrinter()

	// Header
	escpos.AlignCenter()
	escpos.Bold(true)
	escpos.Text("TOKO KASA")
	escpos.LF()

	escpos.Bold(false)
	escpos.Text("Jl. Contoh No. 123")
	escpos.LF()
	escpos.LF()

	// Content
	escpos.AlignLeft()
	escpos.Text("--------------------------------")
	escpos.LF()

	escpos.Text("Kopi Susu")
	escpos.LF()

	escpos.Text("2 x Rp 15.000          Rp 30.000")
	escpos.LF()

	escpos.Text("--------------------------------")
	escpos.LF()

	// Footer
	escpos.AlignCenter()
	escpos.Bold(true)
	escpos.Text("TERIMA KASIH")
	escpos.LF()

	escpos.Bold(false)
	escpos.Feed(3)

	data := escpos.Bytes()
	totalWritten := 0

	for totalWritten < len(data) {
		n, err := printer.Write(data[totalWritten:])
		if err != nil {
			panic(fmt.Sprintf("failed to write to printer: %v", err))
		}

		totalWritten += n
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Printf(
		"ESC/POS data sent successfully: %d/%d bytes\n",
		totalWritten,
		len(data),
	)

	fmt.Println("ESC/POS formatting test printed successfully.")
}
