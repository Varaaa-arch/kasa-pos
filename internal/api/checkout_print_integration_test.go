package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/product"
	domaintransaction "pos-system/internal/domain/transaction"
	"pos-system/internal/printer/agent"
	"pos-system/internal/repository/postgres"
	checkoutsvc "pos-system/internal/service/checkout"
	productservice "pos-system/internal/service/product"
	receiptsvc "pos-system/internal/service/receipt"
	transactionservice "pos-system/internal/service/transaction"
)

func setupCheckoutIntegrationRouter(
	t *testing.T,
	printAgentURL string,
) (*httptest.Server, *sql.DB, *postgres.ProductRepository, *postgres.TransactionRepository, func()) {
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

	atomicSvc := checkoutsvc.NewAtomicService(database, transactionRepo, productRepo)
	orchestratorSvc := checkoutsvc.NewOrchestratorService(
		atomicSvc,
		receiptsvc.NewPrintService(),
		agent.NewHTTPClient(printAgentURL),
		checkoutsvc.DefaultReceiptDefaults(),
	)

	productHandler := NewProductHandler(productservice.NewService(productRepo))
	checkoutHandler := NewCheckoutHandler(orchestratorSvc, productRepo)
	transactionHandler := NewTransactionHandler(
		transactionservice.NewService(transactionRepo),
		orchestratorSvc,
	)

	server := httptest.NewServer(
		NewRouter(productHandler, transactionHandler, checkoutHandler, database),
	)

	cleanup := func() {
		server.Close()
		database.Close()
	}

	return server, database, productRepo, transactionRepo, cleanup
}

func TestCheckoutIntegrationPrintAgentSuccess(t *testing.T) {
	var received checkoutPrintPayload
	var idempotencyKey string

	printAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/print" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		idempotencyKey = r.Header.Get("Idempotency-Key")

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agent.PrintResponse{
			JobID:   "PJ-integration-001",
			Message: "receipt printed successfully",
		})
	}))
	defer printAgent.Close()

	apiServer, database, productRepo, transactionRepo, cleanup := setupCheckoutIntegrationRouter(
		t,
		printAgent.URL,
	)
	defer cleanup()

	ctx := context.Background()
	productID := uuid.NewString()
	invoice := "INT-PRINT-OK-" + uuid.NewString()

	p := product.Product{
		ID:    productID,
		SKU:   "INT-PRINT-OK-" + uuid.NewString()[:8],
		Name:  "Kopi Integration OK",
		Price: 25000,
		Stock: 12,
	}

	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("create product: %v", err)
	}

	t.Cleanup(func() {
		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM transactions WHERE invoice_number = $1`,
			invoice,
		)
		_ = productRepo.Delete(ctx, productID)
	})

	body := map[string]any{
		"items": []map[string]any{
			{"product_id": productID, "quantity": 2},
		},
		"paid_amount":    60000,
		"payment_method": "CASH",
		"invoice_number": invoice,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	resp, err := http.Post(
		apiServer.URL+"/checkout",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		t.Fatalf("post checkout: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var checkoutResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&checkoutResp); err != nil {
		t.Fatalf("decode checkout response: %v", err)
	}

	transactionID, ok := checkoutResp["transaction_id"].(string)
	if !ok || transactionID == "" {
		t.Fatalf("missing transaction_id")
	}

	if checkoutResp["status"] != "COMPLETED" {
		t.Fatalf("status = %v, want COMPLETED", checkoutResp["status"])
	}

	printJob, ok := checkoutResp["print_job"].(map[string]any)
	if !ok {
		t.Fatalf("missing print_job")
	}

	if printJob["status"] != "COMPLETED" {
		t.Fatalf("print_job status = %v, want COMPLETED", printJob["status"])
	}

	updated, err := productRepo.GetByID(ctx, productID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}

	if updated.Stock != 10 {
		t.Fatalf("stock = %d, want 10", updated.Stock)
	}

	stored, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}

	if stored.Status != domaintransaction.StatusCompleted {
		t.Fatalf("stored status = %q, want COMPLETED", stored.Status)
	}

	if idempotencyKey != transactionID {
		t.Fatalf("idempotency key = %q, want %q", idempotencyKey, transactionID)
	}

	if received.Transaction.ID != transactionID {
		t.Fatalf("receipt transaction id = %q, want %q", received.Transaction.ID, transactionID)
	}

	if received.Transaction.InvoiceNumber != invoice {
		t.Fatalf("receipt invoice = %q, want %q", received.Transaction.InvoiceNumber, invoice)
	}

	if len(received.Items) != 1 {
		t.Fatalf("receipt items = %d, want 1", len(received.Items))
	}

	if received.Items[0].SKU != p.SKU {
		t.Fatalf("receipt sku = %q, want %q", received.Items[0].SKU, p.SKU)
	}

	if received.Items[0].Quantity != 2 {
		t.Fatalf("receipt quantity = %d, want 2", received.Items[0].Quantity)
	}

	if received.Items[0].UnitPrice != 25000 {
		t.Fatalf("receipt unit price = %d, want 25000", received.Items[0].UnitPrice)
	}

	if received.Summary.Subtotal != 50000 {
		t.Fatalf("receipt subtotal = %d, want 50000", received.Summary.Subtotal)
	}

	if received.Summary.Total != 50000 {
		t.Fatalf("receipt total = %d, want 50000", received.Summary.Total)
	}

	if received.Payment.Method != "CASH" {
		t.Fatalf("receipt payment method = %q, want CASH", received.Payment.Method)
	}

	if received.Payment.Paid != 60000 {
		t.Fatalf("receipt paid = %d, want 60000", received.Payment.Paid)
	}

	if received.Payment.Change != 10000 {
		t.Fatalf("receipt change = %d, want 10000", received.Payment.Change)
	}

	if received.Store.Name != "TOKO KASA" {
		t.Fatalf("receipt store = %q, want TOKO KASA", received.Store.Name)
	}

	if received.Footer.Message != "Terima kasih" {
		t.Fatalf("receipt footer = %q, want Terima kasih", received.Footer.Message)
	}
}

func TestCheckoutIntegrationPrintAgentUnavailable(t *testing.T) {
	unavailable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusInternalServerError)
	}))
	defer unavailable.Close()

	apiServer, database, productRepo, transactionRepo, cleanup := setupCheckoutIntegrationRouter(
		t,
		unavailable.URL,
	)
	defer cleanup()

	ctx := context.Background()
	productID := uuid.NewString()
	invoice := "INT-PRINT-FAIL-" + uuid.NewString()

	p := product.Product{
		ID:    productID,
		SKU:   "INT-PRINT-FAIL-" + uuid.NewString()[:8],
		Name:  "Kopi Integration Fail",
		Price: 18000,
		Stock: 8,
	}

	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("create product: %v", err)
	}

	t.Cleanup(func() {
		_, _ = database.ExecContext(
			ctx,
			`DELETE FROM transactions WHERE invoice_number = $1`,
			invoice,
		)
		_ = productRepo.Delete(ctx, productID)
	})

	body := map[string]any{
		"items": []map[string]any{
			{"product_id": productID, "quantity": 1},
		},
		"paid_amount":    25000,
		"payment_method": "CASH",
		"invoice_number": invoice,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	resp, err := http.Post(
		apiServer.URL+"/checkout",
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		t.Fatalf("post checkout: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var checkoutResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&checkoutResp); err != nil {
		t.Fatalf("decode checkout response: %v", err)
	}

	if checkoutResp["status"] != "COMPLETED" {
		t.Fatalf("checkout status = %v, want COMPLETED", checkoutResp["status"])
	}

	printJob, ok := checkoutResp["print_job"].(map[string]any)
	if !ok {
		t.Fatalf("missing print_job")
	}

	if printJob["status"] != "FAILED" {
		t.Fatalf("print_job status = %v, want FAILED", printJob["status"])
	}

	if printJob["error"] == nil || printJob["error"] == "" {
		t.Fatalf("expected print_job error")
	}

	transactionID, ok := checkoutResp["transaction_id"].(string)
	if !ok || transactionID == "" {
		t.Fatalf("missing transaction_id")
	}

	stored, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}

	if stored.Status != domaintransaction.StatusCompleted {
		t.Fatalf("stored status = %q, want COMPLETED", stored.Status)
	}

	if stored.InvoiceNumber != invoice {
		t.Fatalf("invoice = %q, want %q", stored.InvoiceNumber, invoice)
	}

	updated, err := productRepo.GetByID(ctx, productID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}

	if updated.Stock != 7 {
		t.Fatalf("stock = %d, want 7", updated.Stock)
	}
}
