package trx_test

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/pkg/trx"
)

func TestGetDB_Fallback(t *testing.T) {
	ctx := context.Background()
	dummyFallback := &gorm.DB{}

	db := trx.GetDB(ctx, dummyFallback)
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	if trx.InTransaction(ctx) {
		t.Error("expected false for InTransaction on clean context")
	}
}

func TestNewGormManager(t *testing.T) {
	dummyDB := &gorm.DB{}
	mgr := trx.NewGormManager(dummyDB)
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}
