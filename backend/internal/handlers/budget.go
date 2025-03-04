package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

type CreateBudgetRequest struct {
	CategoryID  *uint   `json:"categoryId"` // nullable => global if null
	LimitAmount float64 `json:"limitAmount" binding:"required,gt=0"`
	StartDate   string  `json:"startDate" binding:"required"`
	EndDate     string  `json:"endDate" binding:"required"`
}

// recalcBudgetRemaining sums all transactions in the budget's date range & category
// and sets remaining_amount = limit_amount - total_spent
func recalcBudgetRemaining(budget *models.Budget, log *zap.Logger) error {
	var sumResult struct {
		Total float64
	}

	//query to sum transactions in [start_date, end_date] for partcular user & specific category
	query := db.DB.Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND transaction_date >= ? AND transaction_date <= ? AND deleted_at IS NULL",
			budget.UserID, budget.StartDate, budget.EndDate)

	if budget.CategoryID != nil {
		query = query.Where("category_id = ?", *budget.CategoryID)
	} else {
		query = query.Where("category_id IS NULL")
	}

	if err := query.Scan(&sumResult).Error; err != nil {
		log.Error("Failed to sum transactions for budget recalc", zap.Error(err))
		return err
	}

	budget.RemainingAmount = budget.LimitAmount - sumResult.Total
	if err := db.DB.Save(&budget).Error; err != nil {
		log.Error("Failed to update budget remaining_amount", zap.Error(err))
		return err
	}
	return nil
}

// CreateBudget either overwrites an existing budget if (user_id, category_id, start_date, end_date)
// matches, or creates a new record. Then we recalc remaining_amount from transactions.
func CreateBudget(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	var req CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid budget creation data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the date, force midnight local time to avoid offsets
	parsedStart, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date"})
		return
	}
	start := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, time.Local)

	parsedEnd, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date"})
		return
	}
	end := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 0, 0, 0, 0, time.Local)

	if end.Before(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	// Check for an existing budget
	var existing models.Budget
	findErr := db.DB.Where("user_id = ? AND category_id <=> ? AND start_date = ? AND end_date = ?",
		userID, req.CategoryID, start, end).First(&existing).Error

	if findErr == nil {
		// Overwrite existing record
		existing.LimitAmount = req.LimitAmount
		existing.StartDate = start
		existing.EndDate = end

		if err := db.DB.Save(&existing).Error; err != nil {
			log.Error("Failed to overwrite existing budget", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not overwrite budget"})
			return
		}
		if err := recalcBudgetRemaining(&existing, log); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalc budget"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Budget already exists, overwriting.",
			"budget":  existing,
		})
		return
	}

	if errors.Is(findErr, gorm.ErrRecordNotFound) {
		// Create new budget
		newBudget := models.Budget{
			UserID:      userID,
			CategoryID:  req.CategoryID,
			LimitAmount: req.LimitAmount,
			StartDate:   start,
			EndDate:     end,
		}
		if err := db.DB.Create(&newBudget).Error; err != nil {
			log.Error("Failed to create new budget", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create budget"})
			return
		}
		if err := recalcBudgetRemaining(&newBudget, log); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalc budget"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "Budget created successfully",
			"budget":  newBudget,
		})
		return
	}

	log.Error("Database error checking existing budget", zap.Error(findErr))
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
}

// GetBudgets lists all budgets for the authenticated user
func GetBudgets(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	var budgets []models.Budget
	if err := db.DB.Where("user_id = ?", userID).Find(&budgets).Error; err != nil {
		log.Error("Failed to fetch budgets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch budgets"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"budgets": budgets})
}

// UpdateBudget modifies the budget's limit or date range, then recalculates remaining_amount.
func UpdateBudget(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)
	budgetID := c.Param("id")

	var existing models.Budget
	if err := db.DB.Where("id = ? AND user_id = ?", budgetID, userID).First(&existing).Error; err != nil {
		log.Warn("Budget not found or unauthorized", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	var req CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid budget update data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedStart, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date"})
		return
	}
	start := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, time.Local)

	parsedEnd, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date"})
		return
	}
	end := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 0, 0, 0, 0, time.Local)

	if end.Before(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	existing.CategoryID = req.CategoryID
	existing.LimitAmount = req.LimitAmount
	existing.StartDate = start
	existing.EndDate = end

	if err := db.DB.Save(&existing).Error; err != nil {
		log.Error("Failed to update budget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update budget"})
		return
	}
	// Recalculate after updating
	if err := recalcBudgetRemaining(&existing, log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalc budget"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Budget updated successfully",
		"budget":  existing,
	})
}

// DeleteBudget removes a budget owned by the authenticated user
func DeleteBudget(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)
	budgetID := c.Param("id")

	if err := db.DB.Where("id = ? AND user_id = ?", budgetID, userID).Delete(&models.Budget{}).Error; err != nil {
		log.Error("Failed to delete budget", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found or could not be deleted"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted successfully"})
}

// RecalculateAllBudgets recalculates remaining amounts for all budgets in the system.
// This is called during server startup to ensure budget amounts are accurate.
func RecalculateAllBudgets(logger *zap.Logger) error {
	var budgets []models.Budget
	if err := db.DB.Find(&budgets).Error; err != nil {
		logger.Error("Failed to fetch budgets for recalculation", zap.Error(err))
		return err
	}

	for i := range budgets {
		if err := recalcBudgetRemaining(&budgets[i], logger); err != nil {
			logger.Error("Failed to recalculate budget",
				zap.Error(err),
				zap.Uint("budgetID", budgets[i].ID),
				zap.Uint("userID", budgets[i].UserID))
			continue
		}
	}

	logger.Info("Successfully recalculated all budget remaining amounts")
	return nil
}
