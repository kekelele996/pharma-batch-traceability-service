package device

import (
	"fmt"
	"time"
)

type Device struct {
	ID              string    `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Kind            string    `json:"kind"`
	WarehouseID     string    `json:"warehouse_id"`
	LastCalibration time.Time `json:"last_calibration"`
	CalibrationDue  time.Time `json:"calibration_due"`
	Status          string    `json:"status"`
}

var kinds = map[string]bool{"thermometer": true, "hygrometer": true, "logger": true}
var statuses = map[string]bool{"active": true, "faulty": true, "retired": true}

func Validate(d Device) error {
	if d.Code == "" || d.Name == "" {
		return fmt.Errorf("device: code and name required")
	}
	if !kinds[d.Kind] {
		return fmt.Errorf("device: unknown kind %q", d.Kind)
	}
	if !statuses[d.Status] {
		return fmt.Errorf("device: unknown status %q", d.Status)
	}
	if !d.CalibrationDue.IsZero() && !d.LastCalibration.IsZero() && d.CalibrationDue.Before(d.LastCalibration) {
		return fmt.Errorf("device: calibration due cannot precede last calibration")
	}
	return nil
}
