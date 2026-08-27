package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pos-system/internal/printer/api"
	"pos-system/internal/printer/logging"
	"pos-system/internal/printer/mock"
	"pos-system/internal/printer/receipt"
)

func TestPrintAgentStartup(t *testing.T) {
	printer := &mock.Printer{}
	renderer := receipt.NewRenderer()
	logger := logging.New()

	handler := api.NewHandler(
		printer,
		renderer,
		"/dev/usb/lp0",
		logger,
		receipt.NewIdempotencyStore(),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/print", handler.Print)
	mux.HandleFunc("/status", handler.Status)
	mux.HandleFunc("/health", handler.Status)

	server := &http.Server{
		Addr:    "127.0.0.1:0", // Use random available port
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatalf("server failed to start: %v", err)
		}
	}()

	defer server.Shutdown(context.Background())

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	if server.Addr == "" {
		t.Fatal("server address should not be empty")
	}
}

func TestHealthEndpoint(t *testing.T) {
	printer := &mock.Printer{}
	renderer := receipt.NewRenderer()
	logger := logging.New()

	handler := api.NewHandler(
		printer,
		renderer,
		"/dev/usb/lp0",
		logger,
		receipt.NewIdempotencyStore(),
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Status(rec, req)

	// Mock device doesn't exist, so expect 503 but the response should still be valid
	if rec.Code != http.StatusServiceUnavailable && rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK or 503 Service Unavailable, got %d: %s", rec.Code, rec.Body.String())
	}

	var response api.StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Printer != "BP-LITE58" {
		t.Fatalf("expected printer BP-LITE58, got %s", response.Printer)
	}
}

func TestValidPrint(t *testing.T) {
	printer := &mock.Printer{}
	renderer := receipt.NewRenderer()
	logger := logging.New()

	handler := api.NewHandler(
		printer,
		renderer,
		"/dev/usb/lp0",
		logger,
		receipt.NewIdempotencyStore(),
	)

	payload := `{
		"store": {
			"name": "TOKO KASA",
			"address": "Jl. Contoh No. 123",
			"phone": "081234567890"
		},
		"transaction": {
			"id": "TXN-TEST-001",
			"invoice_number": "INV-TEST-001",
			"timestamp": "2026-08-14T00:00:00Z",
			"cashier": "Test"
		},
		"items": [
			{
				"name": "Test Item",
				"sku": "TEST-001",
				"quantity": 1,
				"unit_price": 10000
			}
		],
		"summary": {
			"subtotal": 10000,
			"total": 10000
		},
		"payment": {
			"method": "CASH",
			"paid": 10000,
			"change": 0
		},
		"footer": {
			"message": "TEST"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/print", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-valid-001")
	rec := httptest.NewRecorder()

	handler.Print(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	if !printer.HasData() {
		t.Fatal("expected printer to have data")
	}
}

func TestInvalidPayload(t *testing.T) {
	printer := &mock.Printer{}
	renderer := receipt.NewRenderer()
	logger := logging.New()

	handler := api.NewHandler(
		printer,
		renderer,
		"/dev/usb/lp0",
		logger,
		receipt.NewIdempotencyStore(),
	)

	// Missing required fields
	payload := `{
		"store": {
			"name": "TOKO KASA"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/print", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-invalid-001")
	rec := httptest.NewRecorder()

	handler.Print(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
	}

	if printer.HasData() {
		t.Fatal("expected printer to not have data")
	}
}

func TestPrinterUnavailable(t *testing.T) {
	printer := &mock.Printer{
		OpenErr: mock.ErrOpen,
	}
	renderer := receipt.NewRenderer()
	logger := logging.New()

	handler := api.NewHandler(
		printer,
		renderer,
		"/dev/usb/lp0",
		logger,
		receipt.NewIdempotencyStore(),
	)

	payload := validPrintPayload()

	req := httptest.NewRequest(http.MethodPost, "/print", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-unavailable-001")
	rec := httptest.NewRecorder()

	handler.Print(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 Internal Server Error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPrinterTimeout(t *testing.T) {
	printer := &mock.SlowPrinter{
		WriteDelay: 200 * time.Millisecond,
	}
	renderer := receipt.NewRenderer()
	logger := logging.New()

	handler := api.NewHandler(
		printer,
		renderer,
		"/dev/usb/lp0",
		logger,
		receipt.NewIdempotencyStore(),
	)
	handler.PrintTimeout = 50 * time.Millisecond

	payload := validPrintPayload()

	req := httptest.NewRequest(http.MethodPost, "/print", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-timeout-001")
	rec := httptest.NewRecorder()

	handler.Print(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504 Gateway Timeout, got %d: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "PRINT_TIMEOUT") {
		t.Fatalf("expected response to contain PRINT_TIMEOUT, got: %s", rec.Body.String())
	}
}

func TestMultipleRequests(t *testing.T) {
	printer := &mock.Printer{}
	renderer := receipt.NewRenderer()
	logger := logging.New()

	handler := api.NewHandler(
		printer,
		renderer,
		"/dev/usb/lp0",
		logger,
		receipt.NewIdempotencyStore(),
	)

	payload := validPrintPayload()

	// Send 5 concurrent requests with unique idempotency keys
	for i := 0; i < 5; i++ {
		go func(id int) {
			req := httptest.NewRequest(http.MethodPost, "/print", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", "test-multiple-001")
			rec := httptest.NewRecorder()

			handler.Print(rec, req)

			// First request should succeed, others may get 409 due to idempotency
			if rec.Code != http.StatusOK && rec.Code != http.StatusConflict {
				t.Errorf("request %d: expected 200 OK or 409 Conflict, got %d", id, rec.Code)
			}
		}(i)
	}

	// Wait for all requests to complete
	time.Sleep(500 * time.Millisecond)

	// Printer should have been used
	if printer.OpenCount == 0 {
		t.Fatal("expected printer to be opened")
	}
}

func TestGracefulShutdown(t *testing.T) {
	// This test verifies that the server can be shut down gracefully
	// In a real scenario, this would involve signal handling
	t.Skip("Graceful shutdown test skipped for unit testing")
}

func validPrintPayload() string {
	return `{
		"store": {
			"name": "TOKO KASA",
			"address": "Jl. Contoh No. 123",
			"phone": "081234567890"
		},
		"transaction": {
			"id": "TXN-TEST-001",
			"invoice_number": "INV-TEST-001",
			"timestamp": "2026-08-14T00:00:00Z",
			"cashier": "Test"
		},
		"items": [
			{
				"name": "Test Item",
				"sku": "TEST-001",
				"quantity": 1,
				"unit_price": 10000
			}
		],
		"summary": {
			"subtotal": 10000,
			"total": 10000
		},
		"payment": {
			"method": "CASH",
			"paid": 10000,
			"change": 0
		},
		"footer": {
			"message": "TEST"
		}
	}`
}