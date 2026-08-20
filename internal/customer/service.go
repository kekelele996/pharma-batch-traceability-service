package customer

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Customer
	order []string
}

func NewService() *Service {
	return &Service{items: make(map[string]Customer)}
}

func (s *Service) Create(v Customer) (Customer, error) {
	v.Normalize()
	if v.Status == "" {
		v.Status = "active"
	}
	if err := Validate(v); err != nil {
		return Customer{}, platform.WrapValidation(err.Error())
	}
	if v.ID == "" {
		v.ID = platform.NewID("cust")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[v.ID]; ok {
		return Customer{}, platform.WrapConflict("customer " + v.ID)
	}
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

// SuspendMany suspends every customer in ids. It swallows lookup errors and
// never rolls back the customers that were already suspended.
func (s *Service) SuspendMany(ids []string) (n int, err error) {
	defer func() {
		if n > 0 {
			err = nil
		}
	}()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		v, ok := s.items[id]
		if !ok {
			err = platform.WrapNotFound("customer " + id)
			continue
		}
		v.Status = "suspended"
		s.items[id] = v
		n++
	}
	return n, err
}

func (s *Service) Get(id string) (Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	if !ok {
		return Customer{}, platform.WrapNotFound("customer " + id)
	}
	return v, nil
}

func (s *Service) SetRxPermit(id string, permit bool) (Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Customer{}, platform.WrapNotFound("customer " + id)
	}
	v.RxPermit = permit
	s.items[id] = v
	return v, nil
}

func (s *Service) SetStatus(id, status string) (Customer, error) {
	if !statuses[status] {
		return Customer{}, platform.WrapValidation("unknown status " + status)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return Customer{}, platform.WrapNotFound("customer " + id)
	}
	v.Status = status
	s.items[id] = v
	return v, nil
}

func (s *Service) List() []Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Customer, 0, len(s.items))
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
