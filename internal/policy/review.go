package policy

import "fmt"

// ReviewState is the intermediate state a quality record enters while a batch
// is being re-checked.
const ReviewState = "review"

var reviewStatuses = map[string]bool{
	"pending": true,
	"pass":    true,
	"fail":    true,
}

var reviewTransitions = map[string][]string{
	"pending": {"pass", "fail"},
}

var reviewLabels = map[string]string{
	"pending": "待检",
	"pass":    "合格",
	"fail":    "不合格",
}

// ValidReviewState reports whether a result state may be assigned to a record.
func ValidReviewState(s string) bool {
	return reviewStatuses[s]
}

// ReviewStateLabel returns the human readable label of a review state.
func ReviewStateLabel(s string) string {
	return reviewLabels[s]
}

// CanReviewTransition reports whether a record may move from one result state
// to another.
func CanReviewTransition(from, to string) (bool, error) {
	if !reviewStatuses[from] {
		return false, fmt.Errorf("policy: unknown review state %q", from)
	}
	for _, t := range reviewTransitions[from] {
		if t == to {
			return true, nil
		}
	}
	return false, nil
}
