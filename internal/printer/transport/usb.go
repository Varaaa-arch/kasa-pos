package transport

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// ErrPrintTimeout is returned when a write to the printer device exceeds
// the configured WriteTimeout. The printer is likely stuck or offline.
var ErrPrintTimeout = errors.New("print timeout: printer did not respond in time")

// DefaultWriteTimeout is the per-chunk write deadline applied to each
// os.File.Write call. 5 seconds is generous for a 64-byte chunk —
// if the printer hasn't accepted it by then, it's stuck.
const DefaultWriteTimeout = 5 * time.Second

type USBPrinter struct {
	devicePath   string
	file         *os.File
	WriteTimeout time.Duration // 0 means no deadline
}

// Compile-time check: USBPrinter must implement Printer.
var _ Printer = (*USBPrinter)(nil)

func NewUSBPrinter(devicePath string) *USBPrinter {
	return &USBPrinter{
		devicePath:   devicePath,
		WriteTimeout: DefaultWriteTimeout,
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

		// Set per-chunk write deadline so a stuck printer is detected fast.
		if p.WriteTimeout > 0 {
			deadline := time.Now().Add(p.WriteTimeout)
			if err := p.file.SetWriteDeadline(deadline); err != nil {
				// SetWriteDeadline may fail on non-network files on some OS.
				// Log and skip — better than crashing.
				_ = err
			}
		}

		n, err := p.file.Write(data[total:end])
		if err != nil {
			if isTimeout(err) {
				return total, fmt.Errorf("%w: %s", ErrPrintTimeout, err.Error())
			}
			return total, fmt.Errorf("failed to write to printer: %w", err)
		}

		total += n

		// Give the printer time to process the current chunk.
		time.Sleep(chunkDelay)
	}

	// Clear deadline after successful write.
	if p.WriteTimeout > 0 && p.file != nil {
		_ = p.file.SetWriteDeadline(time.Time{})
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
		return fmt.Errorf("failed to close printer: %w", err)
	}

	return nil
}

// isTimeout reports whether err is an OS-level timeout error
// (syscall.ETIMEDOUT wrapped in *os.PathError, os.ErrDeadlineExceeded, etc).
func isTimeout(err error) bool {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	var timeoutErr interface{ Timeout() bool }
	if errors.As(err, &timeoutErr) {
		return timeoutErr.Timeout()
	}
	return false
}
