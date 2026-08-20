package audit

import (
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

// TestAuditSnapshotStableP15 调用方修改 List 返回值不能污染内部审计记录。
func TestAuditSnapshotStableR011E(t *testing.T) {
	svc := NewService(platform.NewClock())
	svc.Append("alice", "create", "drug-1", "first")
	svc.Append("bob", "release", "batch-1", "second")
	got := svc.List(0)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	got[0].Actor = "tampered"
	fresh := svc.List(0)
	if fresh[0].Actor == "tampered" {
		t.Fatal("caller mutation leaked into the audit store")
	}
}

// TestAuditTargetFilterStableP17 按目标过滤后内部记录不能被改写或丢失。
func TestAuditTargetFilterStableR011G(t *testing.T) {
	svc := NewService(platform.NewClock())
	svc.Append("a1", "act", "x", "d1")
	svc.Append("b1", "act", "y", "d2")
	svc.Append("c1", "act", "y", "d3")
	got := svc.ListByTarget("y")
	if len(got) != 2 {
		t.Fatalf("expected 2 entries for y, got %d", len(got))
	}
	all := svc.ListByTarget("")
	if len(all) != 3 {
		t.Fatalf("expected 3 entries total, got %d", len(all))
	}
	found := false
	for _, e := range all {
		if e.Actor == "a1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("entry a1 was lost after a filtered read")
	}
}
