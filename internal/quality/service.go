package quality

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Record
	order []string
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Record), clock: clock}
}

func (s *Service) Create(r Record) (Record, error) {
	if r.Result == "" {
		r.Result = ResultPending
	}
	if err := Validate(r); err != nil {
		return Record{}, platform.WrapValidation(err.Error())
	}
	if r.ID == "" {
		r.ID = platform.NewID("qc")
	}
	r.InspectedAt = s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[r.ID]; ok {
		return Record{}, platform.WrapConflict("quality record " + r.ID)
	}
	s.items[r.ID] = r
	s.order = append(s.order, r.ID)
	return r, nil
}

func (s *Service) Decide(id, result, inspector, note string) (Record, error) {
	if !results[result] {
		return Record{}, platform.WrapValidation("unknown result " + result)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.items[id]
	if !ok {
		return Record{}, platform.WrapNotFound("quality record " + id)
	}
	r.Result = result
	r.Inspector = inspector
	r.Note = note
	r.InspectedAt = s.clock.Now()
	s.items[id] = r
	return r, nil
}

func (s *Service) Get(id string) (Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.items[id]
	if !ok {
		return Record{}, platform.WrapNotFound("quality record " + id)
	}
	return r, nil
}

func (s *Service) ListByBatch(batchID string) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Record
	for _, id := range s.order {
		r := s.items[id]
		if batchID == "" || r.BatchID == batchID {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].InspectedAt.Before(out[j].InspectedAt) })
	return out
}

func (s *Service) ListByInbound(inboundID string) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Record
	for _, id := range s.order {
		r := s.items[id]
		if inboundID == "" || r.InboundID == inboundID {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].InspectedAt.Before(out[j].InspectedAt) })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

