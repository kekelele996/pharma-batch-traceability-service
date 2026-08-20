package outbound

import (
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
)

func newOutboundService(t *testing.T) (*Service, string, string) {
	t.Helper()
	clock := platform.NewClock()
	b := production.NewService(clock)
	st := stock.NewService(clock)
	sh := shipment.NewService(clock)
	svc := NewService(clock, b, st, sh)
	bat, err := b.Create(production.Batch{
		DrugID: "drug-1", BatchNo: "BT-001",
		ProductionDate: time.Now().Add(-24 * time.Hour),
		ExpiryDate:     time.Now().Add(720 * time.Hour),
		Quantity:       1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.Release(bat.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Receive(bat.ID, "wh-1", 100); err != nil {
		t.Fatal(err)
	}
	return svc, bat.ID, "wh-1"
}

// TestAllocateNoScratchAliasP11 连续分配两张出库单，第一张的分配明细不能被第二张覆盖。
func TestAllocateNoScratchAliasR014A(t *testing.T) {
	svc, _, wh := newOutboundService(t)
	mk := func(qty int) string {
		v, err := svc.Create(Outbound{CustomerID: "c1", WarehouseID: wh,
			Items: []OutboundItem{{DrugID: "drug-1", Qty: qty}}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Allocate(v.ID); err != nil {
			t.Fatal(err)
		}
		return v.ID
	}
	idA := mk(10)
	idB := mk(20)
	_ = idB
	gotA, err := svc.Get(idA)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotA.Allocations) != 1 || gotA.Allocations[0].Qty != 10 {
		t.Fatalf("order A allocations corrupted after allocating order B: %+v", gotA.Allocations)
	}
}

// TestCreateItemsIsolatedP12 创建单据后调用方修改自己的输入切片不能串改单据内容。
func TestCreateItemsIsolatedR014B(t *testing.T) {
	svc, _, wh := newOutboundService(t)
	items := []OutboundItem{{DrugID: "drug-1", Qty: 10}}
	v, err := svc.Create(Outbound{CustomerID: "c1", WarehouseID: wh, Items: items})
	if err != nil {
		t.Fatal(err)
	}
	items[0].Qty = 999
	got, err := svc.Get(v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Items[0].Qty != 10 {
		t.Fatalf("stored items were mutated by the caller: got qty %d", got.Items[0].Qty)
	}
}

// TestValidateNoMutationP13 Validate 不能原地改写调用方传入的 items 切片。
func TestValidateNoMutationR014C(t *testing.T) {
	items := []OutboundItem{{DrugID: "drug-b", Qty: 2}, {DrugID: "drug-a", Qty: 3}}
	v := Outbound{Status: StatusDraft, CustomerID: "c1", WarehouseID: "wh-1", Items: items}
	if err := Validate(v); err != nil {
		t.Fatal(err)
	}
	if items[0].DrugID != "drug-b" {
		t.Fatalf("Validate mutated the caller slice: %+v", items)
	}
}

// TestLastAllocationsSnapshotP15 拿到的最近分配快照不能被后续分配改写。
func TestLastAllocationsSnapshotR014E(t *testing.T) {
	svc, _, wh := newOutboundService(t)
	mk := func(qty int) string {
		v, err := svc.Create(Outbound{CustomerID: "c1", WarehouseID: wh,
			Items: []OutboundItem{{DrugID: "drug-1", Qty: qty}}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Allocate(v.ID); err != nil {
			t.Fatal(err)
		}
		return v.ID
	}
	_ = mk(10)
	snap := svc.LastAllocations()
	_ = mk(20)
	if len(snap) != 1 || snap[0].Qty != 10 {
		t.Fatalf("allocations snapshot was overwritten by a later allocation: %+v", snap)
	}
}
