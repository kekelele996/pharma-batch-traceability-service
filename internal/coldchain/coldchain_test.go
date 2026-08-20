package coldchain

import (
	"sync"
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
)

// TestColdchainConcurrentMinP11 并发写温度记录的同时读取最值，不能出现 data race。
func TestColdchainConcurrentMinR011A(t *testing.T) {
	svc := NewService(platform.NewClock())
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 60; j++ {
				_, _ = svc.Record("ship-1", "dev-1", "cold", 4.0, 50.0, time.Now())
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 60; j++ {
				_, _ = svc.Min("ship-1")
			}
		}()
	}
	close(start)
	wg.Wait()
}

// TestColdchainConcurrentListP12 并发写温度记录的同时读取按时间排序的列表，不能出现 data race。
func TestColdchainConcurrentListR011B(t *testing.T) {
	svc := NewService(platform.NewClock())
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 60; j++ {
				_, _ = svc.Record("ship-2", "dev-2", "cold", 5.0, 55.0, time.Now())
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 60; j++ {
				_ = svc.ListByShipment("ship-2")
			}
		}()
	}
	close(start)
	wg.Wait()
}

// TestColdchainConcurrentListAllP16 并发写温度记录的同时读取全量列表，不能出现 data race。
func TestColdchainConcurrentListAllR011F(t *testing.T) {
	svc := NewService(platform.NewClock())
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 60; j++ {
				_, _ = svc.Record("ship-3", "dev-3", "cold", 6.0, 52.0, time.Now())
			}
		}()
	}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 60; j++ {
				_ = svc.List()
			}
		}()
	}
	close(start)
	wg.Wait()
}
