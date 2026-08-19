package stock

import (
	"sync"
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func TestConcDeductRace01(t *testing.T) {
	svc := NewService(platform.NewClock())
	if _, err := svc.Receive("b1", "wh1", 100); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Deduct("b1", "wh1", 20)
		}()
	}
	close(start)
	wg.Wait()
	row, err := svc.Get("b1", "wh1")
	if err != nil {
		t.Fatal(err)
	}
	if row.Quantity != 0 {
		t.Fatalf("expected 0 remaining after 10x20 deducts from 100, got %d", row.Quantity)
	}
}

func TestConcLockRace03(t *testing.T) {
	svc := NewService(platform.NewClock())
	if _, err := svc.Receive("b1", "wh1", 100); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Lock("b1", "wh1", 5)
		}()
	}
	close(start)
	wg.Wait()
	row, _ := svc.Get("b1", "wh1")
	if row.Locked != 50 {
		t.Fatalf("expected locked 50, got %d", row.Locked)
	}
}

func TestConcUnlockRace04(t *testing.T) {
	svc := NewService(platform.NewClock())
	if _, err := svc.Receive("b1", "wh1", 100); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Lock("b1", "wh1", 50); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Unlock("b1", "wh1", 5)
		}()
	}
	close(start)
	wg.Wait()
	row, _ := svc.Get("b1", "wh1")
	if row.Locked != 0 {
		t.Fatalf("expected locked 0, got %d", row.Locked)
	}
}




