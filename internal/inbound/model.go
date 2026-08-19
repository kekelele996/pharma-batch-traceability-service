package inbound

import (
	"fmt"
	"time"
)

const (
	StatusDraft    = "draft"
	StatusAccepted = "accepted"
	StatusPutaway  = "putaway"
	StatusRejected = "rejected"
)

type InboundItem struct {
	DrugID  string `json:"drug_id"`
	BatchID string `json:"batch_id"`
	Qty     int    `json:"qty"`
}

type Inbound struct {
	ID          string        `json:"id"`
	No          string        `json:"no"`
	SupplierID  string        `json:"supplier_id"`
	WarehouseID string        `json:"warehouse_id"`
	Items       []InboundItem `json:"items"`
	Status      string        `json:"status"`
	QCResult    string        `json:"qc_result"`
	CreatedAt   time.Time     `json:"created_at"`
}

var statuses = map[string]bool{StatusDraft: true, StatusAccepted: true, StatusPutaway: true, StatusRejected: true}

func Validate(v Inbound) error {
	if v.SupplierID == "" || v.WarehouseID == "" {
		return fmt.Errorf("inbound: supplier and warehouse required")
	}
	if len(v.Items) == 0 {
		return fmt.Errorf("inbound: at least one item required")
	}
	seen := map[string]bool{}
	for i, it := range v.Items {
		if it.DrugID == "" || it.BatchID == "" || it.Qty <= 0 {
			return fmt.Errorf("inbound: invalid item %d", i)
		}
		k := it.DrugID + "/" + it.BatchID
		if seen[k] {
			return fmt.Errorf("inbound: duplicate item %d", i)
		}
		seen[k] = true
	}
	if !statuses[v.Status] {
		return fmt.Errorf("inbound: unknown status %q", v.Status)
	}
	return nil
}
