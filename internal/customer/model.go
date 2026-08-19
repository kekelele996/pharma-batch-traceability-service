package customer

import (
	"fmt"
	"strings"
)

type Customer struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreditCode string `json:"credit_code"`
	LicenseNo  string `json:"license_no"`
	RxPermit   bool   `json:"rx_permit"`
	Status     string `json:"status"`
}

var statuses = map[string]bool{"active": true, "suspended": true}

func Validate(c Customer) error {
	c.Name = strings.TrimSpace(c.Name)
	c.CreditCode = strings.TrimSpace(c.CreditCode)
	if c.Name == "" {
		return fmt.Errorf("customer: name required")
	}
	if len(c.CreditCode) != 18 {
		return fmt.Errorf("customer: credit code must be 18 chars")
	}
	if !statuses[c.Status] {
		return fmt.Errorf("customer: unknown status %q", c.Status)
	}
	return nil
}

func (c *Customer) CanSellRx() bool {
	return c.Status == "active" && c.RxPermit
}

func (c *Customer) Normalize() {
	c.Name = strings.TrimSpace(c.Name)
	c.CreditCode = strings.TrimSpace(c.CreditCode)
	c.LicenseNo = strings.TrimSpace(c.LicenseNo)
	c.Status = strings.ToLower(strings.TrimSpace(c.Status))
}
