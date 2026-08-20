package audit

import (
	"sort"
	"sync"
	"time"

	"pharma-batch-traceability-service/internal/platform"
)

type Entry struct {
	ID     string    `json:"id"`
	Actor  string    `json:"actor"`
	Action string    `json:"action"`
	Target string    `json:"target"`
	Detail string    `json:"detail"`
	At     time.Time `json:"at"`
}

type Service struct {
	mu    sync.RWMutex
	items []Entry
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{clock: clock}
}

func (s *Service) Append(actor, action, target, detail string) Entry {
	e := Entry{
		ID:     platform.NewID("audit"),
		Actor:  actor,
		Action: action,
		Target: target,
		Detail: detail,
		At:     s.clock.Now(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, e)
	return e
}

func (s *Service) List(limit int) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, len(s.items))
	copy(out, s.items)
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (s *Service) ListByTarget(target string) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Entry
	for _, e := range s.items {
		if target == "" || e.Target == target {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}
