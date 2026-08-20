package production

import (
	"fmt"
	"time"
)

const (
	StatusPending    = "pending"
	StatusQuarantine = "quarantined"
	StatusReleased   = "released"
	StatusRejected   = "rejected"
)

type Batch struct {
	ID             string    `json:"id"`
	DrugID         string    `json:"drug_id"`
	BatchNo        string    `json:"batch_no"`
	ProductionDate time.Time `json:"production_date"`
	ExpiryDate     time.Time `json:"expiry_date"`
	Quantity       int       `json:"quantity"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

func Validate(b Batch) error {
	if b.DrugID == "" {
		return fmt.Errorf("batch: drug id required")
	}
	if b.BatchNo == "" {
		return fmt.Errorf("batch: batch number required")
	}
	if b.ProductionDate.IsZero() || b.ExpiryDate.IsZero() {
		return fmt.Errorf("batch: production and expiry dates required")
	}
	if !b.ExpiryDate.After(b.ProductionDate) {
		return fmt.Errorf("batch: expiry must be after production")
	}
	if b.Quantity < 0 {
		return fmt.Errorf("batch: quantity must be non-negative")
	}
	if !validStatus(b.Status) {
		return fmt.Errorf("batch: unknown status %q", b.Status)
	}
	return nil
}

func validStatus(s string) bool {
	switch s {
	case StatusPending, StatusQuarantine, StatusReleased, StatusRejected:
		return true
	}
	return false
}

// AllowedTransitions maps a status to the set of reachable statuses.
var AllowedTransitions = map[string][]string{
	StatusPending:    {StatusQuarantine, StatusReleased, StatusRejected},
	StatusQuarantine: {StatusReleased, StatusRejected},
	StatusReleased:   {StatusQuarantine},
	StatusRejected:   {},
}
