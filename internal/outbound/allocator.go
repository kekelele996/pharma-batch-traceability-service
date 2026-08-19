package outbound

import (
	"sort"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/stock"
	"pharma-batch-traceability-service/internal/util"
)

// AllocationPlan maps a drug to the batches that will satisfy its demand,
// chosen first-expired-first-out among released batches with stock on hand.
type AllocationPlan struct {
	DrugID string
	Lines  []Allocation
}

// Plan allocates demand across candidate batches using FEFO ordering. It
// returns the plan and the remaining unsatisfied quantity.
func Plan(demand int, batches []production.Batch, stockSvc *stock.Service, warehouseID string) (AllocationPlan, int) {
	released := util.Filter(batches, func(b production.Batch) bool { return b.Status == production.StatusReleased })
	sort.SliceStable(released, func(i, j int) bool {
		if released[i].ExpiryDate.Equal(released[j].ExpiryDate) {
			return released[i].CreatedAt.Before(released[j].CreatedAt)
		}
		return released[i].ExpiryDate.Before(released[j].ExpiryDate)
	})
	plan := AllocationPlan{Lines: make([]Allocation, len(released))}
	remaining := demand
	idx := 0
	for _, b := range released {
		if remaining <= 0 {
			break
		}
		avail := stockSvc.TotalAvailable(b.ID)
		if warehouseID != "" {
			if row, err := stockSvc.Get(b.ID, warehouseID); err == nil {
				avail = row.Available()
			} else {
				avail = 0
			}
		}
		if avail <= 0 {
			continue
		}
		take := avail
		if take > remaining {
			take = remaining
		}
		plan.Lines[idx] = Allocation{BatchID: b.ID, Qty: take}
		idx++
		remaining -= take
	}
	return plan, remaining
}
