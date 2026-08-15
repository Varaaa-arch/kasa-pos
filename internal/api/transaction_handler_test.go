package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domaintransaction "pos-system/internal/domain/transaction"
	postgres "pos-system/internal/repository/postgres"
	transactionservice "pos-system/internal/service/transaction"
)

type fakeTransactionRepo struct {
	transactions []domaintransaction.Transaction
	listErr      error
	getErr       error
}

func (f *fakeTransactionRepo) Create(ctx context.Context, tx domaintransaction.Transaction) error {
	f.transactions = append(f.transactions, tx)
	return nil
}

func (f *fakeTransactionRepo) GetByID(ctx context.Context, id string) (domaintransaction.Transaction, error) {
	if f.getErr != nil {
		return domaintransaction.Transaction{}, f.getErr
	}
	for _, tx := range f.transactions {
		if tx.ID == id {
			return tx, nil
		}
	}
	return domaintransaction.Transaction{}, postgres.ErrTransactionNotFound
}

func (f *fakeTransactionRepo) List(ctx context.Context) ([]domaintransaction.Transaction, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.transactions, nil
}

func newTestTransactionHandler(repo domaintransaction.Repository) *TransactionHandler {
	svc := transactionservice.NewService(repo)
	return NewTransactionHandler(svc)
}

func TestTransactionHandler_List_Success(t *testing.T) {
	repo := &fakeTransactionRepo{
		transactions: []domaintransaction.Transaction{
			{
				ID:            "tx-1",
				InvoiceNumber: "INV-001",
				Total:         50000,
				PaidAmount:    50000,
				PaymentMethod: "CASH",
				Status:        domaintransaction.StatusCompleted,
				CreatedAt:     time.Now().UTC(),
			},
			{
				ID:            "tx-2",
				InvoiceNumber: "INV-002",
				Total:         30000,
				PaidAmount:    50000,
				Change:        20000,
				PaymentMethod: "CASH",
				Status:        domaintransaction.StatusCompleted,
				CreatedAt:     time.Now().UTC(),
			},
		},
	}

	h := newTestTransactionHandler(repo)
	router := NewRouter(nil, h, nil)

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res []domaintransaction.Transaction
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(res))
	}
	if res[0].InvoiceNumber != "INV-001" {
		t.Errorf("expected invoice INV-001, got %s", res[0].InvoiceNumber)
	}
}

func TestTransactionHandler_List_Error(t *testing.T) {
	repo := &fakeTransactionRepo{
		listErr: errors.New("database error"),
	}

	h := newTestTransactionHandler(repo)
	router := NewRouter(nil, h, nil)

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestTransactionHandler_GetByID_Success(t *testing.T) {
	repo := &fakeTransactionRepo{
		transactions: []domaintransaction.Transaction{
			{
				ID:            "tx-123",
				InvoiceNumber: "INV-123",
				Total:         75000,
				PaidAmount:    100000,
				Change:        25000,
				PaymentMethod: "CASH",
				Status:        domaintransaction.StatusCompleted,
				CreatedAt:     time.Now().UTC(),
			},
		},
	}

	h := newTestTransactionHandler(repo)
	router := NewRouter(nil, h, nil)

	req := httptest.NewRequest(http.MethodGet, "/transactions/tx-123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res domaintransaction.Transaction
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.ID != "tx-123" {
		t.Errorf("expected ID tx-123, got %s", res.ID)
	}
	if res.InvoiceNumber != "INV-123" {
		t.Errorf("expected InvoiceNumber INV-123, got %s", res.InvoiceNumber)
	}
}

func TestTransactionHandler_GetByID_NotFound(t *testing.T) {
	repo := &fakeTransactionRepo{
		transactions: []domaintransaction.Transaction{},
	}

	h := newTestTransactionHandler(repo)
	router := NewRouter(nil, h, nil)

	req := httptest.NewRequest(http.MethodGet, "/transactions/non-existent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTransactionHandler_GetByID_InternalError(t *testing.T) {
	repo := &fakeTransactionRepo{
		getErr: errors.New("db failure"),
	}

	h := newTestTransactionHandler(repo)
	router := NewRouter(nil, h, nil)

	req := httptest.NewRequest(http.MethodGet, "/transactions/tx-123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
