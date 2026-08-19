package metric

import "testing"

func TestPbtcIncDimNoPanic(t *testing.T) {
	svc := NewService()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("IncDim panicked: %v", r)
		}
	}()
	svc.IncDim("api", "hits", 1)
	if svc.Dims()["api"]["hits"] != 1 {
		t.Fatal("expected dim value 1")
	}
}

func TestPbtcDimsIsCopy(t *testing.T) {
	svc := NewService()
	svc.IncDim("api", "hits", 1)
	d := svc.Dims()
	d["api"]["hits"] = 999
	if svc.Dims()["api"]["hits"] != 1 {
		t.Fatalf("internal dims mutated, got %d", svc.Dims()["api"]["hits"])
	}
}

func TestPbtcSnapshotIsCopy(t *testing.T) {
	svc := NewService()
	svc.Inc("counter", 5)
	snap := svc.Snapshot()
	snap["counter"] = 999
	if svc.Get("counter") != 5 {
		t.Fatalf("internal counters mutated, got %d", svc.Get("counter"))
	}
}
