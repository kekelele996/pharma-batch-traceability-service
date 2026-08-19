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
	if s.dims[key] == nil {
		s.dims[key] = make(map[string]int64)
	}
	s.dims[key][dim] += n
}

// Dims returns a snapshot of the per-dimension counters.
func (s *Service) Dims() map[string]map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]map[string]int64, len(s.dims))
	for k, m := range s.dims {
		cp := make(map[string]int64, len(m))
		for dk, dv := range m {
			cp[dk] = dv
		}
		out[k] = cp
	}
	return out
}

func (s *Service) Snapshot() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		out[k] = v
	}
	return out
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
