package receipt

import (
	"errors"
	"fmt"

	domainreceipt "pos-system/internal/domain/receipt"
	"pos-system/internal/printer/retry"
)

type Logger interface {
	Printf(format string, args ...interface{})
}

func Print(
	printer PrintTarget,
	renderer *Renderer,
	input domainreceipt.Receipt,
	logger Logger,
) error {
	// Validate receipt before touching the printer.
	validator := NewValidator()

	if err := validator.Validate(input); err != nil {
		// Wrap validation errors as non-retryable
		var validationErr ValidationError
		if errors.As(err, &validationErr) {
			return fmt.Errorf("%w: %v", retry.ErrValidation, err)
		}
		return err
	}

	// Retry printer opening because this operation may
	// temporarily fail during reconnect/startup.
	config := retry.DefaultConfig()

	// Add retry logging if logger is provided
	if logger != nil {
		config.Logger = func(attempt int, err error) {
			logger.Printf(
				"printer open retry: attempt=%d error=%v",
				attempt,
				err,
			)
		}
	}

	// Use default retryable error classification
	config.IsRetryable = retry.DefaultIsRetryable

	err := retry.Do(
		config,
		func() error {
			return printer.Open()
		},
	)

	if err != nil {
		return fmt.Errorf("printer open failed after retries: %w", err)
	}

	defer printer.Close()

	// Render only after validation and successful printer open.
	data := renderer.Render(input)

	// Write exactly once.
	//
	// Do not automatically retry Write because the printer may
	// have already received some/all of the receipt data.
	_, err = printer.Write(data)
	if err != nil {
		return fmt.Errorf("printer write failed: %w", err)
	}

	return nil
}

func CreatePrintJob(
	id string,
	input domainreceipt.Receipt,
) PrintJob {
	return NewPrintJob(id, input)
}

func (j *PrintJob) Run(
	printer PrintTarget,
	renderer *Renderer,
	logger Logger,
) error {
	if err := j.Start(); err != nil {
		return err
	}

	if err := Print(
		printer,
		renderer,
		j.Receipt,
		logger,
	); err != nil {
		_ = j.Fail(err)
		return err
	}

	if err := j.Complete(); err != nil {
		return err
	}

	return nil
}
