package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ─── Helper ───────────────────────────────────────────────────────────────────

// decodeErrorResponse decodes the structured error envelope from a response
// body. Tests use this instead of string matching.
func decodeErrorResponse(t *testing.T, body string) errorResponse {
	t.Helper()

	var resp errorResponse
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&resp); err != nil {
		t.Fatalf("decodeErrorResponse: failed to parse JSON: %v\nbody: %s", err, body)
	}

	return resp
}

// ─── RequestIDFromContext ─────────────────────────────────────────────────────

func TestRequestIDFromContext_Missing(t *testing.T) {
	id := RequestIDFromContext(context.Background())
	if id != "" {
		t.Fatalf("expected empty string, got %q", id)
	}
}

func TestRequestIDFromContext_Set(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIDKey, "req_abc123")
	id := RequestIDFromContext(ctx)
	if id != "req_abc123" {
		t.Fatalf("expected req_abc123, got %q", id)
	}
}

// ─── RequestID middleware ─────────────────────────────────────────────────────

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	var capturedID string

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	RequestID(inner).ServeHTTP(rec, req)

	if capturedID == "" {
		t.Fatal("expected a generated request ID in context")
	}

	if rec.Header().Get("X-Request-ID") != capturedID {
		t.Fatalf("X-Request-ID header %q does not match context ID %q",
			rec.Header().Get("X-Request-ID"), capturedID)
	}
}

func TestRequestIDMiddleware_PassthroughClientID(t *testing.T) {
	var capturedID string

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "client-supplied-id")
	rec := httptest.NewRecorder()

	RequestID(inner).ServeHTTP(rec, req)

	if capturedID != "client-supplied-id" {
		t.Fatalf("expected client-supplied-id, got %q", capturedID)
	}

	if rec.Header().Get("X-Request-ID") != "client-supplied-id" {
		t.Fatal("X-Request-ID header should echo the client-provided value")
	}
}

func TestRequestIDMiddleware_UniquePerRequest(t *testing.T) {
	ids := make(map[string]bool)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids[RequestIDFromContext(r.Context())] = true
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		RequestID(inner).ServeHTTP(rec, req)
	}

	if len(ids) < 9 {
		// Allow at most 1 collision in 10 runs (extremely unlikely).
		t.Fatalf("expected ~10 unique IDs, got %d", len(ids))
	}
}

// ─── WriteError ───────────────────────────────────────────────────────────────

func TestWriteError_Structure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/products/nope", nil)
	rec := httptest.NewRecorder()

	WriteError(rec, req, http.StatusNotFound, ErrCodeProductNotFound, "Product not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	resp := decodeErrorResponse(t, rec.Body.String())

	if resp.Error.Code != ErrCodeProductNotFound {
		t.Fatalf("expected code %q, got %q", ErrCodeProductNotFound, resp.Error.Code)
	}

	if resp.Error.Message != "Product not found" {
		t.Fatalf("expected message %q, got %q", "Product not found", resp.Error.Message)
	}
}

func TestWriteError_IncludesRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), requestIDKey, "req_test999")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	WriteError(rec, req, http.StatusBadRequest, ErrCodeInvalidBody, "Invalid request body")

	resp := decodeErrorResponse(t, rec.Body.String())

	if resp.Error.RequestID != "req_test999" {
		t.Fatalf("expected request_id req_test999, got %q", resp.Error.RequestID)
	}
}

func TestWriteError_OmitsRequestIDWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	WriteError(rec, req, http.StatusInternalServerError, ErrCodeInternal, "Internal server error")

	body := rec.Body.String()
	resp := decodeErrorResponse(t, body)

	// request_id must be omitted (not present as empty string) when not set.
	if resp.Error.RequestID != "" {
		t.Fatalf("expected no request_id, got %q", resp.Error.RequestID)
	}

	if strings.Contains(body, "request_id") {
		t.Fatal("request_id field should be omitted from JSON when not set")
	}
}

func TestWriteError_AllErrorCodes(t *testing.T) {
	cases := []struct {
		code   ErrorCode
		status int
		msg    string
	}{
		{ErrCodeInvalidBody, http.StatusBadRequest, "Invalid request body"},
		{ErrCodeValidation, http.StatusBadRequest, "Validation error"},
		{ErrCodeEmptyCart, http.StatusBadRequest, "Cart is empty"},
		{ErrCodeInsufficientPay, http.StatusBadRequest, "Payment is insufficient"},
		{ErrCodeInsufficientStk, http.StatusBadRequest, "Insufficient stock"},
		{ErrCodeProductNotFound, http.StatusNotFound, "Product not found"},
		{ErrCodeTransactionNotFound, http.StatusNotFound, "Transaction not found"},
		{ErrCodeInternal, http.StatusInternalServerError, "Internal server error"},
	}

	for _, tc := range cases {
		t.Run(string(tc.code), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			WriteError(rec, req, tc.status, tc.code, tc.msg)

			if rec.Code != tc.status {
				t.Fatalf("expected status %d, got %d", tc.status, rec.Code)
			}

			resp := decodeErrorResponse(t, rec.Body.String())

			if resp.Error.Code != tc.code {
				t.Fatalf("expected code %q, got %q", tc.code, resp.Error.Code)
			}

			if resp.Error.Message != tc.msg {
				t.Fatalf("expected message %q, got %q", tc.msg, resp.Error.Message)
			}
		})
	}
}

// ─── newRequestID format ─────────────────────────────────────────────────────

func TestNewRequestID_Format(t *testing.T) {
	id := newRequestID()

	if !strings.HasPrefix(id, "req_") {
		t.Fatalf("expected prefix req_, got %q", id)
	}

	// req_ (4) + 8 hex chars = 12
	if len(id) != 12 {
		t.Fatalf("expected length 12, got %d (%q)", len(id), id)
	}
}
