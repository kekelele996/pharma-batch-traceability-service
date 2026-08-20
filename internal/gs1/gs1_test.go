package gs1

import "testing"

// TestGTIN14EmptyP11 空 item reference 必须返回错误，不能返回空字符串当成功。
func TestGTIN14EmptyR018A(t *testing.T) {
	gtin, err := GTIN14("")
	if err == nil {
		t.Fatalf("expected an error for an empty item reference, got %q", gtin)
	}
}

// TestSSCCBadRefP12 坏扩展位或坏 reference 必须返回错误。
func TestSSCCBadRefR018B(t *testing.T) {
	if _, err := SSCC("x", "123"); err == nil {
		t.Fatal("expected an error for a bad extension digit")
	}
	if _, err := SSCC("1", "123"); err == nil {
		t.Fatal("expected an error for a short reference")
	}
}

// TestNormalizeBatchNoControlP13 含控制字符或空的批次号必须返回错误。
func TestNormalizeBatchNoControlR018C(t *testing.T) {
	if _, err := NormalizeBatchNo("BT-1\x01"); err == nil {
		t.Fatal("expected an error for a control character")
	}
	if _, err := NormalizeBatchNo("   "); err == nil {
		t.Fatal("expected an error for an empty batch number")
	}
}

// TestComputeCheckDigitInvalidP15 非法前缀必须返回错误，不能返回 0 当成功。
func TestComputeCheckDigitInvalidR018E(t *testing.T) {
	if _, err := ComputeCheckDigit("abc"); err == nil {
		t.Fatal("expected an error for a non-digit prefix")
	}
}

// TestValidGTIN14LenP16 长度不对的 GTIN 必须判为无效。
func TestValidGTIN14LenR018F(t *testing.T) {
	if ValidGTIN14("123") {
		t.Fatal("expected a 3-char string to be invalid")
	}
}
