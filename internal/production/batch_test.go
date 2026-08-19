package production

import (
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
)

func mk(t *testing.T) *Service {
	t.Helper()
	return NewService(platform.NewClock())
}

func createBatch(t *testing.T, svc *Service, no string) Batch {
	t.Helper()
	b, err := svc.Create(Batch{DrugID: "d1", BatchNo: no,
		ProductionDate: time.Now().AddDate(0, -1, 0),
		ExpiryDate:     time.Now().AddDate(0, 1, 0), Quantity: 100})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBatchHoldB21(t *testing.T) {
	svc := mk(t)
	b := createBatch(t, svc, "R1")
	if _, err := svc.Release(b.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Quarantine(b.ID); err != nil {
		t.Fatalf("released batch should be quarantinable: %v", err)
	}
	got, _ := svc.Get(b.ID)
	if got.Status != StatusQuarantine {
		t.Fatalf("expected quarantined, got %s", got.Status)
	}
}

func TestBatchUnblockB22(t *testing.T) {
	svc := mk(t)
	b := createBatch(t, svc, "Q1")
	if _, err := svc.Quarantine(b.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Release(b.ID); err != nil {
		t.Fatalf("quarantined batch should be releasable: %v", err)
	}
	got, _ := svc.Get(b.ID)
	if got.Status != StatusReleased {
		t.Fatalf("expected released, got %s", got.Status)
	}
}

func TestBatchDenyB23(t *testing.T) {
	svc := mk(t)
	b := createBatch(t, svc, "J1")
	if _, err := svc.Reject(b.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.Get(b.ID)
	if got.Status != StatusRejected {
		t.Fatalf("expected rejected, got %s", got.Status)
	}
}
