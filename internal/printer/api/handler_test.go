package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos-system/internal/printer/logging"
	"pos-system/internal/printer/mock"
	"pos-system/internal/printer/receipt"
)

func newTestHandler(printer *mock.Printer) *Handler {
	return NewHandler(
		printer,
		receipt.NewRenderer(),
		"/dev/usb/lp0",
		logging.New(),
	)
}

func validPrintJSON() string {
    return `{
        "store": {
            "name": "TOKO KASA",
            "address": "Jl. Contoh No. 123",
            "phone": "081234567890"
        },
        "transaction": {
            "id": "TXN-HTTP-001",
            "invoice_number": "INV-HTTP-001",
            "timestamp": "2026-08-14T00:00:00Z",
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
        "summary": {
            "subtotal": 30000,
            "total": 30000
        },
        "payment": {
            "method": "CASH",
            "paid": 50000,
            "change": 20000
        },
        "footer": {
            "message": "PRINT AGENT OK"
        }
    }`
}

func TestPrintHandler(t *testing.T) {
	printer := &mock.Printer{}
	handler := newTestHandler(printer)

	req := httptest.NewRequest(
		http.MethodPost,
		"/print",
		strings.NewReader(validPrintJSON()),
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

	if printer.OpenCount != 1 {
		t.Fatalf(
			"expected 1 open, got %d",
			printer.OpenCount,
		)
	}

	if printer.WriteCount != 1 {
		t.Fatalf(
			"expected 1 write, got %d",
			printer.WriteCount,
		)
	}

	if printer.CloseCount != 1 {
		t.Fatalf(
			"expected 1 close, got %d",
			printer.CloseCount,
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"job_id"`,
	) {
		t.Fatalf(
			"response missing job_id: %s",
			rec.Body.String(),
		)
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
	printer := &mock.Printer{}
	handler := newTestHandler(printer)

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

	if printer.OpenCount != 0 {
		t.Fatal(
			"printer should not be opened for invalid JSON",
		)
	}
}

func TestPrintHandlerMethodNotAllowed(t *testing.T) {
	printer := &mock.Printer{}
	handler := newTestHandler(printer)

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
	printer := &mock.Printer{
		OpenErr: mock.ErrOpen,
	}

	handler := newTestHandler(printer)

	body := `{
		"store": {
			"name": "TOKO KASA"
		},
		"transaction": {
			"id": "TXN-OPEN-ERROR-001",
			"invoice_number": "INV-OPEN-ERROR-001",
			"timestamp": "2026-08-14T00:00:00Z",
			"cashier": "Bizar"
		},
		"items": [
			{
				"name": "Kopi Susu",
				"sku": "KOPI-001",
				"quantity": 1,
				"unit_price": 15000
			}
		],
		"summary": {
			"subtotal": 15000,
			"total": 15000
		},
		"payment": {
			"method": "CASH",
			"paid": 15000,
			"change": 0
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

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d: %s",
			rec.Code,
			rec.Body.String(),
		)
	}

	if printer.WriteCount != 0 {
		t.Fatalf(
			"expected no write attempts, got %d",
			printer.WriteCount,
		)
	}
}

func TestPrintHandlerValidationError(t *testing.T) {
	printer := &mock.Printer{}
	handler := newTestHandler(printer)

	body := `{
		"store": {
			"name": "TOKO KASA"
		},
		"transaction": {
			"id": "TXN-INVALID-001",
			"invoice_number": "",
			"timestamp": "2026-08-14T00:00:00Z",
			"cashier": "Bizar"
		},
		"items": [
			{
				"name": "Kopi Susu",
				"quantity": 1,
				"unit_price": 15000
			}
		],
		"summary": {
			"subtotal": 15000,
			"total": 15000
		},
		"payment": {
			"method": "CASH",
			"paid": 15000,
			"change": 0
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

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d: %s",
			rec.Code,
			rec.Body.String(),
		)
	}

	if printer.OpenCount != 0 {
		t.Fatalf(
			"printer should not be opened for invalid receipt, got %d attempts",
			printer.OpenCount,
		)
	}

	if printer.WriteCount != 0 {
		t.Fatalf(
			"printer should not be written to, got %d attempts",
			printer.WriteCount,
		)
	}
}

func TestStatusHandler(t *testing.T) {
	printer := &mock.Printer{}
	handler := newTestHandler(printer)

	req := httptest.NewRequest(
		http.MethodGet,
		"/status",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Status(rec, req)

	if rec.Code != http.StatusOK &&
		rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"unexpected status code: %d",
			rec.Code,
		)
	}

	body := rec.Body.String()

	if !strings.Contains(body, `"printer"`) {
		t.Fatalf(
			"response missing printer field: %s",
			body,
		)
	}

	if !strings.Contains(body, `"device"`) {
		t.Fatalf(
			"response missing device field: %s",
			body,
		)
	}

	if !strings.Contains(body, `"connected"`) {
		t.Fatalf(
			"response missing connected field: %s",
			body,
		)
	}
}

func TestGenerateJobID(t *testing.T) {
	id, err := generateJobID()

	if err != nil {
		t.Fatalf(
			"generateJobID() returned error: %v",
			err,
		)
	}

	if len(id) != 19 {
		t.Fatalf(
			"unexpected job ID length: %d (%q)",
			len(id),
			id,
		)
	}

	if id[:3] != "PJ-" {
		t.Fatalf(
			"unexpected job ID prefix: %q",
			id,
		)
	}
}

func TestGenerateJobIDsAreDifferent(t *testing.T) {
	first, err := generateJobID()

	if err != nil {
		t.Fatal(err)
	}

	second, err := generateJobID()

	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Fatalf(
			"generated duplicate job IDs: %q",
			first,
		)
	}
}
