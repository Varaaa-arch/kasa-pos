package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/product"
)

func TestProductRepositoryCRUD(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	database, err := db.OpenPostgres(dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()

	repo := NewProductRepository(database)
	ctx := context.Background()

	now := time.Now().UTC()

	productID := uuid.NewString()
	sku := "TEST-" + uuid.NewString()

	input := product.Product{
		ID:        productID,
		SKU:       sku,
		Name:      "Kopi Susu",
		Price:     15000,
		Stock:     100,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Cleanup dipasang segera — walaupun assertion gagal, row tetap terhapus.
	t.Cleanup(func() {
		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM products WHERE id = $1`,
			productID,
		)
	})

	// CREATE
	if err := repo.Create(ctx, input); err != nil {
		t.Fatalf("create: %v", err)
	}

	// GET BY ID
	got, err := repo.GetByID(ctx, input.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}

	if got.SKU != input.SKU {
		t.Fatalf("SKU = %q, want %q", got.SKU, input.SKU)
	}

	if got.Name != input.Name {
		t.Fatalf("Name = %q, want %q", got.Name, input.Name)
	}

	if got.Price != input.Price {
		t.Fatalf("Price = %d, want %d", got.Price, input.Price)
	}

	if got.Stock != input.Stock {
		t.Fatalf("Stock = %d, want %d", got.Stock, input.Stock)
	}

	// GET BY SKU
	got, err = repo.GetBySKU(ctx, input.SKU)
	if err != nil {
		t.Fatalf("get by sku: %v", err)
	}

	if got.ID != input.ID {
		t.Fatalf("ID = %q, want %q", got.ID, input.ID)
	}

	// LIST
	products, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	found := false

	for _, p := range products {
		if p.ID == input.ID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("created product not found in list")
	}

	// UPDATE
	input.Name = "Kopi Susu Gula Aren"
	input.Price = 18000
	input.Stock = 95
	input.UpdatedAt = time.Now().UTC()

	if err := repo.Update(ctx, input); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err = repo.GetByID(ctx, input.ID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}

	if got.Name != input.Name {
		t.Fatalf("updated Name = %q, want %q", got.Name, input.Name)
	}

	if got.Price != input.Price {
		t.Fatalf("updated Price = %d, want %d", got.Price, input.Price)
	}

	if got.Stock != input.Stock {
		t.Fatalf("updated Stock = %d, want %d", got.Stock, input.Stock)
	}

	// DELETE
	if err := repo.Delete(ctx, input.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = repo.GetByID(ctx, input.ID)
	if err != ErrProductNotFound {
		t.Fatalf(
			"expected ErrProductNotFound, got %v",
			err,
		)
	}
}
