package receipt

import (
	"bytes"
)

const (
	esc = 0x1B
	gs  = 0x1D
	lf  = 0x0A
)

type ESCPosPrinter struct {
	buffer bytes.Buffer
}

func NewESCPosPrinter() *ESCPosPrinter {
	e := &ESCPosPrinter{}
	e.Init()

	return e
}

func (e *ESCPosPrinter) Init() {
	e.buffer.Write([]byte{esc, 0x40}) // ESC @ - Initialize printer
}

func (e *ESCPosPrinter) AlignLeft() {
	e.buffer.Write([]byte{esc, 0x61, 0x00}) // ESC a 0 - Left alignment
}

func (e *ESCPosPrinter) AlignCenter() {
	e.buffer.Write([]byte{esc, 0x61, 0x01}) // ESC a 1 - Center alignment
}

func (e *ESCPosPrinter) AlignRight() {
	e.buffer.Write([]byte{esc, 0x61, 0x02}) // ESC a 2 - Right alignment
}

func (e *ESCPosPrinter) Bold(enabled bool) {
	value := byte(0)

	if enabled {
		value = 1
	}

	e.buffer.Write([]byte{esc, 0x45, value}) // ESC E n - Turn emphasized mode on/off
}

func (e *ESCPosPrinter) Text(text string) {
	e.buffer.WriteString(text)
}

func (e *ESCPosPrinter) LF() {
	e.buffer.WriteByte(lf) // LF - Line feed
}

func (e *ESCPosPrinter) Feed(lines int) {
	for i := 0; i < lines; i++ {
		e.LF()
	}
}

func (e *ESCPosPrinter) Bytes() []byte {
	return e.buffer.Bytes()
}
