package receipt

import "testing"

func TestDefaultReceiptTemplate(t *testing.T) {
	template := DefaultReceiptTemplate()

	if template.Name != "default" {
		t.Fatalf(
			"expected template name %q, got %q",
			"default",
			template.Name,
		)
	}

	if !template.Header {
		t.Fatal("expected header to be enabled")
	}

	if !template.Transaction {
		t.Fatal("expected transaction to be enabled")
	}

	if !template.Items {
		t.Fatal("expected items to be enabled")
	}

	if !template.Summary {
		t.Fatal("expected summary to be enabled")
	}

	if !template.Payment {
		t.Fatal("expected payment to be enabled")
	}

	if !template.Footer {
		t.Fatal("expected footer to be enabled")
	}
}

func TestCompactReceiptTemplate(t *testing.T) {
	template := CompactReceiptTemplate()

	if template.Name != "compact" {
		t.Fatalf(
			"expected template name %q, got %q",
			"compact",
			template.Name,
		)
	}

	if !template.Header {
		t.Fatal("expected header to be enabled")
	}

	if !template.Transaction {
		t.Fatal("expected transaction to be enabled")
	}

	if !template.Items {
		t.Fatal("expected items to be enabled")
	}

	if !template.Summary {
		t.Fatal("expected summary to be enabled")
	}

	if !template.Payment {
		t.Fatal("expected payment to be enabled")
	}

	if template.Footer {
		t.Fatal("expected footer to be disabled")
	}
}

func TestDetailedReceiptTemplate(t *testing.T) {
	template := DetailedReceiptTemplate()

	if template.Name != "detailed" {
		t.Fatalf(
			"expected template name %q, got %q",
			"detailed",
			template.Name,
		)
	}

	if !template.Header {
		t.Fatal("expected header to be enabled")
	}

	if !template.Transaction {
		t.Fatal("expected transaction to be enabled")
	}

	if !template.Items {
		t.Fatal("expected items to be enabled")
	}

	if !template.Summary {
		t.Fatal("expected summary to be enabled")
	}

	if !template.Payment {
		t.Fatal("expected payment to be enabled")
	}

	if !template.Footer {
		t.Fatal("expected footer to be enabled")
	}
}

func TestNewTemplate(t *testing.T) {
	template := NewTemplate("custom")

	if template.Name != "custom" {
		t.Fatalf(
			"expected template name %q, got %q",
			"custom",
			template.Name,
		)
	}

	if !template.Header {
		t.Fatal("expected header to be enabled")
	}

	if !template.Transaction {
		t.Fatal("expected transaction to be enabled")
	}

	if !template.Items {
		t.Fatal("expected items to be enabled")
	}

	if !template.Summary {
		t.Fatal("expected summary to be enabled")
	}

	if !template.Payment {
		t.Fatal("expected payment to be enabled")
	}

	if !template.Footer {
		t.Fatal("expected footer to be enabled")
	}
}
