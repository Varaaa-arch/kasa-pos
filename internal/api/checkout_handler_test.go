package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"

	"pos-system/internal/db"
	"pos-system/internal/domain/product"
	"pos-system/internal/printer/agent"
	"pos-system/internal/repository/postgres"
	"pos-system/internal/service/checkout"
	receiptsvc "pos-system/internal/service/receipt"
)

type checkoutPrintPayload struct {
	Store struct {
		Name    string `json:"name"`
		Address string `json:"address"`
		Phone   string `json:"phone"`
	} `json:"store"`
	Transaction struct {
		ID            string `json:"id"`
		InvoiceNumber string `json:"invoice_number"`
	} `json:"transaction"`
	Items []struct {
		ProductID string `json:"product_id"`
		SKU       string `json:"sku"`
		Name      string `json:"name"`
		Quantity  int    `json:"quantity"`
		UnitPrice int64  `json:"unit_price"`
	} `json:"items"`
	Summary struct {
		Subtotal      int64 `json:"subtotal"`
		Discount      int64 `json:"discount"`
		Tax           int64 `json:"tax"`
		ServiceCharge int64 `json:"service_charge"`
		Total         int64 `json:"total"`
	} `json:"summary"`
	Payment struct {
		Method string `json:"method"`
		Paid   int64  `json:"paid"`
		Change int64  `json:"change"`
	} `json:"payment"`
	Footer struct {
		Message string `json:"message"`
	} `json:"footer"`
}

func newTestPrintAgentServer(t *testing.T) (*httptest.Server, *checkoutPrintPayload, *string) {
	t.Helper()

	var received checkoutPrintPayload
	var idempotencyKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/print" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		idempotencyKey = r.Header.Get("Idempotency-Key")

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode print request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agent.PrintResponse{
			JobID:   "PJ-test-handler",
			Message: "receipt printed successfully",
		})
	}))

	return server, &received, &idempotencyKey
}

func setupTestCheckoutHandler(
	t *testing.T,
	printClient agent.PrintAgentClient,
) (*CheckoutHandler, *postgres.ProductRepository, func()) {
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
	atomicSvc := checkout.NewAtomicService(database, transactionRepo, productRepo)
	orchestratorSvc := checkout.NewOrchestratorService(
		atomicSvc,
		receiptsvc.NewPrintService(),
		printClient,
		checkout.DefaultReceiptDefaults(),
	)
	handler := NewCheckoutHandler(orchestratorSvc, productRepo)

	cleanup := func() {
		database.Close()
	}

	return handler, productRepo, cleanup
}

func TestCheckoutHandlerInvalidBody(t *testing.T) {
	h := NewCheckoutHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBufferString(`{invalid-json`))
	rec := httptest.NewRecorder()

	h.Checkout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorCode(t, rec, ErrCodeInvalidBody)
}

func TestCheckoutHandlerEmptyItems(t *testing.T) {
	h := NewCheckoutHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBufferString(`{"items":[]}`))
	rec := httptest.NewRecorder()

	h.Checkout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorCode(t, rec, ErrCodeValidation)
}

func TestCheckoutHandlerSuccess(t *testing.T) {
	printServer, received, idempotencyKey := newTestPrintAgentServer(t)
	defer printServer.Close()

	handler, productRepo, cleanup := setupTestCheckoutHandler(
		t,
		agent.NewHTTPClient(printServer.URL),
	)
	defer cleanup()

	ctx := context.Background()
	productID := uuid.NewString()
	p := product.Product{
		ID:    productID,
		SKU:   "CHK-HANDLER-" + uuid.NewString()[:8],
		Name:  "Kopi Susu Handler",
		Price: 20000,
		Stock: 10,
	}

	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("create test product: %v", err)
	}
	defer func() {
		_ = productRepo.Delete(ctx, productID)
	}()

	body := `{
		"items": [
			{"product_id": "` + productID + `", "quantity": 2}
		],
		"paid_amount": 50000,
		"payment_method": "CASH"
	}`

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Checkout(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["transaction_id"] == nil || resp["transaction_id"] == "" {
		t.Errorf("expected non-empty transaction_id")
	}
	if resp["invoice_number"] == nil || resp["invoice_number"] == "" {
		t.Errorf("expected non-empty invoice_number")
	}
	if resp["total"] != float64(40000) {
		t.Errorf("expected total 40000, got %v", resp["total"])
	}
	if resp["paid_amount"] != float64(50000) {
		t.Errorf("expected paid_amount 50000, got %v", resp["paid_amount"])
	}
	if resp["change"] != float64(10000) {
		t.Errorf("expected change 10000, got %v", resp["change"])
	}
	if resp["status"] != "COMPLETED" {
		t.Errorf("expected status COMPLETED, got %v", resp["status"])
	}

	printJob, ok := resp["print_job"].(map[string]any)
	if !ok {
		t.Fatalf("expected print_job in response")
	}

	if printJob["status"] != "COMPLETED" {
		t.Errorf("expected print_job status COMPLETED, got %v", printJob["status"])
	}

	transactionID, ok := resp["transaction_id"].(string)
	if !ok || transactionID == "" {
		t.Fatalf("expected transaction_id string")
	}

	if *idempotencyKey != transactionID {
		t.Errorf("idempotency key = %q, want %q", *idempotencyKey, transactionID)
	}

	if received.Transaction.ID != transactionID {
		t.Errorf("print receipt transaction id = %q, want %q", received.Transaction.ID, transactionID)
	}

	if len(received.Items) != 1 {
		t.Fatalf("expected 1 print item, got %d", len(received.Items))
	}

	if received.Items[0].SKU != p.SKU {
		t.Errorf("print item sku = %q, want %q", received.Items[0].SKU, p.SKU)
	}

	if received.Summary.Total != 40000 {
		t.Errorf("print total = %d, want 40000", received.Summary.Total)
	}

	if received.Store.Name != "TOKO KASA" {
		t.Errorf("print store name = %q, want TOKO KASA", received.Store.Name)
	}
}

func TestCheckoutHandlerPrintFailureKeepsTransaction(t *testing.T) {
	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed to print receipt", http.StatusInternalServerError)
	}))
	defer failingServer.Close()

	handler, productRepo, cleanup := setupTestCheckoutHandler(
		t,
		agent.NewHTTPClient(failingServer.URL),
	)
	defer cleanup()

	ctx := context.Background()
	productID := uuid.NewString()
	p := product.Product{
		ID:    productID,
		SKU:   "CHK-PRINT-FAIL-" + uuid.NewString()[:8],
		Name:  "Kopi Print Fail",
		Price: 15000,
		Stock: 10,
	}

	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("create test product: %v", err)
	}
	defer func() {
		_ = productRepo.Delete(ctx, productID)
	}()

	body := `{
		"items": [
			{"product_id": "` + productID + `", "quantity": 1}
		],
		"paid_amount": 20000,
		"payment_method": "CASH"
	}`

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Checkout(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["status"] != "COMPLETED" {
		t.Errorf("expected status COMPLETED, got %v", resp["status"])
	}

	printJob, ok := resp["print_job"].(map[string]any)
	if !ok {
		t.Fatalf("expected print_job in response")
	}

	if printJob["status"] != "FAILED" {
		t.Errorf("expected print_job status FAILED, got %v", printJob["status"])
	}

	if printJob["error"] == nil || printJob["error"] == "" {
		t.Errorf("expected print_job error")
	}

	updated, err := productRepo.GetByID(ctx, productID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}

	if updated.Stock != 9 {
		t.Errorf("stock = %d, want 9", updated.Stock)
	}
}

func TestCheckoutHandlerProductNotFound(t *testing.T) {
	printServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer printServer.Close()

	handler, _, cleanup := setupTestCheckoutHandler(
		t,
		agent.NewHTTPClient(printServer.URL),
	)
	defer cleanup()

	body := `{
		"items": [
			{"product_id": "` + uuid.NewString() + `", "quantity": 1}
		],
		"paid_amount": 50000,
		"payment_method": "CASH"
	}`

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Checkout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorCode(t, rec, ErrCodeProductNotFound)
}

func TestCheckoutHandlerInsufficientPayment(t *testing.T) {
	printServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer printServer.Close()

	handler, productRepo, cleanup := setupTestCheckoutHandler(
		t,
		agent.NewHTTPClient(printServer.URL),
	)
	defer cleanup()

	ctx := context.Background()
	productID := uuid.NewString()
	p := product.Product{
		ID:    productID,
		SKU:   "CHK-INSUF-" + uuid.NewString()[:8],
		Name:  "Kopi Kurang Bayar",
		Price: 30000,
		Stock: 10,
	}

	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("create test product: %v", err)
	}
	defer func() {
		_ = productRepo.Delete(ctx, productID)
	}()

	body := `{
		"items": [
			{"product_id": "` + productID + `", "quantity": 2}
		],
		"paid_amount": 20000,
		"payment_method": "CASH"
	}`

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Checkout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorCode(t, rec, ErrCodeInsufficientPay)
}

func TestCheckoutHandlerInsufficientStock(t *testing.T) {
	printServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer printServer.Close()

	handler, productRepo, cleanup := setupTestCheckoutHandler(
		t,
		agent.NewHTTPClient(printServer.URL),
	)
	defer cleanup()

	ctx := context.Background()
	productID := uuid.NewString()
	p := product.Product{
		ID:    productID,
		SKU:   "CHK-STOCK-" + uuid.NewString()[:8],
		Name:  "Kopi Kurang Stok",
		Price: 15000,
		Stock: 2,
	}

	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("create test product: %v", err)
	}
	defer func() {
		_ = productRepo.Delete(ctx, productID)
	}()

	body := `{
		"items": [
			{"product_id": "` + productID + `", "quantity": 5}
		],
		"paid_amount": 100000,
		"payment_method": "CASH"
	}`

	req := httptest.NewRequest(http.MethodPost, "/checkout", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Checkout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorCode(t, rec, ErrCodeInsufficientStk)
}
