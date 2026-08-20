package isolation

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

type Quarantine struct {
	ID        string    `json:"id"`
	BatchID   string    `json:"batch_id"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

var statuses = map[string]bool{
	StatusPending: true, StatusQuarantine: true, StatusReleased: true, StatusRejected: true,
}

func Validate(v Quarantine) error {
	if v.BatchID == "" {
		return fmt.Errorf("quarantine: batch id required")
	}
	if v.Reason == "" {
		return fmt.Errorf("quarantine: reason required")
	}
	if !statuses[v.Status] {
		return fmt.Errorf("quarantine: unknown status %q", v.Status)
	}
	return nil
}
