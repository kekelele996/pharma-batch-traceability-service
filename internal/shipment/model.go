package shipment

import (
	"fmt"
	"time"
)

const (
	NodeProduction = "production"
	NodeInbound    = "inbound"
	NodeOutbound   = "outbound"
	NodeSale       = "sale"
	NodeRecall     = "recall"
	NodeDispatch   = "dispatch"
)

type Shipment struct {
	ID         string    `json:"id"`
	BatchID    string    `json:"batch_id"`
	SerialNo   string    `json:"serial_no"`
	FromID     string    `json:"from_id"`
	ToID       string    `json:"to_id"`
	NodeType   string    `json:"node_type"`
	TempC      float64   `json:"temp_c"`
	VehicleID  string    `json:"vehicle_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

var nodeTypes = map[string]bool{
	NodeProduction: true, NodeInbound: true, NodeOutbound: true,
	NodeSale: true, NodeRecall: true, NodeDispatch: true,
}

func Validate(s Shipment) error {
	if s.BatchID == "" {
		return fmt.Errorf("shipment: batch id required")
	}
	if !nodeTypes[s.NodeType] {
		return fmt.Errorf("shipment: unknown node type %q", s.NodeType)
	}
	if s.OccurredAt.IsZero() {
		return fmt.Errorf("shipment: occurred_at required")
	}
	return nil
}
