package manufacturer

import (
	"context"
	"testing"
)

func newTestMfr(i int) Manufacturer {
	return Manufacturer{Name: "mfr-" + string(rune('A'+i)), CreditCode: "913100001234567896",
		GMPNo: "GMP1", Status: "active"}
}

// TestManufacturerVerifyCancelP12 已取消的 context 必须立即中止批量核验。
func TestManufacturerVerifyCancelR015B(t *testing.T) {
	svc := NewService()
	ids := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		m, err := svc.Create(newTestMfr(i))
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, m.ID)
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
