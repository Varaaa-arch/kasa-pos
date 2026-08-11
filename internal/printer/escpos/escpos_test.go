package escpos

import (
	"bytes"
	"testing"
)

func TestInitialize(t *testing.T) {
	expected := []byte{
		0x1B, 0x40,
	}

	if !bytes.Equal(Initialize(), expected) {
		t.Fatalf(
			"unexpected initialize command: % X",
			Initialize(),
		)
	}
}

func TestText(t *testing.T) {
	expected := []byte("HELLO WORLD")

	if !bytes.Equal(Text("HELLO WORLD"), expected) {
		t.Fatalf(
			"unexpected text output: % X",
			Text("HELLO WORLD"),
		)
	}
}

func TestLF(t *testing.T) {
	expected := []byte{
		0x0A,
	}

	if !bytes.Equal(LF(), expected) {
		t.Fatalf(
			"unexpected LF command: % X",
			LF(),
		)
	}
}

func TestBold(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected []byte
	}{
		{
			name:    "enable",
			enabled: true,
			expected: []byte{
				0x1B, 0x45, 0x01,
			},
		},
		{
			name:    "disable",
			enabled: false,
			expected: []byte{
				0x1B, 0x45, 0x00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !bytes.Equal(Bold(tt.enabled), tt.expected) {
				t.Fatalf(
					"unexpected bold command: % X",
					Bold(tt.enabled),
				)
			}
		})
	}
}

func TestAlignment(t *testing.T) {
	tests := []struct {
		name     string
		command  []byte
		expected []byte
	}{
		{
			name:    "left",
			command: AlignLeft(),
			expected: []byte{
				0x1B, 0x61, 0x00,
			},
		},
		{
			name:    "center",
			command: AlignCenter(),
			expected: []byte{
				0x1B, 0x61, 0x01,
			},
		},
		{
			name:    "right",
			command: AlignRight(),
			expected: []byte{
				0x1B, 0x61, 0x02,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !bytes.Equal(tt.command, tt.expected) {
				t.Fatalf(
					"unexpected alignment command: % X",
					tt.command,
				)
			}
		})
	}
}

func TestFontSize(t *testing.T) {
	tests := []struct {
		name     string
		width    byte
		height   byte
		expected []byte
	}{
		{
			name:   "normal",
			width:  1,
			height: 1,
			expected: []byte{
				0x1D, 0x21, 0x00,
			},
		},
		{
			name:   "double",
			width:  2,
			height: 2,
			expected: []byte{
				0x1D, 0x21, 0x11,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !bytes.Equal(
				FontSize(tt.width, tt.height),
				tt.expected,
			) {
				t.Fatalf(
					"unexpected font size command: % X",
					FontSize(tt.width, tt.height),
				)
			}
		})
	}
}

func TestFeed(t *testing.T) {
	expected := []byte{
		0x1B, 0x64, 0x03,
	}

	if !bytes.Equal(Feed(3), expected) {
		t.Fatalf(
			"unexpected feed command: % X",
			Feed(3),
		)
	}
}
