package supplier

import (
	"fmt"
	"strings"
)

type Supplier struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreditCode string `json:"credit_code"`
	LicenseNo  string `json:"license_no"`
	Status     string `json:"status"`
}

var statuses = map[string]bool{"active": true, "suspended": true}

func Validate(s Supplier) error {
	s.Name = strings.TrimSpace(s.Name)
	s.CreditCode = strings.TrimSpace(s.CreditCode)
	if s.Name == "" {
		return fmt.Errorf("supplier: name required")
	}
	if len(s.CreditCode) != 18 {
		return fmt.Errorf("supplier: credit code must be 18 chars")
	}
	if !statuses[s.Status] {
		return fmt.Errorf("supplier: unknown status %q", s.Status)
	}
	return nil
}

func (s *Supplier) Normalize() {
	s.Name = strings.TrimSpace(s.Name)
	s.CreditCode = strings.TrimSpace(s.CreditCode)
	s.LicenseNo = strings.TrimSpace(s.LicenseNo)
	s.Status = strings.ToLower(strings.TrimSpace(s.Status))
}
