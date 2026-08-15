package transaction

import (
	"context"
	"errors"
	"testing"

	domaintransaction "pos-system/internal/domain/transaction"
)

type fakeRepository struct {
	created domaintransaction.Transaction
	list    []domaintransaction.Transaction
}

func (f *fakeRepository) Create(
	ctx context.Context,
	t domaintransaction.Transaction,
) error {
	f.created = t
	return nil
}

func (f *fakeRepository) GetByID(
	ctx context.Context,
	id string,
) (domaintransaction.Transaction, error) {
	return f.created, nil
}

func (f *fakeRepository) List(
	ctx context.Context,
) ([]domaintransaction.Transaction, error) {
	return f.list, nil
}

func TestServiceCreate(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	input := domaintransaction.Transaction{
		ID:            "trx-001",
		InvoiceNumber: "INV-0001",
		Total:         30000,
		PaidAmount:    50000,
		Change:        20000,
		PaymentMethod: "CASH",
	}

	err := service.Create(
		context.Background(),
		input,
	)

	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repo.created.InvoiceNumber != "INV-0001" {
		t.Fatalf(
			"unexpected invoice: %q",
			repo.created.InvoiceNumber,
		)
	}

	if repo.created.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt")
	}

	if repo.created.Status != domaintransaction.StatusCompleted {
		t.Fatalf(
			"unexpected status: %q",
			repo.created.Status,
		)
	}
}

func TestServiceCreateValidation(t *testing.T) {
	tests := []struct {
		name string
		data domaintransaction.Transaction
		want error
	}{
		{
			name: "missing invoice",
			data: domaintransaction.Transaction{
				ID:            "trx-001",
				PaymentMethod: "CASH",
			},
			want: ErrInvoiceRequired,
		},
		{
			name: "missing payment method",
			data: domaintransaction.Transaction{
				ID:            "trx-001",
				InvoiceNumber: "INV-001",
			},
			want: ErrPaymentMethodMissing,
		},
		{
			name: "negative total",
			data: domaintransaction.Transaction{
				ID:            "trx-001",
				InvoiceNumber: "INV-001",
				PaymentMethod: "CASH",
				Total:         -1,
			},
			want: ErrTransactionInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{}
			service := NewService(repo)

			err := service.Create(
				context.Background(),
				tt.data,
			)

			if !errors.Is(err, tt.want) {
				t.Fatalf(
					"expected %v, got %v",
					tt.want,
					err,
				)
			}
		})
	}
}
