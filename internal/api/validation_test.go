package api

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

// ─── validateCheckoutRequest ─────────────────────────────────────────────────

func TestValidateCheckoutRequest_Valid(t *testing.T) {
	req := CheckoutRequest{
		Items: []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{
			{ProductID: "prod-001", Quantity: 2},
		},
		PaidAmount: 50000,
	}

	if err := validateCheckoutRequest(req); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateCheckoutRequest_EmptyItems(t *testing.T) {
	req := CheckoutRequest{
		Items:      nil,
		PaidAmount: 50000,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "")
}

func TestValidateCheckoutRequest_EmptyItemsSlice(t *testing.T) {
	req := CheckoutRequest{
		Items:      []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{},
		PaidAmount: 50000,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "")
}

func TestValidateCheckoutRequest_ProductIDEmpty(t *testing.T) {
	req := CheckoutRequest{
		Items: []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{
			{ProductID: "", Quantity: 1},
		},
		PaidAmount: 50000,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "items[0].product_id")
}

func TestValidateCheckoutRequest_ProductIDBlankSpaces(t *testing.T) {
	req := CheckoutRequest{
		Items: []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{
			{ProductID: "   ", Quantity: 1},
		},
		PaidAmount: 50000,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "items[0].product_id")
}

func TestValidateCheckoutRequest_QuantityZero(t *testing.T) {
	req := CheckoutRequest{
		Items: []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{
			{ProductID: "prod-001", Quantity: 0},
		},
		PaidAmount: 50000,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "items[0].quantity")
}

func TestValidateCheckoutRequest_QuantityNegative(t *testing.T) {
	req := CheckoutRequest{
		Items: []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{
			{ProductID: "prod-001", Quantity: -3},
		},
		PaidAmount: 50000,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "items[0].quantity")
}

func TestValidateCheckoutRequest_SecondItemInvalid(t *testing.T) {
	req := CheckoutRequest{
		Items: []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{
			{ProductID: "prod-001", Quantity: 1},
			{ProductID: "", Quantity: 2}, // invalid
		},
		PaidAmount: 50000,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "items[1].product_id")
}

func TestValidateCheckoutRequest_PaidAmountZero(t *testing.T) {
	req := CheckoutRequest{
		Items: []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{
			{ProductID: "prod-001", Quantity: 1},
		},
		PaidAmount: 0,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "paid_amount")
}

func TestValidateCheckoutRequest_PaidAmountNegative(t *testing.T) {
	req := CheckoutRequest{
		Items: []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		}{
			{ProductID: "prod-001", Quantity: 1},
		},
		PaidAmount: -1000,
	}

	err := validateCheckoutRequest(req)
	assertValidationError(t, err, "paid_amount")
}

// ─── ValidationError type ────────────────────────────────────────────────────

func TestValidationError_WithField(t *testing.T) {
	err := &ValidationError{Field: "paid_amount", Message: "must be > 0"}
	want := "paid_amount: must be > 0"
	if err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestValidationError_WithoutField(t *testing.T) {
	err := &ValidationError{Message: "cart is empty"}
	if err.Error() != "cart is empty" {
		t.Fatalf("expected %q, got %q", "cart is empty", err.Error())
	}
}

func TestIsValidationError_True(t *testing.T) {
	err := &ValidationError{Message: "test"}
	if !isValidationError(err) {
		t.Fatal("expected true for *ValidationError")
	}
}

func TestIsValidationError_False(t *testing.T) {
	err := errors.New("some other error")
	if isValidationError(err) {
		t.Fatal("expected false for non-ValidationError")
	}
}

// ─── Handler integration (no DB) ─────────────────────────────────────────────

func TestCheckoutHandler_ValidationErrors(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		wantCode ErrorCode
	}{
		{
			name:     "empty items",
			body:     `{"items":[],"paid_amount":50000}`,
			wantCode: ErrCodeValidation,
		},
		{
			name:     "product_id blank",
			body:     `{"items":[{"product_id":"","quantity":1}],"paid_amount":50000}`,
			wantCode: ErrCodeValidation,
		},
		{
			name:     "quantity zero",
			body:     `{"items":[{"product_id":"prod-001","quantity":0}],"paid_amount":50000}`,
			wantCode: ErrCodeValidation,
		},
		{
			name:     "quantity negative",
			body:     `{"items":[{"product_id":"prod-001","quantity":-1}],"paid_amount":50000}`,
			wantCode: ErrCodeValidation,
		},
		{
			name:     "paid_amount zero",
			body:     `{"items":[{"product_id":"prod-001","quantity":1}],"paid_amount":0}`,
			wantCode: ErrCodeValidation,
		},
		{
			name:     "paid_amount negative",
			body:     `{"items":[{"product_id":"prod-001","quantity":1}],"paid_amount":-1}`,
			wantCode: ErrCodeValidation,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewCheckoutHandler(nil, nil)
			req := httptest.NewRequest("POST", "/checkout", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			h.Checkout(rec, req)

			if rec.Code != 400 {
				t.Fatalf("expected 400, got %d", rec.Code)
			}

			assertErrorCode(t, rec, tc.wantCode)
		})
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// assertValidationError asserts err is a *ValidationError.
// If field is non-empty, also asserts the field name matches.
func assertValidationError(t *testing.T, err error, field string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected ValidationError, got nil")
	}

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}

	if field != "" && ve.Field != field {
		t.Errorf("expected field %q, got %q (message: %s)", field, ve.Field, ve.Message)
	}
}
