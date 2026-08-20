package outbound

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
)

type Service struct {
	mu       sync.RWMutex
	items    map[string]Outbound
	order    []string
	clock    platform.Clock
	batch    *production.Service
	stock    *stock.Service
	shipment *shipment.Service
}

func NewService(clock platform.Clock, b *production.Service, st *stock.Service, sh *shipment.Service) *Service {
	return &Service{items: make(map[string]Outbound), clock: clock, batch: b, stock: st, shipment: sh}
}

func (s *Service) Create(v Outbound) (Outbound, error) {
	v.Status = StatusDraft
	if err := Validate(v); err != nil {
		return Outbound{}, platform.WrapValidation(err.Error())
	}
	if v.ID == "" {
		v.ID = platform.NewID("out")
	}
	if v.No == "" {
		v.No = "OUT-" + v.ID[len(v.ID)-6:]
	}
	v.CreatedAt = s.clock.Now()
	v.Items = append([]OutboundItem(nil), v.Items...)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

// Allocate computes FEFO allocations for every line, then locks the stock.
func (s *Service) Allocate(id string) (Outbound, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Outbound{}, platform.WrapNotFound("outbound " + id)
	}
	if v.Status != StatusDraft {
		return Outbound{}, platform.WrapConflict("outbound " + id + " is " + v.Status)
	}
	var all []Allocation
	for _, item := range v.Items {
		batches := s.batch.ListByDrug(item.DrugID)
		plan, remaining := Plan(item.Qty, batches, s.stock, v.WarehouseID)
		if remaining > 0 {
			return Outbound{}, platform.WrapConflict("short of stock for drug " + item.DrugID + " by " + itoa(remaining))
		}
		for _, line := range plan.Lines {
			if _, err := s.stock.Lock(line.BatchID, v.WarehouseID, line.Qty); err != nil {
				return Outbound{}, err
			}
		}
		all = append(all, plan.Lines...)
	}
	v.Allocations = all
	v.Status = StatusAllocated
	s.items[id] = v
	return v, nil
}

// Ship converts locked allocations into actual deductions and records an
// outbound movement for every allocated production.
func (s *Service) Ship(id string) (Outbound, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Outbound{}, platform.WrapNotFound("outbound " + id)
	}
	if v.Status != StatusAllocated {
		return Outbound{}, platform.WrapConflict("outbound " + id + " is " + v.Status)
	}
	for _, line := range v.Allocations {
		if _, err := s.stock.Unlock(line.BatchID, v.WarehouseID, line.Qty); err != nil {
			return Outbound{}, err
		}
		if _, err := s.stock.Deduct(line.BatchID, v.WarehouseID, line.Qty); err != nil {
			return Outbound{}, err
		}
		if _, err := s.shipment.Append(shipment.Shipment{
			BatchID:   line.BatchID,
			FromID:    v.WarehouseID,
			ToID:      v.CustomerID,
			NodeType:  shipment.NodeOutbound,
			OccurredAt: s.clock.Now(),
		}); err != nil {
			return Outbound{}, err
		}
	}
	v.Status = StatusShipped
	s.items[id] = v
	return v, nil
}

// Cancel releases any locked allocations and moves the order to cancelled.
func (s *Service) Cancel(id string) (Outbound, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Outbound{}, platform.WrapNotFound("outbound " + id)
	}
	if v.Status == StatusShipped {
		return Outbound{}, platform.WrapConflict("outbound " + id + " already shipped")
	}
	if v.Status == StatusAllocated {
		for _, line := range v.Allocations {
			if _, err := s.stock.Unlock(line.BatchID, v.WarehouseID, line.Qty); err != nil {
				return Outbound{}, err
			}
		}
	}
	v.Status = StatusCancelled
	s.items[id] = v
	return v, nil
}

// LastAllocations returns the allocations of the most recently allocated order.
func (s *Service) LastAllocations() []Allocation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.order) - 1; i >= 0; i-- {
		v := s.items[s.order[i]]
		if v.Status == StatusAllocated {
			out := make([]Allocation, len(v.Allocations))
			copy(out, v.Allocations)
			return out
		}
	}
	return nil
}

func (s *Service) Get(id string) (Outbound, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Outbound{}, platform.WrapNotFound("outbound " + id)
	}
	return v, nil
}

func (s *Service) List() []Outbound {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Outbound, 0, len(s.items))
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

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

