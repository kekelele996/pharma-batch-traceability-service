package policy

import "fmt"

// BatchReleaseRule checks that a batch may be released: quantity must be
// positive and the batch must not be past its expiry date.
func BatchReleaseRule(quantity int, expired bool) error {
	if quantity <= 0 {
		return fmt.Errorf("policy: batch quantity must be positive")
	}
	if expired {
		return fmt.Errorf("policy: batch already expired")
	}
	return nil
}
