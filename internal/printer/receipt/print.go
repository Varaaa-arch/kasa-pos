package receipt

import (
	"pos-system/internal/printer/retry"
	"pos-system/internal/printer/transport"
)

func Print(
	printer transport.Printer,
	renderer *Renderer,
	receipt Receipt,
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

	data := renderer.Render(receipt)

	_, err = printer.Write(data)
	if err != nil {
		return err
	}

	return nil
}
