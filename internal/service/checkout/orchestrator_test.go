package checkout

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domainreceipt "pos-system/internal/domain/receipt"
	domainproduct "pos-system/internal/domain/product"
	domaintransaction "pos-system/internal/domain/transaction"
	"pos-system/internal/printer/agent"
	printerreceipt "pos-system/internal/printer/receipt"
	receiptsvc "pos-system/internal/service/receipt"
)

type mockPrintAgent struct {
	resp          agent.PrintResponse
	err           error
	calls         int
	lastReceipt   domainreceipt.Receipt
	lastKey       string
}

func (m *mockPrintAgent) Print(
	_ context.Context,
	receipt domainreceipt.Receipt,
	idempotencyKey string,
) (agent.PrintResponse, error) {
	m.calls++
	m.lastReceipt = receipt
	m.lastKey = idempotencyKey
	return m.resp, m.err
}

func TestOrchestratorPrintSuccess(t *testing.T) {
	atomic := &stubAtomicService{}
	printAgent := &mockPrintAgent{
		resp: agent.PrintResponse{
			JobID:   "PJ-remote-001",
			Message: "receipt printed successfully",
		},
	}

	orchestrator := NewOrchestratorService(
		atomic,
		receiptsvc.NewPrintService(),
		printAgent,
		DefaultReceiptDefaults(),
	)

	tx := domaintransaction.Transaction{
		ID:            uuid.NewString(),
		InvoiceNumber: "INV-ORCH-001",
		Status:        domaintransaction.StatusCompleted,
		CreatedAt:     time.Now().UTC(),
		Subtotal:      30000,
		Total:         30000,
		PaidAmount:    50000,
		Change:        20000,
		PaymentMethod: "CASH",
		Items: []domaintransaction.Item{
			{
				ProductID: "prod-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: 15000,
				Subtotal:  30000,
			},
		},
	}

	atomic.executeFn = func(
		_ context.Context,
		_ AtomicRequest,
	) (domaintransaction.Transaction, error) {
		return tx, nil
	}

	result, err := orchestrator.Execute(
		context.Background(),
		AtomicRequest{
			PaidAmount: 50000,
		},
	)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if result.Transaction.ID != tx.ID {
		t.Fatalf("unexpected transaction id")
	}

	if result.PrintJob.Status != printerreceipt.PrintJobCompleted {
		t.Fatalf("print job status = %q, want COMPLETED", result.PrintJob.Status)
	}

	if printAgent.calls != 1 {
		t.Fatalf("print agent calls = %d, want 1", printAgent.calls)
	}

	if printAgent.lastKey != tx.ID {
		t.Fatalf("unexpected idempotency key: %q", printAgent.lastKey)
	}

	if printAgent.lastReceipt.Store.Name != "TOKO KASA" {
		t.Fatalf("store not enriched: %+v", printAgent.lastReceipt.Store)
	}

	if printAgent.lastReceipt.Footer.Message != "Terima kasih" {
		t.Fatalf("footer not enriched: %+v", printAgent.lastReceipt.Footer)
	}
}

func TestOrchestratorPrintFailureKeepsTransaction(t *testing.T) {
	atomic := &stubAtomicService{}
	printAgent := &mockPrintAgent{
		err: errors.New("print agent unavailable"),
	}

	orchestrator := NewOrchestratorService(
		atomic,
		receiptsvc.NewPrintService(),
		printAgent,
		DefaultReceiptDefaults(),
	)

	tx := domaintransaction.Transaction{
		ID:            uuid.NewString(),
		InvoiceNumber: "INV-ORCH-002",
		Status:        domaintransaction.StatusCompleted,
		CreatedAt:     time.Now().UTC(),
		Subtotal:      15000,
		Total:         15000,
		PaidAmount:    20000,
		Change:        5000,
		PaymentMethod: "CASH",
		Items: []domaintransaction.Item{
			{
				ProductID: "prod-002",
				SKU:       "ROTI-001",
				Name:      "Roti",
				Quantity:  1,
				UnitPrice: 15000,
				Subtotal:  15000,
			},
		},
	}

	atomic.executeFn = func(
		_ context.Context,
		_ AtomicRequest,
	) (domaintransaction.Transaction, error) {
		return tx, nil
	}

	result, err := orchestrator.Execute(
		context.Background(),
		AtomicRequest{
			PaidAmount: 20000,
		},
	)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if result.Transaction.Status != domaintransaction.StatusCompleted {
		t.Fatalf("transaction status = %q, want COMPLETED", result.Transaction.Status)
	}

	if result.PrintJob.Status != printerreceipt.PrintJobFailed {
		t.Fatalf("print job status = %q, want FAILED", result.PrintJob.Status)
	}

	if result.PrintJob.Error == "" {
		t.Fatal("expected print job error")
	}
}

type stubAtomicService struct {
	executeFn func(context.Context, AtomicRequest) (domaintransaction.Transaction, error)
}

func (s *stubAtomicService) Execute(
	ctx context.Context,
	req AtomicRequest,
) (domaintransaction.Transaction, error) {
	if s.executeFn != nil {
		return s.executeFn(ctx, req)
	}
	return domaintransaction.Transaction{}, errors.New("not implemented")
}

// Ensure stub compiles against orchestrator usage by shadowing AtomicService method in tests only.
func init() {
	_ = domainproduct.Product{}
}
