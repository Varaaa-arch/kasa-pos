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

func TestCheckoutWithAdjustments(t *testing.T) {
	c := cart.New()

	p := product.Product{
		ID:    "p1",
		Name:  "Kopi Susu",
		Price: 30000,
		Stock: 100,
	}

	if err := c.AddItem(p, 2); err != nil {
		t.Fatal(err)
	}

	result, err := NewService().Execute(
		context.Background(),
		Request{
			Cart:          c,
			PaidAmount:    70000,
			Discount:      5000,
			Tax:           3000,
			ServiceCharge: 2000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	// 60.000 - 5.000 + 3.000 + 2.000 = 60.000
	if result.Subtotal != 60000 {
		t.Fatalf(
			"subtotal = %d, want 60000",
			result.Subtotal,
		)
	}

	if result.Total != 60000 {
		t.Fatalf(
			"total = %d, want 60000",
			result.Total,
		)
	}

	if result.Change != 10000 {
		t.Fatalf(
			"change = %d, want 10000",
			result.Change,
		)
	}
}

func TestCheckoutExactPayment(t *testing.T) {
	c := cart.New()

	p := product.Product{
		ID:    "p1",
		Name:  "Roti",
		Price: 10000,
		Stock: 10,
	}

	if err := c.AddItem(p, 2); err != nil {
		t.Fatal(err)
	}

	result, err := NewService().Execute(
		context.Background(),
		Request{
			Cart:       c,
			PaidAmount: 20000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.Change != 0 {
		t.Fatalf(
			"change = %d, want 0",
			result.Change,
		)
	}
}

func TestCheckoutDoesNotMutateCart(t *testing.T) {
	c := cart.New()

	p := product.Product{
		ID:    "p1",
		Name:  "Kopi",
		Price: 15000,
		Stock: 10,
	}

	if err := c.AddItem(p, 2); err != nil {
		t.Fatal(err)
	}

	before := c.Total

	_, err := NewService().Execute(
		context.Background(),
		Request{
			Cart:       c,
			PaidAmount: 30000,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if c.Total != before {
		t.Fatalf(
			"checkout mutated cart total: got %d, want %d",
			c.Total,
			before,
		)
	}
}
