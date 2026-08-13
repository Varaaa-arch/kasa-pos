package receipt

import "testing"

func TestNewSection(t *testing.T) {
	section := NewSection(
		SectionHeader,
		"TOKO KASA",
		"Jl. Contoh No. 123",
	)

	if section.Type != SectionHeader {
		t.Fatalf(
			"expected section type %q, got %q",
			SectionHeader,
			section.Type,
		)
	}

	if len(section.Lines) != 2 {
		t.Fatalf(
			"expected 2 lines, got %d",
			len(section.Lines),
		)
	}

	if section.Lines[0] != "TOKO KASA" {
		t.Fatalf(
			"unexpected first line: %q",
			section.Lines[0],
		)
	}

	if section.Lines[1] != "Jl. Contoh No. 123" {
		t.Fatalf(
			"unexpected second line: %q",
			section.Lines[1],
		)
	}
}

func TestNewSectionCopiesLines(t *testing.T) {
	lines := []string{
		"TOKO KASA",
		"TERIMA KASIH",
	}

	section := NewSection(
		SectionFooter,
		lines...,
	)

	lines[0] = "CHANGED"

	if section.Lines[0] != "TOKO KASA" {
		t.Fatalf(
			"section lines were not copied: %q",
			section.Lines[0],
		)
	}
}

func TestSectionIsEmpty(t *testing.T) {
	empty := NewSection(SectionItems)

	if !empty.IsEmpty() {
		t.Fatal("expected empty section")
	}

	nonEmpty := NewSection(
		SectionItems,
		"Kopi Susu",
	)

	if nonEmpty.IsEmpty() {
		t.Fatal("expected non-empty section")
	}
}

func TestSectionTypes(t *testing.T) {
	tests := []struct {
		name string
		got  SectionType
		want SectionType
	}{
		{
			name: "header",
			got:  SectionHeader,
			want: "HEADER",
		},
		{
			name: "transaction",
			got:  SectionTransaction,
			want: "TRANSACTION",
		},
		{
			name: "items",
			got:  SectionItems,
			want: "ITEMS",
		},
		{
			name: "summary",
			got:  SectionSummary,
			want: "SUMMARY",
		},
		{
			name: "payment",
			got:  SectionPayment,
			want: "PAYMENT",
		},
		{
			name: "footer",
			got:  SectionFooter,
			want: "FOOTER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf(
					"got %q, want %q",
					tt.got,
					tt.want,
				)
			}
		})
	}
}
