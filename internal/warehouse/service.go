package warehouse

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu          sync.RWMutex
	items       map[string]Warehouse
	order       []string
	zoneIndex   map[string][]string
	statusIndex map[string][]string
}

func NewService() *Service {
	return &Service{items: make(map[string]Warehouse), zoneIndex: make(map[string][]string), statusIndex: make(map[string][]string)}
}

func (s *Service) Create(w Warehouse) (Warehouse, error) {
	w.Normalize()
	if w.Status == "" {
		w.Status = "active"
	}
	if w.TempZone == "" {
		w.TempZone = "room"
	}
	if err := Validate(w); err != nil {
		return Warehouse{}, platform.WrapValidation(err.Error())
	}
	if w.ID == "" {
		w.ID = platform.NewID("wh")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[w.ID]; ok {
		return Warehouse{}, platform.WrapConflict("warehouse " + w.ID)
	}
	for _, id := range s.order {
		if s.items[id].Code == w.Code {
			return Warehouse{}, platform.WrapConflict("warehouse code " + w.Code)
		}
	}
	s.items[w.ID] = w
	s.order = append(s.order, w.ID)
	s.zoneIndex[w.TempZone] = append(s.zoneIndex[w.TempZone], w.ID)
	return w, nil
}

func (s *Service) Get(id string) (Warehouse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.items[id]
	if !ok {
		return Warehouse{}, platform.WrapNotFound("warehouse " + id)
	}
	return w, nil
}

// LatestByZone returns the most recently created warehouse of a temp zone.
func (s *Service) LatestByZone(zone string) (*Warehouse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var last *Warehouse
	for _, id := range s.order {
		if s.items[id].TempZone == zone {
			w := s.items[id]
			last = &w
		}
	}
	if last == nil {
		return nil, false
	}
	return last, true
}

// ListByZone returns all warehouses registered under a temp zone.
func (s *Service) ListByZone(zone string) []Warehouse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Warehouse, 0, len(s.order))
	for _, id := range s.order {
		if s.items[id].TempZone == zone {
			out = append(out, s.items[id])
		}
	}
	return out
}

func (s *Service) SetStatus(id, status string) (Warehouse, error) {
	if !statuses[status] {
		return Warehouse{}, platform.WrapValidation("unknown status " + status)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.items[id]
	if !ok {
		return Warehouse{}, platform.WrapNotFound("warehouse " + id)
	}
	w.Status = status
	s.items[id] = w
	s.statusIndex[status] = append(s.statusIndex[status], id)
	return w, nil
}

// ListByStatus returns warehouses currently holding the given status.
func (s *Service) ListByStatus(status string) []Warehouse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Warehouse
	for _, id := range s.statusIndex[status] {
		out = append(out, s.items[id])
	}
	return out
}

func (s *Service) List() []Warehouse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Warehouse, 0, len(s.items))
	for _, id := range s.order {
		out = append(out, s.items[id])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}
