package util

import (
	"reflect"
	"testing"
)

func TestUtilFilterP31(t *testing.T) {
	in := []int{1, 2, 3, 4, 5}
	out := Filter(in, func(v int) bool { return v%2 == 0 })
	if !reflect.DeepEqual(in, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("Filter mutated input: %v", in)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 kept elements, got %d", len(out))
	}
}

func TestUtilDedupeP32(t *testing.T) {
	in := []string{"a", "b", "a", "c"}
	out := Dedupe(in)
	if !reflect.DeepEqual(in, []string{"a", "b", "a", "c"}) {
		t.Fatalf("Dedupe mutated input: %v", in)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 unique elements, got %d", len(out))
	}
}
