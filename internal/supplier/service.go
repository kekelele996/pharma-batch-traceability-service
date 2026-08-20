package supplier

import (
	"context"
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Supplier
	order []string
}

func NewService() *Service {
	return &Service{items: make(map[string]Supplier)}
}

func (s *Service) Create(v Supplier) (Supplier, error) {
	v.Normalize()
	if v.Status == "" {
		v.Status = "active"
	}
	if err := Validate(v); err != nil {
		return Supplier{}, platform.WrapValidation(err.Error())
	}
	if v.ID == "" {
		v.ID = platform.NewID("sup")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[v.ID]; ok {
		return Supplier{}, platform.WrapConflict("supplier " + v.ID)
	}
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

func (s *Service) Get(id string) (Supplier, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Supplier{}, platform.WrapNotFound("supplier " + id)
	}
	return v, nil
}

func (s *Service) SetStatus(id, status string) (Supplier, error) {
	if !statuses[status] {
		return Supplier{}, platform.WrapValidation("unknown status " + status)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Supplier{}, platform.WrapNotFound("supplier " + id)
	}
	v.Status = status
	s.items[id] = v
	return v, nil
}

// VerifyAll checks that every supplier id exists and is active, honouring
// context cancellation between iterations.
func (s *Service) VerifyAll(ctx context.Context, ids []string) (int, error) {
	verified := 0
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return verified, err
		}
		v, err := s.Get(id)
		if err != nil {
			return verified, err
		}
		if v.Status != "active" {
			return verified, platform.WrapConflict("supplier " + id + " not active")
		}
		verified++
	}
	return verified, nil
}

func (s *Service) List() []Supplier {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Supplier, 0, len(s.items))
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
