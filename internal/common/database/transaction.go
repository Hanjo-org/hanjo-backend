package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxManager manages database transactions using the Unit of Work pattern.
// It ensures that all repository operations within a use case share the same transaction.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager creates a new transaction manager.
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// ctxKey is a custom type for context keys to avoid collisions.
type ctxKey string

const txKey ctxKey = "db_tx"

// WithTransaction executes the given function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// If the function succeeds, the transaction is committed.
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Store the transaction in context
	txCtx := context.WithValue(ctx, txKey, tx)

	// Execute the function
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit on success
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTx extracts the transaction from context. If no transaction is found,
// it returns nil (caller should use the pool directly for read-only queries).
func GetTx(ctx context.Context) pgx.Tx {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return nil
	}
	return tx
}

// DBExecutor is an interface that both pgxpool.Pool and pgx.Tx implement.
type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// GetExecutor returns the transaction from context if available, otherwise the pool.
func GetExecutor(ctx context.Context, pool *pgxpool.Pool) DBExecutor {
	if tx := GetTx(ctx); tx != nil {
		return tx
	}
	return pool
}

