package inbound

import (
	"testing"
	"time"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/stock"
)

func TestInboundRollbackP81(t *testing.T) {
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)
	svc := NewService(clock, batches, stocks)

	b1, _ := batches.Create(production.Batch{DrugID: "d1", BatchNo: "R1",
		ProductionDate: time.Now().AddDate(0, -1, 0), ExpiryDate: time.Now().AddDate(0, 1, 0), Quantity: 100})
	batches.Release(b1.ID)
	b2, _ := batches.Create(production.Batch{DrugID: "d1", BatchNo: "P1",
		ProductionDate: time.Now().AddDate(0, -1, 0), ExpiryDate: time.Now().AddDate(0, 1, 0), Quantity: 100})

	in, err := svc.Create(Inbound{SupplierID: "s1", WarehouseID: "wh1",
		Items: []InboundItem{{DrugID: "d1", BatchID: b1.ID, Qty: 10}, {DrugID: "d1", BatchID: b2.ID, Qty: 10}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Accept(in.ID, "passed"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Putaway(in.ID); err == nil {
		t.Fatal("expected putaway to fail on an unreleased batch")
	}
	row, err := stocks.Get(b1.ID, "wh1")
	if err != nil {
		t.Fatal(err)
	}
	if row.Quantity != 0 {
		t.Fatalf("expected rolled back stock 0, got %d", row.Quantity)
	}
}

func TestInboundMissingP82(t *testing.T) {
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)
	svc := NewService(clock, batches, stocks)

	in, err := svc.Create(Inbound{SupplierID: "s1", WarehouseID: "wh1",
		Items: []InboundItem{{DrugID: "d1", BatchID: "missing", Qty: 10}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Accept(in.ID, "passed"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Putaway(in.ID); err == nil {
		t.Fatal("expected putaway to fail on a missing batch")
	}
}


func TestInboundCreateValidateP83(t *testing.T) {
	clock := platform.NewClock()
	batches := production.NewService(clock)
	stocks := stock.NewService(clock)
	svc := NewService(clock, batches, stocks)

	if _, err := svc.Create(Inbound{SupplierID: "", WarehouseID: "wh1",
		Items: []InboundItem{{DrugID: "d1", BatchID: "b1", Qty: 10}}}); err == nil {
		t.Fatal("expected creating an inbound without supplier to fail")
	}
}
