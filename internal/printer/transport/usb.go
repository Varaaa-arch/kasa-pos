package transport

import (
	"fmt"
	"os"
	"time"
)

type USBPrinter struct {
	devicePath string
	file       *os.File
}

// Compile-time check:
// memastikan USBPrinter mengimplementasikan Printer.
var _ Printer = (*USBPrinter)(nil)

func NewUSBPrinter(devicePath string) *USBPrinter {
	return &USBPrinter{
		devicePath: devicePath,
	}
}

func (p *USBPrinter) Open() error {
	if p.file != nil {
		return fmt.Errorf("printer is already open")
	}

	file, err := os.OpenFile(
		p.devicePath,
		os.O_WRONLY,
		0,
	)
	if err != nil {
		return fmt.Errorf("open printer device: %w", err)
	}

	p.file = file

	return nil
}

func (p *USBPrinter) Write(data []byte) (int, error) {
	if p.file == nil {
		return 0, fmt.Errorf("printer is not open")
	}

	const chunkSize = 64
	const chunkDelay = 10 * time.Millisecond

	total := 0

	for total < len(data) {
		end := total + chunkSize

		if end > len(data) {
			end = len(data)
		}

		n, err := p.file.Write(data[total:end])
		if err != nil {
			return total, fmt.Errorf(
				"failed to write to printer: %w",
				err,
			)
		}

		total += n

		// Give the printer time to process the current chunk.
		time.Sleep(chunkDelay)
	}

	return total, nil
}

func (p *USBPrinter) Close() error {
	if p.file == nil {
		return nil
	}

	err := p.file.Close()
	p.file = nil

	if err != nil {
		return fmt.Errorf(
			"failed to close printer: %w",
			err,
		)
	}

	return nil
}
