package isolation

import (
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/stock"
)

func newFixture(t *testing.T) (*Service, string) {
	t.Helper()
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)
	svc := NewService(clock, batches, stocks)
	b, err := batches.Create(production.Batch{
		DrugID: "d1", BatchNo: "BT001",
		ProductionDate: time.Now().AddDate(0, -1, 0),
		ExpiryDate:     time.Now().AddDate(0, 1, 0),
		Quantity:       100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := batches.Release(b.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := stocks.Receive(b.ID, "wh1", 100); err != nil {
		t.Fatal(err)
	}
	return svc, b.ID
}

func TestIsolationDenyB24(t *testing.T) {
	svc, batchID := newFixture(t)
	q, err := svc.Place(batchID, "quality hold")
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Reject(q.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusRejected {
		t.Fatalf("expected rejected, got %s", got.Status)
	}
}
