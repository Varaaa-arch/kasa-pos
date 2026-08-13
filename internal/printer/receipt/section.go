package receipt

type SectionType string

const (
	SectionHeader      SectionType = "HEADER"
	SectionTransaction SectionType = "TRANSACTION"
	SectionItems       SectionType = "ITEMS"
	SectionSummary     SectionType = "SUMMARY"
	SectionPayment     SectionType = "PAYMENT"
	SectionFooter      SectionType = "FOOTER"
)

type Section struct {
	Type  SectionType
	Lines []string
}

func NewSection(
	sectionType SectionType,
	lines ...string,
) Section {
	return Section{
		Type:  sectionType,
		Lines: append([]string(nil), lines...),
	}
}

func (s Section) IsEmpty() bool {
	return len(s.Lines) == 0
}
