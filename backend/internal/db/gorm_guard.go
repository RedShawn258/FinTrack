package db

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FORBIDDEN_GORM_TABLES lists table names that MUST NOT be accessed via GORM APIs.
// These tables use reserved keywords or require explicit SQL control for correctness.
// Using GORM APIs (db.Model, db.Where, db.Create, etc.) on these tables can cause
// GORM to generate incorrect SQL queries. See GORM_ELIMINATION_SUMMARY.md.
var FORBIDDEN_GORM_TABLES = []string{
	"idempotency_keys",
	"journal_entries",
	"journal_lines",
	"outbox_events",
}

// CheckForbiddenGORMUsage checks if a GORM query is attempting to access a forbidden table.
// This is a runtime guard to detect accidental use of GORM APIs on correctness-critical tables.
// Call this function in development/testing mode to catch violations early.
func CheckForbiddenGORMUsage(db *gorm.DB, logger *zap.Logger) error {
	// Extract table name from GORM statement if available
	if db.Statement != nil && db.Statement.Table != "" {
		tableName := db.Statement.Table
		for _, forbidden := range FORBIDDEN_GORM_TABLES {
			if tableName == forbidden {
				logger.Error("FORBIDDEN GORM USAGE DETECTED",
					zap.String("table", tableName),
					zap.String("reason", "This table must use pure database/sql via repository layer"),
					zap.String("solution", "Use repositories.LedgerRepository instead of GORM APIs"),
				)
				return fmt.Errorf(
					"PRODUCTION GUARDRAIL VIOLATION: Table '%s' MUST NOT be accessed via GORM APIs. "+
						"Use repositories.LedgerRepository with pure database/sql instead. "+
						"See ARCHITECTURE.md and GORM_ELIMINATION_SUMMARY.md",
					tableName,
				)
			}
		}
	}
	return nil
}

// NOTE: This is a passive check that must be called explicitly.
// In a production system, you might use GORM hooks or middleware to automatically
// check all queries, but that adds overhead. For now, this serves as documentation
// and can be called manually in test code or during code review.
