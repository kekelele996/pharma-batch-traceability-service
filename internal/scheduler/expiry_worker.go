package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"pharma-batch-traceability-service/internal/expiry"
)

// ExpiryWorker periodically scans for expired batches and freezes their stock
// so they can no longer be shipped. It is driven by a cancellable context and
// stops cleanly when the context is done.
type ExpiryWorker struct {
	interval time.Duration
	service  *expiry.Service
	wg       sync.WaitGroup
}

func NewExpiryWorker(interval time.Duration, svc *expiry.Service) *ExpiryWorker {
	if interval <= 0 {
		interval = time.Minute
	}
	return &ExpiryWorker{interval: interval, service: svc}
}

// Start launches the scan loop in the background and returns immediately.
func (w *ExpiryWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if ctx.Err() != nil {
					return
				}
				if _, err := w.service.LockExpired(ctx, time.Now().UTC()); err != nil {
					log.Printf("expiry worker: lock expired: %v", err)
				}
			}
		}
	}()
}

// Wait blocks until the scan loop has exited.
func (w *ExpiryWorker) Wait() { w.wg.Wait() }
