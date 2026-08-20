package manufacturer

import (
	"fmt"
	"strings"
)

type Manufacturer struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreditCode string `json:"credit_code"`
	GMPNo      string `json:"gmp_no"`
	Scope      string `json:"scope"`
	Status     string `json:"status"`
}

var statuses = map[string]bool{"active": true, "suspended": true, "revoked": true}

func Validate(m Manufacturer) error {
	m.Name = strings.TrimSpace(m.Name)
	m.CreditCode = strings.TrimSpace(m.CreditCode)
	m.GMPNo = strings.TrimSpace(m.GMPNo)
	if m.Name == "" {
		return fmt.Errorf("manufacturer: name required")
	}
	if !validCreditCode(m.CreditCode) {
		return fmt.Errorf("manufacturer: invalid unified credit code %q", m.CreditCode)
	}
	if !statuses[m.Status] {
		return fmt.Errorf("manufacturer: unknown status %q", m.Status)
	}
	return nil
}

// validCreditCode validates an 18-character Chinese unified social credit code:
// 18 chars, alphanumeric, with a mod-31 checksum over the first 17 chars.
func validCreditCode(code string) bool {
	if len(code) != 18 {
		return false
	}
	weights := []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}
	charset := "0123456789ABCDEFGHJKLMNPQRTUWXY"
	val := func(r byte) int {
		for i := 0; i < len(charset); i++ {
			if charset[i] == r {
				return i
			}
		}
		return -1
	}
	sum := 0
	for i := 0; i < 17; i++ {
		v := val(code[i])
		if v < 0 {
			return false
		}
		sum += v * weights[i]
	}
	mod := sum % 31
	check := code[17]
	if mod == 0 {
		return check == '0'
	}
	idx := (31 - mod) % 31
	if idx == 0 {
		return check == '0'
	}
	return charset[idx] == check
}

func (m *Manufacturer) Normalize() {
	m.Name = strings.TrimSpace(m.Name)
	m.CreditCode = strings.TrimSpace(m.CreditCode)
	m.GMPNo = strings.TrimSpace(m.GMPNo)
	m.Scope = strings.TrimSpace(m.Scope)
	m.Status = strings.ToLower(strings.TrimSpace(m.Status))
}
