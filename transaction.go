package database

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// TransactionFunc is a function that runs within a transaction.
type TransactionFunc func(*gorm.DB) error

// WithTransaction runs a function within a transaction.
// Automatically commits on success, rolls back on error.
func WithTransaction(db *gorm.DB, fn TransactionFunc) error {
	return db.Transaction(fn)
}

// WithTransactionContext runs a function within a transaction with context.
func WithTransactionContext(ctx context.Context, db *gorm.DB, fn TransactionFunc) error {
	return db.WithContext(ctx).Transaction(fn)
}

// BeginTransaction starts a new transaction.
func BeginTransaction(db *gorm.DB) *gorm.DB {
	return db.Begin()
}

// BeginTransactionWithOptions starts a transaction with options.
func BeginTransactionWithOptions(db *gorm.DB, opts *sql.TxOptions) *gorm.DB {
	return db.Begin(opts)
}

// Commit commits a transaction.
func Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

// Rollback rolls back a transaction.
func Rollback(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// SavePoint creates a savepoint.
func SavePoint(tx *gorm.DB, name string) error {
	return tx.SavePoint(name).Error
}

// RollbackTo rolls back to a savepoint.
func RollbackTo(tx *gorm.DB, name string) error {
	return tx.RollbackTo(name).Error
}

// TransactionHelper provides transaction helper methods.
type TransactionHelper struct {
	db *gorm.DB
}

// NewTransactionHelper creates a new transaction helper.
func NewTransactionHelper(db *gorm.DB) *TransactionHelper {
	return &TransactionHelper{db: db}
}

// Run runs a function within a transaction.
func (h *TransactionHelper) Run(fn TransactionFunc) error {
	return WithTransaction(h.db, fn)
}

// RunWithContext runs a function within a transaction with context.
func (h *TransactionHelper) RunWithContext(ctx context.Context, fn TransactionFunc) error {
	return WithTransactionContext(ctx, h.db, fn)
}

// Begin starts a new transaction.
func (h *TransactionHelper) Begin() *gorm.DB {
	return BeginTransaction(h.db)
}

// BeginWithOptions starts a transaction with options.
func (h *TransactionHelper) BeginWithOptions(opts *sql.TxOptions) *gorm.DB {
	return BeginTransactionWithOptions(h.db, opts)
}

// Manager transaction methods

// TX returns a transaction helper for the manager.
func (m *Manager) TX() *TransactionHelper {
	return NewTransactionHelper(m.db)
}

// WithTx runs a function within a transaction using the manager.
func (m *Manager) WithTx(fn TransactionFunc) error {
	return WithTransaction(m.db, fn)
}

// WithTxContext runs a function within a transaction with context.
func (m *Manager) WithTxContext(ctx context.Context, fn TransactionFunc) error {
	return WithTransactionContext(ctx, m.db, fn)
}
