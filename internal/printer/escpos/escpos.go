package escpos

const (
	esc = 0x1B
	gs  = 0x1D
)

// Initialize resets the printer to its default state.
func Initialize() []byte {
	return []byte{
		esc, 0x40,
	}
}

// Text returns raw text data for the printer.
func Text(text string) []byte {
	return []byte(text)
}

// LF returns a line feed command.
func LF() []byte {
	return []byte{
		0x0A,
	}
}

// Bold enables or disables bold text.
func Bold(enabled bool) []byte {
	if enabled {
		return []byte{
			esc, 0x45, 0x01,
		}
	}

	return []byte{
		esc, 0x45, 0x00,
	}
}

// AlignLeft sets left alignment.
func AlignLeft() []byte {
	return []byte{
		esc, 0x61, 0x00,
	}
}

// AlignCenter sets center alignment.
func AlignCenter() []byte {
	return []byte{
		esc, 0x61, 0x01,
	}
}

// AlignRight sets right alignment.
func AlignRight() []byte {
	return []byte{
		esc, 0x61, 0x02,
	}
}

// FontSize sets the printer character size.
//
// width and height:
// 1 = normal
// 2 = double
func FontSize(width, height byte) []byte {
	if width < 1 {
		width = 1
	}

	if width > 8 {
		width = 8
	}

	if height < 1 {
		height = 1
	}

	if height > 8 {
		height = 8
	}

	size := ((width - 1) << 4) | (height - 1)

	return []byte{
		gs, 0x21, size,
	}
}

// Feed feeds n lines of paper.
func Feed(lines byte) []byte {
	return []byte{
		esc, 0x64, lines,
	}
}