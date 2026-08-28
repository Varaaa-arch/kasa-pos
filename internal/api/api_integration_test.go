package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"

	"pos-system/internal/domain/product"
	domaintransaction "pos-system/internal/domain/transaction"
	"pos-system/internal/printer/agent"
)

type recordedPrintCall struct {
	path           string
	idempotencyKey string
	payload        checkoutPrintPayload
}

func newRecordingPrintAgent(t *testing.T) (*httptest.Server, *[]recordedPrintCall) {
	t.Helper()

	var mu sync.Mutex
	calls := make([]recordedPrintCall, 0, 2)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var payload checkoutPrintPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		mu.Lock()
		calls = append(calls, recordedPrintCall{
			path:           r.URL.Path,
			idempotencyKey: r.Header.Get("Idempotency-Key"),
			payload:        payload,
		})
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(agent.PrintResponse{
			JobID:   "PJ-api-int-" + uuid.NewString()[:8],
			Message: "receipt printed successfully",
		})
	}))

	return server, &calls
}

func TestAPIIntegrationCheckoutTransactionsReprint(t *testing.T) {
	printAgent, calls := newRecordingPrintAgent(t)
	defer printAgent.Close()

	apiServer, database, productRepo, transactionRepo, cleanup := setupCheckoutIntegrationRouter(
		t,
		printAgent.URL,
	)
	defer cleanup()

	ctx := context.Background()
	productID := uuid.NewString()
	invoice := "API-INT-" + uuid.NewString()

	p := product.Product{
		ID:    productID,
		SKU:   "API-INT-" + uuid.NewString()[:8],
		Name:  "Kopi API Integration",
		Price: 22000,
		Stock: 9,
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

	checkoutBody, err := json.Marshal(map[string]any{
		"items": []map[string]any{
			{"product_id": productID, "quantity": 2},
		},
		"paid_amount":    50000,
		"payment_method": "CASH",
		"invoice_number": invoice,
	})
	if err != nil {
		t.Fatalf("marshal checkout: %v", err)
	}

	checkoutResp, err := http.Post(
		apiServer.URL+"/checkout",
		"application/json",
		bytes.NewReader(checkoutBody),
	)
	if err != nil {
		t.Fatalf("POST /checkout: %v", err)
	}
	defer checkoutResp.Body.Close()

	if checkoutResp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /checkout status = %d, want 201 body=%s", checkoutResp.StatusCode, readBody(t, checkoutResp))
	}

	var checkoutJSON map[string]any
	if err := json.NewDecoder(checkoutResp.Body).Decode(&checkoutJSON); err != nil {
		t.Fatalf("decode checkout: %v", err)
	}

	transactionID, _ := checkoutJSON["transaction_id"].(string)
	if transactionID == "" {
		t.Fatal("POST /checkout missing transaction_id")
	}

	if checkoutJSON["status"] != "COMPLETED" {
		t.Fatalf("checkout status = %v, want COMPLETED", checkoutJSON["status"])
	}

	printJob, _ := checkoutJSON["print_job"].(map[string]any)
	if printJob["status"] != "COMPLETED" {
		t.Fatalf("checkout print_job.status = %v, want COMPLETED", printJob["status"])
	}

	storedAfterCheckout, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("repository GetByID after checkout: %v", err)
	}

	if storedAfterCheckout.InvoiceNumber != invoice {
		t.Fatalf("stored invoice = %q, want %q", storedAfterCheckout.InvoiceNumber, invoice)
	}

	if storedAfterCheckout.Status != domaintransaction.StatusCompleted {
		t.Fatalf("stored status = %q, want COMPLETED", storedAfterCheckout.Status)
	}

	if storedAfterCheckout.Total != 44000 {
		t.Fatalf("stored total = %d, want 44000", storedAfterCheckout.Total)
	}

	if storedAfterCheckout.Change != 6000 {
		t.Fatalf("stored change = %d, want 6000", storedAfterCheckout.Change)
	}

	if len(storedAfterCheckout.Items) != 1 || storedAfterCheckout.Items[0].Quantity != 2 {
		t.Fatalf("stored items = %+v, want 1 item qty 2", storedAfterCheckout.Items)
	}

	updatedProduct, err := productRepo.GetByID(ctx, productID)
	if err != nil {
		t.Fatalf("get product after checkout: %v", err)
	}

	if updatedProduct.Stock != 7 {
		t.Fatalf("stock after checkout = %d, want 7", updatedProduct.Stock)
	}

	listResp, err := http.Get(apiServer.URL + "/transactions")
	if err != nil {
		t.Fatalf("GET /transactions: %v", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("GET /transactions status = %d, want 200 body=%s", listResp.StatusCode, readBody(t, listResp))
	}

	var listed []domaintransaction.Transaction
	if err := json.NewDecoder(listResp.Body).Decode(&listed); err != nil {
		t.Fatalf("decode GET /transactions: %v", err)
	}

	found := false
	for _, tx := range listed {
		if tx.ID == transactionID && tx.InvoiceNumber == invoice {
			found = true
			if tx.Total != 44000 {
				t.Fatalf("listed total = %d, want 44000", tx.Total)
			}
			break
		}
	}

	if !found {
		t.Fatalf("GET /transactions did not include checkout transaction %s", transactionID)
	}

	getResp, err := http.Get(apiServer.URL + "/transactions/" + transactionID)
	if err != nil {
		t.Fatalf("GET /transactions/:id: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET /transactions/:id status = %d, want 200 body=%s", getResp.StatusCode, readBody(t, getResp))
	}

	var fetched domaintransaction.Transaction
	if err := json.NewDecoder(getResp.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode GET /transactions/:id: %v", err)
	}

	if fetched.ID != transactionID {
		t.Fatalf("fetched id = %q, want %q", fetched.ID, transactionID)
	}

	if fetched.InvoiceNumber != invoice {
		t.Fatalf("fetched invoice = %q, want %q", fetched.InvoiceNumber, invoice)
	}

	if fetched.Status != domaintransaction.StatusCompleted {
		t.Fatalf("fetched status = %q, want COMPLETED", fetched.Status)
	}

	if fetched.PaidAmount != 50000 {
		t.Fatalf("fetched paid = %d, want 50000", fetched.PaidAmount)
	}

	if len(fetched.Items) != 1 {
		t.Fatalf("fetched items = %d, want 1", len(fetched.Items))
	}

	if fetched.Items[0].ProductID != productID {
		t.Fatalf("fetched product id = %q, want %q", fetched.Items[0].ProductID, productID)
	}

	if fetched.Items[0].SKU != p.SKU {
		t.Fatalf("fetched sku = %q, want %q", fetched.Items[0].SKU, p.SKU)
	}

	reprintResp, err := http.Post(
		apiServer.URL+"/transactions/"+transactionID+"/reprint",
		"application/json",
		http.NoBody,
	)
	if err != nil {
		t.Fatalf("POST /transactions/:id/reprint: %v", err)
	}
	defer reprintResp.Body.Close()

	if reprintResp.StatusCode != http.StatusOK {
		t.Fatalf("POST reprint status = %d, want 200 body=%s", reprintResp.StatusCode, readBody(t, reprintResp))
	}

	var reprintJSON map[string]any
	if err := json.NewDecoder(reprintResp.Body).Decode(&reprintJSON); err != nil {
		t.Fatalf("decode reprint: %v", err)
	}

	if reprintJSON["transaction_id"] != transactionID {
		t.Fatalf("reprint transaction_id = %v, want %s", reprintJSON["transaction_id"], transactionID)
	}

	reprintJob, _ := reprintJSON["print_job"].(map[string]any)
	if reprintJob["status"] != "COMPLETED" {
		t.Fatalf("reprint print_job.status = %v, want COMPLETED", reprintJob["status"])
	}

	if len(*calls) != 2 {
		t.Fatalf("print agent calls = %d, want 2 (checkout + reprint)", len(*calls))
	}

	checkoutCall := (*calls)[0]
	reprintCall := (*calls)[1]

	if checkoutCall.path != "/print" || reprintCall.path != "/print" {
		t.Fatalf("print agent paths = %q, %q, want /print", checkoutCall.path, reprintCall.path)
	}

	if checkoutCall.idempotencyKey != transactionID {
		t.Fatalf("checkout idempotency key = %q, want transaction id %q", checkoutCall.idempotencyKey, transactionID)
	}

	if reprintCall.idempotencyKey == "" {
		t.Fatal("reprint idempotency key is empty")
	}

	if reprintCall.idempotencyKey == transactionID {
		t.Fatal("reprint reused checkout idempotency key")
	}

	if !strings.HasPrefix(reprintCall.idempotencyKey, "reprint-") {
		t.Fatalf("reprint idempotency key = %q, want reprint- prefix", reprintCall.idempotencyKey)
	}

	if reprintCall.payload.Transaction.ID != transactionID {
		t.Fatalf("reprint receipt transaction id = %q, want %q", reprintCall.payload.Transaction.ID, transactionID)
	}

	if reprintCall.payload.Transaction.InvoiceNumber != invoice {
		t.Fatalf("reprint invoice = %q, want %q", reprintCall.payload.Transaction.InvoiceNumber, invoice)
	}

	if reprintCall.payload.Summary.Total != 44000 {
		t.Fatalf("reprint total = %d, want 44000", reprintCall.payload.Summary.Total)
	}

	stockAfterReprint, err := productRepo.GetByID(ctx, productID)
	if err != nil {
		t.Fatalf("get product after reprint: %v", err)
	}

	if stockAfterReprint.Stock != 7 {
		t.Fatalf("stock after reprint = %d, want 7 (reprint must not change stock)", stockAfterReprint.Stock)
	}

	storedAfterReprint, err := transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		t.Fatalf("repository GetByID after reprint: %v", err)
	}

	if storedAfterReprint.Status != domaintransaction.StatusCompleted {
		t.Fatalf("status after reprint = %q, want COMPLETED", storedAfterReprint.Status)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
