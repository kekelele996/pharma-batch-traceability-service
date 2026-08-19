package relocation

import (
	"fmt"
	"time"
)

const (
	StatusCreated    = "created"
	StatusIntransit  = "intransit"
	StatusArrived    = "arrived"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

type DispatchItem struct {
	DrugID  string `json:"drug_id"`
	BatchID string `json:"batch_id"`
	Qty     int    `json:"qty"`
}

type Dispatch struct {
	ID              string         `json:"id"`
	No              string         `json:"no"`
	FromWarehouseID string         `json:"from_warehouse_id"`
	ToWarehouseID   string         `json:"to_warehouse_id"`
	Items           []DispatchItem `json:"items"`
	Status          string         `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
}

var statuses = map[string]bool{
	StatusCreated: true, StatusIntransit: true, StatusArrived: true,
	StatusCompleted: true, StatusCancelled: true,
}

// CanTransition reports whether a dispatch may move from one status to another.
// A dispatch must be started (in transit) before it can be completed, and only
// a dispatch that has not yet been completed or cancelled may be cancelled.
func CanTransition(from, to string) bool {
	switch from {
	case StatusCreated:
		return to == StatusIntransit || to == StatusCancelled
	case StatusIntransit:
		return to == StatusCompleted || to == StatusCancelled
	case StatusArrived:
		return to == StatusCompleted
	default:
		return false
	}
}

func Validate(v Dispatch) error {
	if v.FromWarehouseID == "" || v.ToWarehouseID == "" {
		return fmt.Errorf("dispatch: source and target warehouses required")
	}
	if v.FromWarehouseID == v.ToWarehouseID {
		return fmt.Errorf("dispatch: source and target warehouses must differ")
	}
	if len(v.Items) == 0 {
		return fmt.Errorf("dispatch: at least one item required")
	}
	for i, it := range v.Items {
		if it.DrugID == "" || it.BatchID == "" || it.Qty <= 0 {
			return fmt.Errorf("dispatch: invalid item %d", i)
		}
	}
	if !statuses[v.Status] {
		return fmt.Errorf("dispatch: unknown status %q", v.Status)
	}
	return nil
}
