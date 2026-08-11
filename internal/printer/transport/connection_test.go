package transport

import (
	"os"
	"testing"
)

func TestPrinterConnection(t *testing.T) {
	devicePath := "/dev/usb/lp0"

	if _, err := os.Stat(devicePath); err != nil {
		t.Skipf(
			"printer device %s is not available: %v",
			devicePath,
			err,
		)
	}

	printer := NewUSBPrinter(devicePath)

	if err := printer.Open(); err != nil {
		t.Fatalf("failed to connect to printer: %v", err)
	}

	defer printer.Close()

	t.Logf("printer connection successful: %s", devicePath)
}
