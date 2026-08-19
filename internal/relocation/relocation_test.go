package relocation

import (
	"testing"

	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
)

func mk(t *testing.T) (*Service, *stock.Service) {
	t.Helper()
	clock := platform.NewClock()
	stocks := stock.NewService(clock)
	shipments := shipment.NewService(clock)
	return NewService(clock, stocks, shipments), stocks
}

func TestRelocationTransA1(t *testing.T) {
	if CanTransition(StatusCreated, StatusCompleted) {
		t.Fatal("created should not transition directly to completed")
	}
	if CanTransition(StatusIntransit, StatusIntransit) {
		t.Fatal("intransit should not transition to itself")
	}
	if CanTransition(StatusArrived, StatusIntransit) {
		t.Fatal("arrived should not transition backwards to intransit")
	}
}

func TestRelocationCompleteA2(t *testing.T) {
	svc, _ := mk(t)
	v, err := svc.Create(Dispatch{FromWarehouseID: "w1", ToWarehouseID: "w2",
		Items: []DispatchItem{{DrugID: "d1", BatchID: "b1", Qty: 10}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Complete(v.ID); err == nil {
		t.Fatal("expected completing a created dispatch to fail")
	} else if !platform.IsConflict(err) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestRelocationCancelA3(t *testing.T) {
	svc, stocks := mk(t)
	v, err := svc.Create(Dispatch{FromWarehouseID: "w1", ToWarehouseID: "w2",
		Items: []DispatchItem{{DrugID: "d1", BatchID: "b1", Qty: 10}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stocks.Receive("b1", "w1", 100); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Start(v.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Complete(v.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Cancel(v.ID); err == nil {
		t.Fatal("expected cancelling a completed dispatch to fail")
	}
}

func TestRelocationSameA4(t *testing.T) {
	svc, _ := mk(t)
	if _, err := svc.Create(Dispatch{FromWarehouseID: "w1", ToWarehouseID: "w1",
		Items: []DispatchItem{{DrugID: "d1", BatchID: "b1", Qty: 10}}}); err == nil {
		t.Fatal("expected creating a same-warehouse dispatch to fail")
	}
}
