package transaction

import (
	"context"
	"errors"
	"strings"
	"time"

	domaintransaction "pos-system/internal/domain/transaction"
)

var (
	ErrInvoiceRequired      = errors.New("invoice number is required")
	ErrTransactionInvalid   = errors.New("transaction is invalid")
	ErrPaymentMethodMissing = errors.New("payment method is required")
)

type Service struct {
	repo domaintransaction.Repository
}

func NewService(
	repo domaintransaction.Repository,
) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	t domaintransaction.Transaction,
) error {
	if strings.TrimSpace(t.ID) == "" {
		return ErrTransactionInvalid
	}

	if strings.TrimSpace(t.InvoiceNumber) == "" {
		return ErrInvoiceRequired
	}

	if strings.TrimSpace(t.PaymentMethod) == "" {
		return ErrPaymentMethodMissing
	}

	if t.Total < 0 ||
		t.PaidAmount < 0 ||
		t.Change < 0 {
		return ErrTransactionInvalid
	}

	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}

	if t.Status == "" {
		t.Status = domaintransaction.StatusCompleted
	}

	return s.repo.Create(ctx, t)
}

func (s *Service) GetByID(
	ctx context.Context,
	id string,
) (domaintransaction.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(
	ctx context.Context,
) ([]domaintransaction.Transaction, error) {
	return s.repo.List(ctx)
}
