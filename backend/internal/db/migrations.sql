-- Explicit SQL migrations for idempotency and double-entry accounting tables
-- These are NOT managed by GORM AutoMigrate to avoid schema validation issues

-- Idempotency keys table
CREATE TABLE IF NOT EXISTS `idempotency_keys` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `idempotency_key` VARCHAR(255) NOT NULL,
  `account_id` INT UNSIGNED NOT NULL,
  `transaction_id` INT UNSIGNED DEFAULT NULL,
  `request_hash` VARCHAR(64) DEFAULT NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'processing',
  `response_code` INT DEFAULT NULL,
  `response_body` TEXT DEFAULT NULL,
  `created_at` DATETIME DEFAULT NULL,
  `updated_at` DATETIME DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  UNIQUE KEY `idx_idempotency_keys_idempotency_key` (`idempotency_key`),
  KEY `idx_account_id` (`account_id`),
  KEY `idx_transaction_id` (`transaction_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Journal entries table (double-entry accounting)
CREATE TABLE IF NOT EXISTS `journal_entries` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `idempotency_key` VARCHAR(255) NOT NULL,
  `description` TEXT DEFAULT NULL,
  `metadata` JSON DEFAULT NULL,
  `created_at` DATETIME DEFAULT NULL,
  `updated_at` DATETIME DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  UNIQUE KEY `idx_journal_entries_idempotency_key` (`idempotency_key`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Journal lines table (double-entry accounting)
CREATE TABLE IF NOT EXISTS `journal_lines` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `entry_id` BIGINT UNSIGNED NOT NULL,
  `account_id` INT UNSIGNED NOT NULL,
  `direction` ENUM('debit', 'credit') NOT NULL,
  `amount` DECIMAL(15,2) NOT NULL,
  `created_at` DATETIME DEFAULT NULL,
  `updated_at` DATETIME DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  KEY `idx_entry_id` (`entry_id`),
  KEY `idx_account_id` (`account_id`),
  KEY `idx_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_journal_lines_entry` FOREIGN KEY (`entry_id`) REFERENCES `journal_entries` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Outbox events table (event-driven architecture)
CREATE TABLE IF NOT EXISTS `outbox_events` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `event_type` VARCHAR(100) NOT NULL,
  `aggregate_type` VARCHAR(100) NOT NULL,
  `aggregate_id` BIGINT UNSIGNED NOT NULL,
  `idempotency_key` VARCHAR(255) DEFAULT NULL,
  `payload_json` JSON NOT NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'PENDING',
  `retry_count` BIGINT NOT NULL DEFAULT 0,
  `next_retry_at` DATETIME DEFAULT NULL,
  `last_error` TEXT DEFAULT NULL,
  `created_at` DATETIME DEFAULT NULL,
  `updated_at` DATETIME DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  KEY `idx_event_type` (`event_type`),
  KEY `idx_aggregate_type` (`aggregate_type`),
  KEY `idx_aggregate_id` (`aggregate_id`),
  KEY `idx_idempotency_key` (`idempotency_key`),
  KEY `idx_status` (`status`),
  KEY `idx_next_retry_at` (`next_retry_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
