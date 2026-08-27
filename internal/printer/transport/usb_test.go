package transport

import (
	"errors"
	"os"
	"testing"
	"time"
)

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

func TestUSBPrinterDefaultWriteTimeout(t *testing.T) {
	if DefaultWriteTimeout != 5*time.Second {
		t.Fatalf("expected DefaultWriteTimeout 5s, got %v", DefaultWriteTimeout)
	}

	printer := NewUSBPrinter("/dev/usb/lp0")
	if printer.WriteTimeout != DefaultWriteTimeout {
		t.Fatalf("expected WriteTimeout %v, got %v", DefaultWriteTimeout, printer.WriteTimeout)
	}
}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestIsTimeout(t *testing.T) {
	if !isTimeout(os.ErrDeadlineExceeded) {
		t.Error("expected isTimeout(os.ErrDeadlineExceeded) to be true")
	}

	if !isTimeout(&timeoutError{}) {
		t.Error("expected isTimeout(&timeoutError{}) to be true")
	}

	if isTimeout(errors.New("generic error")) {
		t.Error("expected isTimeout(generic error) to be false")
	}
}