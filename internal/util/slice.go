package util

// Filter returns the elements of in for which keep returns true. It allocates
// a fresh backing array so callers that still hold the original slice are not
// affected by later mutation of the result.
func Filter[T any](in []T, keep func(T) bool) []T {
	out := make([]T, 0, len(in))
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Transform maps every element of in through f, returning a new slice.
func Transform[T, U any](in []T, f func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, f(v))
	}
	return out
}

// Dedupe returns the first occurrence of each distinct element, preserving
// input order.
func Dedupe[T comparable](in []T) []T {
	seen := make(map[T]struct{}, len(in))
	out := make([]T, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// Contains reports whether needle appears in in.
func Contains[T comparable](in []T, needle T) bool {
	for _, v := range in {
		if v == needle {
			return true
		}
	}
	return false
}
