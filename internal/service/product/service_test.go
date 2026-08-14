package product

import (
	"context"
	"errors"
	"testing"
	"time"

	"pos-system/internal/domain/product"
)

type fakeRepository struct {
	created  []product.Product
	updated  []product.Product
	products []product.Product

	getByIDResult  product.Product
	getByIDErr     error
	getBySKUResult product.Product
	getBySKUErr    error

	deleteErr error
}

func (f *fakeRepository) Create(
	ctx context.Context,
	p product.Product,
) error {
	f.created = append(f.created, p)
	return nil
}

func (f *fakeRepository) GetByID(
	ctx context.Context,
	id string,
) (product.Product, error) {
	return f.getByIDResult, f.getByIDErr
}

func (f *fakeRepository) GetBySKU(
	ctx context.Context,
	sku string,
) (product.Product, error) {
	return f.getBySKUResult, f.getBySKUErr
}

func (f *fakeRepository) List(
	ctx context.Context,
) ([]product.Product, error) {
	return f.products, nil
}

func (f *fakeRepository) Update(
	ctx context.Context,
	p product.Product,
) error {
	f.updated = append(f.updated, p)
	return nil
}

func (f *fakeRepository) Delete(
	ctx context.Context,
	id string,
) error {
	return f.deleteErr
}

func TestServiceCreate(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	input := product.Product{
		ID:    "product-001",
		SKU:   "KOPI-001",
		Name:  "Kopi Susu",
		Price: 15000,
		Stock: 100,
	}

	err := service.Create(context.Background(), input)

	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf(
			"expected 1 created product, got %d",
			len(repo.created),
		)
	}

	if repo.created[0].CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be populated")
	}

	if repo.created[0].UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be populated")
	}
}

func TestServiceCreateValidation(t *testing.T) {
	tests := []struct {
		name string
		data product.Product
		want error
	}{
		{
			name: "missing sku",
			data: product.Product{
				Name: "Kopi",
			},
			want: ErrSKURequired,
		},
		{
			name: "missing name",
			data: product.Product{
				SKU: "KOPI-001",
			},
			want: ErrNameRequired,
		},
		{
			name: "negative price",
			data: product.Product{
				SKU:   "KOPI-001",
				Name:  "Kopi",
				Price: -1,
			},
			want: ErrPriceNegative,
		},
		{
			name: "negative stock",
			data: product.Product{
				SKU:   "KOPI-001",
				Name:  "Kopi",
				Stock: -1,
			},
			want: ErrStockNegative,
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

			if len(repo.created) != 0 {
				t.Fatal(
					"repository should not be called for invalid product",
				)
			}
		})
	}
}

func TestServiceGetAndList(t *testing.T) {
	now := time.Now()

	expected := product.Product{
		ID:        "product-001",
		SKU:       "KOPI-001",
		Name:      "Kopi",
		Price:     15000,
		Stock:     10,
		CreatedAt: now,
		UpdatedAt: now,
	}

	repo := &fakeRepository{
		getByIDResult:  expected,
		getBySKUResult: expected,
		products:       []product.Product{expected},
	}

	service := NewService(repo)
	ctx := context.Background()

	got, err := service.GetByID(ctx, expected.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID != expected.ID {
		t.Fatalf("unexpected product ID: %q", got.ID)
	}

	got, err = service.GetBySKU(ctx, expected.SKU)
	if err != nil {
		t.Fatalf("GetBySKU() error = %v", err)
	}

	if got.SKU != expected.SKU {
		t.Fatalf("unexpected SKU: %q", got.SKU)
	}

	list, err := service.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 product, got %d", len(list))
	}
}

func TestServiceUpdate(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	input := product.Product{
		ID:    "product-001",
		SKU:   "KOPI-001",
		Name:  "Kopi Updated",
		Price: 18000,
		Stock: 90,
	}

	err := service.Update(
		context.Background(),
		input,
	)

	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if len(repo.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(repo.updated))
	}

	if repo.updated[0].UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be updated")
	}
}

func TestServiceDelete(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		"product-001",
	)

	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
