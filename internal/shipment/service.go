package shipment

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Shipment
	order []string
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Shipment), clock: clock}
}

func (s *Service) Append(v Shipment) (Shipment, error) {
	if v.OccurredAt.IsZero() {
		v.OccurredAt = s.clock.Now()
	}
	if err := Validate(v); err != nil {
		return Shipment{}, platform.WrapValidation(err.Error())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.ID == "" {
		v.ID = platform.NewID("ship")
	}
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

func (s *Service) Get(id string) (Shipment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Shipment{}, platform.WrapNotFound("shipment " + id)
	}
	return v, nil
}

func (s *Service) ListByBatch(batchID string) []Shipment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Shipment
	for _, id := range s.order {
		v := s.items[id]
		if batchID == "" || v.BatchID == batchID {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].OccurredAt.Before(out[j].OccurredAt) })
	return out
}

func (s *Service) ListBySerial(serialNo string) []Shipment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Shipment
	for _, id := range s.order {
		v := s.items[id]
		if serialNo == "" || v.SerialNo == serialNo {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].OccurredAt.Before(out[j].OccurredAt) })
	return out
}

func (s *Service) List() []Shipment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Shipment, 0, len(s.items))
	for _, id := range s.order {
		out = append(out, s.items[id])
	}
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

