package policy

import (
	"fmt"

	"pharma-batch-traceability-service/internal/warehouse"
)

// StorageRule verifies that a warehouse temp zone can hold a drug with the
// given storage condition.
func StorageRule(drugStorage, zone string) error {
	if warehouse.TempZoneCompatible(drugStorage, zone) {
		return nil
	}
	return fmt.Errorf("policy: warehouse zone %q cannot store %q", zone, drugStorage)
}
