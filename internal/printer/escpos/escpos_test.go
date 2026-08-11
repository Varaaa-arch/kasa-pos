package escpos

import (
	"bytes"
	"testing"
)

func assertBytes(t *testing.T, got, expected []byte) {
	t.Helper()

	if !bytes.Equal(got, expected) {
		t.Fatalf(
			"unexpected bytes:\n got:  % X\n want: % X",
			got,
			expected,
		)
	}
}

func TestInitialize(t *testing.T) {
	expected := []byte{0x1B, 0x40}

	assertBytes(t, Initialize(), expected)
}

func TestText(t *testing.T) {
	expected := []byte("HELLO")

	assertBytes(t, Text("HELLO"), expected)
}

func TestLF(t *testing.T) {
	expected := []byte{0x0A}

	assertBytes(t, LF(), expected)
}

func TestTextAndLF(t *testing.T) {
	var data []byte
	data = append(data, Text("HELLO")...)
	data = append(data, LF()...)
	data = append(data, Text("WORLD")...)

	expected := []byte{
		'H', 'E', 'L', 'L', 'O',
		0x0A,
		'W', 'O', 'R', 'L', 'D',
	}

	assertBytes(t, data, expected)
}

func TestBoldEnable(t *testing.T) {
	expected := []byte{0x1B, 0x45, 0x01}

	assertBytes(t, Bold(true), expected)
}

func TestBoldDisable(t *testing.T) {
	expected := []byte{0x1B, 0x45, 0x00}

	assertBytes(t, Bold(false), expected)
}

func TestBoldText(t *testing.T) {
	var data []byte
	data = append(data, Bold(true)...)
	data = append(data, Text("TOTAL")...)
	data = append(data, Bold(false)...)

	expected := []byte{
		0x1B, 0x45, 0x01,
		'T', 'O', 'T', 'A', 'L',
		0x1B, 0x45, 0x00,
	}

	assertBytes(t, data, expected)
}

func TestAlignLeft(t *testing.T) {
	expected := []byte{0x1B, 0x61, 0x00}

	assertBytes(t, AlignLeft(), expected)
}

func TestAlignCenter(t *testing.T) {
	expected := []byte{0x1B, 0x61, 0x01}

	assertBytes(t, AlignCenter(), expected)
}

func TestAlignRight(t *testing.T) {
	expected := []byte{0x1B, 0x61, 0x02}

	assertBytes(t, AlignRight(), expected)
}

func TestFontSize(t *testing.T) {
	expected := []byte{0x1D, 0x21, 0x11}

	assertBytes(t, FontSize(2, 2), expected)
}

func TestFeed(t *testing.T) {
	expected := []byte{0x1B, 0x64, 0x03}

	assertBytes(t, Feed(3), expected)
}

func TestFeedZero(t *testing.T) {
	expected := []byte{0x1B, 0x64, 0x00}

	assertBytes(t, Feed(0), expected)
}

func TestMultipleCommands(t *testing.T) {
	var data []byte
	data = append(data, Initialize()...)
	data = append(data, AlignCenter()...)
	data = append(data, Bold(true)...)
	data = append(data, Text("TOKO KASA")...)
	data = append(data, LF()...)
	data = append(data, Bold(false)...)

	expected := []byte{
		// Initialize
		0x1B, 0x40,

		// Center
		0x1B, 0x61, 0x01,

		// Bold ON
		0x1B, 0x45, 0x01,

		// Text
		'T', 'O', 'K', 'O', ' ',
		'K', 'A', 'S', 'A',

		// LF
		0x0A,

		// Bold OFF
		0x1B, 0x45, 0x00,
	}

	assertBytes(t, data, expected)
}

func TestBytes(t *testing.T) {
	got := Text("KASA")

	expected := []byte("KASA")

	assertBytes(t, got, expected)
}

func TestEmptyText(t *testing.T) {
	got := Text("")

	expected := []byte{}

	assertBytes(t, got, expected)
}
