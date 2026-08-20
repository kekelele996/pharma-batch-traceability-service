package po

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]PurchaseOrder
	order []string
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]PurchaseOrder), clock: clock}
}

func (s *Service) Create(v PurchaseOrder) (PurchaseOrder, error) {
	v.Status = StatusDraft
	if err := Validate(v); err != nil {
		return PurchaseOrder{}, platform.WrapValidation(err.Error())
	}
	if v.ID == "" {
		v.ID = platform.NewID("po")
	}
	if v.No == "" {
		v.No = "PO-" + v.ID[len(v.ID)-6:]
	}
	v.CreatedAt = s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

func (s *Service) Approve(id string) (PurchaseOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return PurchaseOrder{}, platform.WrapNotFound("po " + id)
	}
	if !CanTransition(v.Status, StatusApproved) {
		return PurchaseOrder{}, platform.WrapConflict("po " + id + " is " + v.Status)
	}
	v.Status = StatusApproved
	s.items[id] = v
	return v, nil
}

func (s *Service) Order(id string) (PurchaseOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return PurchaseOrder{}, platform.WrapNotFound("po " + id)
	}
	if !CanTransition(v.Status, StatusOrdered) {
		return PurchaseOrder{}, platform.WrapConflict("po " + id + " is " + v.Status)
	}
	v.Status = StatusOrdered
	s.items[id] = v
	return v, nil
}

func (s *Service) Receive(id string) (PurchaseOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return PurchaseOrder{}, platform.WrapNotFound("po " + id)
	}
	if !CanTransition(v.Status, StatusReceived) {
		return PurchaseOrder{}, platform.WrapConflict("po " + id + " is " + v.Status)
	}
	v.Status = StatusReceived
	s.items[id] = v
	return v, nil
}

func (s *Service) Cancel(id string) (PurchaseOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return PurchaseOrder{}, platform.WrapNotFound("po " + id)
	}
	if !CanTransition(v.Status, StatusCancelled) {
		return PurchaseOrder{}, platform.WrapConflict("po " + id + " is " + v.Status)
	}
	v.Status = StatusCancelled
	s.items[id] = v
	return v, nil
}

func (s *Service) Get(id string) (PurchaseOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return PurchaseOrder{}, platform.WrapNotFound("po " + id)
	}
	return v, nil
}

func (s *Service) List() []PurchaseOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PurchaseOrder, 0, len(s.items))
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
