package receipt

import (
	"testing"

	domainreceipt "pos-system/internal/domain/receipt"
)

func TestBuildSectionsWithDefaultTemplate(t *testing.T) {
	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name: "TOKO KASA",
		},
	}

	template := DefaultReceiptTemplate()

	sections := BuildSectionsWithTemplate(
		input,
		template,
	)

	if len(sections) != 6 {
		t.Fatalf(
			"expected 6 sections, got %d",
			len(sections),
		)
	}

	expected := []SectionType{
		SectionHeader,
		SectionTransaction,
		SectionItems,
		SectionSummary,
		SectionPayment,
		SectionFooter,
	}

	for i, expectedType := range expected {
		if sections[i].Type != expectedType {
			t.Fatalf(
				"section %d: expected %q, got %q",
				i,
				expectedType,
				sections[i].Type,
			)
		}
	}
}

func TestBuildSectionsWithCompactTemplate(t *testing.T) {
	input := domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name: "TOKO KASA",
		},

		Footer: domainreceipt.Footer{
			Message: "TERIMA KASIH",
		},
	}

	template := CompactReceiptTemplate()

	sections := BuildSectionsWithTemplate(
		input,
		template,
	)

	if len(sections) != 5 {
		t.Fatalf(
			"expected 5 sections, got %d",
			len(sections),
		)
	}

	for _, section := range sections {
		if section.Type == SectionFooter {
			t.Fatal(
				"compact template should not contain FOOTER",
			)
		}
	}
}

func TestBuildSectionsWithCustomTemplate(t *testing.T) {
	input := domainreceipt.Receipt{}

	template := Template{
		Name:        "items-only",
		Header:      false,
		Transaction: false,
		Items:       true,
		Summary:     false,
		Payment:     false,
		Footer:      false,
	}

	sections := BuildSectionsWithTemplate(
		input,
		template,
	)

	if len(sections) != 1 {
		t.Fatalf(
			"expected 1 section, got %d",
			len(sections),
		)
	}

	if sections[0].Type != SectionItems {
		t.Fatalf(
			"expected ITEMS section, got %q",
			sections[0].Type,
		)
	}
}

func TestBuildSectionsWithEmptyTemplate(t *testing.T) {
	input := domainreceipt.Receipt{}

	template := Template{
		Name: "empty",
	}

	sections := BuildSectionsWithTemplate(
		input,
		template,
	)

	if len(sections) != 0 {
		t.Fatalf(
			"expected 0 sections, got %d",
			len(sections),
		)
	}
}

func TestTemplateAllowsSection(t *testing.T) {
	template := Template{
		Header:      true,
		Transaction: false,
		Items:       true,
		Summary:     false,
		Payment:     true,
		Footer:      false,
	}

	tests := []struct {
		name     string
		section  SectionType
		expected bool
	}{
		{
			name:     "header enabled",
			section:  SectionHeader,
			expected: true,
		},
		{
			name:     "transaction disabled",
			section:  SectionTransaction,
			expected: false,
		},
		{
			name:     "items enabled",
			section:  SectionItems,
			expected: true,
		},
		{
			name:     "summary disabled",
			section:  SectionSummary,
			expected: false,
		},
		{
			name:     "payment enabled",
			section:  SectionPayment,
			expected: true,
		},
		{
			name:     "footer disabled",
			section:  SectionFooter,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := templateAllowsSection(
				template,
				tt.section,
			)

			if got != tt.expected {
				t.Fatalf(
					"got %v, want %v",
					got,
					tt.expected,
				)
			}
		})
	}
}
