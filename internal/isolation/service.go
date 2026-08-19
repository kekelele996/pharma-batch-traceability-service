package isolation

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/stock"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Quarantine
	order []string
	clock platform.Clock
	batch *production.Service
	stock *stock.Service
}

func NewService(clock platform.Clock, b *production.Service, st *stock.Service) *Service {
	return &Service{items: make(map[string]Quarantine), clock: clock, batch: b, stock: st}
}

// Place quarantines a batch: freeze every stock row and move the batch status
// to quarantined so it can no longer be shipped.
func (s *Service) Place(batchID, reason string) (Quarantine, error) {
	if reason == "" {
		return Quarantine{}, platform.WrapValidation("reason required")
	}
	v := Quarantine{
		ID:      platform.NewID("qa"),
		BatchID: batchID,
		Reason:  reason,
		Status:  StatusQuarantine,
	}
	if err := Validate(v); err != nil {
		return Quarantine{}, platform.WrapValidation(err.Error())
	}
	v.CreatedAt = s.clock.Now()
	for _, row := range s.stock.ListByBatch(batchID) {
		if _, err := s.stock.Freeze(row.BatchID, row.WarehouseID); err != nil {
			return Quarantine{}, err
		}
	}
	if _, err := s.batch.Quarantine(batchID); err != nil {
		return Quarantine{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

func (s *Service) Release(id string) (Quarantine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Quarantine{}, platform.WrapNotFound("quarantine " + id)
	}
	if v.Status != StatusQuarantine {
		return Quarantine{}, platform.WrapConflict("quarantine " + id + " is " + v.Status)
	}
	for _, row := range s.stock.ListByBatch(v.BatchID) {
		if _, err := s.stock.Unfreeze(row.BatchID, row.WarehouseID); err != nil {
			return Quarantine{}, err
		}
	}
	if _, err := s.batch.Release(v.BatchID); err != nil {
		return Quarantine{}, err
	}
	v.Status = StatusReleased
	s.items[id] = v
	return v, nil
}

func (s *Service) Reject(id string) (Quarantine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Quarantine{}, platform.WrapNotFound("quarantine " + id)
	}
	if v.Status != StatusQuarantine {
		return Quarantine{}, platform.WrapConflict("quarantine " + id + " is " + v.Status)
	}
	if _, err := s.batch.Reject(v.BatchID); err != nil {
		return Quarantine{}, err
	}
	v.Status = StatusRejected
	s.items[id] = v
	return v, nil
}

func (s *Service) Get(id string) (Quarantine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Quarantine{}, platform.WrapNotFound("quarantine " + id)
	}
	return v, nil
}

func (s *Service) ListByBatch(batchID string) []Quarantine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Quarantine
	for _, id := range s.order {
		v := s.items[id]
		if batchID == "" || v.BatchID == batchID {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

