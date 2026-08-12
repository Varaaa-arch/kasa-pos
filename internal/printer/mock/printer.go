package mock

import (
	"errors"
)

type Printer struct {
	OpenErr  error
	WriteErr error
	CloseErr error

	OpenCount  int
	WriteCount int
	CloseCount int

	Opened bool
	Closed bool

	Data []byte
}

func (p *Printer) Open() error {
	p.OpenCount++

	if p.OpenErr != nil {
		return p.OpenErr
	}

	p.Opened = true
	p.Closed = false

	return nil
}

func (p *Printer) Write(data []byte) (int, error) {
	p.WriteCount++

	if p.WriteErr != nil {
		return 0, p.WriteErr
	}

	p.Data = append(p.Data, data...)

	return len(data), nil
}

func (p *Printer) Close() error {
	p.CloseCount++

	if p.CloseErr != nil {
		return p.CloseErr
	}

	p.Closed = true
	p.Opened = false

	return nil
}

func (p *Printer) Reset() {
	p.OpenErr = nil
	p.WriteErr = nil
	p.CloseErr = nil

	p.OpenCount = 0
	p.WriteCount = 0
	p.CloseCount = 0

	p.Opened = false
	p.Closed = false

	p.Data = nil
}

func (p *Printer) HasData() bool {
	return len(p.Data) > 0
}

func (p *Printer) LastError() error {
	if p.OpenErr != nil {
		return p.OpenErr
	}

	if p.WriteErr != nil {
		return p.WriteErr
	}

	if p.CloseErr != nil {
		return p.CloseErr
	}

	return nil
}

var (
	ErrOpen  = errors.New("mock printer open error")
	ErrWrite = errors.New("mock printer write error")
	ErrClose = errors.New("mock printer close error")
)
