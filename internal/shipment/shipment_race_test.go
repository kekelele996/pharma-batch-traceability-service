package shipment

import (
	"sync"
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
)

func TestConcShipmentAppendRace(t *testing.T) {
	svc := NewService(platform.NewClock())
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 50; j++ {
				_, _ = svc.Append(Shipment{BatchID: "b1", NodeType: NodeOutbound, OccurredAt: time.Now()})
			}
		}()
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 50; j++ {
				_ = svc.ListByBatch("b1")
			}
		}()
	}
	close(start)
	wg.Wait()
}
