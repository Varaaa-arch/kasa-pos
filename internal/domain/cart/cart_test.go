package cart

import (
	"pos-system/internal/domain/product"
	"testing"
)

func TestCartAddItem(t *testing.T) {
	cart := New()

	p := product.Product{
		ID:    "p1",
		Name:  "Kopi Susu",
		Price: 15000,
		Stock: 100,
	}

	if err := cart.AddItem(p, 2); err != nil {
		t.Fatal(err)
	}

	if len(cart.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(cart.Items))
	}

	if cart.Items[0].Quantity != 2 {
		t.Fatalf(
			"expected quantity 2, got %d",
			cart.Items[0].Quantity,
		)
	}

	if cart.Items[0].Subtotal != 30000 {
		t.Fatalf(
			"expected subtotal 30000, got %d",
			cart.Items[0].Subtotal,
		)
	}

	if cart.Total != 30000 {
		t.Fatalf(
			"expected total 30000, got %d",
			cart.Total,
		)
	}
}

func TestCartUpdateAndRemove(t *testing.T) {
	cart := New()

	p := product.Product{
		ID:    "p1",
		Name:  "Kopi Susu",
		Price: 15000,
		Stock: 100,
	}

	if err := cart.AddItem(p, 2); err != nil {
		t.Fatal(err)
	}

	if err := cart.UpdateQuantity("p1", 5); err != nil {
		t.Fatal(err)
	}

	if cart.Total != 75000 {
		t.Fatalf(
			"expected total 75000, got %d",
			cart.Total,
		)
	}

	if err := cart.RemoveItem("p1"); err != nil {
		t.Fatal(err)
	}

	if len(cart.Items) != 0 {
		t.Fatal("expected empty cart")
	}

	if cart.Total != 0 {
		t.Fatal("expected total 0")
	}
}

func TestCartRejectsInvalidQuantity(t *testing.T) {
	cart := New()

	p := product.Product{
		ID:    "p1",
		Price: 15000,
		Stock: 10,
	}

	if err := cart.AddItem(p, 0); err != ErrInvalidQuantity {
		t.Fatalf(
			"expected ErrInvalidQuantity, got %v",
			err,
		)
	}

	if err := cart.AddItem(p, 11); err != ErrInsufficientStock {
		t.Fatalf(
			"expected ErrInsufficientStock, got %v",
			err,
		)
	}
}
