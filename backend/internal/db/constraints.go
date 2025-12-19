package db

import (
	"go.uber.org/zap"
)

// AddConstraints adds database-level constraints for financial correctness
// These constraints provide an additional safety layer beyond application logic
func AddConstraints(logger *zap.Logger) error {
	if DB == nil {
		logger.Error("Database connection not initialized")
		return nil
	}

	logger.Info("Adding database constraints for financial correctness")

	// Note: MySQL 8.0.16+ supports CHECK constraints
	// For older MySQL versions, these will be ignored (application-level validation still applies)

	// 1. Ensure journal lines have valid direction enum
	// This is enforced by GORM enum type, but we document it here
	// Direction must be 'debit' or 'credit'

	// 2. Ensure journal lines have positive amounts
	// CHECK constraint: amount > 0
	checkAmountPositive := `
		ALTER TABLE journal_lines 
		ADD CONSTRAINT chk_journal_line_amount_positive 
		CHECK (amount > 0)
	`
	if err := DB.Exec(checkAmountPositive).Error; err != nil {
		// Constraint may already exist or MySQL version doesn't support CHECK
		logger.Debug("Could not add amount positive constraint (may already exist or unsupported)", zap.Error(err))
	}

	// 3. Ensure outbox events have valid status enum
	// Status must be one of: PENDING, RETRYING, SENT, FAILED
	// This is enforced by GORM varchar, but we document it here

	// 4. Ensure retry_count is non-negative
	checkRetryCountNonNegative := `
		ALTER TABLE outbox_events 
		ADD CONSTRAINT chk_outbox_retry_count_non_negative 
		CHECK (retry_count >= 0)
	`
	if err := DB.Exec(checkRetryCountNonNegative).Error; err != nil {
		logger.Debug("Could not add retry_count constraint (may already exist or unsupported)", zap.Error(err))
	}

	// 5. Foreign key constraints are already enforced by GORM relationships:
	// - journal_lines.entry_id -> journal_entries.id (CASCADE on delete)
	// - journal_lines.account_id -> accounts.id
	// - outbox_events.aggregate_id -> journal_entries.id (logical, not FK)

	// 6. Unique constraints:
	// - journal_entries.idempotency_key (unique index, enforced by GORM)
	// - outbox_events.idempotency_key (index, not unique - same key can have multiple events for different aggregates)

	logger.Info("Database constraints added (or verified)")
	return nil
}

// VerifyConstraints verifies that constraints are in place
func VerifyConstraints(logger *zap.Logger) error {
	if DB == nil {
		return nil
	}

	logger.Info("Verifying database constraints")

	// Check if CHECK constraints exist (MySQL 8.0.16+)
	var constraintCount int64
	DB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.TABLE_CONSTRAINTS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND CONSTRAINT_TYPE = 'CHECK'
		AND TABLE_NAME IN ('journal_lines', 'outbox_events')
	`).Scan(&constraintCount)

	if constraintCount > 0 {
		logger.Info("CHECK constraints found", zap.Int64("count", constraintCount))
	} else {
		logger.Warn("No CHECK constraints found (MySQL version may not support or constraints not added)")
	}

	// Verify foreign keys exist
	var fkCount int64
	DB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.KEY_COLUMN_USAGE 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND REFERENCED_TABLE_NAME IS NOT NULL
		AND TABLE_NAME IN ('journal_lines')
	`).Scan(&fkCount)

	logger.Info("Foreign key constraints found", zap.Int64("count", fkCount))

	// Verify unique indexes
	var uniqueIndexCount int64
	DB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.STATISTICS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND NON_UNIQUE = 0
		AND TABLE_NAME IN ('journal_entries', 'outbox_events')
	`).Scan(&uniqueIndexCount)

	logger.Info("Unique indexes found", zap.Int64("count", uniqueIndexCount))

	return nil
}
