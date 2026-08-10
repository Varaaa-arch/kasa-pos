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
		panic(fmt.Sprintf("failed to open printer device: %v", err))
	}
	defer printer.Close()

	// ESC @ Initialize printer
	init := []byte{0x1B, 0x40}

	// print test
	text := []byte(
		"\n" +
			"TOKO KASA\n" +
			"PAPER CUT TEST\n" +
			"------------------------------\n" +
			"Jika printer memiliki cutter,\n" +
			"kertas seharusnya terpotong.\n\n",
	)

	// Feed beberapa baris sebelum cut
	feed := []byte{0x1B, 0x64, 0x03}

	// GS V 0 (Full Cut)
	cut := []byte{0x1D, 0x56, 0x00}

	commands := [][]byte{
		init,
		text,
		feed,
		cut,
	}

	for _, data := range commands {
		_, err = printer.Write(data)
		if err != nil {
			panic(fmt.Sprintf("failed to write to printer: %v", err))
		}
	}

	fmt.Println("Paper cut command successfully")
}