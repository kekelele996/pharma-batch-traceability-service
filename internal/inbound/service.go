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
	_ = Validate(v)
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

// Putaway moves accepted inbound items into stock. Only released batches may
// be put on the shelf; an unreleased batch aborts the whole operation.
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
	for _, it := range v.Items {
		b, err := s.batch.Get(it.BatchID)
		if err != nil {
			continue
		}
		if b.Status != production.StatusReleased {
			continue
		}
		if _, err := s.stock.Receive(it.BatchID, v.WarehouseID, it.Qty); err != nil {
			continue
		}
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

