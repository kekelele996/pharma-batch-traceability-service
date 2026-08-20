package drug

import (
	"sort"
	"strings"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Drug
	order []string
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Drug), clock: clock}
}

func (s *Service) Create(d Drug) (Drug, error) {
	d.Normalize()
	if err := Validate(d); err != nil {
		return Drug{}, platform.WrapValidation(err.Error())
	}
	if d.ID == "" {
		d.ID = platform.NewID("drug")
	}
	d.CreatedAt = s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[d.ID]; ok {
		return Drug{}, platform.WrapConflict("drug id " + d.ID)
	}
	s.items[d.ID] = d
	s.order = append(s.order, d.ID)
	return d, nil
}

func (s *Service) Update(id string, patch Drug) (Drug, error) {
	patch.Normalize()
	if err := Validate(patch); err != nil {
		return Drug{}, platform.WrapValidation(err.Error())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.items[id]
	if !ok {
		return Drug{}, platform.WrapNotFound("drug " + id)
	}
	patch.ID = cur.ID
	patch.CreatedAt = cur.CreatedAt
	s.items[id] = patch
	return patch, nil
}

func (s *Service) Get(id string) (Drug, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.items[id]
	if !ok {
		return Drug{}, platform.WrapNotFound("drug " + id)
	}
	return d, nil
}

func (s *Service) List() []Drug {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Drug, 0, len(s.items))
	for _, id := range s.order {
		out = append(out, s.items[id])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (s *Service) Search(q string) []Drug {
	q = strings.ToLower(strings.TrimSpace(q))
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Drug
	for _, id := range s.order {
		d := s.items[id]
		if q == "" || strings.Contains(strings.ToLower(d.GenericName), q) ||
			strings.Contains(strings.ToLower(d.TradeName), q) ||
			strings.Contains(strings.ToLower(d.Code), q) {
			out = append(out, d)
		}
	}
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *Service) StorageFor(id string) (string, error) {
	d, err := s.Get(id)
	if err != nil {
		return "", err
	}
	if d.Storage == "" {
		return "", nil
	}
	return d.Storage, nil
}

