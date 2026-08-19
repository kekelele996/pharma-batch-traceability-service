package recall

import (
	"fmt"
	"time"
)

const (
	StatusIssued    = "issued"
	StatusExecuting = "executing"
	StatusCompleted = "completed"
)

type Recall struct {
	ID        string    `json:"id"`
	BatchID   string    `json:"batch_id"`
	Level     int       `json:"level"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

var statuses = map[string]bool{StatusIssued: true, StatusExecuting: true, StatusCompleted: true}

func Validate(v Recall) error {
	if v.BatchID == "" {
		return fmt.Errorf("recall: batch id required")
	}
	if v.Level < 1 || v.Level > 3 {
		return fmt.Errorf("recall: level must be 1..3")
	}
	if v.Reason == "" {
		return fmt.Errorf("recall: reason required")
	}
	if !statuses[v.Status] {
		return fmt.Errorf("recall: unknown status %q", v.Status)
	}
	return nil
}
