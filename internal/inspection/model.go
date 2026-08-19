package inspection

import (
	"fmt"
	"time"
)

const (
	ResultPending = "pending"
	ResultPass    = "pass"
	ResultFail    = "fail"
)

type Inspection struct {
	ID          string    `json:"id"`
	TargetType  string    `json:"target_type"`
	TargetID    string    `json:"target_id"`
	Inspector   string    `json:"inspector"`
	Result      string    `json:"result"`
	Findings    []string  `json:"findings"`
	InspectedAt time.Time `json:"inspected_at"`
}

var targetTypes = map[string]bool{"manufacturer": true, "warehouse": true, "batch": true}
var results = map[string]bool{ResultPending: true, ResultPass: true, ResultFail: true}

func Validate(v Inspection) error {
	if !targetTypes[v.TargetType] {
		return fmt.Errorf("inspection: unknown target type %q", v.TargetType)
	}
	if v.TargetID == "" {
		return fmt.Errorf("inspection: target id required")
	}
	if !results[v.Result] {
		return fmt.Errorf("inspection: unknown result %q", v.Result)
	}
	return nil
}
