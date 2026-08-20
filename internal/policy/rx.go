package policy

import "fmt"

// RxSaleRule checks whether a customer is allowed to buy a drug of the given
// prescription category. Prescription drugs require an active customer with an
// rx permit.
func RxSaleRule(drugRxCategory string, customerActive, customerRxPermit bool) error {
	if drugRxCategory != "rx" {
		return nil
	}
	if !customerActive {
		return fmt.Errorf("policy: customer is not active")
	}
	if !customerRxPermit {
		return fmt.Errorf("policy: customer lacks rx permit for prescription drug")
	}
	return nil
}
