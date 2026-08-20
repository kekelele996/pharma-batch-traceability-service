package warehouse

import (
	"fmt"
	"strings"
)

type Warehouse struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	TempZone string `json:"temp_zone"`
	Province string `json:"province"`
	City     string `json:"city"`
	Status   string `json:"status"`
}

var tempZones = map[string]bool{"room": true, "cool": true, "cold": true, "frozen": true}
var statuses = map[string]bool{"active": true, "closed": true}

func Validate(w Warehouse) error {
	w.Code = strings.TrimSpace(w.Code)
	w.Name = strings.TrimSpace(w.Name)
	if w.Code == "" || w.Name == "" {
		return fmt.Errorf("warehouse: code and name required")
	}
	if !tempZones[w.TempZone] {
		return fmt.Errorf("warehouse: unknown temp zone %q", w.TempZone)
	}
	if !statuses[w.Status] {
		return fmt.Errorf("warehouse: unknown status %q", w.Status)
	}
	return nil
}

// TempZoneCompatible reports whether a warehouse temp zone satisfies a drug
// storage condition. Colder zones can always host warmer requirements.
func TempZoneCompatible(storage, zone string) bool {
	order := map[string]int{"room": 0, "cool": 1, "cold": 2, "frozen": 3}
	need, ok1 := order[storage]
	have, ok2 := order[zone]
	if !ok1 || !ok2 {
		return false
	}
	if storage == "light-proof" {
		return ok2 && zone != ""
	}
	return have >= need
}

func (w *Warehouse) Normalize() {
	w.Code = strings.TrimSpace(w.Code)
	w.Name = strings.TrimSpace(w.Name)
	w.TempZone = strings.ToLower(strings.TrimSpace(w.TempZone))
	w.Province = strings.TrimSpace(w.Province)
	w.City = strings.TrimSpace(w.City)
	w.Status = strings.ToLower(strings.TrimSpace(w.Status))
}
