package label

import (
	"fmt"
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Label
	order []string
	clock platform.Clock
	batch *production.Service
}

func NewService(clock platform.Clock, b *production.Service) *Service {
	return &Service{items: make(map[string]Label), clock: clock, batch: b}
}

// Generate renders qty traceability labels for a batch, embedding its batch
// number, drug and expiry date in a fixed label template.
func (s *Service) Generate(batchID string, qty int) ([]Label, error) {
	if qty <= 0 || qty > 100000 {
		return nil, platform.WrapValidation("label quantity must be in (0, 100000]")
	}
	b, err := s.batch.Get(batchID)
	if err != nil {
		return nil, err
	}
	content := fmt.Sprintf("药品追溯标签|批号:%s|药品:%s|有效期:%s", b.BatchNo, b.DrugID, b.ExpiryDate.Format("2006-01-02"))
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Label, 0, qty)
	for i := 0; i < qty; i++ {
		l := Label{
			ID: platform.NewID("lbl"), BatchID: batchID, Content: content,
			Status: StatusUnprinted, Qty: 1, CreatedAt: s.clock.Now(),
		}
		s.items[l.ID] = l
		s.order = append(s.order, l.ID)
		out = append(out, l)
	}
	return out, nil
}

func (s *Service) Print(id string) (Label, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.items[id]
	if !ok {
		return Label{}, platform.WrapNotFound("label " + id)
	}
	if l.Status == StatusPrinted {
		return Label{}, platform.WrapConflict("label " + id + " already printed")
	}
	l.Status = StatusPrinted
	s.items[id] = l
	return l, nil
}

func (s *Service) ListByBatch(batchID string) []Label {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Label
	for _, id := range s.order {
		l := s.items[id]
		if batchID == "" || l.BatchID == batchID {
			out = append(out, l)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

