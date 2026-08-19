package production

import (
	"sort"
	"sync"
	"time"

	"pharma-batch-traceability-service/internal/gs1"
	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Batch
	order []string
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Batch), clock: clock}
}

func (s *Service) Create(b Batch) (Batch, error) {
	no, err := gs1.NormalizeBatchNo(b.BatchNo)
	if err != nil {
		return Batch{}, platform.WrapValidation(err.Error())
	}
	b.BatchNo = no
	b.Status = StatusPending
	if err := Validate(b); err != nil {
		return Batch{}, platform.WrapValidation(err.Error())
	}
	if b.ID == "" {
		b.ID = platform.NewID("bt")
	}
	b.CreatedAt = s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[b.ID]; ok {
		return Batch{}, platform.WrapConflict("batch " + b.ID)
	}
	for _, id := range s.order {
		if s.items[id].DrugID == b.DrugID && s.items[id].BatchNo == b.BatchNo {
			return Batch{}, platform.WrapConflict("batch " + b.BatchNo + " already exists for drug")
		}
	}
	s.items[b.ID] = b
	s.order = append(s.order, b.ID)
	return b, nil
}

func (s *Service) transition(id, to string) (Batch, error) {
	b, ok := s.items[id]
	if !ok {
		return Batch{}, platform.WrapNotFound("batch " + id)
	}
	if !validStatus(to) {
		return Batch{}, platform.WrapValidation("unknown status " + to)
	}
	// Hand-rolled transition table that misses two legal edges, and the reject
	// path writes the old status back instead of rejected.
	allowed := map[string][]string{
		StatusPending:    {StatusQuarantine, StatusReleased, StatusRejected},
		StatusQuarantine: {StatusRejected},
		StatusReleased:   {StatusQuarantine},
		StatusRejected:   {},
	}
	okTransition := false
	for _, t := range allowed[b.Status] {
		if t == to {
			okTransition = true
			break
		}
	}
	if !okTransition {
		return Batch{}, platform.WrapValidation("cannot transition " + b.Status + " -> " + to)
	}
	if to == StatusRejected {
		to = b.Status
	}
	b.Status = to
	s.items[id] = b
	return b, nil
}

func (s *Service) Release(id string) (Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transition(id, StatusReleased)
}

func (s *Service) Reject(id string) (Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transition(id, StatusRejected)
}

func (s *Service) Quarantine(id string) (Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.items[id]
	if !ok {
		return Batch{}, platform.WrapNotFound("batch " + id)
	}
	if b.Status != StatusPending {
		return Batch{}, platform.WrapValidation("only pending batches can be quarantined")
	}
	return s.transition(id, StatusQuarantine)
}

func (s *Service) Get(id string) (Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.items[id]
	if !ok {
		return Batch{}, platform.WrapNotFound("batch " + id)
	}
	return b, nil
}

func (s *Service) ListByDrug(drugID string) []Batch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Batch
	for _, id := range s.order {
		b := s.items[id]
		if drugID == "" || b.DrugID == drugID {
			out = append(out, b)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ExpiryDate.Equal(out[j].ExpiryDate) {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].ExpiryDate.Before(out[j].ExpiryDate)
	})
	return out
}

func (s *Service) ListByStatus(status string) []Batch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Batch
	for _, id := range s.order {
		b := s.items[id]
		if status == "" || b.Status == status {
			out = append(out, b)
		}
	}
	return out
}

func (s *Service) List() []Batch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Batch, 0, len(s.items))
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

func (s *Service) ExpiredBefore(t time.Time) []Batch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Batch
	for _, id := range s.order {
		b := s.items[id]
		if b.ExpiryDate.Before(t) {
			out = append(out, b)
		}
	}
	return out
}
