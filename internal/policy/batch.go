package policy

import "fmt"

// BatchReleaseRule checks that a batch may be released: quantity must be
// positive, the batch must not be past its expiry date, and a batch under
// review must not be released.
func BatchReleaseRule(quantity int, expired bool, inReview bool) error {
	if quantity <= 0 {
		return fmt.Errorf("policy: batch quantity must be positive")
	}
	if expired {
		return fmt.Errorf("policy: batch already expired")
	}
	if inReview {
		return fmt.Errorf("policy: batch under review cannot be released")
	}
	return nil
}
