package receipt

type Template struct {
	Name string

	Header      bool
	Transaction bool
	Items       bool
	Summary     bool
	Payment     bool
	Footer      bool
}

func NewTemplate(name string) Template {
	return Template{
		Name:        name,
		Header:      true,
		Transaction: true,
		Items:       true,
		Summary:     true,
		Payment:     true,
		Footer:      true,
	}
}

func DefaultReceiptTemplate() Template {
	return NewTemplate("default")
}

func CompactReceiptTemplate() Template {
	return Template{
		Name:        "compact",
		Header:      true,
		Transaction: true,
		Items:       true,
		Summary:     true,
		Payment:     true,
		Footer:      false,
	}
}

func DetailedReceiptTemplate() Template {
	return Template{
		Name:        "detailed",
		Header:      true,
		Transaction: true,
		Items:       true,
		Summary:     true,
		Payment:     true,
		Footer:      true,
	}
}
