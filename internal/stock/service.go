package stock

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Stock
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Stock), clock: clock}
}

func (s *Service) get(k string) (Stock, bool) {
	v, ok := s.items[k]
	return v, ok
}

func (s *Service) set(k string, v Stock) {
	v.UpdatedAt = s.clock.Now()
	s.items[k] = v
}

// Receive adds on-hand quantity for a batch in a warehouse, creating a row when
// none exists yet.
func (s *Service) Receive(batchID, warehouseID string, qty int) (Stock, error) {
	if qty <= 0 {
		return Stock{}, platform.WrapValidation("receive quantity must be positive")
	}
	k := Key(batchID, warehouseID)
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.get(k)
	if !ok {
		v = Stock{BatchID: batchID, WarehouseID: warehouseID}
	}
	v.Quantity += qty
	s.set(k, v)
	return v, nil
}

// Deduct removes available quantity, refusing to over-issue. The read, the
// availability check, and the mutation all happen under the write lock so that
// concurrent deducts cannot race on a stale snapshot and over-issue stock.
func (s *Service) Deduct(batchID, warehouseID string, qty int) (Stock, error) {
	if qty <= 0 {
		return Stock{}, platform.WrapValidation("deduct quantity must be positive")
	}
	k := Key(batchID, warehouseID)
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.get(k)
	if !ok {
		return Stock{}, platform.WrapNotFound("stock " + k)
	}
	if v.Available() < qty {
		return Stock{}, platform.WrapConflict("insufficient available stock for " + k)
	}
	v.Quantity -= qty
	s.set(k, v)
	return v, nil
}

// Lock reserves quantity for an in-flight outbound or relocation. The check
// and the mutation run under the write lock so concurrent locks cannot reserve
// more than what is actually available.
func (s *Service) Lock(batchID, warehouseID string, qty int) (Stock, error) {
	if qty <= 0 {
		return Stock{}, platform.WrapValidation("lock quantity must be positive")
	}
	k := Key(batchID, warehouseID)
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.get(k)
	if !ok {
		return Stock{}, platform.WrapNotFound("stock " + k)
	}
	if v.Available() < qty {
		return Stock{}, platform.WrapConflict("insufficient stock to lock " + k)
	}
	v.Locked += qty
	s.set(k, v)
	return v, nil
}

// Unlock releases previously locked quantity. The check and the mutation run
// under the write lock so concurrent unlocks cannot drive the locked count
// below zero.
func (s *Service) Unlock(batchID, warehouseID string, qty int) (Stock, error) {
	if qty <= 0 {
		return Stock{}, platform.WrapValidation("unlock quantity must be positive")
	}
	k := Key(batchID, warehouseID)
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.get(k)
	if !ok {
		return Stock{}, platform.WrapNotFound("stock " + k)
	}
	if v.Locked < qty {
		return Stock{}, platform.WrapConflict("cannot unlock more than locked for " + k)
	}
	v.Locked -= qty
	s.set(k, v)
	return v, nil
}

func (s *Service) Freeze(batchID, warehouseID string) (Stock, error) {
	k := Key(batchID, warehouseID)
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.get(k)
	if !ok {
		return Stock{}, platform.WrapNotFound("stock " + k)
	}
	v.Frozen = true
	s.set(k, v)
	return v, nil
}

func (s *Service) Unfreeze(batchID, warehouseID string) (Stock, error) {
	k := Key(batchID, warehouseID)
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.get(k)
	if !ok {
		return Stock{}, platform.WrapNotFound("stock " + k)
	}
	v.Frozen = false
	s.set(k, v)
	return v, nil
}

func (s *Service) Get(batchID, warehouseID string) (Stock, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.get(Key(batchID, warehouseID))
	if !ok {
		return Stock{}, platform.WrapNotFound("stock " + Key(batchID, warehouseID))
	}
	return v, nil
}

func (s *Service) ListByBatch(batchID string) []Stock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Stock
	for _, v := range s.items {
		if batchID == "" || v.BatchID == batchID {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].WarehouseID < out[j].WarehouseID })
	return out
}

func (s *Service) ListByWarehouse(warehouseID string) []Stock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Stock
	for _, v := range s.items {
		if warehouseID == "" || v.WarehouseID == warehouseID {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].BatchID < out[j].BatchID })
	return out
}

func (s *Service) List() []Stock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Stock, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].BatchID == out[j].BatchID {
			return out[i].WarehouseID < out[j].WarehouseID
		}
		return out[i].BatchID < out[j].BatchID
	})
	return out
}

func (s *Service) TotalAvailable(batchID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := 0
	for _, v := range s.items {
		if batchID == "" || v.BatchID == batchID {
			total += v.Available()
		}
	}
	return total
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

