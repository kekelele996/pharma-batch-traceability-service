package stock

import (
	"fmt"
	"time"
)

type Stock struct {
	BatchID     string    `json:"batch_id"`
	WarehouseID string    `json:"warehouse_id"`
	Quantity    int       `json:"quantity"`
	Locked      int       `json:"locked"`
	Frozen      bool      `json:"frozen"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func Key(batchID, warehouseID string) string {
	return batchID + "/" + warehouseID
}

// Available returns the quantity that can actually be moved out: on-hand minus
// locked, and zero when the batch is frozen.
func (s Stock) Available() int {
	if s.Frozen {
		return 0
	}
	avail := s.Quantity - s.Locked
	if avail < 0 {
		return 0
	}
	return avail
}

func (s Stock) Validate() error {
	if s.BatchID == "" || s.WarehouseID == "" {
		return fmt.Errorf("stock: batch and warehouse required")
	}
	if s.Quantity < 0 || s.Locked < 0 {
		return fmt.Errorf("stock: quantity and locked must be non-negative")
	}
	if s.Locked > s.Quantity {
		return fmt.Errorf("stock: locked cannot exceed quantity")
	}
	return nil
}
