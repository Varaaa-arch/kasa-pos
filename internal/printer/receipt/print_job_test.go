package receipt

import (
	"errors"
	"strings"
	"testing"
)

func TestPrintJobLifecycle(t *testing.T) {
	input := validPrintReceipt()

	job := NewPrintJob(
		"PJ-TEST-001",
		input,
	)

	if job.ID != "PJ-TEST-001" {
		t.Fatalf(
			"expected job ID PJ-TEST-001, got %q",
			job.ID,
		)
	}

	if job.Status != PrintJobPending {
		t.Fatalf(
			"expected PENDING, got %q",
			job.Status,
		)
	}

	if job.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}

	if job.PrintedAt != nil {
		t.Fatal("expected PrintedAt to be nil")
	}

	if err := job.Start(); err != nil {
		t.Fatal(err)
	}

	if job.Status != PrintJobPrinting {
		t.Fatalf(
			"expected PRINTING, got %q",
			job.Status,
		)

		if job.StartedAt == nil {
			t.Fatal("expected StartedAt to be set")
		}

		if job.CompletedAt != nil {
			t.Fatal("CompletedAt should be nil while printing")
		}
	}

	if err := job.Complete(); err != nil {
		t.Fatal(err)
	}

	if job.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set")
	}

	if job.CompletedAt.Before(*job.StartedAt) {
		t.Fatal("CompletedAt must not be before StartedAt")
	}

	if job.Status != PrintJobCompleted {
		t.Fatalf(
			"expected COMPLETED, got %q",
			job.Status,
		)
	}

	if job.PrintedAt == nil {
		t.Fatal("expected PrintedAt to be set")
	}
}

func TestPrintJobFailure(t *testing.T) {
	input := validPrintReceipt()

	job := NewPrintJob(
		"PJ-FAIL-001",
		input,
	)

	if err := job.Start(); err != nil {
		t.Fatal(err)
	}

	expectedErr := errors.New("printer unavailable")

	if err := job.Fail(expectedErr); err != nil {
		t.Fatal(err)
	}

	if job.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set for failed job")
	}

	if job.Error != "printer unavailable" {
		t.Fatalf(
			"expected error %q, got %q",
			"printer unavailable",
			job.Error,
		)
	}

	if job.Status != PrintJobFailed {
		t.Fatalf(
			"expected FAILED, got %q",
			job.Status,
		)
	}

	if job.PrintedAt != nil {
		t.Fatal(
			"expected PrintedAt to remain nil",
		)
	}
}

func TestPrintJobInvalidTransitions(t *testing.T) {
	job := NewPrintJob(
		"PJ-TRANSITION-001",
		validPrintReceipt(),
	)

	if err := job.Complete(); !errors.Is(
		err,
		ErrInvalidPrintJobTransition,
	) {
		t.Fatalf(
			"expected invalid transition error, got %v",
			err,
		)
	}

	if err := job.Start(); err != nil {
		t.Fatal(err)
	}

	if err := job.Start(); !errors.Is(
		err,
		ErrInvalidPrintJobTransition,
	) {
		t.Fatalf(
			"expected invalid transition error, got %v",
			err,
		)
	}

	if err := job.Complete(); err != nil {
		t.Fatal(err)
	}

	// Ditambahkan errors.New("test error") sebagai argumen
	if err := job.Fail(errors.New("test error")); !errors.Is(
		err,
		ErrInvalidPrintJobTransition,
	) {
		t.Fatalf(
			"expected invalid transition error, got %v",
			err,
		)
	}
}

func TestPrintJobFailureStoresError(t *testing.T) {
	job := NewPrintJob(
		"PJ-FAIL-002",
		validPrintReceipt(),
	)

	if err := job.Start(); err != nil {
		t.Fatal(err)
	}

	expectedErr := errors.New("printer disconnected")

	if err := job.Fail(expectedErr); err != nil {
		t.Fatal(err)
	}

	if job.Status != PrintJobFailed {
		t.Fatalf(
			"expected FAILED, got %q",
			job.Status,
		)
	}

	if job.Error != expectedErr.Error() {
		t.Fatalf(
			"expected error %q, got %q",
			expectedErr.Error(),
			job.Error,
		)
	}

	if job.PrintedAt != nil {
		t.Fatal("failed job should not have PrintedAt")
	}
}

func TestPrintJobRunSuccess(t *testing.T) {
	job := NewPrintJob(
		"PJ-RUN-001",
		validPrintReceipt(),
	)

	printer := &failingPrinter{}
	renderer := NewRenderer()

	err := job.Run(
		printer,
		renderer,
		nil,
	)

	if err != nil {
		t.Fatalf(
			"expected success, got %v",
			err,
		)
	}

	if job.Status != PrintJobCompleted {
		t.Fatalf(
			"expected COMPLETED, got %q",
			job.Status,
		)
	}

	if job.StartedAt == nil {
		t.Fatal("expected StartedAt")
	}

	if job.CompletedAt == nil {
		t.Fatal("expected CompletedAt")
	}

	if job.PrintedAt == nil {
		t.Fatal("expected PrintedAt")
	}

	if printer.writeCount != 1 {
		t.Fatalf(
			"expected 1 write, got %d",
			printer.writeCount,
		)
	}
}

func TestPrintJobRunFailure(t *testing.T) {
	expectedErr := errors.New("printer unavailable")

	job := NewPrintJob(
		"PJ-RUN-002",
		validPrintReceipt(),
	)

	printer := &failingPrinter{
		openErr: expectedErr,
	}

	renderer := NewRenderer()

	err := job.Run(
		printer,
		renderer,
		nil,
	)

	// The error is now wrapped, just check that it's not nil and contains relevant info
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "printer unavailable") {
		t.Fatalf("expected error to contain 'printer unavailable', got %v", err)
	}

	if job.Status != PrintJobFailed {
		t.Fatalf(
			"expected FAILED, got %q",
			job.Status,
		)
	}

	// The job.Error field contains the error message
	if job.Error == "" {
		t.Fatal("expected job.Error to be set")
	}
	if !strings.Contains(job.Error, "printer unavailable") {
		t.Fatalf("expected job.Error to contain 'printer unavailable', got %q", job.Error)
	}

	if job.CompletedAt == nil {
		t.Fatal("expected CompletedAt")
	}

	if job.PrintedAt != nil {
		t.Fatal("failed job should not have PrintedAt")
	}
}
