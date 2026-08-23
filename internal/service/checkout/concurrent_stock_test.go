package checkout

// Unit test concurrent stock tanpa database.
//
// Pendekatan:
//   Karena SELECT FOR UPDATE hanya ada di PostgreSQL dan SQLite tidak mendukungnya,
//   race simulation dilakukan dengan fakeStockRepo yang:
//   - punya mutex internal (seperti yang dilakukan DB dengan row lock)
//   - hanya mengizinkan satu goroutine mengurangi stock ketika stock = 1
//
// Ini membuktikan bahwa AtomicService + ReduceStockTx menghasilkan
// exactly-once semantics: A ✅, B ❌, stock = 0.

import (
	"context"
	"sync"
	"testing"

	"pos-system/internal/db"
	"pos-system/internal/domain/cart"
	"pos-system/internal/domain/product"
	"pos-system/internal/repository/postgres"
)

// serialStockRepo mensimulasikan SELECT FOR UPDATE dengan mutex:
// hanya satu goroutine yang bisa cek+kurangi stock pada satu waktu.
// Ini setara dengan apa yang dilakukan PostgreSQL row lock.
type serialStockRepo struct {
	mu    sync.Mutex
	stock int
}

func (r *serialStockRepo) ReduceStockTx(
	_ context.Context,
	_ db.DBTX,
	_ string,
	quantity int,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stock < quantity {
		return postgres.ErrInsufficientStock
	}

	r.stock -= quantity
	return nil
}

// TestConcurrentStock_ExactlyOne membuktikan bahwa ketika stock = 1
// dan dua goroutine checkout bersamaan, hanya satu yang berhasil.
func TestConcurrentStock_ExactlyOne(t *testing.T) {
	stockRepo := &serialStockRepo{stock: 1}
	txRepo := &fakeTransactionRepo{}
	svc := NewAtomicService(openSQLiteDB(t), txRepo, stockRepo)

	makeCart := func() *cart.Cart {
		c := cart.New()
		p := product.Product{
			ID:    "prod-race",
			SKU:   "RACE-001",
			Name:  "Race Product",
			Price: 10000,
			Stock: 1,
		}
		if err := c.AddItem(p, 1); err != nil {
			t.Fatal(err)
		}
		return c
	}

	type result struct {
		err error
	}

	results := make(chan result, 2)

	// Jalankan 2 checkout bersamaan
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := svc.Execute(context.Background(), AtomicRequest{
				Cart:          makeCart(),
				PaidAmount:    10000,
				PaymentMethod: "CASH",
				InvoiceNumber: "INV-RACE-" + string(rune('A'+i)),
			})
			results <- result{err: err}
		}(i)
	}

	wg.Wait()
	close(results)

	successes, failures := 0, 0
	for r := range results {
		if r.err == nil {
			successes++
		} else {
			failures++
		}
	}

	// Exactly one success, one failure
	if successes != 1 {
		t.Errorf("expected 1 success, got %d", successes)
	}
	if failures != 1 {
		t.Errorf("expected 1 failure, got %d", failures)
	}

	// Stock harus 0 setelah satu berhasil
	if stockRepo.stock != 0 {
		t.Errorf("expected stock=0, got %d", stockRepo.stock)
	}
}

// TestConcurrentStock_AllFail membuktikan bahwa ketika stock = 0
// dari awal, semua checkout gagal.
func TestConcurrentStock_AllFail(t *testing.T) {
	stockRepo := &serialStockRepo{stock: 0}
	txRepo := &fakeTransactionRepo{}
	svc := NewAtomicService(openSQLiteDB(t), txRepo, stockRepo)

	makeCart := func() *cart.Cart {
		c := cart.New()
		p := product.Product{
			ID:    "prod-empty",
			SKU:   "EMPTY-001",
			Name:  "Empty Stock",
			Price: 5000,
			Stock: 1, // cart check pakai in-memory stock
		}
		if err := c.AddItem(p, 1); err != nil {
			t.Fatal(err)
		}
		return c
	}

	results := make(chan error, 3)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := svc.Execute(context.Background(), AtomicRequest{
				Cart:          makeCart(),
				PaidAmount:    5000,
				PaymentMethod: "CASH",
				InvoiceNumber: "INV-EMPTY-" + string(rune('A'+i)),
			})
			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	for err := range results {
		if err == nil {
			t.Error("expected all checkouts to fail when stock=0")
		}
	}

	if stockRepo.stock != 0 {
		t.Errorf("stock should remain 0, got %d", stockRepo.stock)
	}
}

// TestConcurrentStock_MultipleAvailable membuktikan bahwa ketika stock = 3
// dan 3 goroutine checkout bersamaan, semua berhasil.
func TestConcurrentStock_MultipleAvailable(t *testing.T) {
	stockRepo := &serialStockRepo{stock: 3}
	txRepo := &fakeTransactionRepo{}
	svc := NewAtomicService(openSQLiteDB(t), txRepo, stockRepo)

	makeCart := func() *cart.Cart {
		c := cart.New()
		p := product.Product{
			ID:    "prod-multi",
			SKU:   "MULTI-001",
			Name:  "Multi Stock",
			Price: 8000,
			Stock: 3,
		}
		if err := c.AddItem(p, 1); err != nil {
			t.Fatal(err)
		}
		return c
	}

	results := make(chan error, 3)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := svc.Execute(context.Background(), AtomicRequest{
				Cart:          makeCart(),
				PaidAmount:    8000,
				PaymentMethod: "CASH",
				InvoiceNumber: "INV-MULTI-" + string(rune('A'+i)),
			})
			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			t.Errorf("expected all to succeed, got: %v", err)
		}
	}

	if stockRepo.stock != 0 {
		t.Errorf("expected stock=0 after 3 checkouts, got %d", stockRepo.stock)
	}
}
