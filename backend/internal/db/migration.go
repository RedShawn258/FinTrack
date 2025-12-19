package db

import (
	"strings"

	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// RunMigrations performs database migrations for all models
func RunMigrations(logger *zap.Logger) error {
	if DB == nil {
		logger.Error("Database connection not initialized")
		return nil
	}

	logger.Info("Running database migrations")

	// PRODUCTION GUARDRAIL: Validate forbidden models are not included
	// This prevents accidental re-addition of correctness-critical models to AutoMigrate
	migrateModels := []interface{}{
		&models.User{},
		&models.Category{},
		&models.Budget{},
		&models.Transaction{},
		&models.Badge{},
		&models.UserBadge{},
		&models.UserPoints{},
		&models.RefreshToken{},
		// Accounts/Ledger models
		&models.Account{},
		&models.LedgerTransaction{}, // Legacy - kept for backward compatibility
		// FORBIDDEN: IdempotencyKey, JournalEntry, JournalLine, OutboxEvent MUST NOT be here
		// These models must use explicit SQL migrations via RunRawMigrations() below
		// See GORM_ELIMINATION_SUMMARY.md and ARCHITECTURE.md for details
	}

	// Runtime guard: Fail fast if forbidden models are added to AutoMigrate
	if err := validateAutoMigrateModels(migrateModels, logger); err != nil {
		logger.Fatal("PRODUCTION GUARDRAIL VIOLATION", zap.Error(err))
		return err
	}

	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes
	// NOTE: IdempotencyKey, JournalEntry, JournalLine, and OutboxEvent are EXCLUDED from AutoMigrate
	// to prevent GORM from generating queries with reserved keywords (e.g., 'key'). These tables
	// are managed via explicit SQL migrations in RunRawMigrations() to ensure complete control over schema.
	err := DB.AutoMigrate(migrateModels...)

	if err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		return err
	}

	// Run explicit SQL migrations for idempotency and double-entry tables
	// These are NOT managed by GORM AutoMigrate to avoid schema validation issues
	if err := RunRawMigrations(logger); err != nil {
		logger.Error("Failed to run raw SQL migrations", zap.Error(err))
		return err
	}

	// Seed default badges if they don't exist
	seedDefaultBadges(logger)

	// Add database constraints for financial correctness
	if err := AddConstraints(logger); err != nil {
		logger.Warn("Failed to add database constraints", zap.Error(err))
		// Don't fail migration if constraints can't be added (may be unsupported MySQL version)
	}

	// Verify constraints are in place
	if err := VerifyConstraints(logger); err != nil {
		logger.Warn("Failed to verify database constraints", zap.Error(err))
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// seedDefaultBadges adds the default badges to the database if they don't exist
func seedDefaultBadges(logger *zap.Logger) {
	var count int64
	DB.Model(&models.Badge{}).Count(&count)

	// Only seed if no badges exist
	if count == 0 {
		badges := []models.Badge{
			{
				Name:        "Budget Beginner",
				Description: "Created your first budget",
				ImageURL:    "/images/badges/budget_beginner.png",
				Category:    "budgeting",
				Threshold:   10,
			},
			{
				Name:        "Tracking Pro",
				Description: "Tracked expenses for 7 consecutive days",
				ImageURL:    "/images/badges/tracking_pro.png",
				Category:    "consistency",
				Threshold:   50,
			},
			{
				Name:        "Savings Star",
				Description: "Saved 10% of your income",
				ImageURL:    "/images/badges/savings_star.png",
				Category:    "savings",
				Threshold:   100,
			},
			{
				Name:        "Budget Master",
				Description: "Stayed under budget for 3 consecutive months",
				ImageURL:    "/images/badges/budget_master.png",
				Category:    "budgeting",
				Threshold:   150,
			},
			{
				Name:        "Finance Ninja",
				Description: "Created budgets in all essential categories",
				ImageURL:    "/images/badges/finance_ninja.png",
				Category:    "budgeting",
				Threshold:   200,
			},
			{
				Name:        "Expense Tracker",
				Description: "Logged 100 transactions",
				ImageURL:    "/images/badges/expense_tracker.png",
				Category:    "tracking",
				Threshold:   250,
			},
			{
				Name:        "Financial Wizard",
				Description: "Reached 500 total points",
				ImageURL:    "/images/badges/financial_wizard.png",
				Category:    "achievement",
				Threshold:   500,
			},
		}

		for _, badge := range badges {
			if err := DB.Create(&badge).Error; err != nil {
				logger.Error("Failed to seed badge", zap.Error(err), zap.String("badge", badge.Name))
			}
		}

		logger.Info("Seeded default badges")
	}
}

// RunRawMigrations executes explicit SQL migrations for tables that GORM should not manage
// This prevents GORM from generating queries with reserved keywords (e.g., 'key' column name)
// These tables are managed via explicit SQL DDL statements, not GORM AutoMigrate
func RunRawMigrations(logger *zap.Logger) error {
	if DB == nil {
		logger.Error("Database connection not initialized")
		return nil
	}

	logger.Info("Running explicit SQL migrations for idempotency and double-entry tables")

	// Execute CREATE TABLE IF NOT EXISTS statements directly
	// These tables are NOT managed by GORM AutoMigrate to avoid schema validation issues
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			idempotency_key VARCHAR(255) NOT NULL,
			account_id INT UNSIGNED NOT NULL,
			transaction_id INT UNSIGNED DEFAULT NULL,
			request_hash VARCHAR(64) DEFAULT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'processing',
			response_code INT DEFAULT NULL,
			response_body TEXT DEFAULT NULL,
			created_at DATETIME DEFAULT NULL,
			updated_at DATETIME DEFAULT NULL,
			deleted_at DATETIME DEFAULT NULL,
			UNIQUE KEY idx_idempotency_keys_idempotency_key (idempotency_key),
			KEY idx_account_id (account_id),
			KEY idx_transaction_id (transaction_id),
			KEY idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`,
		`CREATE TABLE IF NOT EXISTS journal_entries (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			idempotency_key VARCHAR(255) NOT NULL,
			description TEXT DEFAULT NULL,
			metadata JSON DEFAULT NULL,
			created_at DATETIME DEFAULT NULL,
			updated_at DATETIME DEFAULT NULL,
			deleted_at DATETIME DEFAULT NULL,
			UNIQUE KEY idx_journal_entries_idempotency_key (idempotency_key),
			KEY idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`,
		`CREATE TABLE IF NOT EXISTS journal_lines (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			entry_id BIGINT UNSIGNED NOT NULL,
			account_id INT UNSIGNED NOT NULL,
			direction ENUM('debit', 'credit') NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			created_at DATETIME DEFAULT NULL,
			updated_at DATETIME DEFAULT NULL,
			deleted_at DATETIME DEFAULT NULL,
			KEY idx_entry_id (entry_id),
			KEY idx_account_id (account_id),
			KEY idx_deleted_at (deleted_at),
			CONSTRAINT fk_journal_lines_entry FOREIGN KEY (entry_id) REFERENCES journal_entries (id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`,
		`CREATE TABLE IF NOT EXISTS outbox_events (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			event_type VARCHAR(100) NOT NULL,
			aggregate_type VARCHAR(100) NOT NULL,
			aggregate_id BIGINT UNSIGNED NOT NULL,
			idempotency_key VARCHAR(255) DEFAULT NULL,
			payload_json JSON NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
			retry_count BIGINT NOT NULL DEFAULT 0,
			next_retry_at DATETIME DEFAULT NULL,
			last_error TEXT DEFAULT NULL,
			created_at DATETIME DEFAULT NULL,
			updated_at DATETIME DEFAULT NULL,
			deleted_at DATETIME DEFAULT NULL,
			KEY idx_event_type (event_type),
			KEY idx_aggregate_type (aggregate_type),
			KEY idx_aggregate_id (aggregate_id),
			KEY idx_idempotency_key (idempotency_key),
			KEY idx_status (status),
			KEY idx_next_retry_at (next_retry_at),
			KEY idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci`,
	}

	for _, stmt := range migrations {
		if err := DB.Exec(stmt).Error; err != nil {
			// Ignore "table already exists" errors (idempotent)
			if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "Duplicate key") {
				logger.Warn("Migration statement failed (may already exist)", zap.Error(err))
			}
		}
	}

	logger.Info("Explicit SQL migrations completed")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
