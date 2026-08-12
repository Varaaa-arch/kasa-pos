package receipt

import (
	"errors"
	"testing"
)

type failingPrinter struct {
	openErr  error
	writeErr error
	closeErr error
	closed   bool
}

func (p *failingPrinter) Open() error {
	return p.openErr
}

func (p *failingPrinter) Write(data []byte) (int, error) {
	if p.writeErr != nil {
		return 0, p.writeErr
	}

	return len(data), nil
}

func (p *failingPrinter) Close() error {
	p.closed = true
	return p.closeErr
}

func TestPrintOpenError(t *testing.T) {
	expectedErr := errors.New("printer connection failed")

	printer := &failingPrinter{
		openErr: expectedErr,
	}

	renderer := NewRenderer()

	err := Print(
		printer,
		renderer,
		Receipt{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Print() error = %v, want %v",
			err,
			expectedErr,
		)
	}
}

func TestPrintWriteError(t *testing.T) {
	expectedErr := errors.New("printer write failed")

	printer := &failingPrinter{
		writeErr: expectedErr,
	}

	renderer := NewRenderer()

	err := Print(
		printer,
		renderer,
		Receipt{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Print() error = %v, want %v",
			err,
			expectedErr,
		)
	}
}

func TestPrintSuccess(t *testing.T) {
	printer := &failingPrinter{}

	renderer := NewRenderer()

	err := Print(
		printer,
		renderer,
		Receipt{},
	)

	if err != nil {
		t.Fatalf(
			"Print() returned unexpected error: %v",
			err,
		)
	}
}

func TestPrintClosesPrinter(t *testing.T) {
	printer := &failingPrinter{}

	renderer := NewRenderer()

	err := Print(
		printer,
		renderer,
		Receipt{},
	)

	if err != nil {
		t.Fatalf(
			"Print() returned unexpected error: %v",
			err,
		)
	}

	if !printer.closed {
		t.Fatal("expected printer to be closed")
	}
}
