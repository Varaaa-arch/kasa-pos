package receipt

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestIdempotencyStoreClaim(t *testing.T) {
	store := NewIdempotencyStore()

	if err := store.Claim("KEY-001"); err != nil {
		t.Fatalf(
			"first claim failed: %v",
			err,
		)
	}

	if err := store.Claim("KEY-001"); !errors.Is(
		err,
		ErrDuplicatePrintJob,
	) {
		t.Fatalf(
			"expected duplicate error, got %v",
			err,
		)
	}
}

func TestIdempotencyStoreStatus(t *testing.T) {
	store := NewIdempotencyStore()

	if err := store.Claim("KEY-001"); err != nil {
		t.Fatal(err)
	}

	store.SetStatus(
		"KEY-001",
		PrintJobCompleted,
	)

	status, exists := store.Status("KEY-001")

	if !exists {
		t.Fatal("expected key to exist")
	}

	if status != PrintJobCompleted {
		t.Fatalf(
			"expected COMPLETED, got %q",
			status,
		)
	}
}

func TestIdempotencyStoreMissingKey(t *testing.T) {
	store := NewIdempotencyStore()

	_, exists := store.Status("UNKNOWN")

	if exists {
		t.Fatal("expected unknown key to not exist")
	}
}

func TestIdempotencyStoreConcurrentClaim(t *testing.T) {
	store := NewIdempotencyStore()

	const workers = 20
	const key = "CONCURRENT-KEY"

	var wg sync.WaitGroup
	var successfulClaims int32

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			if err := store.Claim(key); err == nil {
				atomic.AddInt32(
					&successfulClaims,
					1,
				)
			}
		}()
	}

	wg.Wait()

	if successfulClaims != 1 {
		t.Fatalf(
			"expected exactly 1 successful claim, got %d",
			successfulClaims,
		)
	}

	status, exists := store.Status(key)

	if !exists {
		t.Fatal("expected key to exist")
	}

	if status != PrintJobPending {
		t.Fatalf(
			"expected PENDING, got %q",
			status,
		)
	}
}
