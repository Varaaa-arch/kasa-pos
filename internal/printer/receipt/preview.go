package receipt

import (
	"strings"

	domainreceipt "pos-system/internal/domain/receipt"
)

type Preview struct {
	Width int
	Lines []string
}

func NewPreview(width int) *Preview {
	if width <= 0 {
		width = DefaultReceiptWidth
	}

	return &Preview{
		Width: width,
		Lines: make([]string, 0),
	}
}

func (p *Preview) Render(
	input domainreceipt.Receipt,
	template Template,
) {
	p.Lines = p.Lines[:0]

	sections := BuildSectionsWithTemplate(
		input,
		template,
	)

	layout := NewLayout(p.Width)

	for _, section := range sections {
		if section.IsEmpty() {
			continue
		}

		p.renderSection(
			layout,
			section,
		)
	}
}

func (p *Preview) String() string {
	return strings.Join(
		p.Lines,
		"\n",
	)
}

func (p *Preview) renderSection(
	layout Layout,
	section Section,
) {
	switch section.Type {
	case SectionHeader:
		p.Lines = append(
			p.Lines,
			layout.SeparatorLine('='),
		)

		for _, line := range section.Lines {
			p.Lines = append(
				p.Lines,
				layout.Center(line),
			)
		}

		p.Lines = append(
			p.Lines,
			layout.SeparatorLine('='),
		)

	case SectionTransaction:
		for _, line := range section.Lines {
			p.Lines = append(
				p.Lines,
				layout.Left(line),
			)
		}

	case SectionItems:
		p.Lines = append(
			p.Lines,
			layout.SeparatorLine('-'),
		)

		for _, line := range section.Lines {
			if line == "" {
				p.Lines = append(
					p.Lines,
					layout.BlankLine(),
				)

				continue
			}

			p.Lines = append(
				p.Lines,
				layout.Left(line),
			)
		}

	case SectionSummary:
		p.Lines = append(
			p.Lines,
			layout.SeparatorLine('-'),
		)

		for _, line := range section.Lines {
			p.Lines = append(
				p.Lines,
				layout.Left(line),
			)
		}

	case SectionPayment:
		p.Lines = append(
			p.Lines,
			layout.SeparatorLine('-'),
		)

		for _, line := range section.Lines {
			p.Lines = append(
				p.Lines,
				layout.Left(line),
			)
		}

	case SectionFooter:
		p.Lines = append(
			p.Lines,
			layout.SeparatorLine('-'),
		)

		p.Lines = append(
			p.Lines,
			layout.BlankLine(),
		)

		for _, line := range section.Lines {
			p.Lines = append(
				p.Lines,
				layout.Center(line),
			)
		}
	}
}
