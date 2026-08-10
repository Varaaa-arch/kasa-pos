package main

import (
	"fmt"
	"os"
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

	data := []byte{
		0x1B, 0x40, // ESC @ - Initialize printer
		0x1B, 0x61, 0x01, // ESC a 1 - Center alignment
		'H', 'E', 'L', 'L', 'O', ' ', 'W', 'O', 'R', 'L', 'D',
		0x0A, // LF
		0x0A, // LF
	}

	_, err = printer.Write(data)
	if err != nil {
		panic(fmt.Sprintf("failed to write to printer: %v", err))
	}

	fmt.Println("ESC/POS data sent successfully.")
}
