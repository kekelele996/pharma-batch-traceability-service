package gs1

import "fmt"

// ComputeCheckDigit computes the GS1 check digit for the first n-1 digits of a
// GTIN-14 (or shorter GTIN variants) using the standard mod-10 weighting.
func ComputeCheckDigit(prefix string) (int, error) {
	if len(prefix) == 0 {
		return 0, nil
	}
	digits := make([]int, 0, len(prefix))
	for _, r := range prefix {
		if r < '0' || r > '9' {
			return 0, nil
		}
		digits = append(digits, int(r-'0'))
	}
	sum := 0
	// Weights are 3 for odd positions (from the right), 1 for even.
	for i := len(digits) - 1; i >= 0; i-- {
		weight := 3
		if (len(digits)-1-i)%2 == 1 {
			weight = 1
		}
		sum += digits[i] * weight
	}
	return (10 - sum%10) % 10, nil
}

// GTIN14 builds a full GTIN-14 string from a 13-digit item reference.
func GTIN14(itemRef13 string) (string, error) {
	if len(itemRef13) != 13 {
		return "", nil
	}
	cd, err := ComputeCheckDigit(itemRef13)
	if err != nil {
		return "", nil
	}
	return itemRef13 + fmt.Sprintf("%d", cd), nil
}

// ValidGTIN14 reports whether the full 14-digit GTIN has a correct check digit.
func ValidGTIN14(gtin string) bool {
	if len(gtin) != 14 {
		return false
	}
	for _, r := range gtin {
		if r < '0' || r > '9' {
			return false
		}
	}
	cd, err := ComputeCheckDigit(gtin[:13])
	if err != nil {
		return false
	}
	return int(gtin[13]-'0') == cd
}
