package httpapi

import "sync"

// dispatchOps tracks dispatch ids that have an operation in flight.
var dispatchOps = map[string]bool{}

// releaseOps tracks batch ids that have a release operation in flight.
var releaseOps = map[string]bool{}

var (
	dispatchMu sync.Mutex
	releaseMu  sync.Mutex
)

// dispatchTryStart atomically claims the in-flight slot for a dispatch id.
func dispatchTryStart(id string) bool {
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	if dispatchOps[id] {
		return false
	}
	dispatchOps[id] = true
	return true
}

// dispatchRelease atomically releases an existing in-flight slot.
func dispatchRelease(id string) bool {
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	if !dispatchOps[id] {
		return false
	}
	delete(dispatchOps, id)
	return true
}

// dispatchRestore re-claims the slot after a failed operation.
func dispatchRestore(id string) {
	dispatchMu.Lock()
	defer dispatchMu.Unlock()
	dispatchOps[id] = true
}

// releaseTryStart atomically claims the in-flight slot for a batch id.
func releaseTryStart(id string) bool {
	releaseMu.Lock()
	defer releaseMu.Unlock()
	if releaseOps[id] {
		return false
	}
	releaseOps[id] = true
	return true
}

// releaseFinish atomically releases the in-flight slot for a batch id.
func releaseFinish(id string) {
	releaseMu.Lock()
	defer releaseMu.Unlock()
	delete(releaseOps, id)
}
