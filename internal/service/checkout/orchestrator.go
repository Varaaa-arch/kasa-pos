package checkout

import (
	"context"
	"log/slog"

	domainreceipt "pos-system/internal/domain/receipt"
	domaintransaction "pos-system/internal/domain/transaction"
	applogger "pos-system/internal/logger"
	"pos-system/internal/printer/agent"
	printerreceipt "pos-system/internal/printer/receipt"
	receiptsvc "pos-system/internal/service/receipt"
)

type ReceiptDefaults struct {
	Store  domainreceipt.Store
	Footer domainreceipt.Footer
}

func DefaultReceiptDefaults() ReceiptDefaults {
	return ReceiptDefaults{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},
		Footer: domainreceipt.Footer{
			Message: "Terima kasih",
		},
	}
}

type AtomicExecutor interface {
	Execute(
		ctx context.Context,
		req AtomicRequest,
	) (domaintransaction.Transaction, error)
}

type OrchestratorResult struct {
	Transaction domaintransaction.Transaction
	PrintJob    printerreceipt.PrintJob
}

type OrchestratorService struct {
	atomic       AtomicExecutor
	printService *receiptsvc.PrintService
	printAgent   agent.PrintAgentClient
	defaults     ReceiptDefaults
}

func NewOrchestratorService(
	atomic AtomicExecutor,
	printService *receiptsvc.PrintService,
	printAgent agent.PrintAgentClient,
	defaults ReceiptDefaults,
) *OrchestratorService {
	return &OrchestratorService{
		atomic:       atomic,
		printService: printService,
		printAgent:   printAgent,
		defaults:     defaults,
	}
}

func (s *OrchestratorService) Execute(
	ctx context.Context,
	req AtomicRequest,
) (OrchestratorResult, error) {
	reqID := applogger.RequestIDFromContext(ctx)

	slog.InfoContext(ctx, applogger.EventCheckoutStarted,
		"event", applogger.EventCheckoutStarted,
		"request_id", reqID,
	)

	tx, err := s.atomic.Execute(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, applogger.EventCheckoutFailed,
			"event", applogger.EventCheckoutFailed,
			"request_id", reqID,
			"error", err.Error(),
		)
		return OrchestratorResult{}, err
	}

	slog.InfoContext(ctx, applogger.EventCheckoutCompleted,
		"event", applogger.EventCheckoutCompleted,
		"request_id", reqID,
		"transaction_id", tx.ID,
		"invoice_number", tx.InvoiceNumber,
		"total", tx.Total,
	)

	slog.InfoContext(ctx, applogger.EventPrintStarted,
		"event", applogger.EventPrintStarted,
		"request_id", reqID,
		"transaction_id", tx.ID,
	)

	job, err := s.printService.CreateJob(tx, "PJ-"+tx.ID)
	if err != nil {
		slog.ErrorContext(ctx, applogger.EventPrintFailed,
			"event", applogger.EventPrintFailed,
			"request_id", reqID,
			"transaction_id", tx.ID,
			"error", err.Error(),
		)
		return OrchestratorResult{
			Transaction: tx,
			PrintJob: printerreceipt.PrintJob{
				ID:     "PJ-" + tx.ID,
				Status: printerreceipt.PrintJobFailed,
				Error:  err.Error(),
			},
		}, nil
	}

	receipt := enrichReceipt(job.Receipt, s.defaults)

	if err := job.Start(); err != nil {
		slog.ErrorContext(ctx, applogger.EventPrintFailed,
			"event", applogger.EventPrintFailed,
			"request_id", reqID,
			"transaction_id", tx.ID,
			"error", err.Error(),
		)
		_ = job.Fail(err)
		return OrchestratorResult{
			Transaction: tx,
			PrintJob:    job,
		}, nil
	}

	printResp, err := s.printAgent.Print(ctx, receipt, tx.ID)
	if err != nil {
		slog.ErrorContext(ctx, applogger.EventPrintFailed,
			"event", applogger.EventPrintFailed,
			"request_id", reqID,
			"transaction_id", tx.ID,
			"invoice_number", tx.InvoiceNumber,
			"error", err.Error(),
		)
		_ = job.Fail(err)
		return OrchestratorResult{
			Transaction: tx,
			PrintJob:    job,
		}, nil
	}

	if printResp.JobID != "" {
		job.ID = printResp.JobID
	}

	if err := job.Complete(); err != nil {
		slog.ErrorContext(ctx, applogger.EventPrintFailed,
			"event", applogger.EventPrintFailed,
			"request_id", reqID,
			"transaction_id", tx.ID,
			"error", err.Error(),
		)
		_ = job.Fail(err)
	}

	slog.InfoContext(ctx, applogger.EventPrintCompleted,
		"event", applogger.EventPrintCompleted,
		"request_id", reqID,
		"transaction_id", tx.ID,
		"print_job_id", job.ID,
	)

	return OrchestratorResult{
		Transaction: tx,
		PrintJob:    job,
	}, nil
}

func enrichReceipt(
	receipt domainreceipt.Receipt,
	defaults ReceiptDefaults,
) domainreceipt.Receipt {
	if receipt.Store.Name == "" {
		receipt.Store = defaults.Store
	}

	if receipt.Footer.Message == "" {
		receipt.Footer = defaults.Footer
	}

	return receipt
}
