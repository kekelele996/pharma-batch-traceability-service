package manufacturer

import (
	"context"
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Manufacturer
	order []string
}

func NewService() *Service {
	return &Service{items: make(map[string]Manufacturer)}
}

func (s *Service) Create(m Manufacturer) (Manufacturer, error) {
	m.Normalize()
	if m.Status == "" {
		m.Status = "active"
	}
	if err := Validate(m); err != nil {
		return Manufacturer{}, platform.WrapValidation(err.Error())
	}
	if m.ID == "" {
		m.ID = platform.NewID("mfr")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[m.ID]; ok {
		return Manufacturer{}, platform.WrapConflict("manufacturer " + m.ID)
	}
	s.items[m.ID] = m
	s.order = append(s.order, m.ID)
	return m, nil
}

func (s *Service) Get(id string) (Manufacturer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.items[id]
	if !ok {
		return Manufacturer{}, platform.WrapNotFound("manufacturer " + id)
	}
	return m, nil
}

func (s *Service) SetStatus(id, status string) (Manufacturer, error) {
	if !statuses[status] {
		return Manufacturer{}, platform.WrapValidation("unknown status " + status)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.items[id]
	if !ok {
		return Manufacturer{}, platform.WrapNotFound("manufacturer " + id)
	}
	m.Status = status
	s.items[id] = m
	return m, nil
}

// VerifyAll checks that every manufacturer id exists and is active, honouring
// context cancellation between iterations.
func (s *Service) VerifyAll(ctx context.Context, ids []string) (int, error) {
	verified := 0
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return verified, err
		}
		m, err := s.Get(id)
		if err != nil {
			return verified, err
		}
		if m.Status != "active" {
			return verified, platform.WrapConflict("manufacturer " + id + " not active")
		}
		verified++
	}
	return verified, nil
}

func (s *Service) List() []Manufacturer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Manufacturer, 0, len(s.items))
	for _, id := range s.order {
		out = append(out, s.items[id])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}
