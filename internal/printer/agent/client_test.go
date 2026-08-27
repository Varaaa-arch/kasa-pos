package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

func sampleReceipt() domainreceipt.Receipt {
	return domainreceipt.Receipt{
		Store: domainreceipt.Store{
			Name:    "TOKO KASA",
			Address: "Jl. Contoh No. 123",
			Phone:   "081234567890",
		},
		Transaction: domainreceipt.Transaction{
			ID:            "txn-001",
			InvoiceNumber: "INV-001",
			Timestamp:     time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC),
			Cashier:       "Kasir",
		},
		Items: []domainreceipt.Item{
			{
				ProductID: "prod-001",
				SKU:       "KOPI-001",
				Name:      "Kopi Susu",
				Quantity:  2,
				UnitPrice: domainreceipt.NewMoney(15000, domainreceipt.IDR),
			},
		},
		Summary: domainreceipt.Summary{
			Subtotal: domainreceipt.NewMoney(30000, domainreceipt.IDR),
			Total:    domainreceipt.NewMoney(30000, domainreceipt.IDR),
		},
		Payment: domainreceipt.Payment{
			Method: "CASH",
			Paid:   domainreceipt.NewMoney(50000, domainreceipt.IDR),
			Change: domainreceipt.NewMoney(20000, domainreceipt.IDR),
		},
		Footer: domainreceipt.Footer{
			Message: "Terima kasih",
		},
	}
}

func TestHTTPClientPrintSuccess(t *testing.T) {
	var received printRequest
	var idempotencyKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/print" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		idempotencyKey = r.Header.Get("Idempotency-Key")

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PrintResponse{
			JobID:   "PJ-test-001",
			Message: "receipt printed successfully",
		})
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	resp, err := client.Print(
		context.Background(),
		sampleReceipt(),
		"txn-001",
	)
	if err != nil {
		t.Fatalf("print failed: %v", err)
	}

	if resp.JobID != "PJ-test-001" {
		t.Fatalf("unexpected job id: %q", resp.JobID)
	}

	if idempotencyKey != "txn-001" {
		t.Fatalf("unexpected idempotency key: %q", idempotencyKey)
	}

	if received.Transaction.ID != "txn-001" {
		t.Fatalf("unexpected transaction id: %q", received.Transaction.ID)
	}

	if len(received.Items) != 1 || received.Items[0].SKU != "KOPI-001" {
		t.Fatalf("unexpected items: %+v", received.Items)
	}

	if received.Summary.Total != 30000 {
		t.Fatalf("unexpected total: %d", received.Summary.Total)
	}
}

func TestHTTPClientPrintFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed to print receipt", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	_, err := client.Print(
		context.Background(),
		sampleReceipt(),
		"txn-002",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "print agent returned 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClientPrintTimeoutResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "PRINT_TIMEOUT", http.StatusGatewayTimeout)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	_, err := client.Print(
		context.Background(),
		sampleReceipt(),
		"txn-timeout-001",
	)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !strings.Contains(err.Error(), "PRINT_TIMEOUT") {
		t.Fatalf("expected error containing PRINT_TIMEOUT, got: %v", err)
	}
}

func TestHTTPClientContextDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := client.Print(
		ctx,
		sampleReceipt(),
		"txn-timeout-002",
	)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !strings.Contains(err.Error(), "PRINT_TIMEOUT") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}
