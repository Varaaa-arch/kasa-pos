package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"pos-system/internal/domain/product"
	domaintransaction "pos-system/internal/domain/transaction"
	postgres "pos-system/internal/repository/postgres"
	productservice "pos-system/internal/service/product"
	transactionservice "pos-system/internal/service/transaction"
)

type smokeProductRepo struct {
	mu       sync.RWMutex
	products map[string]product.Product
}

func newSmokeProductRepo() *smokeProductRepo {
	return &smokeProductRepo{
		products: make(map[string]product.Product),
	}
}

func (r *smokeProductRepo) Create(ctx context.Context, p product.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[p.ID] = p
	return nil
}

func (r *smokeProductRepo) GetByID(ctx context.Context, id string) (product.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id]
	if !ok {
		return product.Product{}, postgres.ErrProductNotFound
	}
	return p, nil
}

func (r *smokeProductRepo) GetBySKU(ctx context.Context, sku string) (product.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.products {
		if p.SKU == sku {
			return p, nil
		}
	}
	return product.Product{}, postgres.ErrProductNotFound
}

func (r *smokeProductRepo) List(ctx context.Context) ([]product.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]product.Product, 0, len(r.products))
	for _, p := range r.products {
		list = append(list, p)
	}
	return list, nil
}

func (r *smokeProductRepo) Update(ctx context.Context, p product.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.products[p.ID]; !ok {
		return postgres.ErrProductNotFound
	}
	r.products[p.ID] = p
	return nil
}

func (r *smokeProductRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.products[id]; !ok {
		return postgres.ErrProductNotFound
	}
	delete(r.products, id)
	return nil
}

type smokeTransactionRepo struct {
	mu           sync.RWMutex
	transactions map[string]domaintransaction.Transaction
}

func newSmokeTransactionRepo() *smokeTransactionRepo {
	return &smokeTransactionRepo{
		transactions: make(map[string]domaintransaction.Transaction),
	}
}

func (r *smokeTransactionRepo) Create(ctx context.Context, tx domaintransaction.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transactions[tx.ID] = tx
	return nil
}

func (r *smokeTransactionRepo) GetByID(ctx context.Context, id string) (domaintransaction.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, ok := r.transactions[id]
	if !ok {
		return domaintransaction.Transaction{}, postgres.ErrTransactionNotFound
	}
	return tx, nil
}

func (r *smokeTransactionRepo) List(ctx context.Context) ([]domaintransaction.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domaintransaction.Transaction, 0, len(r.transactions))
	for _, tx := range r.transactions {
		list = append(list, tx)
	}
	return list, nil
}

func setupSmokeRouter() (http.Handler, *smokeProductRepo, *smokeTransactionRepo) {
	prodRepo := newSmokeProductRepo()
	txRepo := newSmokeTransactionRepo()

	prodSvc := productservice.NewService(prodRepo)
	txSvc := transactionservice.NewService(txRepo)

	prodHandler := NewProductHandler(prodSvc)
	txHandler := NewTransactionHandler(txSvc)

	router := NewRouter(prodHandler, txHandler, nil, nil)
	return router, prodRepo, txRepo
}

func TestAPISmokeHealth(t *testing.T) {
	router, _, _ := setupSmokeRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Fatalf("expected status ok, got %q", response.Status)
	}
}

func TestAPISmokeFullFlow(t *testing.T) {
	router, _, _ := setupSmokeRouter()

	// 1. GET /products -> 200 (initially empty)
	{
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET /products: expected 200, got %d", rec.Code)
		}
	}

	// 2. POST /products -> 201
	var createdID string
	{
		payload := `{"sku":"SMOKE-01","name":"Kopi Smoke","price":18000,"stock":50}`
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("POST /products: expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var created product.Product
		if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
			t.Fatalf("decode created product: %v", err)
		}
		if created.ID == "" {
			t.Fatal("expected non-empty product ID")
		}
		createdID = created.ID
	}

	// 3. GET /products/{id} -> 200
	{
		req := httptest.NewRequest(http.MethodGet, "/products/"+createdID, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET /products/{id}: expected 200, got %d", rec.Code)
		}

		var fetched product.Product
		if err := json.NewDecoder(rec.Body).Decode(&fetched); err != nil {
			t.Fatalf("decode fetched product: %v", err)
		}
		if fetched.SKU != "SMOKE-01" {
			t.Fatalf("expected SKU SMOKE-01, got %s", fetched.SKU)
		}
	}

	// 4. PUT /products/{id} -> 200
	{
		updatePayload := `{"sku":"SMOKE-01","name":"Kopi Smoke Updated","price":20000,"stock":45}`
		req := httptest.NewRequest(http.MethodPut, "/products/"+createdID, bytes.NewBufferString(updatePayload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("PUT /products/{id}: expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// 5. DELETE /products/{id} -> 204
	{
		req := httptest.NewRequest(http.MethodDelete, "/products/"+createdID, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("DELETE /products/{id}: expected 204, got %d", rec.Code)
		}
	}

	// 6. GET /products/{id} -> 404
	{
		req := httptest.NewRequest(http.MethodGet, "/products/"+createdID, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET deleted /products/{id}: expected 404, got %d", rec.Code)
		}
	}

	// 7. GET /transactions -> 200
	{
		req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET /transactions: expected 200, got %d", rec.Code)
		}
	}

	// 8. GET /transactions/{id} -> 404
	{
		req := httptest.NewRequest(http.MethodGet, "/transactions/non-existent-id", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET /transactions/non-existent-id: expected 404, got %d", rec.Code)
		}
	}
}
