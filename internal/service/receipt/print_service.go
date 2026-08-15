package receipt

import (
	"errors"

	domaintransaction "pos-system/internal/domain/transaction"
	printerreceipt "pos-system/internal/printer/receipt"
)

var ErrTransactionNotCompleted = errors.New(
	"transaction is not completed",
)

type PrintService struct{}

func NewPrintService() *PrintService {
	return &PrintService{}
}

func (s *PrintService) CreateJob(
	tx domaintransaction.Transaction,
	jobID string,
) (printerreceipt.PrintJob, error) {
	if tx.Status != domaintransaction.StatusCompleted {
		return printerreceipt.PrintJob{}, ErrTransactionNotCompleted
	}

	r := FromTransaction(tx)

	return printerreceipt.NewPrintJob(
		jobID,
		r,
	), nil
}
