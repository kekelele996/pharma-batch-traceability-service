package coldchain

import "time"

type Reading struct {
	ID         string    `json:"id"`
	ShipmentID string    `json:"shipment_id"`
	DeviceID   string    `json:"device_id"`
	TempC      float64   `json:"temp_c"`
	Humidity   float64   `json:"humidity"`
	At         time.Time `json:"at"`
}
