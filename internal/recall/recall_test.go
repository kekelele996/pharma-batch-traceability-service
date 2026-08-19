package recall

import (
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/stock"
)

func TestRecallStopB25(t *testing.T) {
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)
	svc := NewService(clock, stocks)

	b, _ := batches.Create(production.Batch{DrugID: "d1", BatchNo: "R1",
		ProductionDate: time.Now().AddDate(0, -1, 0), ExpiryDate: time.Now().AddDate(0, 1, 0), Quantity: 100})
	batches.Release(b.ID)
	stocks.Receive(b.ID, "wh1", 100)
	stocks.Receive(b.ID, "wh2", 100)

	if _, err := svc.Issue(b.ID, 1, "contamination"); err != nil {
		t.Fatal(err)
	}
	for _, wh := range []string{"wh1", "wh2"} {
		row, err := stocks.Get(b.ID, wh)
		if err != nil {
			t.Fatal(err)
		}
		if !row.Frozen {
			t.Fatalf("warehouse %s not frozen", wh)
		}
	}
}
