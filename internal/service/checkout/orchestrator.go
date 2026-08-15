package checkout

import (
	"context"
	"log"

	domainreceipt "pos-system/internal/domain/receipt"
	domaintransaction "pos-system/internal/domain/transaction"
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
	tx, err := s.atomic.Execute(ctx, req)
	if err != nil {
		return OrchestratorResult{}, err
	}

	job, err := s.printService.CreateJob(
		tx,
		"PJ-"+tx.ID,
	)
	if err != nil {
		log.Printf(
			"checkout print job creation failed for transaction %s: %v",
			tx.ID,
			err,
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
		log.Printf(
			"checkout print job start failed for transaction %s: %v",
			tx.ID,
			err,
		)
		_ = job.Fail(err)
		return OrchestratorResult{
			Transaction: tx,
			PrintJob:    job,
		}, nil
	}

	printResp, err := s.printAgent.Print(
		ctx,
		receipt,
		tx.ID,
	)
	if err != nil {
		log.Printf(
			"checkout print failed for transaction %s invoice %s: %v",
			tx.ID,
			tx.InvoiceNumber,
			err,
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
		log.Printf(
			"checkout print job complete failed for transaction %s: %v",
			tx.ID,
			err,
		)
		_ = job.Fail(err)
	}

	log.Printf(
		"checkout print completed for transaction %s job_id=%s",
		tx.ID,
		job.ID,
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
