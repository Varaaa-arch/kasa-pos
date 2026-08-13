package receipt

import domainreceipt "pos-system/internal/domain/receipt"

// BuildSectionsWithTemplate builds receipt sections according
// to the enabled sections in the provided template.
func BuildSectionsWithTemplate(
	input domainreceipt.Receipt,
	template Template,
) []Section {
	allSections := BuildSections(input)

	var sections []Section

	for _, section := range allSections {
		if !templateAllowsSection(template, section.Type) {
			continue
		}

		sections = append(
			sections,
			section,
		)
	}

	return sections
}

func templateAllowsSection(
	template Template,
	sectionType SectionType,
) bool {
	switch sectionType {
	case SectionHeader:
		return template.Header

	case SectionTransaction:
		return template.Transaction

	case SectionItems:
		return template.Items

	case SectionSummary:
		return template.Summary

	case SectionPayment:
		return template.Payment

	case SectionFooter:
		return template.Footer

	default:
		return false
	}
}
