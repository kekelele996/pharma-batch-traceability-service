package scheduler

import (
	"context"
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/expiry"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/stock"
)

func TestCtxWorkerStops03(t *testing.T) {
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)
	expiries := expiry.NewService(batches, stocks)
	w := NewExpiryWorker(5*time.Millisecond, expiries)
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)
	time.Sleep(20 * time.Millisecond)
	cancel()
	done := make(chan struct{})
	go func() { w.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(800 * time.Millisecond):
		t.Fatal("worker did not stop after cancel")
	}
}
