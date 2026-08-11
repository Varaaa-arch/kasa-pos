package transport

import (
	"fmt"
	"os"
)

type USBPrinter struct {
	devicePath string
	file       *os.File
}

func NewUSBPrinter(devicePath string) *USBPrinter {
	return &USBPrinter{
		devicePath: devicePath,
	}
}

func (p *USBPrinter) Open() error {
	file, err := os.OpenFile(
		p.devicePath,
		os.O_WRONLY,
		0,
	)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	p.file = file
	return nil
}

func (p *USBPrinter) Write(data []byte) (int, error) {
	if p.file == nil {
		return 0, fmt.Errorf("printer is not open")
	}
	n, err := p.file.Write(data)
	if err != nil {
		return n, fmt.Errorf("Failed to write to printer: %w", err)
	}
	return n, nil
}

func (p *USBPrinter) Close() error {
	if p.file == nil {
		return nil
	}
	err := p.file.Close()
	p.file = nil

	if err != nil {
		return fmt.Errorf("failed to close printer: %w", err)
	}
	return nil
}
