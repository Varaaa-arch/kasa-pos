package checkout

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/cart"
	domainproduct "pos-system/internal/domain/product"
	domaintransaction "pos-system/internal/domain/transaction"
	"pos-system/internal/printer/mock"
	printerreceipt "pos-system/internal/printer/receipt"
	productrepo "pos-system/internal/repository/postgres"
	transactionrepo "pos-system/internal/repository/postgres"
	receiptsvc "pos-system/internal/service/receipt"
)

func TestEndToEndCheckout(t *testing.T) {
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
	invoice := "E2E-" + uuid.NewString()

	productRepo := productrepo.NewProductRepository(database)
	transactionRepo := transactionrepo.NewTransactionRepository(database)

	product := domainproduct.Product{
		ID:    productID,
		SKU:   "E2E-" + uuid.NewString(),
		Name:  "Kopi E2E",
		Price: 15000,
		Stock: 10,
	}

	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM transactions WHERE invoice_number = $1`,
			invoice,
		)

		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM products WHERE id = $1`,
			productID,
		)
	})

	// Product → Cart
	c := cart.New()

	if err := c.AddItem(product, 2); err != nil {
		t.Fatal(err)
	}

	if c.Total != 30000 {
		t.Fatalf(
			"cart total = %d, want 30000",
			c.Total,
		)
	}

	// Cart → Checkout → Transaction + Stock
	atomic := NewAtomicService(
		database,
		transactionRepo,
		productRepo,
	)

	tx, err := atomic.Execute(
		ctx,
		AtomicRequest{
			Cart:          c,
			PaidAmount:    50000,
			PaymentMethod: "CASH",
			InvoiceNumber: invoice,
		},
	)
	if err != nil {
		t.Fatalf(
			"checkout failed: %v",
			err,
		)
	}

	if tx.Status != domaintransaction.StatusCompleted {
		t.Fatalf(
			"status = %q, want COMPLETED",
			tx.Status,
		)
	}

	if tx.Total != 30000 {
		t.Fatalf(
			"transaction total = %d, want 30000",
			tx.Total,
		)
	}

	if tx.Change != 20000 {
		t.Fatalf(
			"change = %d, want 20000",
			tx.Change,
		)
	}

	// Verify stock reduction.
	var stock int

	err = database.QueryRowContext(
		ctx,
		`SELECT stock FROM products WHERE id = $1`,
		productID,
	).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}

	if stock != 8 {
		t.Fatalf(
			"stock = %d, want 8",
			stock,
		)
	}

	// Transaction → Receipt → PrintJob
	printService := receiptsvc.NewPrintService()

	job, err := printService.CreateJob(
		tx,
		"E2E-PRINT-001",
	)
	if err != nil {
		t.Fatal(err)
	}

	if job.Status != printerreceipt.PrintJobPending {
		t.Fatalf(
			"job status = %q, want PENDING",
			job.Status,
		)
	}

	printer := &mock.Printer{}
	renderer := printerreceipt.NewRenderer()

	if err := job.Run(
		printer,
		renderer,
	); err != nil {
		t.Fatalf(
			"print failed: %v",
			err,
		)
	}

	if job.Status != printerreceipt.PrintJobCompleted {
		t.Fatalf(
			"job status = %q, want COMPLETED",
			job.Status,
		)
	}

	if printer.WriteCount != 1 {
		t.Fatalf(
			"write count = %d, want 1",
			printer.WriteCount,
		)
	}

	if len(printer.Data) == 0 {
		t.Fatal("expected printer output")
	}
}
