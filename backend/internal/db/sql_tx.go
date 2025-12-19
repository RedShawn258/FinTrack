package db

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// MustSQLTx extracts the underlying *sql.Tx from a GORM transaction.
// This is used to bypass GORM's query builder for critical operations
// that require explicit SQL (e.g., idempotency, double-entry accounting).
//
// Returns an error if:
//   - The connection pool is not a *sql.Tx (transaction not active)
//   - The connection pool is nil
//
// Usage:
//
//	err := db.DB.Transaction(func(tx *gorm.DB) error {
//	    sqlTx, err := MustSQLTx(tx)
//	    if err != nil {
//	        return err
//	    }
//	    // Use sqlTx for explicit SQL operations
//	    sqlTx.Exec("INSERT INTO ...")
//	    return nil
//	})
func MustSQLTx(tx *gorm.DB) (*sql.Tx, error) {
	if tx == nil {
		return nil, fmt.Errorf("gorm.DB is nil")
	}

	if tx.Statement == nil {
		return nil, fmt.Errorf("gorm.DB.Statement is nil - transaction may not be active")
	}

	connPool := tx.Statement.ConnPool
	if connPool == nil {
		return nil, fmt.Errorf("gorm.DB.Statement.ConnPool is nil - transaction may not be active")
	}

	// Type assertion to *sql.Tx
	sqlTx, ok := connPool.(*sql.Tx)
	if !ok {
		return nil, fmt.Errorf("expected *sql.Tx but got %T - ensure you are inside a db.DB.Transaction() callback", connPool)
	}

	return sqlTx, nil
}
