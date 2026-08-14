package receipt

import (
	domainreceipt "pos-system/internal/domain/receipt"

	"pos-system/internal/printer/retry"
	"pos-system/internal/printer/transport"
)

func Print(
	printer transport.Printer,
	renderer *Renderer,
	input domainreceipt.Receipt,
) error {
	// Validate receipt before touching the printer.
	validator := NewValidator()

	if err := validator.Validate(input); err != nil {
		return err
	}

	// Retry printer opening because this operation may
	// temporarily fail during reconnect/startup.
	config := retry.DefaultConfig()

	err := retry.Do(
		config,
		func() error {
			return printer.Open()
		},
	)

	if err != nil {
		return err
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
		return err
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
	printer transport.Printer,
	renderer *Renderer,
) error {
	if err := j.Start(); err != nil {
		return err
	}

	if err := Print(
		printer,
		renderer,
		j.Receipt,
	); err != nil {
		_ = j.Fail(err)
		return err
	}

	if err := j.Complete(); err != nil {
		return err
	}

	return nil
}
