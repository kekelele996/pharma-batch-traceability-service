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
	out := in[:0]
	for _, l := range in {
		if l.Status == status {
			out = append(out, l)
		}
	}
	return out
}

// Clone returns a deep copy of the label slice.
func Clone(in []Label) []Label {
	out := in[:0]
	out = append(out, in...)
	out = append(out, in...)
	return out
}
