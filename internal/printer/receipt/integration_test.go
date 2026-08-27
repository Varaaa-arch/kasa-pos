package receipt

import (
	"bytes"
	"testing"

	domainreceipt "pos-system/internal/domain/receipt"
	"pos-system/internal/printer/mock"
)

func TestTransactionToReceiptToPrinter(t *testing.T) {
	input := validPrintReceipt()

	job := NewPrintJob(
		"PJ-INTEGRATION-001",
		input,
	)

	printer := &mock.Printer{}
	renderer := NewRenderer()

	err := job.Run(
		printer,
		renderer,
		nil,
	)
	if err != nil {
		t.Fatalf("job.Run() error = %v", err)
	}

	if job.Status != PrintJobCompleted {
		t.Fatalf(
			"expected COMPLETED, got %q",
			job.Status,
		)
	}

	if printer.OpenCount != 1 {
		t.Fatalf(
			"expected 1 open, got %d",
			printer.OpenCount,
		)
	}

	if printer.WriteCount != 1 {
		t.Fatalf(
			"expected 1 write, got %d",
			printer.WriteCount,
		)
	}

	if printer.CloseCount != 1 {
		t.Fatalf(
			"expected 1 close, got %d",
			printer.CloseCount,
		)
	}

	if len(printer.Data) == 0 {
		t.Fatal("expected printer data")
	}

	expectedTexts := []string{
		"TOKO KASA",
		"INV-PRINT-001",
		"Kopi Susu",
		"Rp30.000",
		"Rp50.000",
		"Rp20.000",
		"TEST PRINT",
	}

	for _, text := range expectedTexts {
		if !bytes.Contains(
			printer.Data,
			[]byte(text),
		) {
			t.Fatalf(
				"printed data missing %q",
				text,
			)
		}
	}
}

func TestTransactionToReceiptValidationFailure(
	t *testing.T,
) {
	input := validPrintReceipt()

	input.Transaction.InvoiceNumber = ""

	job := NewPrintJob(
		"PJ-INTEGRATION-002",
		input,
	)

	printer := &mock.Printer{}
	renderer := NewRenderer()

	err := job.Run(
		printer,
		renderer,
		nil,
	)

	if err == nil {
		t.Fatal("expected validation error")
	}

	if job.Status != PrintJobFailed {
		t.Fatalf(
			"expected FAILED, got %q",
			job.Status,
		)
	}

	if printer.OpenCount != 0 {
		t.Fatalf(
			"printer should not open, got %d",
			printer.OpenCount,
		)
	}

	if printer.WriteCount != 0 {
		t.Fatalf(
			"printer should not write, got %d",
			printer.WriteCount,
		)
	}
}

func TestTransactionToReceiptUsesDomainData(
	t *testing.T,
) {
	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name: "INTEGRATION STORE",
		},

		Transaction: domainreceipt.Transaction{
			ID:            "TXN-INTEGRATION-001",
			InvoiceNumber: "INV-INTEGRATION-001",
			Timestamp:     validPrintReceipt().Transaction.Timestamp,
			Cashier:       "Integration",
		},

		Items: []domainreceipt.Item{
			{
				Name:     "Produk Integration",
				Quantity: 2,
				UnitPrice: domainreceipt.NewMoney(
					25000,
					domainreceipt.IDR,
				),
			},
		},

		Summary: domainreceipt.Summary{
			Subtotal: domainreceipt.NewMoney(
				50000,
				domainreceipt.IDR,
			),
			Total: domainreceipt.NewMoney(
				50000,
				domainreceipt.IDR,
			),
		},

		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid: domainreceipt.NewMoney(
				100000,
				domainreceipt.IDR,
			),
			Change: domainreceipt.NewMoney(
				50000,
				domainreceipt.IDR,
			),
		},

		Footer: domainreceipt.Footer{
			Message: "INTEGRATION TEST",
		},
	}

	job := NewPrintJob(
		"PJ-INTEGRATION-003",
		input,
	)

	printer := &mock.Printer{}
	renderer := NewRenderer()

	if err := job.Run(
		printer,
		renderer,
		nil,
	); err != nil {
		t.Fatalf(
			"job.Run() error = %v",
			err,
		)
	}

	if !bytes.Contains(
		printer.Data,
		[]byte("INTEGRATION STORE"),
	) {
		t.Fatal("store name missing")
	}

	if !bytes.Contains(
		printer.Data,
		[]byte("INV-INTEGRATION-001"),
	) {
		t.Fatal("invoice missing")
	}

	if !bytes.Contains(
		printer.Data,
		[]byte("Produk Integration"),
	) {
		t.Fatal("item missing")
	}

	if !bytes.Contains(
		printer.Data,
		[]byte("Rp50.000"),
	) {
		t.Fatal("total missing")
	}
}
