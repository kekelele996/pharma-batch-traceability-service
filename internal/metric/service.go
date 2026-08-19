package metric

import (
	"sort"
	"sync"
)

type Service struct {
	mu       sync.RWMutex
	counters map[string]int64
	dims     map[string]map[string]int64
}

func NewService() *Service {
	return &Service{counters: make(map[string]int64), dims: make(map[string]map[string]int64)}
}

func (s *Service) Inc(key string, n int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[key] += n
}

func (s *Service) Set(key string, v int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[key] = v
}

func (s *Service) Get(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counters[key]
}

// IncDim records a value against a key/dimension pair, initialising the inner
// dimension map on first use.
func (s *Service) IncDim(key, dim string, n int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dims[key][dim] += n
}

// Dims returns a snapshot of the per-dimension counters.
func (s *Service) Dims() map[string]map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dims
}

func (s *Service) Snapshot() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counters
}

func (s *Service) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.counters))
	for k := range s.counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
