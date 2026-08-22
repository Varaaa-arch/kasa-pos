package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

// ─── RequestIDFromContext ─────────────────────────────────────────────────────

func TestRequestIDFromContext_Missing(t *testing.T) {
	id := RequestIDFromContext(context.Background())
	if id != "" {
		t.Fatalf("expected empty string for missing key, got %q", id)
	}
}

func TestRequestIDFromContext_Set(t *testing.T) {
	ctx := ContextWithRequestID(context.Background(), "req_abc123")
	id := RequestIDFromContext(ctx)
	if id != "req_abc123" {
		t.Fatalf("expected req_abc123, got %q", id)
	}
}

func TestContextWithRequestID_DoesNotMutateParent(t *testing.T) {
	parent := context.Background()
	child := ContextWithRequestID(parent, "req_xyz")

	// parent tidak boleh ikut berubah
	if RequestIDFromContext(parent) != "" {
		t.Fatal("parent context should not be affected")
	}

	if RequestIDFromContext(child) != "req_xyz" {
		t.Fatal("child context should have the request ID")
	}
}

// ─── Event constants ──────────────────────────────────────────────────────────

func TestEventConstants_NotEmpty(t *testing.T) {
	events := []string{
		EventCheckoutStarted,
		EventCheckoutCompleted,
		EventCheckoutFailed,
		EventPrintStarted,
		EventPrintCompleted,
		EventPrintFailed,
		EventDBError,
	}

	for _, e := range events {
		if e == "" {
			t.Errorf("event constant must not be empty")
		}
	}
}

func TestEventConstants_Unique(t *testing.T) {
	events := []string{
		EventCheckoutStarted,
		EventCheckoutCompleted,
		EventCheckoutFailed,
		EventPrintStarted,
		EventPrintCompleted,
		EventPrintFailed,
		EventDBError,
	}

	seen := make(map[string]bool)
	for _, e := range events {
		if seen[e] {
			t.Errorf("duplicate event constant: %q", e)
		}
		seen[e] = true
	}
}

// ─── Init ─────────────────────────────────────────────────────────────────────

func TestInit_SetsDefaultLogger(t *testing.T) {
	// Redirect slog output to a buffer so we can assert on it.
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	slog.Info("test_event", "event", "test_event", "key", "value")

	out := buf.String()
	if !strings.Contains(out, "test_event") {
		t.Fatalf("expected log output to contain test_event, got: %s", out)
	}
	if !strings.Contains(out, "key=value") {
		t.Fatalf("expected log output to contain key=value, got: %s", out)
	}
}

func TestInit_StructuredFields(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	ctx := ContextWithRequestID(context.Background(), "req_test001")

	slog.InfoContext(ctx, EventCheckoutCompleted,
		"event", EventCheckoutCompleted,
		"request_id", RequestIDFromContext(ctx),
		"transaction_id", "tx_abc",
	)

	out := buf.String()

	checks := []string{
		EventCheckoutCompleted,
		"request_id=req_test001",
		"transaction_id=tx_abc",
	}

	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("expected log to contain %q\ngot: %s", want, out)
		}
	}
}

func TestInit_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	ctx := ContextWithRequestID(context.Background(), "req_err001")

	slog.ErrorContext(ctx, EventCheckoutFailed,
		"event", EventCheckoutFailed,
		"request_id", RequestIDFromContext(ctx),
		"error", "insufficient stock",
	)

	out := buf.String()

	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected level=ERROR in output\ngot: %s", out)
	}
	if !strings.Contains(out, EventCheckoutFailed) {
		t.Errorf("expected event %q in output\ngot: %s", EventCheckoutFailed, out)
	}
	if !strings.Contains(out, "error=") {
		t.Errorf("expected error field in output\ngot: %s", out)
	}
}
