package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// LedgerRepository handles all database operations for the double-entry accounting ledger.
// Uses pure database/sql to avoid GORM query generation issues with reserved keywords.
//
// ARCHITECTURE BOUNDARY: This repository MUST use pure database/sql.
// DO NOT use GORM APIs (db.Model, db.Where, db.Create) on tables managed by this repository.
// See ARCHITECTURE.md for details.
//
// Tables managed: idempotency_keys, journal_entries, journal_lines, outbox_events
type LedgerRepository struct {
	// No fields needed - all operations use *sql.Tx passed as parameter
}

// NewLedgerRepository creates a new ledger repository.
func NewLedgerRepository() *LedgerRepository {
	return &LedgerRepository{}
}

// IdempotencyKeyRecord represents an idempotency key record.
type IdempotencyKeyRecord struct {
	ID             uint
	IdempotencyKey string
	AccountID      uint
	TransactionID  *uint
	RequestHash    string
	Status         string
	ResponseCode   int
	ResponseBody   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// GetOrCreateIdempotencyKey retrieves an existing idempotency key or creates a new one.
// Returns the record and a boolean indicating if it was just created.
func (r *LedgerRepository) GetOrCreateIdempotencyKey(
	ctx context.Context,
	sqlTx *sql.Tx,
	idempotencyKey string,
	accountID uint,
	requestHash string,
) (*IdempotencyKeyRecord, bool, error) {
	// Set timeout for this operation (2 seconds)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Check if key exists
	var record IdempotencyKeyRecord
	row := sqlTx.QueryRowContext(ctx,
		"SELECT id, idempotency_key, account_id, transaction_id, request_hash, status, response_code, response_body, created_at, updated_at FROM idempotency_keys WHERE idempotency_key = ? AND deleted_at IS NULL LIMIT 1",
		idempotencyKey,
	)

	err := row.Scan(
		&record.ID,
		&record.IdempotencyKey,
		&record.AccountID,
		&record.TransactionID,
		&record.RequestHash,
		&record.Status,
		&record.ResponseCode,
		&record.ResponseBody,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err == nil {
		// Record exists - update it to "processing" status
		ctxUpdate, cancelUpdate := context.WithTimeout(ctx, 2*time.Second)
		defer cancelUpdate()

		_, err = sqlTx.ExecContext(ctxUpdate,
			"UPDATE idempotency_keys SET account_id = ?, request_hash = ?, status = ?, response_code = ?, updated_at = NOW() WHERE idempotency_key = ?",
			accountID, requestHash, "processing", 102, idempotencyKey,
		)
		if err != nil {
			return nil, false, fmt.Errorf("failed to update idempotency key: %w", err)
		}

		// Reload the record
		row = sqlTx.QueryRowContext(ctx,
			"SELECT id, idempotency_key, account_id, transaction_id, request_hash, status, response_code, response_body, created_at, updated_at FROM idempotency_keys WHERE idempotency_key = ? AND deleted_at IS NULL LIMIT 1",
			idempotencyKey,
		)
		row.Scan(
			&record.ID,
			&record.IdempotencyKey,
			&record.AccountID,
			&record.TransactionID,
			&record.RequestHash,
			&record.Status,
			&record.ResponseCode,
			&record.ResponseBody,
			&record.CreatedAt,
			&record.UpdatedAt,
		)

		return &record, false, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, fmt.Errorf("failed to query idempotency key: %w", err)
	}

	// Record doesn't exist - create it
	ctxCreate, cancelCreate := context.WithTimeout(ctx, 5*time.Second)
	defer cancelCreate()

	_, err = sqlTx.ExecContext(ctxCreate,
		"INSERT INTO idempotency_keys (idempotency_key, account_id, request_hash, status, response_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())",
		idempotencyKey, accountID, requestHash, "processing", 102,
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create idempotency key: %w", err)
	}

	// Reload the created record
	row = sqlTx.QueryRowContext(ctx,
		"SELECT id, idempotency_key, account_id, transaction_id, request_hash, status, response_code, response_body, created_at, updated_at FROM idempotency_keys WHERE idempotency_key = ? AND deleted_at IS NULL LIMIT 1",
		idempotencyKey,
	)
	err = row.Scan(
		&record.ID,
		&record.IdempotencyKey,
		&record.AccountID,
		&record.TransactionID,
		&record.RequestHash,
		&record.Status,
		&record.ResponseCode,
		&record.ResponseBody,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to reload created idempotency key: %w", err)
	}

	return &record, true, nil
}

// UpdateIdempotencyKeyStatus updates the status of an idempotency key.
func (r *LedgerRepository) UpdateIdempotencyKeyStatus(
	ctx context.Context,
	sqlTx *sql.Tx,
	idempotencyKey string,
	status string,
	transactionID *uint,
	responseCode int,
	responseBody string,
) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := sqlTx.ExecContext(ctx,
		"UPDATE idempotency_keys SET status = ?, transaction_id = ?, response_code = ?, response_body = ?, updated_at = NOW() WHERE idempotency_key = ?",
		status, transactionID, responseCode, responseBody, idempotencyKey,
	)
	if err != nil {
		return fmt.Errorf("failed to update idempotency key status: %w", err)
	}
	return nil
}

// InsertJournalEntry creates a new journal entry and returns its ID.
func (r *LedgerRepository) InsertJournalEntry(
	ctx context.Context,
	sqlTx *sql.Tx,
	idempotencyKey string,
	description string,
) (uint, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := sqlTx.ExecContext(ctx,
		"INSERT INTO journal_entries (idempotency_key, description, created_at, updated_at) VALUES (?, ?, NOW(), NOW())",
		idempotencyKey, description,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert journal entry: %w", err)
	}

	// Get the inserted entry ID
	var entryID uint
	row := sqlTx.QueryRowContext(ctx,
		"SELECT id FROM journal_entries WHERE idempotency_key = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT 1",
		idempotencyKey,
	)
	if err := row.Scan(&entryID); err != nil {
		return 0, fmt.Errorf("failed to get journal entry ID: %w", err)
	}

	return entryID, nil
}

// JournalLine represents a journal line (debit or credit).
type JournalLine struct {
	EntryID   uint
	AccountID uint
	Direction string // "debit" or "credit"
	Amount    int64  // Amount in cents (integer)
}

// InsertJournalLines inserts journal lines for a journal entry.
// Returns the total debits and credits for invariant verification.
func (r *LedgerRepository) InsertJournalLines(
	ctx context.Context,
	sqlTx *sql.Tx,
	lines []JournalLine,
) (totalDebits int64, totalCredits int64, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, line := range lines {
		// Convert cents to decimal for storage (amount is stored as DECIMAL(15,2))
		amountDecimal := float64(line.Amount) / 100.0

		_, err = sqlTx.ExecContext(ctx,
			"INSERT INTO journal_lines (entry_id, account_id, direction, amount, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())",
			line.EntryID, line.AccountID, line.Direction, amountDecimal,
		)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to insert journal line: %w", err)
		}

		if line.Direction == "debit" {
			totalDebits += line.Amount
		} else {
			totalCredits += line.Amount
		}
	}

	// Verify invariant: debits == credits
	if totalDebits != totalCredits {
		return 0, 0, fmt.Errorf("double-entry invariant violated: debits (%d cents) != credits (%d cents)", totalDebits, totalCredits)
	}

	return totalDebits, totalCredits, nil
}

// UpdateAccountBalances updates account balances atomically.
// Amounts are in cents (integer).
// Note: Accounts should already be locked with FOR UPDATE before calling this function.
// This function expects balances to be passed in to avoid re-querying locked rows.
func (r *LedgerRepository) UpdateAccountBalances(
	ctx context.Context,
	sqlTx *sql.Tx,
	fromAccountID uint,
	toAccountID uint,
	fromBalanceCents int64, // Current from account balance in cents
	toBalanceCents int64, // Current to account balance in cents
	amountCents int64, // Amount to transfer in cents
) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Calculate new balances (in cents)
	newFromBalanceCents := fromBalanceCents - amountCents
	newToBalanceCents := toBalanceCents + amountCents

	// Convert back to decimal for storage
	newFromBalance := float64(newFromBalanceCents) / 100.0
	newToBalance := float64(newToBalanceCents) / 100.0

	// Update balances
	_, err1 := sqlTx.ExecContext(ctx,
		"UPDATE accounts SET balance = ?, updated_at = NOW() WHERE id = ?",
		newFromBalance, fromAccountID,
	)
	if err1 != nil {
		return fmt.Errorf("failed to update from account balance: %w", err1)
	}

	_, err2 := sqlTx.ExecContext(ctx,
		"UPDATE accounts SET balance = ?, updated_at = NOW() WHERE id = ?",
		newToBalance, toAccountID,
	)
	if err2 != nil {
		return fmt.Errorf("failed to update to account balance: %w", err2)
	}

	return nil
}

// OutboxEventRecord represents an outbox event record.
type OutboxEventRecord struct {
	ID             uint
	EventType      string
	AggregateType  string
	AggregateID    uint
	IdempotencyKey string
	PayloadJSON    string
	Status         string
	RetryCount     int
	NextRetryAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// GetOutboxEventByKey checks if an outbox event already exists for the given idempotency key.
func (r *LedgerRepository) GetOutboxEventByKey(
	ctx context.Context,
	sqlTx *sql.Tx,
	idempotencyKey string,
) (*OutboxEventRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var record OutboxEventRecord
	row := sqlTx.QueryRowContext(ctx,
		"SELECT id, event_type, aggregate_type, aggregate_id, idempotency_key, payload_json, status, retry_count, next_retry_at, created_at, updated_at FROM outbox_events WHERE idempotency_key = ? AND deleted_at IS NULL LIMIT 1",
		idempotencyKey,
	)

	err := row.Scan(
		&record.ID,
		&record.EventType,
		&record.AggregateType,
		&record.AggregateID,
		&record.IdempotencyKey,
		&record.PayloadJSON,
		&record.Status,
		&record.RetryCount,
		&record.NextRetryAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox event: %w", err)
	}

	return &record, nil
}

// InsertOutboxEvent creates a new outbox event.
func (r *LedgerRepository) InsertOutboxEvent(
	ctx context.Context,
	sqlTx *sql.Tx,
	eventType string,
	aggregateType string,
	aggregateID uint,
	idempotencyKey string,
	payloadJSON string,
) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := sqlTx.ExecContext(ctx,
		"INSERT INTO outbox_events (event_type, aggregate_type, aggregate_id, idempotency_key, payload_json, status, retry_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())",
		eventType, aggregateType, aggregateID, idempotencyKey, payloadJSON, "PENDING", 0,
	)
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return nil
}
