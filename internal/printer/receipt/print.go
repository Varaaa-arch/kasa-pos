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

	data := renderer.Render(input)

	_, err = printer.Write(data)
	if err != nil {
		return err
	}

	return nil
}
