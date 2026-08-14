package receipt

import (
	"errors"
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
	ID          string
	Receipt     domainreceipt.Receipt
	Status      PrintJobStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	PrintedAt   *time.Time
	Error       string
}

var ErrInvalidPrintJobTransition = errors.New(
	"invalid print job status transition",
)

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

func (j *PrintJob) Start() error {
	if !CanTransition(
		j.Status,
		PrintJobPrinting,
	) {
		return ErrInvalidPrintJobTransition
	}

	now := time.Now()

	j.Status = PrintJobPrinting
	j.StartedAt = &now

	return nil
}

func (j *PrintJob) Complete() error {
	if !CanTransition(
		j.Status,
		PrintJobCompleted,
	) {
		return ErrInvalidPrintJobTransition
	}

	now := time.Now()

	j.Status = PrintJobCompleted
	j.CompletedAt = &now
	j.PrintedAt = &now

	return nil
}

func (j *PrintJob) Fail(err error) error {
	if !CanTransition(
		j.Status,
		PrintJobFailed,
	) {
		return ErrInvalidPrintJobTransition
	}

	now := time.Now()

	j.Status = PrintJobFailed
	j.CompletedAt = &now

	if err != nil {
		j.Error = err.Error()
	}

	return nil
}

func CanTransition(
	from PrintJobStatus,
	to PrintJobStatus,
) bool {
	switch from {
	case PrintJobPending:
		return to == PrintJobPrinting

	case PrintJobPrinting:
		return to == PrintJobCompleted ||
			to == PrintJobFailed

	case PrintJobCompleted, PrintJobFailed:
		return false

	default:
		return false
	}
}
