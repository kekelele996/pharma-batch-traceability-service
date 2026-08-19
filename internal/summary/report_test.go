package summary

import (
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/serialization"
	"pharma-batch-traceability-service/internal/stock"
)

func TestReportRejectedP33(t *testing.T) {
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)
	serials := serialization.NewService(clock)

	rel, _ := batches.Create(production.Batch{DrugID: "d1", BatchNo: "R1",
		ProductionDate: time.Now().AddDate(0, -1, 0), ExpiryDate: time.Now().AddDate(0, 1, 0), Quantity: 100})
	batches.Release(rel.ID)
	rej, _ := batches.Create(production.Batch{DrugID: "d1", BatchNo: "J1",
		ProductionDate: time.Now().AddDate(0, -1, 0), ExpiryDate: time.Now().AddDate(0, 1, 0), Quantity: 100})
	batches.Reject(rej.ID)

	sum := NewService(stocks, batches, serials).Summary(time.Now())
	if sum.ReleasedBatches != 1 {
		t.Fatalf("expected 1 released, got %d", sum.ReleasedBatches)
	}
	if sum.RejectedBatches != 1 {
		t.Fatalf("expected 1 rejected, got %d", sum.RejectedBatches)
	}
}
