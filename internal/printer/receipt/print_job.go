package receipt

import (
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

type PrintJobStatus string

const (
	PrintJobPending   PrintJobStatus = "PENDING"
	PrintJobPrinting  PrintJobStatus = "PRINTING"
	PrintJobCompleted PrintJobStatus = "COMPLETED"
	PrintJobFailed    PrintJobStatus = "FAILED"
)

type PrintJob struct {
	ID        string
	Receipt   domainreceipt.Receipt
	Status    PrintJobStatus
	CreatedAt time.Time
	PrintedAt *time.Time
}

func NewPrintJob(
	id string,
	input domainreceipt.Receipt,
) PrintJob {
	return PrintJob{
		ID:        id,
		Receipt:   input,
		Status:    PrintJobPending,
		CreatedAt: time.Now(),
	}
}

func (j *PrintJob) Start() {
	j.Status = PrintJobPrinting
}

func (j *PrintJob) Complete() {
	now := time.Now()

	j.Status = PrintJobCompleted
	j.PrintedAt = &now
}

func (j *PrintJob) Fail() {
	j.Status = PrintJobFailed
}
