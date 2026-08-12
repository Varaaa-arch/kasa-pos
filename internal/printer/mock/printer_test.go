package mock

import (
	"bytes"
	"testing"

	"pos-system/internal/printer/transport"
)

func TestMockPrinterImplementsPrinter(t *testing.T) {
	var printer transport.Printer = &Printer{}

	if printer == nil {
		t.Fatal("expected mock printer implementation")
	}
}

func TestMockPrinterOpen(t *testing.T) {
	printer := &Printer{}

	if err := printer.Open(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if printer.OpenCount != 1 {
		t.Fatalf(
			"expected OpenCount=1, got %d",
			printer.OpenCount,
		)
	}

	if !printer.Opened {
		t.Fatal("expected printer to be opened")
	}
}

func TestMockPrinterWrite(t *testing.T) {
	printer := &Printer{}

	data := []byte("HELLO")

	n, err := printer.Write(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != len(data) {
		t.Fatalf(
			"expected %d bytes, got %d",
			len(data),
			n,
		)
	}

	if !bytes.Equal(printer.Data, data) {
		t.Fatalf(
			"unexpected stored data: %q",
			printer.Data,
		)
	}

	if printer.WriteCount != 1 {
		t.Fatalf(
			"expected WriteCount=1, got %d",
			printer.WriteCount,
		)
	}
}

func TestMockPrinterClose(t *testing.T) {
	printer := &Printer{}

	if err := printer.Open(); err != nil {
		t.Fatal(err)
	}

	if err := printer.Close(); err != nil {
		t.Fatal(err)
	}

	if printer.CloseCount != 1 {
		t.Fatalf(
			"expected CloseCount=1, got %d",
			printer.CloseCount,
		)
	}

	if printer.Opened {
		t.Fatal("expected printer to be closed")
	}

	if !printer.Closed {
		t.Fatal("expected Closed=true")
	}
}

func TestMockPrinterWriteError(t *testing.T) {
	printer := &Printer{
		WriteErr: ErrWrite,
	}

	_, err := printer.Write([]byte("TEST"))

	if err != ErrWrite {
		t.Fatalf(
			"expected %v, got %v",
			ErrWrite,
			err,
		)
	}

	if printer.WriteCount != 1 {
		t.Fatalf(
			"expected WriteCount=1, got %d",
			printer.WriteCount,
		)
	}
}

func TestMockPrinterReset(t *testing.T) {
	printer := &Printer{}

	_ = printer.Open()
	_, _ = printer.Write([]byte("TEST"))
	_ = printer.Close()

	printer.Reset()

	if printer.OpenCount != 0 {
		t.Fatal("OpenCount was not reset")
	}

	if printer.WriteCount != 0 {
		t.Fatal("WriteCount was not reset")
	}

	if printer.CloseCount != 0 {
		t.Fatal("CloseCount was not reset")
	}

	if len(printer.Data) != 0 {
		t.Fatal("Data was not reset")
	}

	if printer.Opened {
		t.Fatal("Opened was not reset")
	}
}
