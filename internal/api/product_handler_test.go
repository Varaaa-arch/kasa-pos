package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pos-system/internal/domain/product"
	postgres "pos-system/internal/repository/postgres"
	productservice "pos-system/internal/service/product"
)

// fakeProductRepo implements product.Repository for handler tests.
type fakeProductRepo struct {
	created []product.Product
	updated []product.Product

	getByIDResult product.Product
	getByIDErr    error

	getBySKUResult product.Product
	getBySKUErr    error

	listResult []product.Product
	deleteErr  error
}

func (f *fakeProductRepo) Create(ctx context.Context, p product.Product) error {
	f.created = append(f.created, p)
	return nil
}

func (f *fakeProductRepo) GetByID(ctx context.Context, id string) (product.Product, error) {
	return f.getByIDResult, f.getByIDErr
}

func (f *fakeProductRepo) GetBySKU(ctx context.Context, sku string) (product.Product, error) {
	return f.getBySKUResult, f.getBySKUErr
}

func (f *fakeProductRepo) List(ctx context.Context) ([]product.Product, error) {
	return f.listResult, nil
}

func (f *fakeProductRepo) Update(ctx context.Context, p product.Product) error {
	f.updated = append(f.updated, p)
	return nil
}

func (f *fakeProductRepo) Delete(ctx context.Context, id string) error {
	return f.deleteErr
}

// newTestHandler builds a ProductHandler backed by a fake repo.
func newTestHandler(repo product.Repository) *ProductHandler {
	svc := productservice.NewService(repo)
	return NewProductHandler(svc)
}

// --- POST /products ---

func TestHandlerCreate(t *testing.T) {
	repo := &fakeProductRepo{}
	h := newTestHandler(repo)

	body := `{"sku":"KOPI-001","name":"Kopi Susu","price":15000,"stock":100}`
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected 1 product created, got %d", len(repo.created))
	}

	if repo.created[0].SKU != "KOPI-001" {
		t.Fatalf("unexpected SKU: %q", repo.created[0].SKU)
	}

	var resp product.Product
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.ID == "" {
		t.Fatal("expected non-empty ID in response")
	}
}

func TestHandlerCreateInvalidBody(t *testing.T) {
	repo := &fakeProductRepo{}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandlerCreateValidationFail(t *testing.T) {
	repo := &fakeProductRepo{}
	h := newTestHandler(repo)

	// SKU kosong → ErrSKURequired dari service
	body := `{"sku":"","name":"Kopi","price":1000,"stock":10}`
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// --- GET /products ---

func TestHandlerList(t *testing.T) {
	now := time.Now()
	repo := &fakeProductRepo{
		listResult: []product.Product{
			{ID: "p1", SKU: "SKU-1", Name: "Produk 1", Price: 1000, Stock: 5, CreatedAt: now, UpdatedAt: now},
			{ID: "p2", SKU: "SKU-2", Name: "Produk 2", Price: 2000, Stock: 3, CreatedAt: now, UpdatedAt: now},
		},
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var list []product.Product
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 products, got %d", len(list))
	}
}

func TestHandlerListEmpty(t *testing.T) {
	repo := &fakeProductRepo{listResult: []product.Product{}}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- GET /products/{id} ---

func TestHandlerGetByID(t *testing.T) {
	now := time.Now()
	expected := product.Product{
		ID: "p1", SKU: "SKU-1", Name: "Kopi", Price: 15000, Stock: 10,
		CreatedAt: now, UpdatedAt: now,
	}
	repo := &fakeProductRepo{getByIDResult: expected}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/products/p1", nil)
	req.SetPathValue("id", "p1")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp product.Product
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.ID != expected.ID {
		t.Fatalf("unexpected ID: %q", resp.ID)
	}
}

func TestHandlerGetByIDNotFound(t *testing.T) {
	repo := &fakeProductRepo{getByIDErr: postgres.ErrProductNotFound}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/products/nope", nil)
	req.SetPathValue("id", "nope")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// --- PUT /products/{id} ---

func TestHandlerUpdate(t *testing.T) {
	repo := &fakeProductRepo{}
	h := newTestHandler(repo)

	body := `{"sku":"KOPI-001","name":"Kopi Updated","price":18000,"stock":90}`
	req := httptest.NewRequest(http.MethodPut, "/products/p1", bytes.NewBufferString(body))
	req.SetPathValue("id", "p1")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(repo.updated))
	}

	if repo.updated[0].ID != "p1" {
		t.Fatalf("unexpected updated ID: %q", repo.updated[0].ID)
	}
}

func TestHandlerUpdateNotFound(t *testing.T) {
	// fakeProductRepo yang Update-nya return ErrProductNotFound
	repoNotFound := &fakeProductRepoUpdateNotFound{}
	hNF := newTestHandler(repoNotFound)

	body := `{"sku":"KOPI-001","name":"Kopi","price":1000,"stock":10}`
	req := httptest.NewRequest(http.MethodPut, "/products/nope", bytes.NewBufferString(body))
	req.SetPathValue("id", "nope")
	rec := httptest.NewRecorder()

	hNF.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandlerUpdateInvalidBody(t *testing.T) {
	repo := &fakeProductRepo{}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodPut, "/products/p1", bytes.NewBufferString("bad"))
	req.SetPathValue("id", "p1")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// fakeProductRepoUpdateNotFound — Update() always returns ErrProductNotFound.
type fakeProductRepoUpdateNotFound struct {
	fakeProductRepo
}

func (f *fakeProductRepoUpdateNotFound) Update(ctx context.Context, p product.Product) error {
	return postgres.ErrProductNotFound
}

// --- DELETE /products/{id} ---

func TestHandlerDelete(t *testing.T) {
	repo := &fakeProductRepo{}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/products/p1", nil)
	req.SetPathValue("id", "p1")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestHandlerDeleteNotFound(t *testing.T) {
	repo := &fakeProductRepo{deleteErr: postgres.ErrProductNotFound}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/products/nope", nil)
	req.SetPathValue("id", "nope")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
