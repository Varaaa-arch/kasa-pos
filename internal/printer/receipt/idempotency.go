package receipt

import (
	"errors"
	"sync"
)

var (
	ErrDuplicatePrintJob = errors.New("duplicate print job")
)

type IdempotencyStore struct {
	mu   sync.Mutex
	jobs map[string]PrintJobStatus
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{
		jobs: make(map[string]PrintJobStatus),
	}
}

func (s *IdempotencyStore) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.jobs[key]

	return exists
}

func (s *IdempotencyStore) Claim(
	key string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[key]; exists {
		return ErrDuplicatePrintJob
	}

	s.jobs[key] = PrintJobPending

	return nil
}

func (s *IdempotencyStore) SetStatus(
	key string,
	status PrintJobStatus,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[key] = status
}

func (s *IdempotencyStore) Status(
	key string,
) (PrintJobStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	status, exists := s.jobs[key]

	return status, exists
}


