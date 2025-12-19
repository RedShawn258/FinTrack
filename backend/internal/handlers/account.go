package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// CreateAccountRequest represents the request to create an account
type CreateAccountRequest struct {
	InitialBalance float64 `json:"initialBalance" binding:"required,gte=0"`
}

// CreateAccountResponse represents the response after creating an account
type CreateAccountResponse struct {
	ID      uint    `json:"id"`
	Balance float64 `json:"balance"`
	Message string  `json:"message"`
}

// CreateAccount creates a new account with an initial balance
// @Summary      Create account
// @Description  Create a new account with an initial balance. Uses transactional SQL to ensure atomicity.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        request  body      CreateAccountRequest  true  "Account creation data"
// @Success      201      {object}  CreateAccountResponse  "Account created successfully"
// @Failure      400      {object}  map[string]string  "Invalid request"
// @Failure      500      {object}  map[string]string  "Internal server error"
// @Router       /accounts [post]
func CreateAccount(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid account creation data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use transaction to ensure atomicity
	var account models.Account
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		account = models.Account{
			Balance: req.InitialBalance,
		}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Error("Failed to create account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create account"})
		return
	}

	log.Info("Account created", zap.Uint("accountId", account.ID), zap.Float64("balance", account.Balance))

	c.JSON(http.StatusCreated, CreateAccountResponse{
		ID:      account.ID,
		Balance: account.Balance,
		Message: "Account created successfully",
	})
}

// GetAccountBalance returns the current balance of an account
// @Summary      Get account balance
// @Description  Returns the current balance of the specified account
// @Tags         accounts
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  map[string]interface{}  "Account balance"
// @Failure      400  {object}  map[string]string  "Invalid account ID"
// @Failure      404  {object}  map[string]string  "Account not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /accounts/{id}/balance [get]
func GetAccountBalance(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	var account models.Account
	if err := db.DB.First(&account, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		log.Error("Failed to fetch account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accountId": account.ID,
		"balance":   account.Balance,
	})
}

// GetAccountLedgerRequest represents query parameters for ledger retrieval
type GetAccountLedgerRequest struct {
	Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int `form:"offset" binding:"omitempty,min=0"`
}

// GetAccountLedger returns recent transactions for an account
// @Summary      Get account ledger
// @Description  Returns recent transactions (credits/debits) for the specified account with pagination
// @Tags         accounts
// @Produce      json
// @Param        id      path      int  true  "Account ID"
// @Param        limit   query     int  false  "Number of transactions to return (default: 20, max: 100)"
// @Param        offset  query     int  false  "Number of transactions to skip (default: 0)"
// @Success      200     {object}  map[string]interface{}  "Account ledger"
// @Failure      400     {object}  map[string]string  "Invalid account ID or query parameters"
// @Failure      404     {object}  map[string]string  "Account not found"
// @Failure      500     {object}  map[string]string  "Internal server error"
// @Router       /accounts/{id}/ledger [get]
func GetAccountLedger(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	// Verify account exists
	var account models.Account
	if err := db.DB.First(&account, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		log.Error("Failed to fetch account", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch account"})
		return
	}

	// Parse query parameters
	var query GetAccountLedgerRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	limit := query.Limit
	if limit == 0 {
		limit = 20
	}
	offset := query.Offset

	// Fetch transactions
	var transactions []models.LedgerTransaction
	var total int64

	// Count total transactions
	db.DB.Model(&models.LedgerTransaction{}).Where("account_id = ?", uint(id)).Count(&total)

	// Fetch paginated transactions
	if err := db.DB.Where("account_id = ?", uint(id)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		log.Error("Failed to fetch ledger transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch ledger"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accountId":    uint(id),
		"balance":      account.Balance,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
		"transactions": transactions,
	})
}
