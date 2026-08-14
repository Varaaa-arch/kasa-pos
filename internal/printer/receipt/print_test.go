package receipt

import (
	"errors"
	"testing"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

func validPrintReceipt() domainreceipt.Receipt {
	return domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name: "TOKO KASA",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-PRINT-001",
			InvoiceNumber: "INV-PRINT-001",
			Timestamp:     time.Now(),
			Cashier:       "Bizar",
		},

		Items: []domainreceipt.Item{
			{
				ProductID: "PROD-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal: domainreceipt.NewMoney(30000, domainreceipt.IDR),
			Total: domainreceipt.NewMoney(30000, domainreceipt.IDR),
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid: domainreceipt.NewMoney(50000, domainreceipt.IDR),
			Change: domainreceipt.NewMoney(20000, domainreceipt.IDR),
		},

		Footer: domainreceipt.Footer{
			Message: "TEST PRINT",
		},
	}
}

type failingPrinter struct {
	openErr  error
	writeErr error
	closeErr error

	opened     bool
	closed     bool
	openCount  int
	writeCount int
}

func (p *failingPrinter) Open() error {
	p.openCount++

	if p.openErr != nil {
		return p.openErr
	}

	p.opened = true
	p.closed = false

	return nil
}

func (p *failingPrinter) Write(data []byte) (int, error) {
	p.writeCount++

	if p.writeErr != nil {
		return 0, p.writeErr
	}

	return len(data), nil
}

func (p *failingPrinter) Close() error {
	p.closed = true
	p.opened = false

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
		validPrintReceipt(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Print() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	if printer.writeCount != 0 {
		t.Fatalf(
			"expected no writes, got %d",
			printer.writeCount,
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
		validPrintReceipt(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Print() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	if printer.openCount < 1 {
		t.Fatal("expected printer to be opened")
	}

	if printer.writeCount != 1 {
		t.Fatalf(
			"expected exactly 1 write attempt, got %d",
			printer.writeCount,
		)
	}
}

func TestPrintSuccess(t *testing.T) {
	printer := &failingPrinter{}

	renderer := NewRenderer()

	err := Print(
		printer,
		renderer,
		validPrintReceipt(),
	)

	if err != nil {
		t.Fatalf(
			"Print() returned unexpected error: %v",
			err,
		)
	}

	if printer.openCount != 1 {
		t.Fatalf(
			"expected 1 open attempt, got %d",
			printer.openCount,
		)
	}

	if printer.writeCount != 1 {
		t.Fatalf(
			"expected 1 write attempt, got %d",
			printer.writeCount,
		)
	}

	if !printer.closed {
		t.Fatal("expected printer to be closed")
	}
}

func TestPrintClosesPrinter(t *testing.T) {
	printer := &failingPrinter{}

	renderer := NewRenderer()

	err := Print(
		printer,
		renderer,
		validPrintReceipt(),
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
		validPrintReceipt(),
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

func TestPrintRejectsInvalidReceipt(t *testing.T) {
	printer := &failingPrinter{}
	renderer := NewRenderer()

	input := validPrintReceipt()
	input.Transaction.InvoiceNumber = ""

	err := Print(
		printer,
		renderer,
		input,
	)

	if !errors.Is(err, ErrInvoiceRequired) {
		t.Fatalf(
			"expected ErrInvoiceRequired, got %v",
			err,
		)
	}

	if printer.openCount != 0 {
		t.Fatalf(
			"printer should not be opened for invalid receipt, got %d open attempts",
			printer.openCount,
		)
	}

	if printer.writeCount != 0 {
		t.Fatalf(
			"printer should not be written to for invalid receipt, got %d writes",
			printer.writeCount,
		)
	}
}
