package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos-system/internal/printer/receipt"
)

type mockPrinter struct {
	openErr  error
	writeErr error
	closeErr error

	opened bool
	closed bool
	data   []byte
}

func (m *mockPrinter) Open() error {
	if m.openErr != nil {
		return m.openErr
	}

	m.opened = true

	return nil
}

func (m *mockPrinter) Write(data []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}

	m.data = append(m.data, data...)

	return len(data), nil
}

func (m *mockPrinter) Close() error {
	m.closed = true

	return m.closeErr
}

func TestPrintHandler(t *testing.T) {
	printer := &mockPrinter{}

	renderer := receipt.NewRenderer()

	handler := NewHandler(
		printer,
		renderer,
	)

	body := `{
		"store": {
			"name": "TOKO KASA",
			"address": "Jl. Contoh No. 123",
			"phone": "081234567890"
		},
		"transaction": {
			"invoice_number": "INV-HTTP-001",
			"cashier": "Bizar"
		},
		"items": [
			{
				"name": "Kopi Susu",
				"sku": "KOPI-001",
				"quantity": 2,
				"unit_price": 15000
			}
		],
		"payment": {
			"method": "CASH",
			"paid": 50000
		},
		"footer": {
			"message": "PRINT AGENT OK"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/print",
		strings.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	handler.Print(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			rec.Code,
			rec.Body.String(),
		)
	}

	if !printer.opened {
		t.Fatal("expected printer to be opened")
	}

	if !printer.closed {
		t.Fatal("expected printer to be closed")
	}

	if len(printer.data) == 0 {
		t.Fatal("expected printer to receive data")
	}

	if !strings.Contains(
		rec.Body.String(),
		"receipt printed successfully",
	) {
		t.Fatalf(
			"unexpected response: %s",
			rec.Body.String(),
		)
	}
}

func TestPrintHandlerInvalidJSON(t *testing.T) {
	printer := &mockPrinter{}
	renderer := receipt.NewRenderer()

	handler := NewHandler(
		printer,
		renderer,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/print",
		strings.NewReader(`{"invalid"`),
	)

	rec := httptest.NewRecorder()

	handler.Print(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			rec.Code,
		)
	}

	if printer.opened {
		t.Fatal("printer should not be opened for invalid JSON")
	}
}

func TestPrintHandlerMethodNotAllowed(t *testing.T) {
	printer := &mockPrinter{}
	renderer := receipt.NewRenderer()

	handler := NewHandler(
		printer,
		renderer,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/print",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Print(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status 405, got %d",
			rec.Code,
		)
	}
}

func TestPrintHandlerOpenError(t *testing.T) {
	expectedErr := errors.New("printer unavailable")

	printer := &mockPrinter{
		openErr: expectedErr,
	}

	renderer := receipt.NewRenderer()

	handler := NewHandler(
		printer,
		renderer,
	)

	body := `{
		"store": {
			"name": "TOKO KASA"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/print",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Print(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			rec.Code,
		)
	}

	if !printer.opened {
		// Expected because Open failed.
		return
	}

	t.Fatal("printer should not be marked as opened")
}
