package retry

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	MaxAttempts   int
	Delay         time.Duration
	BackoffFunc   func(attempt int) time.Duration
	Logger        func(attempt int, err error)
	IsRetryable   func(err error) bool
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		Delay:       1 * time.Second,
		BackoffFunc: func(attempt int) time.Duration {
			// Exponential backoff: 1s, 2s, 4s
			return time.Duration(1<<uint(attempt-1)) * time.Second
		},
		IsRetryable: func(err error) bool {
			// By default, retry all errors
			return true
		},
	}
}

func Do(config Config, operation func() error) error {
	if config.MaxAttempts < 1 {
		config.MaxAttempts = 1
	}

	if config.Delay < 0 {
		config.Delay = 0
	}

	if config.BackoffFunc == nil {
		config.BackoffFunc = func(attempt int) time.Duration {
			return config.Delay
		}
	}

	if config.IsRetryable == nil {
		config.IsRetryable = func(err error) bool {
			return true
		}
	}

	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !config.IsRetryable(err) {
			break
		}

		if attempt == config.MaxAttempts {
			break
		}

		backoff := config.BackoffFunc(attempt)
		if config.Logger != nil {
			config.Logger(attempt, err)
		}

		time.Sleep(backoff)
	}

	return fmt.Errorf("failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// Common non-retryable errors
var (
	ErrValidation = errors.New("validation error")
	ErrTimeout    = errors.New("timeout error")
	ErrPermission = errors.New("permission denied")
)

// DefaultIsRetryable returns true for retryable errors, false for non-retryable ones
func DefaultIsRetryable(err error) bool {
	if errors.Is(err, ErrValidation) {
		return false
	}
	if errors.Is(err, ErrTimeout) {
		return false
	}
	if errors.Is(err, ErrPermission) {
		return false
	}
	return true
}
