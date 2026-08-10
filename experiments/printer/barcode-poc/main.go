package main

import (
	"fmt"
	"os"
)

const PrinterDevice = "/dev/usb/lp0"

func main() {
	printer, err := os.OpenFile(
		PrinterDevice,
		os.O_WRONLY,
		0,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to open printer device: %v", err))
	}
	defer printer.Close()

	// Init
	init := []byte{0x1B, 0x40}

	// Center Allignment
	center := []byte{0x1B, 0x61, 0x01}

	// Print Title
	title := []byte("KASA BARCODE TEST\n\n")

	barcode := []byte{
		0x1D, 0x6B, 0x04,
		'K', 'A', 'S', 'A', '0', '0', '1',
		0x00,
	}

	// Feed
	feed := []byte{
		0x1B, 0x64, 0x04,
	}

	commands := [][]byte{
		init,
		center,
		title,
		barcode,
		feed,
	}

	for _, data := range commands {
		_, err := printer.Write(data)
		if err != nil {
			panic(fmt.Sprintf("failed to write to printer: %v", err))
		}
	}

	fmt.Println("Barcode test command sent successfully.")
}
