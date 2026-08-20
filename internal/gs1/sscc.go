package gs1

import (
	"fmt"
	"strings"
)

// SSCC encodes a Serial Shipping Container Code from an extension digit and a
// 16-digit company/serial reference. The returned 18-digit string ends with the
// mod-10 check digit.
func SSCC(extensionDigit string, reference string) (string, error) {
	if len(extensionDigit) != 1 || extensionDigit < "0" || extensionDigit > "9" {
		return "", fmt.Errorf("gs1: extension digit must be a single digit 0-9")
	}
	if len(reference) != 16 {
		return "", fmt.Errorf("gs1: reference must be 16 digits, got %d", len(reference))
	}
	body := extensionDigit + reference
	cd, err := ComputeCheckDigit(body)
	if err != nil {
		return "", err
	}
	return body + fmt.Sprintf("%d", cd), nil
}

// NormalizeBatchNo trims and upper-cases a batch/lot number, rejecting empty
// or control-character heavy values.
func NormalizeBatchNo(in string) (string, error) {
	v := strings.ToUpper(strings.TrimSpace(in))
	if v == "" {
		return "", fmt.Errorf("gs1: batch number required")
	}
	for _, r := range v {
		if r < 32 || r > 126 {
			return "", fmt.Errorf("gs1: batch number contains control characters")
		}
	}
	return v, nil
}
