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
