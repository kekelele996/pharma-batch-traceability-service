package supplier

import (
	"context"
	"errors"
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func newTestSupplier(i int) Supplier {
	return Supplier{Name: "sup-" + string(rune('A'+i)), CreditCode: "913100001234567890",
		LicenseNo: "L", Status: "active"}
}

// TestSupplierVerifyCancelP11 已取消的 context 必须立即中止批量核验。
func TestSupplierVerifyCancelR015A(t *testing.T) {
	svc := NewService()
	ids := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		v, err := svc.Create(newTestSupplier(i))
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, v.ID)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	verified, err := svc.VerifyAll(ctx, ids)
	if verified != 0 {
		t.Fatalf("expected 0 verified after cancel, got %d", verified)
	}
	if err == nil {
		t.Fatal("expected a cancellation error")
	}
}

// TestSupplierVerifyMissingP13 核验列表中存在不存在的供应商时必须返回错误。
func TestSupplierVerifyMissingR015C(t *testing.T) {
	svc := NewService()
	v, err := svc.Create(newTestSupplier(0))
	if err != nil {
		t.Fatal(err)
	}
	verified, err := svc.VerifyAll(context.Background(), []string{v.ID, "missing-sup"})
	if err == nil {
		t.Fatalf("expected an error, got verified=%d", verified)
	}
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("expected not-found error, got %v", err)
	}
}
