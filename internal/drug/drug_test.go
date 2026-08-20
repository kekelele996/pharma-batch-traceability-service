package drug

import (
	"testing"

	"pharma-batch-traceability-service/internal/platform"
)

func TestDrugStorageForMissingR013C(t *testing.T) {
	svc := NewService(platform.NewClock())
	_, err := svc.StorageFor("missing-drug")
	if err == nil {
		t.Fatal("expected an error for a missing drug")
	}
	if !platform.IsNotFound(err) {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

// TestDrugCreateInvalidR013I 非法药品必须返回校验错误，不能零值当成功。
func TestDrugCreateInvalidR013I(t *testing.T) {
	svc := NewService(platform.NewClock())
	_, err := svc.Create(Drug{Code: "H20240001", GenericName: "g", DosageForm: "tablet",
		RxCategory: "rx", Storage: "bogus"})
	if err == nil {
		t.Fatal("expected a validation error for an invalid drug")
	}
	if !platform.IsValidation(err) {
		t.Fatalf("expected validation sentinel, got %v", err)
	}
}
