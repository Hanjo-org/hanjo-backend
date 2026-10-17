package trx

import (
	"context"

	"gorm.io/gorm"
)

type txContextKey struct{}

var activeTxKey = txContextKey{}

// Manager defines the contract for executing operations within an atomic database transaction.
// Application usecases depend on this interface without importing concrete database drivers.
type Manager interface {
	Execute(ctx context.Context, fn func(txCtx context.Context) error) error
}

type gormTxManager struct {
	db *gorm.DB
}

// NewGormManager creates a GORM-backed Transaction Manager implementation.
func NewGormManager(db *gorm.DB) Manager {
	return &gormTxManager{db: db}
}

// Execute wraps fn in a GORM transaction.
// If an active transaction is already present in ctx, it participates in the existing transaction.
func (m *gormTxManager) Execute(ctx context.Context, fn func(txCtx context.Context) error) error {
	// If already in a transaction, reuse it
	if _, ok := ctx.Value(activeTxKey).(*gorm.DB); ok {
		return fn(ctx)
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, activeTxKey, tx)
		return fn(txCtx)
	})
}

// GetDB retrieves the active transaction *gorm.DB from ctx if present,
// otherwise returns fallbackDB.
// Repositories should use this helper to transparently support atomic transactions.
func GetDB(ctx context.Context, fallbackDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(activeTxKey).(*gorm.DB); ok && tx != nil {
		if ctx != nil && tx.Config != nil {
			return tx.WithContext(ctx)
		}
		return tx
	}
	if fallbackDB != nil && ctx != nil && fallbackDB.Config != nil {
		return fallbackDB.WithContext(ctx)
	}
	return fallbackDB
}

// InTransaction checks whether the provided context is currently executing inside an active transaction.
func InTransaction(ctx context.Context) bool {
	tx, ok := ctx.Value(activeTxKey).(*gorm.DB)
	return ok && tx != nil
}
