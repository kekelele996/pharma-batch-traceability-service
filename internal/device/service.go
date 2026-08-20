package device

import (
	"sort"
	"sync"
	"time"

	"pharma-batch-traceability-service/internal/platform"
)

type Service struct {
	mu    sync.RWMutex
	items map[string]Device
	order []string
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{items: make(map[string]Device), clock: clock}
}

func (s *Service) Create(d Device) (Device, error) {
	if d.Status == "" {
		d.Status = "active"
	}
	if err := Validate(d); err != nil {
		return Device{}, platform.WrapValidation(err.Error())
	}
	if d.ID == "" {
		d.ID = platform.NewID("dev")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[d.ID]; ok {
		return Device{}, platform.WrapConflict("device " + d.ID)
	}
	for _, id := range s.order {
		if s.items[id].Code == d.Code {
			return Device{}, platform.WrapConflict("device code " + d.Code)
		}
	}
	s.items[d.ID] = d
	s.order = append(s.order, d.ID)
	return d, nil
}

func (s *Service) Get(id string) (Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.items[id]
	if !ok {
		return Device{}, platform.WrapNotFound("device " + id)
	}
	return d, nil
}

func (s *Service) MarkFaulty(id string) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.items[id]
	if !ok {
		return Device{}, platform.WrapNotFound("device " + id)
	}
	d.Status = "faulty"
	s.items[id] = d
	return d, nil
}

func (s *Service) Calibrate(id string, due time.Time) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.items[id]
	if !ok {
		return Device{}, platform.WrapNotFound("device " + id)
	}
	d.LastCalibration = s.clock.Now()
	d.CalibrationDue = due
	d.Status = "active"
	s.items[id] = d
	return d, nil
}

// CalibrationDue returns devices whose calibration deadline is on or before now.
func (s *Service) CalibrationDue(now time.Time) []Device {
	var out []Device
	for _, id := range s.order {
		d := s.items[id]
		if !d.CalibrationDue.IsZero() && !d.CalibrationDue.After(now) && d.Status != "retired" {
			out = append(out, d)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CalibrationDue.Before(out[j].CalibrationDue) })
	return out
}

func (s *Service) List() []Device {
	out := make([]Device, 0, len(s.items))
	for _, id := range s.order {
		out = append(out, s.items[id])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}
