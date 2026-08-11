package escpos

import "bytes"

type ESCPosPrinter struct {
	buf bytes.Buffer
}

func NewESCPosPrinter() *ESCPosPrinter {
	return &ESCPosPrinter{}
}

func (p *ESCPosPrinter) Bytes() []byte {
	return p.buf.Bytes()
}

func (p *ESCPosPrinter) Initialize() {
	p.buf.Write([]byte{0x1B, 0x40})
}

func (p *ESCPosPrinter) Text(s string) {
	p.buf.WriteString(s)
}

func (p *ESCPosPrinter) LF() {
	p.buf.WriteByte(0x0A)
}

func (p *ESCPosPrinter) Bold(enable bool) {
	if enable {
		p.buf.Write([]byte{0x1B, 0x45, 0x01})
	} else {
		p.buf.Write([]byte{0x1B, 0x45, 0x00})
	}
}

func (p *ESCPosPrinter) AlignLeft() {
	p.buf.Write([]byte{0x1B, 0x61, 0x00})
}

func (p *ESCPosPrinter) AlignCenter() {
	p.buf.Write([]byte{0x1B, 0x61, 0x01})
}

func (p *ESCPosPrinter) AlignRight() {
	p.buf.Write([]byte{0x1B, 0x61, 0x02})
}

// FontSize sets width/height multiplier (1-8). ESC/POS encodes as
// (width-1) in the high nibble and (height-1) in the low nibble.
func (p *ESCPosPrinter) FontSize(width, height int) {
	n := byte((width-1)<<4 | (height - 1))
	p.buf.Write([]byte{0x1D, 0x21, n})
}

func (p *ESCPosPrinter) Feed(n int) {
	p.buf.Write([]byte{0x1B, 0x64, byte(n)})
}
