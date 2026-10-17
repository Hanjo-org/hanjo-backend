package ptr_test

import (
	"testing"

	"github.com/hanjo-erp/hanjo/pkg/ptr"
)

func TestPtr(t *testing.T) {
	s := ptr.Ptr("hello")
	if *s != "hello" {
		t.Errorf("expected hello, got %s", *s)
	}

	n := ptr.Ptr(42)
	if *n != 42 {
		t.Errorf("expected 42, got %d", *n)
	}
}

func TestValOrDefault(t *testing.T) {
	var s *string
	if val := ptr.ValOrDefault(s, "default"); val != "default" {
		t.Errorf("expected default, got %s", val)
	}

	val := "custom"
	s = &val
	if res := ptr.ValOrDefault(s, "default"); res != "custom" {
		t.Errorf("expected custom, got %s", res)
	}
}

func TestVal(t *testing.T) {
	var num *int
	if val := ptr.Val(num); val != 0 {
		t.Errorf("expected 0, got %d", val)
	}

	fortyTwo := 42
	num = &fortyTwo
	if val := ptr.Val(num); val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestEmptySlice(t *testing.T) {
	var nilSlice []string
	res := ptr.EmptySlice(nilSlice)
	if res == nil {
		t.Error("expected non-nil slice")
	}
	if len(res) != 0 {
		t.Errorf("expected 0 length, got %d", len(res))
	}

	existing := []string{"a", "b"}
	if len(ptr.EmptySlice(existing)) != 2 {
		t.Error("expected unchanged existing slice")
	}
}

func TestEnsureNonNilSlice(t *testing.T) {
	var nilSlice []int
	res := ptr.EnsureNonNilSlice(nilSlice)
	if res == nil {
		t.Error("expected non-nil slice from EnsureNonNilSlice")
	}

	type TestStruct struct {
		Name string
	}
	item := TestStruct{Name: "Go"}
	if ptr.EnsureNonNilSlice(item) != item {
		t.Error("expected non-slice to remain unchanged")
	}
}

func TestEqual(t *testing.T) {
	a := ptr.Ptr("same")
	b := ptr.Ptr("same")
	c := ptr.Ptr("different")

	if !ptr.Equal(a, b) {
		t.Error("expected a and b to be equal")
	}
	if ptr.Equal(a, c) {
		t.Error("expected a and c to be unequal")
	}
	if !ptr.Equal[string](nil, nil) {
		t.Error("expected nil and nil to be equal")
	}
	if ptr.Equal(a, nil) {
		t.Error("expected a and nil to be unequal")
	}
}
