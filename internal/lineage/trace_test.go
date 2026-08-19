package lineage

import (
	"testing"

	"pharma-batch-traceability-service/internal/production"
	"pharma-batch-traceability-service/internal/platform"
	"pharma-batch-traceability-service/internal/serialization"
	"pharma-batch-traceability-service/internal/shipment"
	"pharma-batch-traceability-service/internal/stock"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	clock := platform.NewClock()
	return NewService(
		production.NewService(clock),
		serialization.NewService(clock),
		shipment.NewService(clock),
		stock.NewService(clock),
	)
}

func TestResolveByCodeMissingP11(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.ResolveByCode("0000000000000000000")
	if err == nil {
		t.Fatal("expected an error for a missing serial code")
	}
	if !platform.IsNotFound(err) {
		t.Fatalf("expected a not-found error, got %T: %v", err, err)
	}
}

func TestResolveByBatchMissingP12(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.ResolveByBatch("missing-batch")
	if err == nil {
		t.Fatal("expected an error for a missing batch")
	}
	if !platform.IsNotFound(err) {
		t.Fatalf("expected a not-found error, got %T: %v", err, err)
	}
}
