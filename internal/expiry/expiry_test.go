package expiry

import (
	"context"
	"strings"
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/stock"
)

func TestCtxLockAborts04(t *testing.T) {
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)
	svc := NewService(batches, stocks)

	b, _ := batches.Create(production.Batch{DrugID: "d1", BatchNo: "E1",
		ProductionDate: time.Now().AddDate(0, -2, 0), ExpiryDate: time.Now().AddDate(0, -1, 0), Quantity: 10})
	batches.Release(b.ID)
	stocks.Receive(b.ID, "wh1", 10)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := svc.LockExpired(ctx, time.Now())
	if err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("expected aborted error, got %v", err)
	}
}
