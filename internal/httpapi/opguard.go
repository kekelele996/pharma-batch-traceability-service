package httpapi

import "sync"

// dispatchOps tracks dispatch ids that have an operation in flight.
var dispatchOps = map[string]bool{}

// releaseOps tracks batch ids that have a release operation in flight.
var releaseOps = map[string]bool{}

// opsMu guards all reads and writes to dispatchOps and releaseOps, which are
// mutated concurrently by HTTP handlers.
var opsMu sync.Mutex

func setOpFlag(m map[string]bool, id string) {
	opsMu.Lock()
	defer opsMu.Unlock()
	m[id] = true
}

func clearOpFlag(m map[string]bool, id string) {
	opsMu.Lock()
	defer opsMu.Unlock()
	delete(m, id)
}
