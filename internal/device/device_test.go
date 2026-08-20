package device

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
)

func newTestDevice(i int) Device {
	return Device{
		Code: fmt.Sprintf("TH-%d", i), Name: "thermo",
		Kind: "thermometer", Status: "active",
		CalibrationDue: time.Now().Add(24 * time.Hour),
	}
}

// TestDeviceConcurrentListP13 并发校准/故障标记设备的同时读取列表，不能出现 data race。
func TestDeviceConcurrentListR011C(t *testing.T) {
	svc := NewService(platform.NewClock())
	ids := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		d, err := svc.Create(newTestDevice(i))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		ids = append(ids, d.ID)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			id := ids[i%len(ids)]
			for j := 0; j < 40; j++ {
				if j%2 == 0 {
					_, _ = svc.Calibrate(id, time.Now().Add(time.Hour))
				} else {
					_, _ = svc.MarkFaulty(id)
				}
			}
		}(i)
	}
	for i := 0; i < 4; i++ {
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

// TestDeviceConcurrentDueP14 并发校准设备的同时读取校准到期清单，不能出现 data race。
func TestDeviceConcurrentDueR011D(t *testing.T) {
	svc := NewService(platform.NewClock())
	ids := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		d, err := svc.Create(newTestDevice(i))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		ids = append(ids, d.ID)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			id := ids[i%len(ids)]
			for j := 0; j < 40; j++ {
				_, _ = svc.Calibrate(id, time.Now().Add(time.Duration(j%3)*time.Hour))
			}
		}(i)
	}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 60; j++ {
				_ = svc.CalibrationDue(time.Now().Add(2 * time.Hour))
			}
		}()
	}
	close(start)
	wg.Wait()
}
