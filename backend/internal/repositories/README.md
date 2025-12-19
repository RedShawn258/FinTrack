# Repository Layer

This package contains repository implementations for database access.

## LedgerRepository

**Location:** `ledger_repo.go`

**Purpose:** Handles all database operations for correctness-critical financial tables.

**Key Characteristics:**

- Uses **pure `database/sql`** (no GORM)
- Accepts `*sql.Tx` extracted from GORM transaction via `db.MustSQLTx()`
- Uses `context.Context` for timeouts (2s reads, 5s writes)
- Uses **integer cents** internally for money calculations

**Tables Managed:**

- `idempotency_keys`
- `journal_entries`
- `journal_lines`
- `outbox_events`

**Why Pure SQL?**
See `ARCHITECTURE.md` and `GORM_ELIMINATION_SUMMARY.md` for details. In short: GORM's schema cache can cause SQL syntax errors with reserved keywords, and we need explicit control over every SQL query for financial correctness.

**Usage Example:**

```go
err := db.DB.Transaction(func(tx *gorm.DB) error {
    sqlTx, err := db.MustSQLTx(tx)
    if err != nil {
        return err
    }

    ledgerRepo := repositories.NewLedgerRepository()
    ctx := context.Background()

    // All operations use pure SQL
    record, _, err := ledgerRepo.GetOrCreateIdempotencyKey(ctx, sqlTx, key, accountID, hash)
    entryID, err := ledgerRepo.InsertJournalEntry(ctx, sqlTx, key, description)
    err = ledgerRepo.UpdateAccountBalances(ctx, sqlTx, fromID, toID, fromCents, toCents, amountCents)

    return nil
})
```

## Other Repositories

- `TokenRepository` - Uses GORM (non-critical path, GORM allowed)

## Architecture Boundary

**DO NOT use GORM APIs directly on tables managed by `LedgerRepository`.**

See `ARCHITECTURE.md` for complete boundary rules.
