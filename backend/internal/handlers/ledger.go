package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// CreateTransactionRequest represents a request to create a ledger transaction
type CreateTransactionRequest struct {
	AccountID      uint    `json:"accountId" binding:"required"`
	Type           string  `json:"type" binding:"required,oneof=credit debit"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Description    string  `json:"description"`
	IdempotencyKey string  `json:"idempotencyKey" binding:"required"`
}

// CreateTransactionResponse represents the response after creating a transaction
type CreateTransactionResponse struct {
	TransactionID uint    `json:"transactionId"`
	AccountID     uint    `json:"accountId"`
	NewBalance    float64 `json:"newBalance"`
	Message       string  `json:"message"`
}

// CreateLedgerTransaction creates a credit or debit transaction on an account
// Uses idempotency keys to prevent duplicate processing
// @Summary      Create ledger transaction
// @Description  Creates a credit or debit transaction on an account. Supports idempotency via Idempotency-Key header to prevent duplicate processing. All operations are atomic.
// @Tags         ledger
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string  true  "Idempotency key to prevent duplicate processing"
// @Param        request          body      CreateTransactionRequest  true  "Transaction data"
// @Success      201              {object}  CreateTransactionResponse  "Transaction created successfully"
// @Success      200              {object}  CreateTransactionResponse  "Transaction already processed (idempotent)"
// @Failure      400              {object}  map[string]string  "Invalid request or insufficient balance"
// @Failure      404              {object}  map[string]string  "Account not found"
// @Failure      409              {object}  map[string]string  "Idempotency key conflict"
// @Failure      500              {object}  map[string]string  "Internal server error"
// @Router       /transactions [post]
func CreateLedgerTransaction(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	// Get idempotency key from header (preferred) or request body
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
		return
	}

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid transaction creation data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use the header value if provided, otherwise use body value
	if idempotencyKey == "" {
		idempotencyKey = req.IdempotencyKey
	}

	// Validate transaction type
	if req.Type != "credit" && req.Type != "debit" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be 'credit' or 'debit'"})
		return
	}

	// Create request hash for validation
	requestBody := fmt.Sprintf("%d:%s:%.2f:%s", req.AccountID, req.Type, req.Amount, req.Description)
	requestHash := sha256.Sum256([]byte(requestBody))
	requestHashStr := hex.EncodeToString(requestHash[:])

	// Check if idempotency key already exists (using raw SQL to avoid GORM field name issues)
	var existingKey models.IdempotencyKey
	row := db.DB.Raw("SELECT id, idempotency_key, account_id, transaction_id, request_hash, status, response_code, response_body, created_at, updated_at FROM idempotency_keys WHERE idempotency_key = ? AND deleted_at IS NULL LIMIT 1", idempotencyKey).Row()
	err := row.Scan(&existingKey.ID, &existingKey.IdempotencyKey, &existingKey.AccountID, &existingKey.TransactionID, &existingKey.RequestHash, &existingKey.Status, &existingKey.ResponseCode, &existingKey.ResponseBody, &existingKey.CreatedAt, &existingKey.UpdatedAt)

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
			var transaction models.LedgerTransaction
			if err := db.DB.First(&transaction, *existingKey.TransactionID).Error; err == nil {
				var account models.Account
				db.DB.First(&account, transaction.AccountID)

				c.JSON(http.StatusOK, CreateTransactionResponse{
					TransactionID: transaction.ID,
					AccountID:     transaction.AccountID,
					NewBalance:    account.Balance,
					Message:       "Transaction already processed (idempotent)",
				})
				return
			}
		}

		// Key exists but transaction failed or is still processing
		if existingKey.Status == "processing" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Transaction with this idempotency key is already being processed",
			})
			return
		}
	}

	// Verify account exists
	var account models.Account
	if err := db.DB.First(&account, req.AccountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		log.Error("Failed to fetch account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch account"})
		return
	}

	// Check sufficient balance for debit transactions
	if req.Type == "debit" && account.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Insufficient balance",
			"balance": account.Balance,
			"amount":  req.Amount,
		})
		return
	}

	// Process transaction atomically
	var transaction models.LedgerTransaction
	var newBalance float64

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Create or update idempotency key record
		idempotencyRecord := models.IdempotencyKey{
			IdempotencyKey: idempotencyKey,
			AccountID:      req.AccountID,
			RequestHash:    requestHashStr,
			Status:         "processing",
			ResponseCode:   http.StatusProcessing,
		}

		// Try to create, if exists update status (using raw SQL to avoid GORM field name issues)
		var existingRecordID uint
		row := tx.Raw("SELECT id FROM idempotency_keys WHERE idempotency_key = ? AND deleted_at IS NULL LIMIT 1", idempotencyKey).Row()
		if row.Scan(&existingRecordID) == nil && existingRecordID > 0 {
			// Update existing record
			if err := tx.Exec("UPDATE idempotency_keys SET account_id = ?, request_hash = ?, status = ?, updated_at = NOW() WHERE idempotency_key = ?",
				req.AccountID, requestHashStr, "processing", idempotencyKey).Error; err != nil {
				return err
			}
		} else {
			// Create new record
			if err := tx.Exec("INSERT INTO idempotency_keys (idempotency_key, account_id, request_hash, status, response_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())",
				idempotencyKey, req.AccountID, requestHashStr, "processing", http.StatusProcessing).Error; err != nil {
				return err
			}
		}

		// Lock account row for update
		var lockedAccount models.Account
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&lockedAccount, req.AccountID).Error; err != nil {
			return err
		}

		// Calculate new balance
		if req.Type == "credit" {
			newBalance = lockedAccount.Balance + req.Amount
		} else {
			newBalance = lockedAccount.Balance - req.Amount
		}

		// Update account balance
		if err := tx.Model(&lockedAccount).Update("balance", newBalance).Error; err != nil {
			return err
		}

		// Create transaction record
		transaction = models.LedgerTransaction{
			AccountID:      req.AccountID,
			Type:           models.TransactionType(req.Type),
			Amount:         req.Amount,
			Description:    req.Description,
			IdempotencyKey: idempotencyKey,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		// Update idempotency key with transaction ID and success status
		idempotencyRecord.TransactionID = &transaction.ID
		idempotencyRecord.Status = "completed"
		idempotencyRecord.ResponseCode = http.StatusCreated
		idempotencyRecord.ResponseBody = fmt.Sprintf(`{"transactionId":%d,"accountId":%d,"newBalance":%.2f}`, transaction.ID, req.AccountID, newBalance)

		if err := tx.Save(&idempotencyRecord).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Error("Failed to process transaction", zap.Error(err))

		// Update idempotency key with failure status (using raw SQL to avoid GORM field name issues)
		db.DB.Exec("UPDATE idempotency_keys SET status = ?, response_code = ?, response_body = ?, updated_at = NOW() WHERE idempotency_key = ?",
			"failed", http.StatusInternalServerError, fmt.Sprintf(`{"error":"%s"}`, err.Error()), idempotencyKey)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process transaction"})
		return
	}

	log.Info("Transaction processed",
		zap.Uint("transactionId", transaction.ID),
		zap.Uint("accountId", req.AccountID),
		zap.String("type", req.Type),
		zap.Float64("amount", req.Amount),
		zap.Float64("newBalance", newBalance),
	)

	c.JSON(http.StatusCreated, CreateTransactionResponse{
		TransactionID: transaction.ID,
		AccountID:     req.AccountID,
		NewBalance:    newBalance,
		Message:       "Transaction processed successfully",
	})
}
