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
		panic(fmt.Sprintf(
			"failed to open printer device: %v",
			err,
		))
	}
	defer printer.Close()

	data := []byte(
		"\x1B\x40" +
			"12345678901234567890123456789012\n" +           // 32
			"123456789012345678901234567890123\n" +           // 33
			"123456789012345678901234567890123456789012\n" +  // 42
			"123456789012345678901234567890123456789012345678\n" + // 48
			"1234567890123456789012345678901234567890123456789\n" + // 49
			"\n" +
			"END\n" +
			"\x1B\x64\x03",
	)

	_, err = printer.Write(data)
	if err != nil {
		panic(fmt.Sprintf(
			"failed to write to printer: %v",
			err,
		))
	}

	fmt.Println("Character width test printed successfully.")
}
