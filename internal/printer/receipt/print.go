package receipt

import "pos-system/internal/printer/transport"

func Print(
	printer transport.Printer,
	renderer *Renderer,
	receipt Receipt,
) error {
	if err := printer.Open(); err != nil {
		return err
	}

	defer printer.Close()

	data := renderer.Render(receipt)

	_, err := printer.Write(data)
	if err != nil {
		return err
	}

	return nil
}
