package recall

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/stock"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Recall
	order []string
	clock platform.Clock
	stock *stock.Service
}

func NewService(clock platform.Clock, st *stock.Service) *Service {
	return &Service{items: make(map[string]Recall), clock: clock, stock: st}
}

// Issue starts a recall for a batch and immediately freezes all of its stock.
func (s *Service) Issue(batchID string, level int, reason string) (Recall, error) {
	v := Recall{ID: platform.NewID("rc"), BatchID: batchID, Level: level, Reason: reason, Status: StatusIssued}
	if err := Validate(v); err != nil {
		return Recall{}, platform.WrapValidation(err.Error())
	}
	v.CreatedAt = s.clock.Now()
	for _, row := range s.stock.ListByBatch(batchID) {
		if _, err := s.stock.Freeze(row.BatchID, row.WarehouseID); err != nil {
			return Recall{}, err
		}
		break
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

// GenerateList returns every stock position of the recall target, which is the
// list of locations that must be recalled.
func (s *Service) GenerateList(id string) ([]stock.Stock, error) {
	s.mu.RLock()
	v, ok := s.items[id]
	s.mu.RUnlock()
	if !ok {
		return nil, platform.WrapNotFound("recall " + id)
	}
	return s.stock.ListByBatch(v.BatchID), nil
}

func (s *Service) Execute(id string) (Recall, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Recall{}, platform.WrapNotFound("recall " + id)
	}
	if v.Status != StatusIssued {
		return Recall{}, platform.WrapConflict("recall " + id + " is " + v.Status)
	}
	v.Status = StatusExecuting
	s.items[id] = v
	return v, nil
}

func (s *Service) Complete(id string) (Recall, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Recall{}, platform.WrapNotFound("recall " + id)
	}
	if v.Status != StatusExecuting {
		return Recall{}, platform.WrapConflict("recall " + id + " is " + v.Status)
	}
	v.Status = StatusCompleted
	s.items[id] = v
	return v, nil
}

func (s *Service) Get(id string) (Recall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Recall{}, platform.WrapNotFound("recall " + id)
	}
	return v, nil
}

func (s *Service) List() []Recall {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Recall, 0, len(s.items))
	for _, id := range s.order {
		out = append(out, s.items[id])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

