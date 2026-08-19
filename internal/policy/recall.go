package policy

import "fmt"

// RecallLevelRule maps a risk score to a recall level (1..3), where a higher
// score demands a more severe recall. Scores below the floor are rejected.
func RecallLevelRule(score int) (int, error) {
	if score < 0 {
		return 0, fmt.Errorf("policy: risk score cannot be negative")
	}
	switch {
	case score >= 80:
		return 1, nil
	case score >= 50:
		return 2, nil
	default:
		return 3, nil
	}
}
