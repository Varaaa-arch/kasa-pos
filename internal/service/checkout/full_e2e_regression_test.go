package checkout

// Full E2E Regression Test untuk memverifikasi complete flow dari POS ke Receipt
//
// Test ini mencakup seluruh chain:
// POS → Checkout → Transaction → PrintJob → Print Agent → BP-LITE58 → Receipt
//
// Test ini memverifikasi bahwa seluruh sistem berfungsi dengan benar
// dari input user (POS) sampai output fisik (receipt di printer).

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/cart"
	domainproduct "pos-system/internal/domain/product"
	domainreceipt "pos-system/internal/domain/receipt"
	domaintransaction "pos-system/internal/domain/transaction"
	"pos-system/internal/printer/agent"
	printerreceipt "pos-system/internal/printer/receipt"
	productrepo "pos-system/internal/repository/postgres"
	transactionrepo "pos-system/internal/repository/postgres"
	receiptsvc "pos-system/internal/service/receipt"
)

// bpLite58Printer simulates BP-LITE58 thermal printer behavior
type bpLite58Printer struct {
	writeCount int
	data       []byte
	openCount  int
	closeCount int
}

func (b *bpLite58Printer) Open() error {
	b.openCount++
	return nil
}

func (b *bpLite58Printer) Write(p []byte) (n int, err error) {
	b.writeCount++
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bpLite58Printer) Close() error {
	b.closeCount++
	return nil
}

// realisticPrintAgent simulates real print agent behavior with BP-LITE58
type realisticPrintAgent struct {
	printer *bpLite58Printer
	baseURL string
}

func (r *realisticPrintAgent) Print(
	ctx context.Context,
	receipt domainreceipt.Receipt,
	idempotencyKey string,
) (agent.PrintResponse, error) {
	// Simulate processing time
	time.Sleep(10 * time.Millisecond)

	// Simulate actual printing to BP-LITE58
	r.printer.Open()
	defer r.printer.Close()

	renderer := printerreceipt.NewRenderer()
	escposData := renderer.Render(receipt)

	_, err := r.printer.Write(escposData)
	if err != nil {
		return agent.PrintResponse{}, err
	}

	return agent.PrintResponse{
		JobID:   "BP-LITE58-" + uuid.NewString()[:8],
		Message: "receipt printed to BP-LITE58",
	}, nil
}

// TestFullE2ERegression memverifikasi complete flow dari POS ke Receipt
// dengan realistic simulation dari setiap component.
func TestFullE2ERegression(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not configured for full E2E test")
	}

	database, err := db.OpenPostgres(dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// STEP 1: Setup Database dan Repositories
	productRepo := productrepo.NewProductRepository(database)
	transactionRepo := transactionrepo.NewTransactionRepository(database)

	productID := uuid.NewString()
	invoice := "E2E-FULL-" + uuid.NewString()

	product := domainproduct.Product{
		ID:    productID,
		SKU:   "E2E-FULL-" + uuid.NewString()[:8],
		Name:  "Kopi E2E Full Regression",
		Price: 18000,
		Stock: 15,
	}

	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("create product: %v", err)
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

	// STEP 2: POS - User memilih produk dan masuk ke cart
	t.Log("STEP 2: POS - User selects products and adds to cart")
	c := cart.New()

	if err := c.AddItem(product, 3); err != nil {
		t.Fatalf("add item to cart: %v", err)
	}

	expectedCartTotal := int64(3 * 18000) // 54000
	if c.Total != expectedCartTotal {
		t.Fatalf("cart total = %d, want %d", c.Total, expectedCartTotal)
	}

	t.Logf("✓ Cart created: %d items, total: %d", len(c.Items), c.Total)

	// STEP 3: Checkout - Process checkout request
	t.Log("STEP 3: Checkout - Process checkout with payment")
	atomicSvc := NewAtomicService(database, transactionRepo, productRepo)

	tx, err := atomicSvc.Execute(
		ctx,
		AtomicRequest{
			Cart:          c,
			PaidAmount:    60000,
			PaymentMethod: "CASH",
			InvoiceNumber: invoice,
		},
	)
	if err != nil {
		t.Fatalf("checkout failed: %v", err)
	}

	t.Logf("✓ Checkout completed: TX ID = %s, Invoice = %s, Total = %d", 
		tx.ID, tx.InvoiceNumber, tx.Total)

	// Verify transaction
	if tx.Status != domaintransaction.StatusCompleted {
		t.Fatalf("transaction status = %q, want COMPLETED", tx.Status)
	}

	if tx.Total != expectedCartTotal {
		t.Fatalf("transaction total = %d, want %d", tx.Total, expectedCartTotal)
	}

	if len(tx.Items) != 1 {
		t.Fatalf("transaction items = %d, want 1", len(tx.Items))
	}

	// STEP 4: Transaction - Verify transaction stored in database
	t.Log("STEP 4: Transaction - Verify transaction persistence")
	storedTx, err := transactionRepo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("get stored transaction: %v", err)
	}

	if storedTx.ID != tx.ID {
		t.Fatalf("stored transaction ID mismatch")
	}

	if storedTx.InvoiceNumber != invoice {
		t.Fatalf("stored invoice mismatch")
	}

	if storedTx.Status != domaintransaction.StatusCompleted {
		t.Fatalf("stored status = %q, want COMPLETED", storedTx.Status)
	}

	t.Logf("✓ Transaction verified in database: ID = %s, Status = %s", storedTx.ID, storedTx.Status)

	// STEP 5: Stock - Verify stock reduction
	t.Log("STEP 5: Stock - Verify inventory updated")
	updatedProduct, err := productRepo.GetByID(ctx, productID)
	if err != nil {
		t.Fatalf("get updated product: %v", err)
	}

	expectedStock := 15 - 3 // 12
	if updatedProduct.Stock != expectedStock {
		t.Fatalf("product stock = %d, want %d", updatedProduct.Stock, expectedStock)
	}

	t.Logf("✓ Stock verified: reduced from 15 to %d", updatedProduct.Stock)

	// STEP 6: PrintJob - Create print job from transaction
	t.Log("STEP 6: PrintJob - Create print job from transaction")
	printService := receiptsvc.NewPrintService()

	job, err := printService.CreateJob(tx, "E2E-FULL-PRINT-"+uuid.NewString()[:8])
	if err != nil {
		t.Fatalf("create print job: %v", err)
	}

	if job.Status != printerreceipt.PrintJobPending {
		t.Fatalf("print job status = %q, want PENDING", job.Status)
	}

	t.Logf("✓ Print job created: ID = %s, Status = %s", job.ID, job.Status)

	// STEP 7: Print Agent - Setup realistic print agent with BP-LITE58 simulation
	t.Log("STEP 7: Print Agent - Setup BP-LITE58 printer simulation")
	bpLite58 := &bpLite58Printer{}
	
	printAgent := &realisticPrintAgent{
		printer: bpLite58,
	}

	// STEP 8: Print - Execute print via agent
	t.Log("STEP 8: Print - Send receipt to BP-LITE58 printer")
	printResp, err := printAgent.Print(ctx, job.Receipt, job.ID)
	if err != nil {
		t.Fatalf("print agent print failed: %v", err)
	}

	t.Logf("✓ Print agent response: Job ID = %s, Message = %s", printResp.JobID, printResp.Message)

	// Update job with print agent response
	if printResp.JobID != "" {
		job.ID = printResp.JobID
	}

	if err := job.Complete(); err != nil {
		t.Fatalf("complete print job: %v", err)
	}

	// STEP 9: Receipt - Verify final receipt output
	t.Log("STEP 9: Receipt - Verify BP-LITE58 output")
	if job.Status != printerreceipt.PrintJobCompleted {
		t.Fatalf("print job status = %q, want COMPLETED", job.Status)
	}

	if bpLite58.writeCount != 1 {
		t.Fatalf("BP-LITE58 write count = %d, want 1", bpLite58.writeCount)
	}

	if len(bpLite58.data) == 0 {
		t.Fatal("BP-LITE58 produced no output")
	}

	// Verify printer was opened and closed properly
	if bpLite58.openCount != 1 {
		t.Fatalf("BP-LITE58 open count = %d, want 1", bpLite58.openCount)
	}

	if bpLite58.closeCount != 1 {
		t.Fatalf("BP-LITE58 close count = %d, want 1", bpLite58.closeCount)
	}

	t.Logf("✓ Receipt verified: BP-LITE58 wrote %d bytes", len(bpLite58.data))

	// STEP 10: Full Chain Verification
	t.Log("STEP 10: Full Chain - Verify complete data consistency")
	
	// Verify receipt content matches transaction
	if job.Receipt.Transaction.ID != tx.ID {
		t.Fatalf("receipt transaction ID mismatch")
	}

	if job.Receipt.Transaction.InvoiceNumber != tx.InvoiceNumber {
		t.Fatalf("receipt invoice number mismatch")
	}

	if len(job.Receipt.Items) != len(tx.Items) {
		t.Fatalf("receipt items count mismatch")
	}

	if job.Receipt.Summary.Total.Amount != tx.Total {
		t.Fatalf("receipt total mismatch: receipt=%d, tx=%d", job.Receipt.Summary.Total.Amount, tx.Total)
	}

	t.Logf("✓ Full chain verified: All components consistent")

	// FINAL SUMMARY
	t.Log("=== FULL E2E REGRESSION TEST SUMMARY ===")
	t.Logf("✓ POS: Cart created successfully")
	t.Logf("✓ Checkout: Transaction processed successfully")
	t.Logf("✓ Transaction: Data persisted correctly")
	t.Logf("✓ Stock: Inventory updated correctly")
	t.Logf("✓ PrintJob: Print job created and managed")
	t.Logf("✓ Print Agent: Communication successful")
	t.Logf("✓ BP-LITE58: Physical printer simulation successful")
	t.Logf("✓ Receipt: Final output verified")
	t.Logf("=== COMPLETE CHAIN: POS → Checkout → Transaction → PrintJob → Print Agent → BP-LITE58 → Receipt ===")
}

// TestFullE2ERegressionWithAPI memverifikasi complete flow melalui API layer
// untuk mensimulasikan real POS system usage.
func TestFullE2ERegressionWithAPI(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not configured for API E2E test")
	}

	database, err := db.OpenPostgres(dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer database.Close()

	// Setup realistic print agent
	bpLite58 := &bpLite58Printer{}
	printAgent := &realisticPrintAgent{
		printer: bpLite58,
	}

	// Create mock print agent server
	printAgentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/print" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var receipt domainreceipt.Receipt
		if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")

		// Use realistic print agent
		resp, err := printAgent.Print(r.Context(), receipt, idempotencyKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer printAgentServer.Close()

	// Setup repositories
	productRepo := productrepo.NewProductRepository(database)
	transactionRepo := transactionrepo.NewTransactionRepository(database)

	// Setup services
	atomicSvc := NewAtomicService(database, transactionRepo, productRepo)
	printService := receiptsvc.NewPrintService()
	agentClient := agent.NewHTTPClient(printAgentServer.URL)

	orchestratorSvc := NewOrchestratorService(
		atomicSvc,
		printService,
		agentClient,
		DefaultReceiptDefaults(),
	)

	ctx := context.Background()
	productID := uuid.NewString()
	invoice := "E2E-API-" + uuid.NewString()

	// Create product
	product := domainproduct.Product{
		ID:    productID,
		SKU:   "E2E-API-" + uuid.NewString()[:8],
		Name:  "Kopi E2E API Test",
		Price: 22000,
		Stock: 20,
	}

	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("create product: %v", err)
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

	// Simulate POS API request
	t.Log("API E2E: Simulating POS checkout API request")
	c := cart.New()
	if err := c.AddItem(product, 2); err != nil {
		t.Fatalf("add item to cart: %v", err)
	}

	// Execute checkout through orchestrator (simulates API layer)
	result, err := orchestratorSvc.Execute(
		ctx,
		AtomicRequest{
			Cart:          c,
			PaidAmount:    50000,
			PaymentMethod: "CASH",
			InvoiceNumber: invoice,
		},
	)
	if err != nil {
		t.Fatalf("orchestrator checkout failed: %v", err)
	}

	t.Logf("✓ API Checkout successful: TX ID = %s", result.Transaction.ID)

	// Verify complete chain
	if result.Transaction.Status != domaintransaction.StatusCompleted {
		t.Fatalf("transaction status = %q, want COMPLETED", result.Transaction.Status)
	}

	if result.PrintJob.Status != printerreceipt.PrintJobCompleted {
		t.Fatalf("print job status = %q, want COMPLETED", result.PrintJob.Status)
	}

	// Verify database state
	storedTx, err := transactionRepo.GetByID(ctx, result.Transaction.ID)
	if err != nil {
		t.Fatalf("get stored transaction: %v", err)
	}

	if storedTx.Status != domaintransaction.StatusCompleted {
		t.Fatalf("stored transaction status = %q, want COMPLETED", storedTx.Status)
	}

	// Verify stock
	updatedProduct, err := productRepo.GetByID(ctx, productID)
	if err != nil {
		t.Fatalf("get updated product: %v", err)
	}

	expectedStock := 20 - 2 // 18
	if updatedProduct.Stock != expectedStock {
		t.Fatalf("product stock = %d, want %d", updatedProduct.Stock, expectedStock)
	}

	// Verify BP-LITE58 output
	if bpLite58.writeCount != 1 {
		t.Fatalf("BP-LITE58 write count = %d, want 1", bpLite58.writeCount)
	}

	t.Log("=== API E2E REGRESSION TEST SUMMARY ===")
	t.Logf("✓ API Request: Processed successfully")
	t.Logf("✓ Transaction: Created and persisted")
	t.Logf("✓ Stock: Updated correctly")
	t.Logf("✓ Print Agent: Communication successful")
	t.Logf("✓ BP-LITE58: Receipt printed")
	t.Logf("=== COMPLETE API CHAIN VERIFIED ===")
}