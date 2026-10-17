package gormscope_test

import (
	"testing"
	"time"

	"github.com/hanjo-erp/hanjo/pkg/gormscope"
)

func TestFilterLike(t *testing.T) {
	scope1 := gormscope.FilterLike("name", "john")
	if scope1 == nil {
		t.Fatal("expected non-nil scope for non-empty value")
	}

	scopeEmpty := gormscope.FilterLike("name", "")
	if scopeEmpty == nil {
		t.Fatal("expected non-nil scope function for empty value")
	}
}

func TestFilterExact(t *testing.T) {
	val := "active"
	scope1 := gormscope.FilterExact("status", &val)
	if scope1 == nil {
		t.Fatal("expected non-nil scope for non-nil value")
	}

	scopeNil := gormscope.FilterExact[string]("status", nil)
	if scopeNil == nil {
		t.Fatal("expected non-nil scope function for nil value")
	}
}

func TestFilterIn(t *testing.T) {
	ids := []string{"1", "2", "3"}
	scope1 := gormscope.FilterIn("id", ids)
	if scope1 == nil {
		t.Fatal("expected non-nil scope for non-empty slice")
	}

	scopeEmpty := gormscope.FilterIn("id", []string{})
	if scopeEmpty == nil {
		t.Fatal("expected non-nil scope function for empty slice")
	}
}

func TestFilterNotIn(t *testing.T) {
	ids := []int{1, 2}
	scope := gormscope.FilterNotIn("id", ids)
	if scope == nil {
		t.Fatal("expected non-nil scope")
	}
}

func TestFilterDateRange(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	scope := gormscope.FilterDateRange("created_at", &yesterday, &now)
	if scope == nil {
		t.Fatal("expected non-nil scope")
	}
}

func TestFilterActive(t *testing.T) {
	scopeDefault := gormscope.FilterActive()
	if scopeDefault == nil {
		t.Fatal("expected non-nil default active scope")
	}

	scopeCustom := gormscope.FilterActive("enabled")
	if scopeCustom == nil {
		t.Fatal("expected non-nil custom active scope")
	}
}
