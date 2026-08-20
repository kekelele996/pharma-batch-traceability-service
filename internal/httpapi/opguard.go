package httpapi

// dispatchOps tracks dispatch ids that have an operation in flight.
var dispatchOps = map[string]bool{}

// releaseOps tracks batch ids that have a release operation in flight.
var releaseOps = map[string]bool{}
