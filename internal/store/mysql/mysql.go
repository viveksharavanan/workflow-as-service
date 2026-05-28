package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viveksharavanan/workflow-as-service/internal/store"
)

// DBTX is the common interface satisfied by both *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// MySQLStore implements store.Store backed by MySQL.
type MySQLStore struct {
	db       DBTX
	blobsDir string
}

// Compile-time check that MySQLStore implements store.Store.
var _ store.Store = (*MySQLStore)(nil)

// NewMySQLStore returns a new MySQLStore using the given database handle.
func NewMySQLStore(db *sql.DB, blobsDir string) *MySQLStore {
	return &MySQLStore{db: db, blobsDir: blobsDir}
}

// NewMySQLStoreFromTx returns a MySQLStore that operates within the
// given transaction.  This is useful when the caller manages the
// transaction externally (e.g. backward-compatible shims).
func NewMySQLStoreFromTx(tx *sql.Tx) *MySQLStore {
	return &MySQLStore{db: tx}
}

// WithTx executes fn within a database transaction. If the store is
// already operating inside a transaction, fn runs directly without
// nesting.
func (s *MySQLStore) WithTx(ctx context.Context, fn func(store.Store) error) error {
	// Already inside a transaction -- just call fn.
	if _, ok := s.db.(*sql.Tx); ok {
		return fn(s)
	}

	sqlDB, ok := s.db.(*sql.DB)
	if !ok {
		return fmt.Errorf("mysql: WithTx called on unsupported DBTX type %T", s.db)
	}

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txStore := &MySQLStore{db: tx, blobsDir: s.blobsDir}
	if err := fn(txStore); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
