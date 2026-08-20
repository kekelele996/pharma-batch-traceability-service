package quality

import (
	"errors"
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func TestQualityDecideMissingR012A(t *testing.T) {
	svc := NewService(platform.NewClock())
	_, err := svc.Decide("missing-qc", "pass", "insp", "note")
	if err == nil {
		t.Fatal("expected an error for a missing quality record")
	}
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("expected not-found sentinel, got %v", err)
	}
}

func TestQualityDecideUnknownResultR012C(t *testing.T) {
	svc := NewService(platform.NewClock())
	r, err := svc.Create(Record{BatchID: "b1", SampleSize: 1, Result: "pending"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Decide(r.ID, "bogus", "insp", "x")
	if err == nil {
		t.Fatal("expected an error for an unknown result")
	}
	if !errors.Is(err, platform.ErrValidation) {
		t.Fatalf("expected validation sentinel, got %v", err)
	}
}

func TestQualityCreateConflictR012E(t *testing.T) {
	svc := NewService(platform.NewClock())
	r := Record{ID: "qc-fixed", BatchID: "b1", SampleSize: 1, Result: "pending"}
	if _, err := svc.Create(r); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Create(r)
	if err == nil {
		t.Fatal("expected a conflict for a duplicate id")
	}
	if !errors.Is(err, platform.ErrConflict) {
		t.Fatalf("expected conflict sentinel, got %v", err)
	}
}
