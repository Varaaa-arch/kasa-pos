package api

import (
	"encoding/json"
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

	NewRouter(nil, nil, nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			rec.Code,
		)
	}

	var response HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Fatalf(
			"expected status ok, got %q",
			response.Status,
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

	NewRouter(nil, nil, nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected 405, got %d",
			rec.Code,
		)
	}
}

func TestReadyWithDB(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	rec := httptest.NewRecorder()

	// With nil DB, should return not ready
	NewRouter(nil, nil, nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected 503 without DB, got %d",
			rec.Code,
		)
	}

	var response ReadinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.API != true {
		t.Fatal("expected API to be true")
	}

	if response.DB != false {
		t.Fatal("expected DB to be false without connection")
	}

	if response.Ready != false {
		t.Fatal("expected Ready to be false without DB")
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
