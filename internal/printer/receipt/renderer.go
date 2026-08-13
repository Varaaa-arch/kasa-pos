package receipt

import (
	"fmt"
	"strings"

	domainreceipt "pos-system/internal/domain/receipt"
	"pos-system/internal/printer/escpos"
)

const DefaultLineWidth = 32

type Renderer struct {
	Width int
}

func NewRenderer() *Renderer {
	return &Renderer{
		Width: DefaultLineWidth,
	}
}

func (r *Renderer) Render(
	input domainreceipt.Receipt,
) []byte {
	return r.RenderWithTemplate(
		input,
		DefaultReceiptTemplate(),
	)
}

func (r *Renderer) RenderWithTemplate(
	input domainreceipt.Receipt,
	template Template,
) []byte {
	var data []byte

	layout := NewLayout(r.Width)

	sections := BuildSectionsWithTemplate(
		input,
		template,
	)

	data = append(
		data,
		escpos.Initialize()...,
	)

	for _, section := range sections {
		if section.IsEmpty() {
			continue
		}

		data = append(
			data,
			r.renderSection(section, layout)...,
		)
	}

	data = append(
		data,
		escpos.Feed(3)...,
	)

	return data
}

func (r *Renderer) renderSection(
	section Section,
	layout Layout,
) []byte {
	var data []byte

	switch section.Type {
	case SectionHeader:
		data = append(
			data,
			escpos.AlignCenter()...,
		)

		data = append(
			data,
			escpos.Bold(true)...,
		)

		for i, line := range section.Lines {
			if line == "" {
				data = append(
					data,
					escpos.LF()...,
				)
				continue
			}

			data = append(
				data,
				escpos.Text(line)...,
			)

			data = append(
				data,
				escpos.LF()...,
			)

			if i == 0 {
				data = append(
					data,
					escpos.Bold(false)...,
				)
			}
		}

		data = append(
			data,
			escpos.Bold(false)...,
		)

		data = append(
			data,
			escpos.Text(
				r.separator("="),
			)...,
		)

		data = append(
			data,
			escpos.LF()...,
		)

	case SectionTransaction:
		data = append(
			data,
			escpos.AlignLeft()...,
		)

		for _, line := range section.Lines {
			data = append(
				data,
				escpos.Text(line)...,
			)

			data = append(
				data,
				escpos.LF()...,
			)
		}

	case SectionItems:
		data = append(
			data,
			escpos.AlignLeft()...,
		)

		data = append(
			data,
			escpos.Text(
				r.separator("-"),
			)...,
		)

		data = append(
			data,
			escpos.LF()...,
		)

		for _, line := range section.Lines {
			if line == "" {
				data = append(
					data,
					escpos.LF()...,
				)

				continue
			}

			data = append(
				data,
				escpos.Text(line)...,
			)

			data = append(
				data,
				escpos.LF()...,
			)
		}

	case SectionSummary:
		data = append(
			data,
			escpos.AlignLeft()...,
		)

		data = append(
			data,
			escpos.Text(
				r.separator("-"),
			)...,
		)

		data = append(
			data,
			escpos.LF()...,
		)

		for _, line := range section.Lines {
			if line == "" {
				data = append(
					data,
					escpos.LF()...,
				)

				continue
			}

			isTotal := strings.HasPrefix(
				strings.TrimSpace(line),
				"TOTAL",
			)

			if isTotal {
				data = append(
					data,
					escpos.Bold(true)...,
				)
			}

			data = append(
				data,
				escpos.Text(line)...,
			)

			data = append(
				data,
				escpos.LF()...,
			)

			if isTotal {
				data = append(
					data,
					escpos.Bold(false)...,
				)
			}
		}

	case SectionPayment:
		data = append(
			data,
			escpos.AlignLeft()...,
		)

		data = append(
			data,
			escpos.Text(
				r.separator("-"),
			)...,
		)

		data = append(
			data,
			escpos.LF()...,
		)

		for _, line := range section.Lines {
			data = append(
				data,
				escpos.Text(line)...,
			)

			data = append(
				data,
				escpos.LF()...,
			)
		}

	case SectionFooter:
		data = append(
			data,
			escpos.AlignCenter()...,
		)

		data = append(
			data,
			escpos.Text(
				r.separator("-"),
			)...,
		)

		data = append(
			data,
			escpos.LF()...,
		)

		data = append(
			data,
			escpos.LF()...,
		)

		data = append(
			data,
			escpos.Bold(true)...,
		)

		for _, line := range section.Lines {
			if line == "" {
				data = append(
					data,
					escpos.LF()...,
				)

				continue
			}

			data = append(
				data,
				escpos.Text(line)...,
			)

			data = append(
				data,
				escpos.LF()...,
			)
		}

		data = append(
			data,
			escpos.Bold(false)...,
		)
	}

	return data
}

func (r *Renderer) separator(
	character string,
) string {
	return strings.Repeat(
		character,
		r.Width,
	)
}

func formatLeftRight(
	left string,
	right string,
	width int,
) string {
	if width <= 0 {
		return ""
	}

	if len(right) >= width {
		return right[:width]
	}

	if len(left)+len(right) >= width {
		maxLeft := width - len(right)

		if maxLeft <= 0 {
			return right
		}

		return truncate(left, maxLeft) + right
	}

	spaces := width - len(left) - len(right)

	return left +
		strings.Repeat(" ", spaces) +
		right
}

func formatRupiah(amount int64) string {
	if amount == 0 {
		return "Rp 0"
	}

	sign := ""

	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	digits := []rune(
		fmt.Sprintf("%d", amount),
	)

	var groups []string

	for len(digits) > 3 {
		groups = append(
			[]string{
				string(digits[len(digits)-3:]),
			},
			groups...,
		)

		digits = digits[:len(digits)-3]
	}

	groups = append(
		[]string{
			string(digits),
		},
		groups...,
	)

	return sign +
		"Rp " +
		strings.Join(groups, ".")
}
