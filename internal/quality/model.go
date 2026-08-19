package quality

import (
	"fmt"
	"time"
)

const (
	ResultPending = "pending"
	ResultPass    = "pass"
	ResultFail    = "fail"
)

type Record struct {
	ID          string    `json:"id"`
	InboundID   string    `json:"inbound_id"`
	BatchID     string    `json:"batch_id"`
	SampleSize  int       `json:"sample_size"`
	Result      string    `json:"result"`
	Inspector   string    `json:"inspector"`
	Note        string    `json:"note"`
	InspectedAt time.Time `json:"inspected_at"`
}

var results = map[string]bool{ResultPending: true, ResultPass: true, ResultFail: true}

func Validate(r Record) error {
	if r.BatchID == "" {
		return fmt.Errorf("quality: batch id required")
	}
	if r.SampleSize < 0 {
		return fmt.Errorf("quality: sample size must be non-negative")
	}
	if !results[r.Result] {
		return fmt.Errorf("quality: unknown result %q", r.Result)
	}
	return nil
}
