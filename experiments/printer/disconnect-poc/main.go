package main

import (
	"fmt"
	"os"
)

const printerDevice = "/dev/usb/lp0"

func main() {
	fmt.Println("Opening printer...")

	printer, err := os.OpenFile(
		printerDevice,
		os.O_WRONLY,
		0,
	)
	if err != nil {
		fmt.Printf("Failed to open printer: %v\n", err)
		return
	}

	fmt.Println("Printer connected successfully.")
	fmt.Println("Now disconnect the USB cable and press ENTER.")

	fmt.Scanln()

	_, err = printer.Write([]byte("DISCONNECT TEST\n"))
	if err != nil {
		fmt.Printf("Write failed after disconnect: %v\n", err)
		return
	}

	fmt.Println("Write succeeded.")
}