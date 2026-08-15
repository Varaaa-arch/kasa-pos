package checkout

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/cart"
	"pos-system/internal/domain/product"
	"pos-system/internal/repository/postgres"
)

func setupAtomicTest(t *testing.T) (
	*AtomicService,
	*postgres.ProductRepository,
	*postgres.TransactionRepository,
	func(),
) {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	database, err := db.OpenPostgres(dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	productRepo := postgres.NewProductRepository(database)
	transactionRepo := postgres.NewTransactionRepository(database)
	svc := NewAtomicService(database, transactionRepo, productRepo)

	cleanup := func() {
		database.Close()
	}

	return svc, productRepo, transactionRepo, cleanup
}

// TestAtomicCheckoutSuccess verifies that a successful checkout
// persists the transaction and reduces product stock atomically.
func TestAtomicCheckoutSuccess(t *testing.T) {
	svc, productRepo, transactionRepo, cleanup := setupAtomicTest(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: buat product dengan stock 10.
	p := product.Product{
		ID:    "00000000-0000-0000-0000-000000000011",
		SKU:   "ATOMIC-001",
		Name:  "Kopi Susu",
		Price: 15000,
		Stock: 10,
	}

	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("setup product: %v", err)
	}

	t.Cleanup(func() {
		db, _ := db.OpenPostgres(os.Getenv("DATABASE_URL"))
		if db != nil {
			db.ExecContext(ctx,
				`DELETE FROM products WHERE id = $1`, p.ID,
			)
			db.Close()
		}
	})

	// Buat cart dengan quantity 2.
	c := cart.New()
	if err := c.AddItem(p, 2); err != nil {
		t.Fatalf("add item: %v", err)
	}

	result, err := svc.Execute(ctx, AtomicRequest{
		Cart:          c,
		PaidAmount:    50000,
		PaymentMethod: "CASH",
		InvoiceNumber: "INV-ATOMIC-SUCCESS-001",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Assert: transaction tersimpan.
	got, err := transactionRepo.GetByID(ctx, result.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.InvoiceNumber != "INV-ATOMIC-SUCCESS-001" {
		t.Fatalf(
			"InvoiceNumber = %q, want %q",
			got.InvoiceNumber,
			"INV-ATOMIC-SUCCESS-001",
		)
	}

	if got.Total != 30000 {
		t.Fatalf("Total = %d, want 30000", got.Total)
	}

	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}

	// Assert: stock berkurang dari 10 → 8.
	updatedProduct, err := productRepo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID product: %v", err)
	}

	if updatedProduct.Stock != 8 {
		t.Fatalf(
			"stock = %d, want 8 (was 10, bought 2)",
			updatedProduct.Stock,
		)
	}

	// Cleanup transaction.
	t.Cleanup(func() {
		db, _ := db.OpenPostgres(os.Getenv("DATABASE_URL"))
		if db != nil {
			db.ExecContext(ctx,
				`DELETE FROM transactions WHERE id = $1`, result.ID,
			)
			db.Close()
		}
	})
}

// TestAtomicCheckoutRollbackOnInsufficientStock verifies that when
// stock reduction fails, the transaction is NOT persisted.
// This proves sql.Tx ROLLBACK works correctly.
func TestAtomicCheckoutRollbackOnInsufficientStock(t *testing.T) {
	svc, productRepo, _, cleanup := setupAtomicTest(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: buat product di DB dengan stock 1 saja.
	dbProduct := product.Product{
		ID:    "00000000-0000-0000-0000-000000000012",
		SKU:   "ATOMIC-002",
		Name:  "Roti Bakar",
		Price: 12000,
		Stock: 1,
	}

	if err := productRepo.Create(ctx, dbProduct); err != nil {
		t.Fatalf("setup product: %v", err)
	}

	t.Cleanup(func() {
		db, _ := db.OpenPostgres(os.Getenv("DATABASE_URL"))
		if db != nil {
			db.ExecContext(ctx,
				`DELETE FROM products WHERE id = $1`, dbProduct.ID,
			)
			db.Close()
		}
	})

	// Buat cart dengan product yang stock-nya di-set tinggi di memory
	// supaya cart.AddItem tidak menolak — tapi database punya stock 1.
	// Ini mensimulasikan race condition atau stale data.
	cartProduct := product.Product{
		ID:    dbProduct.ID,
		SKU:   dbProduct.SKU,
		Name:  dbProduct.Name,
		Price: dbProduct.Price,
		Stock: 100, // stock di memory tinggi, tapi DB = 1
	}

	c := cart.New()
	if err := c.AddItem(cartProduct, 5); err != nil {
		t.Fatalf("add item: %v", err)
	}

	var err error
	_, err = svc.Execute(ctx, AtomicRequest{
		Cart:          c,
		PaidAmount:    100000,
		PaymentMethod: "CASH",
		InvoiceNumber: "INV-ATOMIC-ROLLBACK-001",
	})

	// Harus error karena stock DB tidak cukup (1 < 5).
	if !errors.Is(err, postgres.ErrInsufficientStock) {
		t.Fatalf(
			"expected ErrInsufficientStock, got %v",
			err,
		)
	}

	// Bukti ROLLBACK: transaction TIDAK boleh tersimpan.
	// Query langsung via invoice number karena result.ID adalah zero-value.
	dbConn, err := db.OpenPostgres(os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("open db for verification: %v", err)
	}
	defer dbConn.Close()

	var count int
	dbConn.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM transactions WHERE invoice_number = $1`,
		"INV-ATOMIC-ROLLBACK-001",
	).Scan(&count)

	if count != 0 {
		t.Fatalf(
			"transaction should not exist after rollback, found %d row(s)",
			count,
		)
	}

	// Stock harus tetap 1 — tidak berubah.
	unchanged, err := productRepo.GetByID(ctx, dbProduct.ID)
	if err != nil {
		t.Fatalf("GetByID product: %v", err)
	}

	if unchanged.Stock != 1 {
		t.Fatalf(
			"stock = %d, want 1 (should be unchanged after rollback)",
			unchanged.Stock,
		)
	}
}

func TestAtomicCheckoutConcurrentStock(t *testing.T) {
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

	productID := uuid.NewString()
	transactionRepo := postgres.NewTransactionRepository(database)
	productRepo := postgres.NewProductRepository(database)

	now := time.Now().UTC()

	concurrentSKU := "CONCURRENT-" + uuid.NewString()

	_, err = database.ExecContext(
		ctx,
		`
		INSERT INTO products (
			id, sku, name, price, stock, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		productID,
		concurrentSKU,
		"Concurrent Test",
		10000,
		1,
		now,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	testRunID := uuid.NewString()[:8]
	invoicePrefix := fmt.Sprintf("CONCURRENT-%s-", testRunID)

	t.Cleanup(func() {
		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM transaction_items WHERE transaction_id IN (SELECT id FROM transactions WHERE invoice_number LIKE $1)`,
			invoicePrefix+"%",
		)

		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM transactions WHERE invoice_number LIKE $1`,
			invoicePrefix+"%",
		)

		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM products WHERE id = $1`,
			productID,
		)
	})

	makeCart := func() *cart.Cart {
		c := cart.New()

		p := product.Product{
			ID:    productID,
			SKU:   concurrentSKU,
			Name:  "Concurrent Test",
			Price: 10000,
			Stock: 1,
		}

		if err := c.AddItem(p, 1); err != nil {
			t.Fatal(err)
		}

		return c
	}

	service := NewAtomicService(
		database,
		transactionRepo,
		productRepo,
	)

	var wg sync.WaitGroup

	results := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			_, err := service.Execute(
				ctx,
				AtomicRequest{
					Cart:          makeCart(),
					PaidAmount:    10000,
					PaymentMethod: "CASH",
					InvoiceNumber: fmt.Sprintf(
						"%s%d",
						invoicePrefix,
						i,
					),
				},
			)

			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	successes := 0
	failures := 0

	for err := range results {
		if err == nil {
			successes++
			continue
		}

		failures++
	}

	if successes != 1 {
		t.Fatalf(
			"expected exactly 1 successful checkout, got %d",
			successes,
		)
	}

	if failures != 1 {
		t.Fatalf(
			"expected exactly 1 failed checkout, got %d",
			failures,
		)
	}

	var stock int

	err = database.QueryRowContext(
		ctx,
		`SELECT stock FROM products WHERE id = $1`,
		productID,
	).Scan(&stock)

	if err != nil {
		t.Fatal(err)
	}

	if stock != 0 {
		t.Fatalf(
			"expected final stock 0, got %d",
			stock,
		)
	}
}
