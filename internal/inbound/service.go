package inbound

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/stock"
)

type Service struct {
	mu     sync.RWMutex
	items  map[string]Inbound
	order  []string
	clock  platform.Clock
	batch  *production.Service
	stock  *stock.Service
}

func NewService(clock platform.Clock, b *production.Service, st *stock.Service) *Service {
	return &Service{items: make(map[string]Inbound), clock: clock, batch: b, stock: st}
}

func (s *Service) Create(v Inbound) (Inbound, error) {
	v.Status = StatusDraft
	if err := Validate(v); err != nil {
		return Inbound{}, platform.WrapValidation(err.Error())
	}
	if v.ID == "" {
		v.ID = platform.NewID("in")
	}
	if v.No == "" {
		v.No = "IN-" + v.ID[len(v.ID)-6:]
	}
	v.CreatedAt = s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

func (s *Service) Accept(id, qcResult string) (Inbound, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Inbound{}, platform.WrapNotFound("inbound " + id)
	}
	if v.Status != StatusDraft {
		return Inbound{}, platform.WrapConflict("inbound " + id + " is " + v.Status)
	}
	if qcResult == "rejected" {
		v.Status = StatusRejected
	} else {
		v.Status = StatusAccepted
	}
	v.QCResult = qcResult
	s.items[id] = v
	return v, nil
}

// Putaway moves accepted inbound items into stock. Each item must resolve to a
// released batch; if any item is missing or unreleased the whole operation
// aborts and stock already received for earlier items is rolled back, so a
// single bad batch never leaves partial shelves behind.
func (s *Service) Putaway(id string) (Inbound, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Inbound{}, platform.WrapNotFound("inbound " + id)
	}
	if v.Status != StatusAccepted {
		return Inbound{}, platform.WrapConflict("inbound " + id + " is " + v.Status)
	}

	// Receive item by item. On the first missing or unreleased batch, undo
	// the shelves we already filled so the warehouse is not left with partial
	// stock, then surface the error.
	received := make([]InboundItem, 0, len(v.Items))
	for _, it := range v.Items {
		b, err := s.batch.Get(it.BatchID)
		if err != nil {
			s.rollback(v.WarehouseID, received)
			return Inbound{}, platform.WrapNotFound("inbound " + id + ": batch " + it.BatchID)
		}
		if b.Status != production.StatusReleased {
			s.rollback(v.WarehouseID, received)
			return Inbound{}, platform.WrapConflict("inbound " + id + ": batch " + it.BatchID + " is " + b.Status)
		}
		if _, err := s.stock.Receive(it.BatchID, v.WarehouseID, it.Qty); err != nil {
			s.rollback(v.WarehouseID, received)
			return Inbound{}, err
		}
		received = append(received, it)
	}

	v.Status = StatusPutaway
	s.items[id] = v
	return v, nil
}


func (s *Service) Get(id string) (Inbound, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Inbound{}, platform.WrapNotFound("inbound " + id)
	}
	return v, nil
}

func (s *Service) List() []Inbound {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Inbound, 0, len(s.items))
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

