package database

import (
	"context"
	"database/sql"
	"fmt"
)

// txKey is the key type for the context transaction
type txKey struct{}

// TransactionManager defines the interface for database transaction management
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

// PostgresTransactionManager implements TransactionManager for PostgreSQL
type PostgresTransactionManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new TransactionManager instance
func NewTransactionManager(db *sql.DB) TransactionManager {
	return &PostgresTransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// If the function returns nil, the transaction is committed
func (tm *PostgresTransactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	// Check if already in a transaction
	if GetTx(ctx) != nil {
		return fn(ctx)
	}

	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-panic after rollback
		}
	}()

	// Inject transaction into context
	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctxWithTx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %v (rollback failed: %v)", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTx returns the transaction from the context if it exists, nil otherwise
func GetTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}

