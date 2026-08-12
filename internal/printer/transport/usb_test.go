package transport

import "testing"

func TestNewUSBPrinter(t *testing.T) {
	printer := NewUSBPrinter("/dev/usb/lp0")

	if printer == nil {
		t.Fatal("Expected printer instance")
	}
	if printer.devicePath != "/dev/usb/lp0" {
		t.Errorf("Expected device path /dev/usb/lp0, got %s", printer.devicePath)
	}
}

func TestWriteWithoutOpen(t *testing.T) {
	printer := NewUSBPrinter("dev/usb/lp0")
	_, err := printer.Write([]byte("test"))
	if err == nil {
		t.Fatal("expected error when printer is not open")
	}
}

func TestUSBPrinterImplementsPrinter(t *testing.T) {
	var printer Printer

	printer = NewUSBPrinter("/dev/usb/lp0")

	if printer == nil {
		t.Fatal("expected printer implementation")
	}
}

func TestUSBPrinterLifecycle(t *testing.T) {
	printer := NewUSBPrinter("/dev/usb/lp0")

	if printer == nil {
		t.Fatal("expected printer instance")
	}

	// Initial state.
	if printer.file != nil {
		t.Fatal("expected printer to be closed initially")
	}
}