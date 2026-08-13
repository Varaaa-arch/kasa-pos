package receipt

import (
	"errors"
	"testing"

	domainreceipt "pos-system/internal/domain/receipt"
)

type failingPrinter struct {
	openErr  error
	writeErr error
	closeErr error

	closed bool
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
		domainreceipt.Receipt{},
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
		domainreceipt.Receipt{},
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
		domainreceipt.Receipt{},
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
		domainreceipt.Receipt{},
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

type retryPrinter struct {
	openAttempts  int
	writeAttempts int
}

func (p *retryPrinter) Open() error {
	p.openAttempts++

	if p.openAttempts < 3 {
		return errors.New("printer temporarily unavailable")
	}

	return nil
}

func (p *retryPrinter) Write(data []byte) (int, error) {
	p.writeAttempts++

	return len(data), nil
}

func (p *retryPrinter) Close() error {
	return nil
}

func TestPrintRetriesOpen(t *testing.T) {
	printer := &retryPrinter{}

	renderer := NewRenderer()

	err := Print(
		printer,
		renderer,
		domainreceipt.Receipt{},
	)

	if err != nil {
		t.Fatalf(
			"expected print to succeed after retry: %v",
			err,
		)
	}

	if printer.openAttempts != 3 {
		t.Fatalf(
			"expected 3 open attempts, got %d",
			printer.openAttempts,
		)
	}

	if printer.writeAttempts != 1 {
		t.Fatalf(
			"expected exactly 1 write attempt, got %d",
			printer.writeAttempts,
		)
	}
}
