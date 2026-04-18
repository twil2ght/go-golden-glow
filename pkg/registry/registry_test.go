package registry

import (
	"reflect"
	"testing"
)

func TestDefaultRegistry_RegisterGet(t *testing.T) {
	r := New[int]()

	r.Register("first", 1)
	r.Register("second", 2)

	if got := r.Len(); got != 2 {
		t.Fatalf("expected length 2, got %d", got)
	}

	keys := r.Keys()
	expectedKeys := []string{"first", "second"}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Fatalf("expected keys %v, got %v", expectedKeys, keys)
	}

	value, err := r.Get("first")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 1 {
		t.Fatalf("expected value 1, got %d", value)
	}

	value, err = r.Get("second")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 2 {
		t.Fatalf("expected value 2, got %d", value)
	}
}

func TestDefaultRegistry_RegisterDuplicateIgnored(t *testing.T) {
	r := New[string]()
	r.Register("alpha", "one")
	r.Register("alpha", "two")

	if got := r.Len(); got != 1 {
		t.Fatalf("expected length 1 after duplicate register, got %d", got)
	}

	value, err := r.Get("alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "one" {
		t.Fatalf("expected stored value to remain 'one', got %q", value)
	}
}

func TestDefaultRegistry_Unregister(t *testing.T) {
	r := New[int]()
	r.Register("a", 10)
	r.Register("b", 20)

	r.Unregister("a")

	if got := r.Len(); got != 1 {
		t.Fatalf("expected length 1 after unregister, got %d", got)
	}

	keys := r.Keys()
	expectedKeys := []string{"b"}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Fatalf("expected keys %v, got %v", expectedKeys, keys)
	}

	if _, err := r.Get("a"); err == nil {
		t.Fatal("expected error for missing key after unregister")
	}
}

func TestDefaultRegistry_GetMissing(t *testing.T) {
	r := New[float32]()
	_, err := r.Get("missing")
	if err == nil {
		t.Fatal("expected error when getting missing key")
	}
}

func TestDefaultRegistry_Range(t *testing.T) {
	r := New[int]()
	r.Register("x", 1)
	r.Register("y", 2)
	r.Register("z", 3)

	seen := make([]int, 0, 3)
	r.Range(func(value int) bool {
		seen = append(seen, value)
		return value != 2
	})

	expectedSeen := []int{1, 2}
	if !reflect.DeepEqual(seen, expectedSeen) {
		t.Fatalf("expected seen values %v, got %v", expectedSeen, seen)
	}
}
