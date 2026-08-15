package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"pos-system/internal/db"
	"pos-system/internal/domain/transaction"
)

func TestTransactionRepositoryCreateAndRead(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	database, err := db.OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	ctx := context.Background()

	// Setup: insert product dummy yang direferensikan FK.
	productID := "00000000-0000-0000-0000-000000000001"

	_, err = database.ExecContext(
		ctx,
		`INSERT INTO products (id, sku, name, price, stock, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		 ON CONFLICT (id) DO NOTHING`,
		productID,
		"KOPI-001",
		"Kopi Susu",
		15000,
		100,
	)
	if err != nil {
		t.Fatalf("setup product: %v", err)
	}

	// Cleanup product setelah test selesai.
	t.Cleanup(func() {
		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM products WHERE id = $1`,
			productID,
		)
	})

	repo := NewTransactionRepository(database)

	now := time.Now().UTC()

	input := transaction.Transaction{
		ID:            "00000000-0000-0000-0000-000000000101",
		InvoiceNumber: "TRX-TEST-001",
		Subtotal:      30000,
		Discount:      0,
		Tax:           0,
		ServiceCharge: 0,
		Total:         30000,
		PaidAmount:    50000,
		Change:        20000,
		PaymentMethod: "CASH",
		Status:        transaction.StatusCompleted,
		CreatedAt:     now,
		Items: []transaction.Item{
			{
				ID:            "00000000-0000-0000-0000-000000000201",
				TransactionID: "00000000-0000-0000-0000-000000000101",
				ProductID:     productID,
				SKU:           "KOPI-001",
				Name:          "Kopi Susu",
				Quantity:      2,
				UnitPrice:     15000,
				Subtotal:      30000,
			},
		},
	}

	if err := repo.Create(ctx, input); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, input.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID != input.ID {
		t.Fatalf("ID = %q, want %q", got.ID, input.ID)
	}

	if got.InvoiceNumber != input.InvoiceNumber {
		t.Fatalf(
			"InvoiceNumber = %q, want %q",
			got.InvoiceNumber,
			input.InvoiceNumber,
		)
	}

	if got.Total != 30000 {
		t.Fatalf(
			"Total = %d, want 30000",
			got.Total,
		)
	}

	if len(got.Items) != 1 {
		t.Fatalf(
			"expected 1 item, got %d",
			len(got.Items),
		)
	}

	if got.Items[0].Name != "Kopi Susu" {
		t.Fatalf(
			"item name = %q, want %q",
			got.Items[0].Name,
			"Kopi Susu",
		)
	}

	transactions, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	found := false

	for _, trx := range transactions {
		if trx.ID == input.ID {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("created transaction not found in list")
	}

	// Cleanup transaction (items ikut terhapus via ON DELETE CASCADE).
	_, err = database.ExecContext(
		ctx,
		`DELETE FROM transactions WHERE id = $1`,
		input.ID,
	)
	if err != nil {
		t.Fatalf("cleanup transaction: %v", err)
	}
}
