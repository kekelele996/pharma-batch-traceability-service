package policy

import "fmt"

// ColdChainRule checks that a shipment temperature stayed within the range
// required by the drug storage condition. Cold and frozen chains have strict
// limits; room/cool tolerate a broader window.
func ColdChainRule(storage string, tempC float64) error {
	switch storage {
	case "frozen":
		if tempC > -18 {
			return fmt.Errorf("policy: frozen chain breached at %.1fC", tempC)
		}
	case "cold":
		if tempC < 2 || tempC > 8 {
			return fmt.Errorf("policy: cold chain breached at %.1fC", tempC)
		}
	case "cool":
		if tempC > 20 {
			return fmt.Errorf("policy: cool chain breached at %.1fC", tempC)
		}
	}
	return nil
}
