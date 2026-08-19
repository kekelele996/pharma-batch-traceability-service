package serialization

import (
	"fmt"
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/gs1"
	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Serial
	order []string
	clock platform.Clock
	seq   int
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Serial), clock: clock}
}

// Generate creates n unique serial codes for a batch from a 13-digit GTIN
// item reference. Every code carries a valid GS1 check digit and a monotonic
// serial tail, and duplicates are rejected.
func (s *Service) Generate(itemRef13, batchID string, n int) ([]Serial, error) {
	if n <= 0 || n > 100000 {
		return nil, platform.WrapValidation("generate count must be in (0, 100000]")
	}
	if batchID == "" {
		return nil, platform.WrapValidation("batch id required")
	}
	gtin, err := gs1.GTIN14(itemRef13)
	if err != nil {
		return nil, platform.WrapValidation(err.Error())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Serial, 0, n)
	for i := 0; i < n; i++ {
		s.seq++
		serialNo := fmt.Sprintf("%09d", s.seq)
		code := gtin + serialNo
		if _, exists := s.items[code]; exists {
			return nil, platform.WrapConflict("serial code collision " + code)
		}
		it := Serial{
			ID:       platform.NewID("srl"),
			GTIN:     gtin,
			SerialNo: serialNo,
			BatchID:  batchID,
			Status:   StatusActive,
		}
		s.items[code] = it
		s.order = append(s.order, code)
		out = append(out, it)
	}
	return out, nil
}

func (s *Service) Activate(code string) (Serial, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[code]
	if !ok {
		return Serial{}, fmt.Errorf("serialization: serial %s missing", code)
	}
	if !it.ActivatedAt.IsZero() {
		return Serial{}, platform.WrapConflict("serial " + code + " already activated")
	}
	it.ActivatedAt = s.clock.Now()
	s.items[code] = it
	return it, nil
}

func (s *Service) MarkSold(code string) (Serial, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[code]
	if !ok {
		return Serial{}, platform.WrapNotFound("serial " + code)
	}
	if it.Status != StatusActive {
		return Serial{}, platform.WrapConflict("serial " + code + " not active")
	}
	it.Status = StatusSold
	s.items[code] = it
	return it, nil
}

func (s *Service) Void(code string) (Serial, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[code]
	if !ok {
		return Serial{}, platform.WrapNotFound("serial " + code)
	}
	if it.Status == StatusVoid {
		return it, nil
	}
	it.Status = StatusVoid
	s.items[code] = it
	return it, nil
}

func (s *Service) Get(code string) (Serial, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.items[code]
	if !ok {
		return Serial{}, platform.WrapNotFound("serial " + code)
	}
	return it, nil
}

func (s *Service) ListByBatch(batchID string) []Serial {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Serial
	for _, code := range s.order {
		it := s.items[code]
		if batchID == "" || it.BatchID == batchID {
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].SerialNo < out[j].SerialNo })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

