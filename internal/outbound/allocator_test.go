package outbound

import (
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/stock"
)

func TestPlanLinesP34(t *testing.T) {
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)

	a, _ := batches.Create(production.Batch{DrugID: "d1", BatchNo: "A1",
		ProductionDate: time.Now().AddDate(0, -1, 0), ExpiryDate: time.Now().AddDate(0, 1, 0), Quantity: 100})
	batches.Release(a.ID)
	stocks.Receive(a.ID, "wh1", 100)

	b, _ := batches.Create(production.Batch{DrugID: "d1", BatchNo: "B1",
		ProductionDate: time.Now().AddDate(0, -1, 0), ExpiryDate: time.Now().AddDate(0, 1, 0), Quantity: 0})
	batches.Release(b.ID)

	plan, _ := Plan(50, batches.ListByDrug("d1"), stocks, "wh1")
	for i, line := range plan.Lines {
		if line.BatchID == "" {
			t.Fatalf("empty allocation at index %d", i)
		}
	}
	if len(plan.Lines) == 0 {
		t.Fatal("expected at least one allocation")
	}
}
