package coldchain

import (
	"sort"
	"sync"
	"time"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/policy"
)

type Service struct {
	mu    sync.RWMutex
	items []Reading
	clock platform.Clock
}

func NewService(clock platform.Clock) *Service {
	return &Service{clock: clock}
}

// Record appends a temperature reading after validating it against the cold
// chain rule of the required storage condition. A breached reading is rejected
// with a policy error.
func (s *Service) Record(shipmentID, deviceID, storage string, tempC, humidity float64, at time.Time) (Reading, error) {
	if shipmentID == "" || deviceID == "" {
		return Reading{}, platform.WrapValidation("coldchain: shipment and device required")
	}
	if err := policy.ColdChainRule(storage, tempC); err != nil {
		return Reading{}, platform.WrapValidation(err.Error())
	}
	if at.IsZero() {
		at = s.clock.Now()
	}
	r := Reading{
		ID: platform.NewID("cc"), ShipmentID: shipmentID, DeviceID: deviceID,
		TempC: tempC, Humidity: humidity, At: at,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, r)
	return r, nil
}

func (s *Service) ListByShipment(shipmentID string) []Reading {
	var out []Reading
	for _, r := range s.items {
		if shipmentID == "" || r.ShipmentID == shipmentID {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}

func (s *Service) List() []Reading {
	out := make([]Reading, len(s.items))
	copy(out, s.items)
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}

func (s *Service) Min(shipmentID string) (Reading, bool) {
	var list []Reading
	for _, r := range s.items {
		if shipmentID == "" || r.ShipmentID == shipmentID {
			list = append(list, r)
		}
	}
	if len(list) == 0 {
		return Reading{}, false
	}
	min := list[0]
	for _, r := range list[1:] {
		if r.TempC < min.TempC {
			min = r
		}
	}
	return min, true
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

