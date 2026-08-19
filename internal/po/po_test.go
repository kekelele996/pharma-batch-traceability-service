package po

import (
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func mk(t *testing.T) *Service {
	t.Helper()
	return NewService(platform.NewClock())
}

func mkPo(t *testing.T, svc *Service) PurchaseOrder {
	t.Helper()
	v, err := svc.Create(PurchaseOrder{SupplierID: "s1",
		Items: []POItem{{DrugID: "d1", Qty: 1, PriceCents: 100}}})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestPoTransP71(t *testing.T) {
	if CanTransition(StatusDraft, StatusReceived) {
		t.Fatal("draft should not transition directly to received")
	}
	if CanTransition(StatusOrdered, StatusApproved) {
		t.Fatal("ordered should not transition backwards to approved")
	}
	if CanTransition(StatusReceived, StatusCancelled) {
		t.Fatal("received should be final")
	}
}

func TestPoOrderP72(t *testing.T) {
	svc := mk(t)
	v := mkPo(t, svc)
	if _, err := svc.Order(v.ID); err == nil {
		t.Fatal("expected ordering a draft po to fail")
	}
}

func TestPoReceiP73(t *testing.T) {
	svc := mk(t)
	v := mkPo(t, svc)
	if _, err := svc.Receive(v.ID); err == nil {
		t.Fatal("expected receiving a draft po to fail")
	}
}

func TestPoCanceP74(t *testing.T) {
	svc := mk(t)
	v := mkPo(t, svc)
	svc.Approve(v.ID)
	svc.Order(v.ID)
	svc.Receive(v.ID)
	if _, err := svc.Cancel(v.ID); err == nil {
		t.Fatal("expected cancelling a received po to fail")
	}
}
