package warehouse

import (
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func newTestWH(code, zone string) Warehouse {
	return Warehouse{Code: code, Name: "wh-" + code, TempZone: zone, Status: "active"}
}

// TestWarehouseSetStatusNoPanicR013A 首次更新仓库状态不能因 nil map 写入而 panic。
func TestWarehouseSetStatusNoPanicR013A(t *testing.T) {
	svc := NewService()
	w, err := svc.Create(newTestWH("WH-1", "cold"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetStatus(w.ID, "closed"); err != nil {
		t.Fatal(err)
	}
}

// TestWarehouseListByStatusR013K 按状态查询能拿到该状态的仓库。
func TestWarehouseListByStatusR013K(t *testing.T) {
	svc := NewService()
	w, err := svc.Create(newTestWH("WH-2", "frozen"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetStatus(w.ID, "closed"); err != nil {
		t.Fatal(err)
	}
	got := svc.ListByStatus("closed")
	if len(got) != 1 || got[0].ID != w.ID {
		t.Fatalf("expected [%s] for closed status, got %+v", w.ID, got)
	}
}

// TestWarehouseLatestTypedNilR013E 空温区查询不能返回“看似有效”的 nil 指针。
func TestWarehouseLatestTypedNilR013E(t *testing.T) {
	svc := NewService()
	_, ok := svc.LatestByZone("tropical")
	if ok {
		t.Fatal("expected ok=false for a zone with no warehouses")
	}
}

// TestWarehouseDefaultZoneR013F 未指定温区的仓库默认落到 room 温区。
func TestWarehouseDefaultZoneR013F(t *testing.T) {
	svc := NewService()
	w, err := svc.Create(newTestWH("WH-3", ""))
	if err != nil {
		t.Fatal(err)
	}
	if w.TempZone != "room" {
		t.Fatalf("expected default zone room, got %q", w.TempZone)
	}
}

// TestWarehouseGetMissingR013G 查询不存在的仓库必须返回错误，不能零值当成功。
func TestWarehouseGetMissingR013G(t *testing.T) {
	svc := NewService()
	_, err := svc.Get("missing-wh")
	if err == nil {
		t.Fatal("expected an error for a missing warehouse")
	}
	if !platform.IsNotFound(err) {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

// TestWarehouseTempZoneGuardR013J 未知温区对避光药品也不兼容。
func TestWarehouseTempZoneGuardR013J(t *testing.T) {
	if TempZoneCompatible("light-proof", "bogus-zone") {
		t.Fatal("unknown zone must not be compatible with light-proof storage")
	}
}
