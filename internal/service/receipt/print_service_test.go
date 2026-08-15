package receipt

import (
	"errors"
	"testing"
	"time"

	domaintransaction "pos-system/internal/domain/transaction"
	printerreceipt "pos-system/internal/printer/receipt"
)

func TestCreatePrintJobFromTransaction(t *testing.T) {
	tx := domaintransaction.Transaction{
		ID:            "trx-001",
		InvoiceNumber: "INV-001",
		Status:        domaintransaction.StatusCompleted,
		CreatedAt:     time.Now().UTC(),
		Subtotal:      30000,
		Total:         30000,
		PaidAmount:    50000,
		Change:        20000,
		PaymentMethod: "CASH",
		Items: []domaintransaction.Item{
			{
				ID:        "item-001",
				ProductID: "prod-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
				Subtotal:  30000,
			},
		},
	}

	job, err := NewPrintService().CreateJob(
		tx,
		"PJ-001",
	)

	if err != nil {
		t.Fatal(err)
	}

	if job.ID != "PJ-001" {
		t.Fatalf("unexpected job ID: %q", job.ID)
	}

	if job.Status != printerreceipt.PrintJobPending {
		t.Fatalf(
			"expected PENDING, got %q",
			job.Status,
		)
	}

	if job.Receipt.Transaction.InvoiceNumber != "INV-001" {
		t.Fatal("invoice mismatch")
	}

	if job.Receipt.Summary.Total.Amount != 30000 {
		t.Fatalf("unexpected receipt total")
	}

	if job.Receipt.Payment.Change.Amount != 20000 {
		t.Fatalf("unexpected receipt change")
	}
}

func TestCreatePrintJobRejectsIncompleteTransaction(t *testing.T) {
	tx := domaintransaction.Transaction{
		ID:            "trx-002",
		InvoiceNumber: "INV-002",
		Status:        domaintransaction.StatusFailed,
	}

	_, err := NewPrintService().CreateJob(
		tx,
		"PJ-002",
	)

	if !errors.Is(err, ErrTransactionNotCompleted) {
		t.Fatalf(
			"expected ErrTransactionNotCompleted, got %v",
			err,
		)
	}
}
