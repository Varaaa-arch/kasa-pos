package checkout

import (
	"context"
	"errors"

	"pos-system/internal/domain/cart"
)

var (
	ErrEmptyCart        = errors.New("cart is empty")
	ErrInsufficientCash = errors.New("payment is insufficient")
)

type Request struct {
	Cart          *cart.Cart
	PaidAmount    int64
	Discount      int64
	Tax           int64
	ServiceCharge int64
}

type Result struct {
	Subtotal      int64
	Discount      int64
	Tax           int64
	ServiceCharge int64
	Total         int64
	PaidAmount    int64
	Change        int64
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Execute(
	ctx context.Context,
	req Request,
) (Result, error) {
	if req.Cart == nil || len(req.Cart.Items) == 0 {
		return Result{}, ErrEmptyCart
	}

	subtotal := req.Cart.Total

	total := subtotal -
		req.Discount +
		req.Tax +
		req.ServiceCharge

	if total < 0 {
		total = 0
	}

	if req.PaidAmount < total {
		return Result{}, ErrInsufficientCash
	}

	return Result{
		Subtotal:      subtotal,
		Discount:      req.Discount,
		Tax:           req.Tax,
		ServiceCharge: req.ServiceCharge,
		Total:         total,
		PaidAmount:    req.PaidAmount,
		Change:        req.PaidAmount - total,
	}, nil
}
