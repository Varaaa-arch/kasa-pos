package receipt

import (
	"testing"
	"time"

	domaintransaction "pos-system/internal/domain/transaction"
	"pos-system/internal/printer/mock"
	printerreceipt "pos-system/internal/printer/receipt"
)

func TestTransactionToPrintPipeline(t *testing.T) {
	tx := domaintransaction.Transaction{
		ID:            "trx-integration-001",
		InvoiceNumber: "INV-INTEGRATION-001",
		Status:        domaintransaction.StatusCompleted,
		CreatedAt:     time.Now().UTC(),

		Subtotal:      30000,
		Discount:      0,
		Tax:           0,
		ServiceCharge: 0,
		Total:         30000,

		PaidAmount:    50000,
		Change:        20000,
		PaymentMethod: "CASH",

		Items: []domaintransaction.Item{
			{
				ID:            "item-001",
				TransactionID: "trx-integration-001",
				ProductID:     "prod-001",
				SKU:           "KOPI-001",
				Name:          "Kopi Susu",
				Quantity:      2,
				UnitPrice:     15000,
				Subtotal:      30000,
			},
		},
	}

	service := NewPrintService()

	job, err := service.CreateJob(
		tx,
		"PJ-INTEGRATION-001",
	)
	if err != nil {
		t.Fatal(err)
	}

	if job.Status != printerreceipt.PrintJobPending {
		t.Fatalf(
			"expected PENDING, got %q",
			job.Status,
		)
	}

	printer := &mock.Printer{}
	renderer := printerreceipt.NewRenderer()

	if err := job.Run(
		printer,
		renderer,
		nil,
	); err != nil {
		t.Fatalf(
			"print job failed: %v",
			err,
		)
	}

	if job.Status != printerreceipt.PrintJobCompleted {
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
}
