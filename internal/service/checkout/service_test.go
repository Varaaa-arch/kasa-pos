package checkout

import (
	"context"
	"errors"
	"testing"

	"pos-system/internal/domain/cart"
	"pos-system/internal/domain/product"
)

func TestCheckout(t *testing.T) {
	c := cart.New()

	p := product.Product{
		ID:    "p1",
		Name:  "Kopi Susu",
		Price: 15000,
		Stock: 100,
	}

	if err := c.AddItem(p, 2); err != nil {
		t.Fatal(err)
	}

	service := NewService()

	result, err := service.Execute(
		context.Background(),
		Request{
			Cart:       c,
			PaidAmount: 50000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.Subtotal != 30000 {
		t.Fatalf("subtotal = %d, want 30000", result.Subtotal)
	}

	if result.Total != 30000 {
		t.Fatalf("total = %d, want 30000", result.Total)
	}

	if result.Change != 20000 {
		t.Fatalf("change = %d, want 20000", result.Change)
	}
}

func TestCheckoutRejectsInsufficientPayment(t *testing.T) {
	c := cart.New()

	p := product.Product{
		ID:    "p1",
		Name:  "Kopi Susu",
		Price: 15000,
		Stock: 100,
	}

	if err := c.AddItem(p, 1); err != nil {
		t.Fatal(err)
	}

	_, err := NewService().Execute(
		context.Background(),
		Request{
			Cart:       c,
			PaidAmount: 10000,
		},
	)

	if !errors.Is(err, ErrInsufficientCash) {
		t.Fatalf(
			"expected ErrInsufficientCash, got %v",
			err,
		)
	}
}

func TestCheckoutRejectsEmptyCart(t *testing.T) {
	_, err := NewService().Execute(
		context.Background(),
		Request{
			Cart:       cart.New(),
			PaidAmount: 10000,
		},
	)

	if !errors.Is(err, ErrEmptyCart) {
		t.Fatalf(
			"expected ErrEmptyCart, got %v",
			err,
		)
	}
}
