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
			BackoffFunc: func(attempt int) time.Duration {
				return time.Millisecond
			},
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
			BackoffFunc: func(attempt int) time.Duration {
				return time.Millisecond
			},
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
			BackoffFunc: func(attempt int) time.Duration {
				return time.Millisecond
			},
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
			BackoffFunc: func(attempt int) time.Duration {
				return time.Millisecond
			},
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

func TestDoExponentialBackoff(t *testing.T) {
	attempts := 0
	var backoffs []time.Duration

	err := Do(
		Config{
			MaxAttempts: 3,
			Delay:       0,
			BackoffFunc: func(attempt int) time.Duration {
				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
				backoffs = append(backoffs, backoff)
				return time.Microsecond // Use microsecond for test speed
			},
		},
		func() error {
			attempts++
			return errors.New("always fails")
		},
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}

	// Verify backoff pattern: 1s, 2s
	if len(backoffs) != 2 {
		t.Fatalf("expected 2 backoffs, got %d", len(backoffs))
	}

	if backoffs[0] != 1*time.Second {
		t.Fatalf("expected first backoff 1s, got %v", backoffs[0])
	}

	if backoffs[1] != 2*time.Second {
		t.Fatalf("expected second backoff 2s, got %v", backoffs[1])
	}
}

func TestDoRetryLogging(t *testing.T) {
	attempts := 0
	var loggedAttempts []int
	var loggedErrors []error

	config := Config{
		MaxAttempts: 3,
		Delay:       time.Millisecond,
		BackoffFunc: func(attempt int) time.Duration {
			return time.Millisecond
		},
		Logger: func(attempt int, err error) {
			loggedAttempts = append(loggedAttempts, attempt)
			loggedErrors = append(loggedErrors, err)
		},
	}

	err := Do(
		config,
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

	// Should have logged 2 retries (attempts 1 and 2)
	if len(loggedAttempts) != 2 {
		t.Fatalf("expected 2 logged retries, got %d", len(loggedAttempts))
	}

	if loggedAttempts[0] != 1 {
		t.Fatalf("expected first logged attempt 1, got %d", loggedAttempts[0])
	}

	if loggedAttempts[1] != 2 {
		t.Fatalf("expected second logged attempt 2, got %d", loggedAttempts[1])
	}
}

func TestDoNonRetryableError(t *testing.T) {
	attempts := 0

	config := Config{
		MaxAttempts: 3,
		Delay:       time.Millisecond,
		BackoffFunc: func(attempt int) time.Duration {
			return time.Millisecond
		},
		IsRetryable: func(err error) bool {
			return !errors.Is(err, ErrValidation)
		},
	}

	err := Do(
		config,
		func() error {
			attempts++
			return ErrValidation
		},
	)

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	// Should only attempt once since validation error is non-retryable
	if attempts != 1 {
		t.Fatalf("expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestDoRetryableError(t *testing.T) {
	attempts := 0

	config := Config{
		MaxAttempts: 3,
		Delay:       time.Millisecond,
		BackoffFunc: func(attempt int) time.Duration {
			return time.Millisecond
		},
		IsRetryable: func(err error) bool {
			return !errors.Is(err, ErrValidation)
		},
	}

	err := Do(
		config,
		func() error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should retry since temporary error is retryable
	if attempts != 3 {
		t.Fatalf("expected 3 attempts for retryable error, got %d", attempts)
	}
}

func TestDefaultIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "temporary error is retryable",
			err:      errors.New("temporary error"),
			expected: true,
		},
		{
			name:     "validation error is not retryable",
			err:      ErrValidation,
			expected: false,
		},
		{
			name:     "timeout error is not retryable",
			err:      ErrTimeout,
			expected: false,
		},
		{
			name:     "permission error is not retryable",
			err:      ErrPermission,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultIsRetryable(tt.err)
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
