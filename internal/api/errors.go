package api

import (
	"encoding/json"
	"net/http"

	applogger "pos-system/internal/logger"
)

// RequestIDFromContext is a convenience re-export so existing handler/test
// code in the api package doesn't need a new import.
// The canonical implementation lives in internal/logger.
var RequestIDFromContext = applogger.RequestIDFromContext

// ─── Error Codes ─────────────────────────────────────────────────────────────

// ErrorCode is a machine-readable string that identifies the specific error.
type ErrorCode string

const (
	// 400 — client request errors
	ErrCodeInvalidBody     ErrorCode = "INVALID_REQUEST_BODY"
	ErrCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrCodeEmptyCart       ErrorCode = "CART_EMPTY"
	ErrCodeInsufficientPay ErrorCode = "INSUFFICIENT_PAYMENT"
	ErrCodeInsufficientStk ErrorCode = "INSUFFICIENT_STOCK"

	// 404 — not found
	ErrCodeProductNotFound     ErrorCode = "PRODUCT_NOT_FOUND"
	ErrCodeTransactionNotFound ErrorCode = "TRANSACTION_NOT_FOUND"

	// 500 — server errors
	ErrCodeInternal ErrorCode = "INTERNAL_SERVER_ERROR"
)

// ─── Error Response Model ─────────────────────────────────────────────────────

// errorDetail is the inner object within the JSON envelope.
type errorDetail struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id,omitempty"`
}

// errorResponse is the top-level JSON envelope:
//
//	{"error": {"code": "...", "message": "...", "request_id": "..."}}
type errorResponse struct {
	Error errorDetail `json:"error"`
}

// ─── WriteError ───────────────────────────────────────────────────────────────

// WriteError writes a structured JSON error response.
// It reads the request ID from ctx (injected by RequestID middleware) and
// includes it in the response so clients can correlate errors with logs.
//
// Example output:
//
//	HTTP 404
//	{"error":{"code":"PRODUCT_NOT_FOUND","message":"Product not found","request_id":"req_abc123"}}
func WriteError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code ErrorCode,
	message string,
) {
	resp := errorResponse{
		Error: errorDetail{
			Code:      code,
			Message:   message,
			RequestID: RequestIDFromContext(r.Context()),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Encode errors are intentionally swallowed; the header is already sent.
	_ = json.NewEncoder(w).Encode(resp)
}
