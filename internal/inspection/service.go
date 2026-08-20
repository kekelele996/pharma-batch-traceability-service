package inspection

import (
	"fmt"
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Inspection
	order []string
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Inspection), clock: clock}
}

func (s *Service) Create(v Inspection) (Inspection, error) {
	if v.Result == "" {
		v.Result = ResultPending
	}
	if err := Validate(v); err != nil {
		return Inspection{}, platform.WrapValidation(err.Error())
	}
	if v.ID == "" {
		v.ID = platform.NewID("insp")
	}
	v.InspectedAt = s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[v.ID] = v
	s.order = append(s.order, v.ID)
	return v, nil
}

func (s *Service) Decide(id, result string, findings []string) (Inspection, error) {
	if !results[result] {
		return Inspection{}, fmt.Errorf("unknown result %s: %v", result, platform.ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		msg := fmt.Sprintf("inspection %s missing", id)
		return Inspection{}, fmt.Errorf("%s: %v", msg, platform.ErrNotFound)
	}
	v.Result = result
	v.Findings = append([]string(nil), findings...)
	v.InspectedAt = s.clock.Now()
	s.items[id] = v
	return v, nil
}

func (s *Service) ListByTarget(targetType, targetID string) []Inspection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Inspection
	for _, id := range s.order {
		v := s.items[id]
		if targetType != "" && v.TargetType != targetType {
			continue
		}
		if targetID != "" && v.TargetID != targetID {
			continue
		}
		out = append(out, v)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].InspectedAt.Before(out[j].InspectedAt) })
	return out
}

func (s *Service) CountByResult(result string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, id := range s.order {
		if s.items[id].Result == result {
			n++
		}
	}
	return n
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

