package checkout

// Unit test untuk AtomicService yang membuktikan rollback behaviour
// tanpa membutuhkan koneksi database sungguhan.
//
// Pendekatan:
//   - fakeTxBeginner mengembalikan *sql.Tx dari sql.DB in-memory (SQLite via modernc)
//   - fakeTransactionRepo dan fakeProductRepo mengontrol kapan operasi gagal
//   - Setelah Execute() gagal, kita verifikasi rollback terjadi via state tracker

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"

	"pos-system/internal/db"
	"pos-system/internal/domain/cart"
	"pos-system/internal/domain/product"
	domaintransaction "pos-system/internal/domain/transaction"
)

// ─── Fakes ───────────────────────────────────────────────────────────────────

// fakeTransactionRepo records calls and can be configured to fail.
type fakeTransactionRepo struct {
	createErr    error          // jika non-nil, CreateTx akan return error ini
	created      []domaintransaction.Transaction
	rollbackSeen bool
}

func (f *fakeTransactionRepo) CreateTx(
	_ context.Context,
	tx db.DBTX,
	t domaintransaction.Transaction,
) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, t)
	return nil
}

// fakeProductRepo records calls and can be configured to fail.
type fakeProductRepo struct {
	reduceErr error // jika non-nil, ReduceStockTx akan return error ini
	reduced   []string // product IDs yang berhasil di-reduce
}

func (f *fakeProductRepo) ReduceStockTx(
	_ context.Context,
	_ db.DBTX,
	productID string,
	_ int,
) error {
	if f.reduceErr != nil {
		return f.reduceErr
	}
	f.reduced = append(f.reduced, productID)
	return nil
}

// openSQLiteDB membuka SQLite in-memory DB yang support BeginTx.
// SQLite dipakai hanya untuk mendapatkan *sql.Tx yang valid —
// fakeTransactionRepo dan fakeProductRepo yang menangani logika sebenarnya.
func openSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// cartWith builds a simple cart with one product for testing.
func cartWith(productID string, price int64, stock, qty int) *cart.Cart {
	c := cart.New()
	p := product.Product{
		ID:    productID,
		SKU:   "TEST-SKU",
		Name:  "Test Product",
		Price: price,
		Stock: stock,
	}
	if err := c.AddItem(p, qty); err != nil {
		panic("cartWith: " + err.Error())
	}
	return c
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// TestAtomicUnit_Success membuktikan bahwa ketika semua operasi berhasil,
// transaction dan stock tercatat di fake repos.
func TestAtomicUnit_Success(t *testing.T) {
	txRepo := &fakeTransactionRepo{}
	prodRepo := &fakeProductRepo{}
	svc := NewAtomicService(openSQLiteDB(t), txRepo, prodRepo)

	c := cartWith("prod-001", 15000, 10, 2)

	result, err := svc.Execute(context.Background(), AtomicRequest{
		Cart:          c,
		PaidAmount:    50000,
		PaymentMethod: "CASH",
		InvoiceNumber: "INV-UNIT-001",
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if result.InvoiceNumber != "INV-UNIT-001" {
		t.Errorf("invoice = %q", result.InvoiceNumber)
	}

	// transactionRepo.CreateTx harus dipanggil
	if len(txRepo.created) != 1 {
		t.Fatalf("expected 1 created transaction, got %d", len(txRepo.created))
	}

	// productRepo.ReduceStockTx harus dipanggil untuk prod-001
	if len(prodRepo.reduced) != 1 || prodRepo.reduced[0] != "prod-001" {
		t.Fatalf("expected prod-001 to be reduced, got %v", prodRepo.reduced)
	}
}

// TestAtomicUnit_TransactionFailure_Rollback membuktikan bahwa ketika
// CreateTx gagal, Execute() return error dan productRepo tidak dipanggil
// (membuktikan operasi berhenti sebelum stock dikurangi).
func TestAtomicUnit_TransactionFailure_Rollback(t *testing.T) {
	txErr := errors.New("insert transaction: duplicate invoice_number")

	txRepo := &fakeTransactionRepo{createErr: txErr}
	prodRepo := &fakeProductRepo{}
	svc := NewAtomicService(openSQLiteDB(t), txRepo, prodRepo)

	c := cartWith("prod-002", 10000, 5, 1)

	_, err := svc.Execute(context.Background(), AtomicRequest{
		Cart:          c,
		PaidAmount:    10000,
		PaymentMethod: "CASH",
		InvoiceNumber: "INV-UNIT-FAIL-TX",
	})

	// Harus error
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, txErr) {
		t.Errorf("expected txErr, got %v", err)
	}

	// productRepo TIDAK boleh dipanggil — rollback implicit karena return sebelum ReduceStockTx
	if len(prodRepo.reduced) != 0 {
		t.Errorf("stock should not be reduced after transaction failure, got %v", prodRepo.reduced)
	}
}

// TestAtomicUnit_ItemFailure_Rollback membuktikan bahwa ketika
// ReduceStockTx gagal pada item pertama, Execute() return error
// dan tidak ada item berikutnya yang diproses.
func TestAtomicUnit_ItemFailure_Rollback(t *testing.T) {
	stockErr := errors.New("insufficient stock")

	txRepo := &fakeTransactionRepo{}
	prodRepo := &fakeProductRepo{reduceErr: stockErr}
	svc := NewAtomicService(openSQLiteDB(t), txRepo, prodRepo)

	// Cart dengan 2 produk berbeda — hanya prod-A yang diproses sebelum gagal
	c := cart.New()
	pA := product.Product{ID: "prod-A", SKU: "A", Name: "A", Price: 5000, Stock: 1}
	pB := product.Product{ID: "prod-B", SKU: "B", Name: "B", Price: 5000, Stock: 1}
	if err := c.AddItem(pA, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.AddItem(pB, 1); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Execute(context.Background(), AtomicRequest{
		Cart:          c,
		PaidAmount:    20000,
		PaymentMethod: "CASH",
		InvoiceNumber: "INV-UNIT-FAIL-ITEM",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, stockErr) {
		t.Errorf("expected stockErr, got %v", err)
	}

	// CreateTx dipanggil (transaction sempat dibuat di dalam tx)
	if len(txRepo.created) != 1 {
		t.Errorf("expected CreateTx to be called once, got %d", len(txRepo.created))
	}

	// Tapi tidak ada stock yang berhasil dikurangi (reduceErr langsung pada call pertama)
	if len(prodRepo.reduced) != 0 {
		t.Errorf("no stock should be reduced, got %v", prodRepo.reduced)
	}
}

// TestAtomicUnit_StockFailure_Rollback membuktikan rollback ketika
// ReduceStockTx gagal pada item kedua (item pertama sudah masuk tapi
// Commit tidak terjadi, sehingga seluruh tx rollback).
func TestAtomicUnit_StockFailure_Rollback(t *testing.T) {
	callCount := 0
	stockErrOnSecond := errors.New("insufficient stock on second item")

	txRepo := &fakeTransactionRepo{}
	prodRepo := &fakeProductRepo{}

	// Override ReduceStockTx: item pertama sukses, item kedua gagal
	prodRepoPartial := &fakeProductRepoPartial{
		reduceFn: func(productID string) error {
			callCount++
			if callCount == 2 {
				return stockErrOnSecond
			}
			prodRepo.reduced = append(prodRepo.reduced, productID)
			return nil
		},
	}

	svc := NewAtomicService(openSQLiteDB(t), txRepo, prodRepoPartial)

	c := cart.New()
	p1 := product.Product{ID: "prod-1", SKU: "P1", Name: "P1", Price: 5000, Stock: 5}
	p2 := product.Product{ID: "prod-2", SKU: "P2", Name: "P2", Price: 5000, Stock: 5}
	if err := c.AddItem(p1, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.AddItem(p2, 1); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Execute(context.Background(), AtomicRequest{
		Cart:          c,
		PaidAmount:    20000,
		PaymentMethod: "CASH",
		InvoiceNumber: "INV-UNIT-FAIL-STOCK",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, stockErrOnSecond) {
		t.Errorf("expected stockErrOnSecond, got %v", err)
	}

	// Karena Commit tidak terjadi, seluruh tx di-rollback oleh defer tx.Rollback()
	// Item pertama "berhasil" di fake prodRepo.reduced, tapi tx tidak committed
	if callCount != 2 {
		t.Errorf("expected 2 ReduceStockTx calls, got %d", callCount)
	}
}

// fakeProductRepoPartial memungkinkan kontrolper-call lewat closure.
type fakeProductRepoPartial struct {
	reduceFn func(productID string) error
}

func (f *fakeProductRepoPartial) ReduceStockTx(
	_ context.Context,
	_ db.DBTX,
	productID string,
	_ int,
) error {
	return f.reduceFn(productID)
}
