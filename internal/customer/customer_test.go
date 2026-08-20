package customer

import (
	"errors"
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func newTestCustomer(i int) Customer {
	return Customer{Name: "cust-" + string(rune('A'+i)), CreditCode: "913100001234567890",
		LicenseNo: "L", Status: "active"}
}

func mustCreate(t *testing.T, svc *Service, c Customer) string {
	t.Helper()
	v, err := svc.Create(c)
	if err != nil {
		t.Fatal(err)
	}
	return v.ID
}

// TestCustomerSuspendRollbackP11 批量停用中途失败时，已停用的客户必须回滚。
func TestCustomerSuspendRollbackR016A(t *testing.T) {
	svc := NewService()
	g1 := mustCreate(t, svc, newTestCustomer(0))
	_, err := svc.SuspendMany([]string{g1, "missing-cust"})
	if err == nil {
		t.Fatal("expected an error for a missing customer")
	}
	got, _ := svc.Get(g1)
	if got.Status != "active" {
		t.Fatalf("customer should be rolled back to active, got %q", got.Status)
	}
}

// TestCustomerSuspendErrorKeptP12 部分成功时原始 not-found 错误不能被吞掉。
func TestCustomerSuspendErrorKeptR016B(t *testing.T) {
	svc := NewService()
	g1 := mustCreate(t, svc, newTestCustomer(1))
	_, err := svc.SuspendMany([]string{g1, "missing-cust"})
	if err == nil {
		t.Fatal("expected the not-found error to be preserved")
	}
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

// TestCustomerSuspendStopsP13 遇到第一个失败后必须停止处理后续 id。
func TestCustomerSuspendStopsR016C(t *testing.T) {
	svc := NewService()
	g1 := mustCreate(t, svc, newTestCustomer(2))
	g2 := mustCreate(t, svc, newTestCustomer(3))
	_, _ = svc.SuspendMany([]string{g1, "missing-cust", g2})
	got2, _ := svc.Get(g2)
	if got2.Status != "active" {
		t.Fatalf("processing should stop at the first failure, g2 got %q", got2.Status)
	}
}
