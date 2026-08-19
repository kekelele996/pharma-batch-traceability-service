package po

import (
	"fmt"
	"time"
)

const (
	StatusDraft    = "draft"
	StatusApproved = "approved"
	StatusOrdered  = "ordered"
	StatusReceived = "received"
	StatusCancelled = "cancelled"
)

type POItem struct {
	DrugID     string `json:"drug_id"`
	Qty        int    `json:"qty"`
	PriceCents int64  `json:"price_cents"`
}

type PurchaseOrder struct {
	ID         string    `json:"id"`
	No         string    `json:"no"`
	SupplierID string    `json:"supplier_id"`
	Items      []POItem  `json:"items"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

var statuses = map[string]bool{
	StatusDraft: true, StatusApproved: true, StatusOrdered: true,
	StatusReceived: true, StatusCancelled: true,
}

func Validate(v PurchaseOrder) error {
	if v.SupplierID == "" {
		return fmt.Errorf("po: supplier required")
	}
	if len(v.Items) == 0 {
		return fmt.Errorf("po: at least one item required")
	}
	seen := map[string]bool{}
	for i, it := range v.Items {
		if it.DrugID == "" || it.Qty <= 0 {
			return fmt.Errorf("po: invalid item %d", i)
		}
		if it.PriceCents < 0 {
			return fmt.Errorf("po: negative price at item %d", i)
		}
		if seen[it.DrugID] {
			return fmt.Errorf("po: duplicate item %d", i)
		}
		seen[it.DrugID] = true
	}
	if !statuses[v.Status] {
		return fmt.Errorf("po: unknown status %q", v.Status)
	}
	return nil
}

// CanTransition reports whether a purchase order may move from one status to
// another. The lifecycle is draft -> approved -> ordered -> received; either
// draft, approved or ordered may be cancelled, but received and cancelled are
// terminal.
func CanTransition(from, to string) bool {
	switch from {
	case StatusDraft:
		return to == StatusApproved || to == StatusCancelled
	case StatusApproved:
		return to == StatusOrdered || to == StatusCancelled
	case StatusOrdered:
		return to == StatusReceived || to == StatusCancelled
	default:
		return false
	}
}

// TotalCents returns the order total in cents.
func (v PurchaseOrder) TotalCents() int64 {
	var total int64
	for _, it := range v.Items {
		total += it.PriceCents * int64(it.Qty)
	}
	return total
}
