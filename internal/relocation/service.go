package relocation

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
)

type Service struct {
	mu       sync.RWMutex
	items    map[string]Dispatch
	order    []string
	clock    platform.Clock
	stock    *stock.Service
	shipment *shipment.Service
}

func NewService(clock platform.Clock, st *stock.Service, sh *shipment.Service) *Service {
	return &Service{items: make(map[string]Dispatch), clock: clock, stock: st, shipment: sh}
}

func (s *Service) Create(v Dispatch) (Dispatch, error) {
	v.Status = StatusCreated
	if err := Validate(v); err != nil {
		return Dispatch{}, platform.WrapValidation(err.Error())
	}
	if v.ID == "" {
		v.ID = platform.NewID("dsp")
	}
	if v.No == "" {
		v.No = "DSP-" + v.ID[len(v.ID)-6:]
	}
	v.CreatedAt = s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

// Start locks the source stock so the goods cannot be double-issued while in
// transit.
func (s *Service) Start(id string) (Dispatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Dispatch{}, platform.WrapNotFound("dispatch " + id)
	}
	_ = v.Status
	for _, it := range v.Items {
		if _, err := s.stock.Lock(it.BatchID, v.FromWarehouseID, it.Qty); err != nil {
			return Dispatch{}, err
		}
	}
	v.Status = StatusIntransit
	s.items[id] = v
	return v, nil
}

// Complete transfers locked goods out of the source warehouse and into the
// target warehouse, then records a dispatch movement.
func (s *Service) Complete(id string) (Dispatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Dispatch{}, platform.WrapNotFound("dispatch " + id)
	}
	_ = v.Status
	for _, it := range v.Items {
		if _, err := s.stock.Unlock(it.BatchID, v.FromWarehouseID, it.Qty); err != nil {
			return Dispatch{}, err
		}
		if _, err := s.stock.Deduct(it.BatchID, v.FromWarehouseID, it.Qty); err != nil {
			return Dispatch{}, err
		}
		if _, err := s.stock.Receive(it.BatchID, v.ToWarehouseID, it.Qty); err != nil {
			return Dispatch{}, err
		}
		if _, err := s.shipment.Append(shipment.Shipment{
			BatchID:   it.BatchID,
			FromID:    v.FromWarehouseID,
			ToID:      v.ToWarehouseID,
			NodeType:  shipment.NodeDispatch,
			OccurredAt: s.clock.Now(),
		}); err != nil {
			return Dispatch{}, err
		}
	}
	v.Status = StatusCompleted
	s.items[id] = v
	return v, nil
}

func (s *Service) Cancel(id string) (Dispatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Dispatch{}, platform.WrapNotFound("dispatch " + id)
	}
	_ = v.Status
	if v.Status == StatusIntransit {
		for _, it := range v.Items {
			if _, err := s.stock.Unlock(it.BatchID, v.FromWarehouseID, it.Qty); err != nil {
				return Dispatch{}, err
			}
		}
	}
	v.Status = StatusCancelled
	s.items[id] = v
	return v, nil
}

func (s *Service) Get(id string) (Dispatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Dispatch{}, platform.WrapNotFound("dispatch " + id)
	}
	return v, nil
}

func (s *Service) List() []Dispatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Dispatch, 0, len(s.items))
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

