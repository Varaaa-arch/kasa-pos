package main

import (
	"fmt"
	"log"
	"time"

	"pos-system/internal/printer/transport"
)

const printerDevice = "/dev/usb/lp0"

func main() {
	printer := transport.NewUSBPrinter(printerDevice)

	fmt.Println("Step 1: opening printer...")

	if err := printer.Open(); err != nil {
		log.Fatalf("failed to open printer: %v", err)
	}

	fmt.Println("Printer connected.")

	fmt.Println("Step 2: disconnect USB cable now.")
	fmt.Println("Waiting 10 seconds...")

	time.Sleep(10 * time.Second)

	fmt.Println("Step 3: attempting write after disconnect...")

	_, err := printer.Write(
		[]byte("RECONNECT TEST\n"),
	)

	if err != nil {
		fmt.Printf("Expected write error: %v\n", err)
	} else {
		fmt.Println("WARNING: write succeeded after disconnect.")
	}

	fmt.Println("Step 4: closing old printer handle...")

	if err := printer.Close(); err != nil {
		log.Fatalf("failed to close printer: %v", err)
	}

	fmt.Println("Old printer handle closed.")

	fmt.Println("Step 5: reconnect USB printer.")
	fmt.Println("Waiting 5 seconds...")

	time.Sleep(5 * time.Second)

	fmt.Println("Step 6: opening printer again...")

	if err := printer.Open(); err != nil {
		log.Fatalf(
			"failed to reopen printer after reconnect: %v",
			err,
		)
	}

	defer printer.Close()

	fmt.Println("Printer reopened successfully.")

	fmt.Println("Step 7: sending reconnect test...")

	_, err = printer.Write(
		[]byte("KASA RECONNECT OK\n"),
	)

	if err != nil {
		log.Fatalf(
			"failed to write after reconnect: %v",
			err,
		)
	}

	fmt.Println("Reconnect test passed.")
}