package notification

import (
	"sort"
	"sync"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu        sync.RWMutex
	items     map[string]Notification
	order     []string
	clock     platform.Clock
	byChannel map[string]map[string]bool
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Notification), clock: clock, byChannel: make(map[string]map[string]bool)}
}

func (s *Service) Enqueue(channel, recipient, subject, body string) Notification {
	n := Notification{
		ID: platform.NewID("ntf"), Channel: channel, Recipient: recipient,
		Subject: subject, Body: body, Status: StatusPending, CreatedAt: s.clock.Now(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[n.ID] = n
	s.order = append(s.order, n.ID)
	if s.byChannel[channel] == nil {
		s.byChannel[channel] = make(map[string]bool)
	}
	s.byChannel[channel][recipient] = true
	return n
}

// RecipientsByChannel returns the distinct recipients seen on a channel.
func (s *Service) RecipientsByChannel(channel string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []string
	for r := range s.byChannel[channel] {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}

// SendPending marks every pending notification as sent and returns them.
func (s *Service) SendPending() []Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	var sent []Notification
	for _, id := range s.order {
		n := s.items[id]
		if n.Status == StatusPending {
			n.Status = StatusSent
			n.SentAt = s.clock.Now()
			s.items[id] = n
			sent = append(sent, n)
		}
	}
	sort.SliceStable(sent, func(i, j int) bool { return sent[i].CreatedAt.Before(sent[j].CreatedAt) })
	return sent
}

func (s *Service) Fail(id string) (Notification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.items[id]
	if !ok {
		return Notification{}, platform.WrapNotFound("notification " + id)
	}
	n.Status = StatusFailed
	s.items[id] = n
	return n, nil
}

func (s *Service) List(limit int) []Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Notification, 0, len(s.items))
	for _, id := range s.order {
		out = append(out, s.items[id])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

