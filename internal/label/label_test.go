package label

import (
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/production"
)

func mkSvc(t *testing.T) (*Service, string) {
	t.Helper()
	clock := platform.NewClock()
	batches := production.NewService(clock)
	svc := NewService(clock, batches)
	b, err := batches.Create(production.Batch{DrugID: "d1", BatchNo: "B1",
		ProductionDate: time.Now().AddDate(0, -1, 0), ExpiryDate: time.Now().AddDate(0, 1, 0), Quantity: 100})
	if err != nil {
		t.Fatal(err)
	}
	batches.Release(b.ID)
	return svc, b.ID
}

func TestLabelGenerateP90(t *testing.T) {
	svc, id := mkSvc(t)
	out, err := svc.Generate(id, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 5 {
		t.Fatalf("expected 5 labels, got %d", len(out))
	}
}

func TestLabelListP91(t *testing.T) {
	svc, id := mkSvc(t)
	if _, err := svc.Generate(id, 3); err != nil {
		t.Fatal(err)
	}
	list := svc.ListByBatch(id)
	if len(list) != 3 {
		t.Fatalf("expected 3 labels, got %d", len(list))
	}
	for i, l := range list {
		if l.ID == "" {
			t.Fatalf("empty label at index %d", i)
		}
	}
}

func TestLabelCloneP92(t *testing.T) {
	in := []Label{{ID: "a"}, {ID: "b"}}
	out := Clone(in)
	if len(out) != len(in) {
		t.Fatalf("clone length %d != input length %d", len(out), len(in))
	}
	if in[0].ID != "a" || in[1].ID != "b" {
		t.Fatalf("input mutated: %v", in)
	}
}

func TestLabelFilterP93(t *testing.T) {
	in := []Label{{ID: "a", Status: StatusUnprinted}, {ID: "b", Status: StatusPrinted}}
	out := FilterByStatus(in, StatusPrinted)
	if len(out) != 1 {
		t.Fatalf("expected 1 printed label, got %d", len(out))
	}
	if in[0].ID != "a" || in[1].ID != "b" {
		t.Fatalf("input mutated: %v", in)
	}
}
