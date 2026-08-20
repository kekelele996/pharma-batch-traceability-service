package policy

import "fmt"

// RxSaleRule checks whether a customer is allowed to buy a drug of the given
// prescription category. Prescription drugs require an active customer with an
// rx permit, and a batch under review suspends the sale.
func RxSaleRule(drugRxCategory string, customerActive, customerRxPermit bool, inReview bool) error {
	if drugRxCategory != "rx" {
		return nil
	}
	if inReview {
		return fmt.Errorf("policy: rx sale suspended while batch under review")
	}
	if !customerActive {
		return fmt.Errorf("policy: customer is not active")
	}
	if !customerRxPermit {
		return fmt.Errorf("policy: customer lacks rx permit for prescription drug")
	}
	return nil
}
