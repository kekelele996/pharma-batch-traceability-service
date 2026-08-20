package outbound

import (
	"fmt"
	"time"
)

const (
	StatusDraft     = "draft"
	StatusAllocated = "allocated"
	StatusShipped   = "shipped"
	StatusCancelled = "cancelled"
)

type OutboundItem struct {
	DrugID string `json:"drug_id"`
	Qty    int    `json:"qty"`
}

type Allocation struct {
	BatchID string `json:"batch_id"`
	Qty     int    `json:"qty"`
}

type Outbound struct {
	ID          string         `json:"id"`
	No          string         `json:"no"`
	CustomerID  string         `json:"customer_id"`
	WarehouseID string         `json:"warehouse_id"`
	Items       []OutboundItem `json:"items"`
	Allocations []Allocation   `json:"allocations"`
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
}

var statuses = map[string]bool{StatusDraft: true, StatusAllocated: true, StatusShipped: true, StatusCancelled: true}

func Validate(v Outbound) error {
	if v.CustomerID == "" || v.WarehouseID == "" {
		return fmt.Errorf("outbound: customer and warehouse required")
	}
	if len(v.Items) == 0 {
		return fmt.Errorf("outbound: at least one item required")
	}
	seen := map[string]bool{}
	for i, it := range v.Items {
		if it.DrugID == "" || it.Qty <= 0 {
			return fmt.Errorf("outbound: invalid item %d", i)
		}
		if seen[it.DrugID] {
			return fmt.Errorf("outbound: duplicate item %d", i)
		}
		seen[it.DrugID] = true
	}
	if !statuses[v.Status] {
		return fmt.Errorf("outbound: unknown status %q", v.Status)
	}
	return nil
}
