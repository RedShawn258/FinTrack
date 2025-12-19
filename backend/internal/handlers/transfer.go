package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
	"github.com/RedShawn258/FinTrack/backend/internal/repositories"
	"github.com/RedShawn258/FinTrack/backend/internal/utils"
)

// CreateTransferRequest represents a request to transfer money between accounts
type CreateTransferRequest struct {
	FromAccountID uint    `json:"fromAccountId" binding:"required"`
	ToAccountID   uint    `json:"toAccountId" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Description   string  `json:"description"`
}

// CreateTransferResponse represents the response after creating a transfer
type CreateTransferResponse struct {
	EntryID       uint    `json:"entryId"`
	FromAccountID uint    `json:"fromAccountId"`
	ToAccountID   uint    `json:"toAccountId"`
	FromBalance   float64 `json:"fromBalance"`
	ToBalance     float64 `json:"toBalance"`
	Amount        float64 `json:"amount"`
	Message       string  `json:"message"`
}

// CreateTransfer creates a transfer between two accounts using double-entry accounting
// Uses idempotency keys to prevent duplicate processing
// @Summary      Create transfer
// @Description  Transfers money between two accounts using double-entry accounting. Creates a journal entry with two balanced lines (debit from source, credit to destination). Supports idempotency via Idempotency-Key header. All operations are atomic and use consistent locking order to prevent deadlocks.
// @Tags         transfers
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string  true  "Idempotency key to prevent duplicate processing"
// @Param        request          body      CreateTransferRequest  true  "Transfer data"
// @Success      201              {object}  CreateTransferResponse  "Transfer created successfully"
// @Success      200              {object}  CreateTransferResponse  "Transfer already processed (idempotent)"
// @Failure      400              {object}  map[string]string  "Invalid request or insufficient balance"
// @Failure      404              {object}  map[string]string  "Account not found"
// @Failure      409              {object}  map[string]string  "Idempotency key conflict"
// @Failure      500              {object}  map[string]string  "Internal server error"
// @Router       /transfers [post]
func CreateTransfer(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	// Get idempotency key from header
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
		return
	}

	var req CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid transfer creation data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate accounts are different
	if req.FromAccountID == req.ToAccountID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "From and to accounts must be different"})
		return
	}

	// Create request hash for idempotency validation
	requestBody := fmt.Sprintf("%d:%d:%.2f:%s", req.FromAccountID, req.ToAccountID, req.Amount, req.Description)
	requestHash := sha256.Sum256([]byte(requestBody))
	requestHashStr := hex.EncodeToString(requestHash[:])

	// Check if idempotency key already exists (outside transaction for quick check)
	var existingKey models.IdempotencyKey
	row := db.DB.Raw("SELECT id, idempotency_key, account_id, transaction_id, request_hash, status, response_code, response_body, created_at, updated_at FROM idempotency_keys WHERE idempotency_key = ? AND deleted_at IS NULL LIMIT 1", idempotencyKey).Row()
	err := row.Scan(&existingKey.ID, &existingKey.IdempotencyKey, &existingKey.AccountID, &existingKey.TransactionID, &existingKey.RequestHash, &existingKey.Status, &existingKey.ResponseCode, &existingKey.ResponseBody, &existingKey.CreatedAt, &existingKey.UpdatedAt)

	// Handle sql.ErrNoRows as "not found" (not an error - continue to create new)
	if err == nil && existingKey.ID > 0 {
		// Idempotency key exists - check if it's the same request
		if existingKey.RequestHash != requestHashStr {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Idempotency key already used with different request parameters",
			})
			return
		}

		// Same request - return cached response
		if existingKey.Status == "completed" && existingKey.TransactionID != nil {
			var entry models.JournalEntry
			// Use raw SQL to avoid GORM field name issues
			entryRow := db.DB.Raw("SELECT id, idempotency_key, description, metadata, created_at, updated_at FROM journal_entries WHERE id = ? AND deleted_at IS NULL LIMIT 1", *existingKey.TransactionID).Row()
			if err := entryRow.Scan(&entry.ID, &entry.IdempotencyKey, &entry.Description, &entry.Metadata, &entry.CreatedAt, &entry.UpdatedAt); err == nil {
				// Load lines separately using raw SQL
				var lines []models.JournalLine
				linesRows, _ := db.DB.Raw("SELECT id, entry_id, account_id, direction, amount, created_at, updated_at FROM journal_lines WHERE entry_id = ? AND deleted_at IS NULL", entry.ID).Rows()
				defer linesRows.Close()
				for linesRows.Next() {
					var line models.JournalLine
					linesRows.Scan(&line.ID, &line.EntryID, &line.AccountID, &line.Direction, &line.Amount, &line.CreatedAt, &line.UpdatedAt)
					lines = append(lines, line)
				}
				entry.Lines = lines
				// Get current balances using raw SQL
				var fromAccount, toAccount models.Account
				fromRow := db.DB.Raw("SELECT id, balance, created_at, updated_at FROM accounts WHERE id = ? LIMIT 1", req.FromAccountID).Row()
				fromRow.Scan(&fromAccount.ID, &fromAccount.Balance, &fromAccount.CreatedAt, &fromAccount.UpdatedAt)
				toRow := db.DB.Raw("SELECT id, balance, created_at, updated_at FROM accounts WHERE id = ? LIMIT 1", req.ToAccountID).Row()
				toRow.Scan(&toAccount.ID, &toAccount.Balance, &toAccount.CreatedAt, &toAccount.UpdatedAt)

				c.JSON(http.StatusOK, CreateTransferResponse{
					EntryID:       entry.ID,
					FromAccountID: req.FromAccountID,
					ToAccountID:   req.ToAccountID,
					FromBalance:   fromAccount.Balance,
					ToBalance:     toAccount.Balance,
					Amount:        req.Amount,
					Message:       "Transfer already processed (idempotent)",
				})
				return
			}
		}

		// Key exists but transaction failed or is still processing
		if existingKey.Status == "processing" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Transfer with this idempotency key is already being processed",
			})
			return
		}
	}

	// Determine lock order (always lock lower ID first to prevent deadlocks)
	firstID := req.FromAccountID
	secondID := req.ToAccountID
	if req.ToAccountID < req.FromAccountID {
		firstID = req.ToAccountID
		secondID = req.FromAccountID
	}

	// Process transfer atomically
	var entry models.JournalEntry
	var fromBalance, toBalance float64

	// Convert amount to cents for internal calculations (avoids float precision issues)
	amountCents := utils.DollarsToCents(req.Amount)

	// Initialize repository
	ledgerRepo := repositories.NewLedgerRepository()

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Get underlying *sql.Tx safely
		sqlTx, err := db.MustSQLTx(tx)
		if err != nil {
			return fmt.Errorf("failed to extract SQL transaction: %w", err)
		}

		// Use context with timeout
		ctx := context.Background()

		// Get or create idempotency key using repository
		_, _, err = ledgerRepo.GetOrCreateIdempotencyKey(ctx, sqlTx, idempotencyKey, req.FromAccountID, requestHashStr)
		if err != nil {
			return fmt.Errorf("failed to get or create idempotency key: %w", err)
		}

		// Lock accounts in consistent order (ascending ID) to prevent deadlocks
		// Get balances with FOR UPDATE lock (consistent order prevents deadlocks)
		var firstAccount, secondAccount models.Account
		firstRow := tx.Raw("SELECT id, balance, created_at, updated_at FROM accounts WHERE id = ? FOR UPDATE", firstID).Row()
		if err := firstRow.Scan(&firstAccount.ID, &firstAccount.Balance, &firstAccount.CreatedAt, &firstAccount.UpdatedAt); err != nil {
			return fmt.Errorf("failed to lock first account: %w", err)
		}

		secondRow := tx.Raw("SELECT id, balance, created_at, updated_at FROM accounts WHERE id = ? FOR UPDATE", secondID).Row()
		if err := secondRow.Scan(&secondAccount.ID, &secondAccount.Balance, &secondAccount.CreatedAt, &secondAccount.UpdatedAt); err != nil {
			return fmt.Errorf("failed to lock second account: %w", err)
		}

		// Determine current balances in cents based on lock order (for balance check and update)
		var fromBalanceCents, toBalanceCents int64
		if firstID == req.FromAccountID {
			fromBalanceCents = utils.DollarsToCents(firstAccount.Balance)
			toBalanceCents = utils.DollarsToCents(secondAccount.Balance)
		} else {
			fromBalanceCents = utils.DollarsToCents(secondAccount.Balance)
			toBalanceCents = utils.DollarsToCents(firstAccount.Balance)
		}

		// Check sufficient balance (using cents for precision)
		if fromBalanceCents < amountCents {
			return fmt.Errorf("insufficient balance: account %d has %.2f, need %.2f",
				req.FromAccountID, utils.CentsToDollars(fromBalanceCents), req.Amount)
		}

		// Create journal entry using repository
		entryID, err := ledgerRepo.InsertJournalEntry(ctx, sqlTx, idempotencyKey, req.Description)
		if err != nil {
			return fmt.Errorf("failed to insert journal entry: %w", err)
		}
		entry.ID = entryID
		entry.IdempotencyKey = idempotencyKey
		entry.Description = req.Description

		// Create journal lines using repository (amount in cents)
		journalLines := []repositories.JournalLine{
			{EntryID: entryID, AccountID: req.FromAccountID, Direction: "debit", Amount: amountCents},
			{EntryID: entryID, AccountID: req.ToAccountID, Direction: "credit", Amount: amountCents},
		}
		_, _, err = ledgerRepo.InsertJournalLines(ctx, sqlTx, journalLines)
		if err != nil {
			return fmt.Errorf("failed to insert journal lines: %w", err)
		}
		if firstID == req.FromAccountID {
			fromBalanceCents = utils.DollarsToCents(firstAccount.Balance)
			toBalanceCents = utils.DollarsToCents(secondAccount.Balance)
		} else {
			fromBalanceCents = utils.DollarsToCents(secondAccount.Balance)
			toBalanceCents = utils.DollarsToCents(firstAccount.Balance)
		}

		err = ledgerRepo.UpdateAccountBalances(ctx, sqlTx, req.FromAccountID, req.ToAccountID, fromBalanceCents, toBalanceCents, amountCents)
		if err != nil {
			return fmt.Errorf("failed to update account balances: %w", err)
		}

		// Get updated balances for response
		var updatedFromBalance, updatedToBalance float64
		fromBalanceRow := tx.Raw("SELECT balance FROM accounts WHERE id = ?", req.FromAccountID).Row()
		if err := fromBalanceRow.Scan(&updatedFromBalance); err != nil {
			return fmt.Errorf("failed to get updated from balance: %w", err)
		}
		toBalanceRow := tx.Raw("SELECT balance FROM accounts WHERE id = ?", req.ToAccountID).Row()
		if err := toBalanceRow.Scan(&updatedToBalance); err != nil {
			return fmt.Errorf("failed to get updated to balance: %w", err)
		}
		fromBalance = updatedFromBalance
		toBalance = updatedToBalance

		// Update idempotency key with success status using repository
		responseBody := fmt.Sprintf(`{"entryId":%d,"fromAccountId":%d,"toAccountId":%d,"amount":%.2f}`,
			entry.ID, req.FromAccountID, req.ToAccountID, req.Amount)
		entryIDPtr := &entry.ID
		err = ledgerRepo.UpdateIdempotencyKeyStatus(ctx, sqlTx, idempotencyKey, "completed", entryIDPtr, http.StatusCreated, responseBody)
		if err != nil {
			return fmt.Errorf("failed to update idempotency key status: %w", err)
		}

		// Check if outbox event already exists (prevent duplicates)
		existingOutboxEvent, err := ledgerRepo.GetOutboxEventByKey(ctx, sqlTx, idempotencyKey)
		if err != nil {
			return fmt.Errorf("failed to check for existing outbox event: %w", err)
		}
		if existingOutboxEvent != nil {
			// Outbox event already exists for this idempotency key, skip creation
			// This ensures idempotency: same transfer idempotency key = same outbox event
			return nil
		}

		// Create TransferCreated event payload
		eventPayload := fmt.Sprintf(`{
			"eventId": %d,
			"eventType": "TransferCreated",
			"aggregateType": "Transfer",
			"aggregateId": %d,
			"journalEntryId": %d,
			"fromAccountId": %d,
			"toAccountId": %d,
			"amount": %.2f,
			"fromBalance": %.2f,
			"toBalance": %.2f,
			"description": "%s",
			"timestamp": "%s"
		}`, entry.ID, entry.ID, entry.ID, req.FromAccountID, req.ToAccountID, req.Amount, fromBalance, toBalance, req.Description, time.Now().UTC().Format(time.RFC3339))

		// Create outbox event using repository
		err = ledgerRepo.InsertOutboxEvent(ctx, sqlTx, "TransferCreated", "Transfer", entry.ID, idempotencyKey, eventPayload)
		if err != nil {
			return fmt.Errorf("failed to insert outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Error("Failed to process transfer", zap.Error(err))

		// Update idempotency key with failure status using pure database/sql to bypass GORM
		sqlDB, _ := db.DB.DB()
		sqlDB.Exec("UPDATE idempotency_keys SET status = ?, response_code = ?, response_body = ?, updated_at = NOW() WHERE idempotency_key = ?",
			"failed", http.StatusInternalServerError, fmt.Sprintf(`{"error":"%s"}`, err.Error()), idempotencyKey)

		// Check for specific error types
		errStr := err.Error()
		if err == gorm.ErrRecordNotFound || errStr == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "One or both accounts not found"})
			return
		}

		if len(errStr) >= 20 && errStr[:20] == "insufficient balance" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process transfer"})
		return
	}

	log.Info("Transfer processed",
		zap.Uint("entryId", entry.ID),
		zap.Uint("fromAccountId", req.FromAccountID),
		zap.Uint("toAccountId", req.ToAccountID),
		zap.Float64("amount", req.Amount),
		zap.Float64("fromBalance", fromBalance),
		zap.Float64("toBalance", toBalance),
	)

	c.JSON(http.StatusCreated, CreateTransferResponse{
		EntryID:       entry.ID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		FromBalance:   fromBalance,
		ToBalance:     toBalance,
		Amount:        req.Amount,
		Message:       "Transfer processed successfully",
	})
}
