package product

import (
	"context"
	"errors"
	"strings"
	"time"

	"pos-system/internal/domain/product"
)

var (
	ErrSKURequired   = errors.New("sku is required")
	ErrNameRequired  = errors.New("name is required")
	ErrPriceNegative = errors.New("price must be non-negative")
	ErrStockNegative = errors.New("stock must be non-negative")
)

type Service struct {
	repo product.Repository
}

func NewService(repo product.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	p product.Product,
) error {
	if strings.TrimSpace(p.SKU) == "" {
		return ErrSKURequired
	}

	if strings.TrimSpace(p.Name) == "" {
		return ErrNameRequired
	}

	if p.Price < 0 {
		return ErrPriceNegative
	}

	if p.Stock < 0 {
		return ErrStockNegative
	}

	now := time.Now().UTC()

	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}

	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}

	return s.repo.Create(ctx, p)
}

func (s *Service) GetByID(
	ctx context.Context,
	id string,
) (product.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetBySKU(
	ctx context.Context,
	sku string,
) (product.Product, error) {
	return s.repo.GetBySKU(ctx, sku)
}

func (s *Service) List(
	ctx context.Context,
) ([]product.Product, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(
	ctx context.Context,
	p product.Product,
) error {
	if strings.TrimSpace(p.SKU) == "" {
		return ErrSKURequired
	}

	if strings.TrimSpace(p.Name) == "" {
		return ErrNameRequired
	}

	if p.Price < 0 {
		return ErrPriceNegative
	}

	if p.Stock < 0 {
		return ErrStockNegative
	}

	p.UpdatedAt = time.Now().UTC()

	return s.repo.Update(ctx, p)
}

func (s *Service) Delete(
	ctx context.Context,
	id string,
) error {
	return s.repo.Delete(ctx, id)
}
