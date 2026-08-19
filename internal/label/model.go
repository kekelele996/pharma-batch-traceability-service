package label

import "time"

const (
	StatusUnprinted = "unprinted"
	StatusPrinted   = "printed"
)

type Label struct {
	ID        string    `json:"id"`
	BatchID   string    `json:"batch_id"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	Qty       int       `json:"qty"`
	CreatedAt time.Time `json:"created_at"`
}

// FilterByStatus returns labels whose status matches the given status.
func FilterByStatus(in []Label, status string) []Label {
	out := make([]Label, 0, len(in))
	for _, l := range in {
		if l.Status == status {
			out = append(out, l)
		}
	}
	return out
}

// Clone returns a deep copy of the label slice.
func Clone(in []Label) []Label {
	out := make([]Label, len(in))
	copy(out, in)
	return out
}
