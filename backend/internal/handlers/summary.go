package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/cache"
	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

type TransactionSummary struct {
	TotalAmount   float64 `json:"totalAmount"`
	Count         int     `json:"count"`
	StartDate     string  `json:"startDate,omitempty"`
	EndDate       string  `json:"endDate,omitempty"`
	CategoryBreakdown []CategoryAmount `json:"categoryBreakdown,omitempty"`
}

type CategoryAmount struct {
	CategoryID   *uint   `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Amount       float64 `json:"amount"`
	Count        int     `json:"count"`
}

type BudgetSummary struct {
	Month           string  `json:"month"` // YYYY-MM format
	TotalBudget     float64 `json:"totalBudget"`
	TotalSpent      float64 `json:"totalSpent"`
	TotalRemaining  float64 `json:"totalRemaining"`
	BudgetCount     int     `json:"budgetCount"`
	Budgets         []models.Budget `json:"budgets,omitempty"`
}

// GetTransactionSummary returns a summary of transactions with optional date filtering
// @Summary      Get transaction summary
// @Description  Get summary statistics for user's transactions (total amount, count, category breakdown) with optional date filtering
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        startDate    query     string  false  "Start date (YYYY-MM-DD)"
// @Param        endDate      query     string  false  "End date (YYYY-MM-DD)"
// @Success      200          {object}  TransactionSummary  "Transaction summary"
// @Failure      401          {object}  map[string]string  "Unauthorized"
// @Failure      500          {object}  map[string]string  "Internal server error"
// @Router       /transactions/summary [get]
func GetTransactionSummary(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Generate cache key
	cacheKey := cache.GenerateKey("tx:summary", userID, startDate, endDate)

	// Try to get from cache
	var cacheService *cache.CacheService
	if cs, exists := c.Get("cacheService"); exists && cs != nil {
		cacheService = cs.(*cache.CacheService)
	}

	ctx := context.Background()
	var summary TransactionSummary

	if cacheService != nil && cacheService.IsEnabled() {
		if found, err := cacheService.Get(ctx, cacheKey, &summary); err == nil && found {
			log.Info("Cache hit for transaction summary", zap.String("key", cacheKey))
			c.JSON(http.StatusOK, summary)
			return
		}
		log.Debug("Cache miss for transaction summary", zap.String("key", cacheKey))
	}

	// Build query
	query := db.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if startDate != "" {
		query = query.Where("transaction_date >= ?", startDate)
		summary.StartDate = startDate
	}
	if endDate != "" {
		query = query.Where("transaction_date <= ?", endDate)
		summary.EndDate = endDate
	}

	// Get total amount and count
	var totalResult struct {
		Total float64
		Count int
	}
	if err := query.Select("COALESCE(SUM(amount), 0) as total, COUNT(*) as count").
		Scan(&totalResult).Error; err != nil {
		log.Error("Failed to calculate transaction summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not calculate summary"})
		return
	}

	summary.TotalAmount = totalResult.Total
	summary.Count = totalResult.Count

	// Get category breakdown
	var categoryResults []struct {
		CategoryID   *uint
		CategoryName string
		Amount       float64
		Count        int
	}

	categoryQuery := db.DB.Table("transactions").
		Select("transactions.category_id, COALESCE(categories.name, 'Uncategorized') as category_name, COALESCE(SUM(transactions.amount), 0) as amount, COUNT(*) as count").
		Joins("LEFT JOIN categories ON transactions.category_id = categories.id").
		Where("transactions.user_id = ? AND transactions.deleted_at IS NULL", userID).
		Group("transactions.category_id, categories.name")

	if startDate != "" {
		categoryQuery = categoryQuery.Where("transactions.transaction_date >= ?", startDate)
	}
	if endDate != "" {
		categoryQuery = categoryQuery.Where("transactions.transaction_date <= ?", endDate)
	}

	if err := categoryQuery.Scan(&categoryResults).Error; err != nil {
		log.Warn("Failed to get category breakdown", zap.Error(err))
		// Continue without category breakdown
	} else {
		summary.CategoryBreakdown = make([]CategoryAmount, len(categoryResults))
		for i, cr := range categoryResults {
			summary.CategoryBreakdown[i] = CategoryAmount{
				CategoryID:   cr.CategoryID,
				CategoryName: cr.CategoryName,
				Amount:       cr.Amount,
				Count:        cr.Count,
			}
		}
	}

	// Cache the result
	if cacheService != nil && cacheService.IsEnabled() {
		if err := cacheService.Set(ctx, cacheKey, summary, cache.DefaultTTL); err != nil {
			log.Warn("Failed to cache transaction summary", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, summary)
}

// GetBudgetSummary returns a summary of budgets for a specific month
// @Summary      Get budget summary by month
// @Description  Get summary of all budgets for a specific month (YYYY-MM format)
// @Tags         budgets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        month        path      string  true   "Month in YYYY-MM format"
// @Success      200          {object}  BudgetSummary  "Budget summary"
// @Failure      400          {object}  map[string]string  "Invalid month format"
// @Failure      401          {object}  map[string]string  "Unauthorized"
// @Failure      500          {object}  map[string]string  "Internal server error"
// @Router       /budgets/{month} [get]
func GetBudgetSummary(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	month := c.Param("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Month parameter is required (YYYY-MM format)"})
		return
	}

	// Validate month format (YYYY-MM)
	_, err := time.Parse("2006-01", month)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month format. Use YYYY-MM"})
		return
	}

	// Generate cache key
	cacheKey := cache.GenerateKey("budget:summary", userID, month)

	// Try to get from cache
	var cacheService *cache.CacheService
	if cs, exists := c.Get("cacheService"); exists && cs != nil {
		cacheService = cs.(*cache.CacheService)
	}

	ctx := context.Background()
	var summary BudgetSummary

	if cacheService != nil && cacheService.IsEnabled() {
		if found, err := cacheService.Get(ctx, cacheKey, &summary); err == nil && found {
			log.Info("Cache hit for budget summary", zap.String("key", cacheKey))
			c.JSON(http.StatusOK, summary)
			return
		}
		log.Debug("Cache miss for budget summary", zap.String("key", cacheKey))
	}

	// Calculate start and end of month
	monthTime, _ := time.Parse("2006-01", month)
	startOfMonth := time.Date(monthTime.Year(), monthTime.Month(), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Get budgets that overlap with this month
	var budgets []models.Budget
	if err := db.DB.Where("user_id = ? AND start_date <= ? AND end_date >= ?", 
		userID, endOfMonth, startOfMonth).Find(&budgets).Error; err != nil {
		log.Error("Failed to fetch budgets for summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch budget summary"})
		return
	}

	summary.Month = month
	summary.BudgetCount = len(budgets)
	summary.Budgets = budgets

	// Calculate totals
	for _, budget := range budgets {
		summary.TotalBudget += budget.LimitAmount
		summary.TotalSpent += (budget.LimitAmount - budget.RemainingAmount)
		summary.TotalRemaining += budget.RemainingAmount
	}

	// Cache the result
	if cacheService != nil && cacheService.IsEnabled() {
		if err := cacheService.Set(ctx, cacheKey, summary, cache.DefaultTTL); err != nil {
			log.Warn("Failed to cache budget summary", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, summary)
}




