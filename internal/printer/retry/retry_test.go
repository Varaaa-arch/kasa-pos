package retry

import (
	"errors"
	"testing"
	"time"
)

func TestDoSucceedsImmediately(t *testing.T) {
	attempts := 0

	err := Do(
		Config{
			MaxAttempts: 3,
			Delay:       time.Millisecond,
		},
		func() error {
			attempts++
			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 1 {
		t.Fatalf(
			"expected 1 attempt, got %d",
			attempts,
		)
	}
}

func TestDoRetriesUntilSuccess(t *testing.T) {
	attempts := 0

	err := Do(
		Config{
			MaxAttempts: 3,
			Delay:       time.Millisecond,
		},
		func() error {
			attempts++

			if attempts < 3 {
				return errors.New("temporary failure")
			}

			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Fatalf(
			"expected 3 attempts, got %d",
			attempts,
		)
	}
}

func TestDoReturnsLastError(t *testing.T) {
	expectedErr := errors.New("printer unavailable")
	attempts := 0

	err := Do(
		Config{
			MaxAttempts: 3,
			Delay:       time.Millisecond,
		},
		func() error {
			attempts++
			return expectedErr
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	if attempts != 3 {
		t.Fatalf(
			"expected 3 attempts, got %d",
			attempts,
		)
	}
}

func TestDoSingleAttempt(t *testing.T) {
	attempts := 0

	err := Do(
		Config{
			MaxAttempts: 0,
			Delay:       time.Millisecond,
		},
		func() error {
			attempts++
			return errors.New("failed")
		},
	)

	if attempts != 1 {
		t.Fatalf(
			"expected 1 attempt, got %d",
			attempts,
		)
	}

	if err == nil {
		t.Fatal("expected error")
	}
}
