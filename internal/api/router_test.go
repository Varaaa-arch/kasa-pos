package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			rec.Code,
		)
	}

	if rec.Body.String() != `{"status":"ok"}` {
		t.Fatalf(
			"unexpected body: %q",
			rec.Body.String(),
		)
	}
}

func TestHealthMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected 405, got %d",
			rec.Code,
		)
	}
}

func TestRecover(t *testing.T) {
	handler := Recover(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("boom")
		},
	))

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500, got %d",
			rec.Code,
		)
	}
}
