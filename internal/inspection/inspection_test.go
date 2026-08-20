package inspection

import (
	"errors"
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func TestInspectionDecideMissingR012B(t *testing.T) {
	svc := NewService(platform.NewClock())
	_, err := svc.Decide("missing-insp", "pass", []string{"ok"})
	if err == nil {
		t.Fatal("expected an error for a missing inspection")
	}
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("expected not-found sentinel, got %v", err)
	}
}

func TestInspectionCreateConflictR012F(t *testing.T) {
	svc := NewService(platform.NewClock())
	v := Inspection{ID: "insp-fixed", TargetType: "batch", TargetID: "b1", Result: "pending"}
	if _, err := svc.Create(v); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Create(v)
	if err == nil {
		t.Fatal("expected a conflict for a duplicate inspection id")
	}
	if !errors.Is(err, platform.ErrConflict) {
		t.Fatalf("expected conflict sentinel, got %v", err)
	}
}
