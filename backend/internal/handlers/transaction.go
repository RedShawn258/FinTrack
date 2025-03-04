package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

type TransactionRequest struct {
	CategoryID      *uint   `json:"categoryId"` // null => uncategorized
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transactionDate" binding:"required"`
}

// recalcAllBudgetsForTransaction: find budgets that include this transaction's date/category and recalc each.
func recalcAllBudgetsForTransaction(tx models.Transaction, log *zap.Logger) {
	// Budgets that match user_id, date range covers transaction date, and category_id matches or is null for global.
	var budgets []models.Budget
	q := db.DB.Where("user_id = ? AND start_date <= ? AND end_date >= ?", tx.UserID, tx.TransactionDate, tx.TransactionDate)
	if tx.CategoryID != nil {
		q = q.Where("category_id = ?", *tx.CategoryID)
	} else {
		q = q.Where("category_id IS NULL")
	}

	if err := q.Find(&budgets).Error; err != nil {
		log.Error("Failed to find budgets for transaction recalc", zap.Error(err))
		return
	}

	for i := range budgets {
		recalcBudgetRemaining(&budgets[i], log)
	}
}

// CreateTransaction: inserts new record, then recalc budgets that might be affected.
func CreateTransaction(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid transaction creation data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction date format"})
		return
	}
	start := time.Date(txDate.Year(), txDate.Month(), txDate.Day(), 0, 0, 0, 0, time.Local)

	newTx := models.Transaction{
		UserID:          userID,
		CategoryID:      req.CategoryID,
		Amount:          req.Amount,
		Description:     req.Description,
		TransactionDate: start,
	}

	if err := db.DB.Create(&newTx).Error; err != nil {
		log.Error("Failed to create transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create transaction"})
		return
	}

	// Recalc all budgets that might include this transaction
	recalcAllBudgetsForTransaction(newTx, log)

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Transaction created successfully",
		"transaction": newTx,
	})
}

func GetTransactions(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	categoryParam := c.Query("categoryId")

	var transactions []models.Transaction
	query := db.DB.Where("user_id = ?", userID)

	if categoryParam != "" {
		query = query.Where("category_id = ?", categoryParam)
	}
	if startDate != "" {
		query = query.Where("transaction_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("transaction_date <= ?", endDate)
	}

	if err := query.Find(&transactions).Error; err != nil {
		log.Error("Failed to fetch transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// UpdateTransaction: Overwrite existing transaction, then recalc budgets for oldTx and newTx.
func UpdateTransaction(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)
	transactionID := c.Param("id")

	var existing models.Transaction
	if err := db.DB.Where("id = ? AND user_id = ?", transactionID, userID).First(&existing).Error; err != nil {
		log.Warn("Transaction not found or unauthorized", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}
	oldTx := existing

	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid transaction update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction date format"})
		return
	}
	start := time.Date(txDate.Year(), txDate.Month(), txDate.Day(), 0, 0, 0, 0, time.Local)

	// Overwrite
	existing.CategoryID = req.CategoryID
	existing.Amount = req.Amount
	existing.Description = req.Description
	existing.TransactionDate = start

	if err := db.DB.Save(&existing).Error; err != nil {
		log.Error("Failed to update transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update transaction"})
		return
	}

	// Recalc budgets for oldTx (remove its effect)
	recalcAllBudgetsForTransaction(oldTx, log)
	// Recalc budgets for newTx (apply its effect)
	recalcAllBudgetsForTransaction(existing, log)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transaction updated successfully",
		"transaction": existing,
	})
}

// DeleteTransaction: remove the record, then recalc budgets that included it.
func DeleteTransaction(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)
	transactionID := c.Param("id")

	var transaction models.Transaction
	if err := db.DB.Where("id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		log.Warn("Transaction not found or unauthorized", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if err := db.DB.Delete(&transaction).Error; err != nil {
		log.Error("Failed to delete transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete transaction"})
		return
	}

	// Recalc budgets that included this transaction
	recalcAllBudgetsForTransaction(transaction, log)

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}
