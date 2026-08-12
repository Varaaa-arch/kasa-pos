package retry

import "time"

type Config struct {
	MaxAttempts int
	Delay       time.Duration
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		Delay:       500 * time.Millisecond,
	}
}

func Do(config Config, operation func() error) error {
	if config.MaxAttempts < 1 {
		config.MaxAttempts = 1
	}

	if config.Delay < 0 {
		config.Delay = 0
	}

	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt == config.MaxAttempts {
			break
		}

		time.Sleep(config.Delay)
	}

	return lastErr
}
