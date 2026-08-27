package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pos-system/internal/api"
	"pos-system/internal/db"
	applogger "pos-system/internal/logger"
)

func TestHealthEndpoint(t *testing.T) {
	// Test health endpoint without database
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router := api.NewRouter(nil, nil, nil, nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var response api.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Fatalf("expected status ok, got %q", response.Status)
	}
}

func TestReadyEndpointWithoutDB(t *testing.T) {
	// Test readiness endpoint without database connection
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	router := api.NewRouter(nil, nil, nil, nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable without DB, got %d: %s", rec.Code, rec.Body.String())
	}

	var response api.ReadinessResponse
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

func TestReadyEndpointWithDB(t *testing.T) {
	// Test readiness endpoint with database connection
	applogger.Init()

	database, err := db.OpenPostgres("")
	if err != nil {
		t.Skip("Skipping test: database connection failed")
	}
	defer database.Close()

	// Give database time to connect
	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	router := api.NewRouter(nil, nil, nil, database)
	router.ServeHTTP(rec, req)

	var response api.ReadinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.API != true {
		t.Fatal("expected API to be true")
	}

	// DB status depends on actual connection
	if response.DB {
		if !response.Ready {
			t.Fatal("expected Ready to be true when DB is connected")
		}
	} else {
		if response.Ready {
			t.Fatal("expected Ready to be false when DB is not connected")
		}
	}
}

func TestGracefulShutdown(t *testing.T) {
	// Test that database connection can be closed properly
	applogger.Init()

	database, err := db.OpenPostgres("")
	if err != nil {
		t.Skip("Skipping test: database connection failed")
	}

	// Test that database can be closed
	if err := database.Close(); err != nil {
		t.Fatalf("failed to close database: %v", err)
	}
}