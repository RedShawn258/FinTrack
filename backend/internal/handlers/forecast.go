package handlers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// ForecastExpensesHandler handles expense forecasting based on historical data
func ForecastExpensesHandler(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	var req models.ForecastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid forecast request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request parameters
	if req.MonthsAhead <= 0 || req.MonthsAhead > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MonthsAhead must be between 1 and 12"})
		return
	}

	// Default startDate to beginning of current month if not provided
	startDate := time.Now().UTC().AddDate(0, 0, -time.Now().UTC().Day()+1)
	if req.StartDate != nil {
		startDate = *req.StartDate
	}

	// Use the global DB connection
	if db.DB == nil {
		logger.Error("Database connection not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get user's transaction history
	var transactions []models.Transaction
	query := db.DB.Where("user_id = ?", userID)

	// Filter by category if provided
	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}

	// Get transactions from the last 6 months for forecasting
	sixMonthsAgo := startDate.AddDate(0, -6, 0)
	if err := query.Where("transaction_date >= ?", sixMonthsAgo).Find(&transactions).Error; err != nil {
		logger.Error("Failed to retrieve transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction history"})
		return
	}

	// Get categories for reporting
	var categories []models.Category
	if err := db.DB.Where("user_id = ?", userID).Find(&categories).Error; err != nil {
		logger.Error("Failed to retrieve categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	// Create a map of category ID to name
	categoryMap := make(map[uint]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
	}

	// Calculate forecasts
	totalForecast := calculateTotalForecast(transactions, startDate, req.MonthsAhead)
	categoryForecasts := calculateCategoryForecasts(transactions, categoryMap, startDate, req.MonthsAhead)

	response := models.ForecastResponse{
		TotalForecast:     totalForecast,
		CategoryForecasts: categoryForecasts,
		Message:           fmt.Sprintf("Forecast generated for %d months based on your last 6 months of transactions", req.MonthsAhead),
	}

	c.JSON(http.StatusOK, response)
}

// calculateTotalForecast generates an overall expense forecast
func calculateTotalForecast(transactions []models.Transaction, startDate time.Time, monthsAhead int) []models.ForecastPoint {
	// Group transactions by month
	monthlyTotals := make(map[string]float64)

	for _, tx := range transactions {
		month := tx.TransactionDate.Format("2006-01")
		monthlyTotals[month] += tx.Amount
	}

	// Calculate average monthly spending
	var totalSpent float64
	for _, amount := range monthlyTotals {
		totalSpent += amount
	}

	// If we have no transaction history, estimate will be zero
	averageMonthlySpending := 0.0
	confidence := 0.5 // Default medium confidence

	if len(monthlyTotals) > 0 {
		averageMonthlySpending = totalSpent / float64(len(monthlyTotals))

		// Confidence increases with more data points
		confidence = math.Min(float64(len(monthlyTotals))/6.0, 1.0)
	}

	// Generate forecast points
	forecastPoints := make([]models.ForecastPoint, monthsAhead)
	for i := 0; i < monthsAhead; i++ {
		forecastDate := startDate.AddDate(0, i, 0)
		month := forecastDate.Format("2006-01")

		// For future months, add some randomness to the forecast
		// Variance increases with distance into the future
		varianceFactor := 1.0 + (float64(i) * 0.02) // 2% more variance per month
		amount := averageMonthlySpending * (1.0 + (rand() * 0.1 * varianceFactor))

		// Confidence decreases for months further in the future
		monthConfidence := confidence * math.Pow(0.95, float64(i))

		forecastPoints[i] = models.ForecastPoint{
			Month:       month,
			Amount:      math.Round(amount*100) / 100, // Round to 2 decimal places
			Probability: math.Round(monthConfidence*100) / 100,
		}
	}

	return forecastPoints
}

// calculateCategoryForecasts generates category-specific forecasts
func calculateCategoryForecasts(transactions []models.Transaction, categoryMap map[uint]string, startDate time.Time, monthsAhead int) []models.CategoryForecast {
	// Group transactions by category and month
	categoryMonthlyData := make(map[uint]map[string]float64)

	for _, tx := range transactions {
		// Skip transactions with no category
		if tx.CategoryID == nil {
			continue
		}

		catID := *tx.CategoryID
		if _, exists := categoryMonthlyData[catID]; !exists {
			categoryMonthlyData[catID] = make(map[string]float64)
		}
		month := tx.TransactionDate.Format("2006-01")
		categoryMonthlyData[catID][month] += tx.Amount
	}

	// Generate forecasts for each category
	var categoryForecasts []models.CategoryForecast

	for catID, monthlyData := range categoryMonthlyData {
		var categoryTotal float64
		for _, amount := range monthlyData {
			categoryTotal += amount
		}

		// Calculate average monthly spending for this category
		avgCategorySpending := 0.0
		if len(monthlyData) > 0 {
			avgCategorySpending = categoryTotal / float64(len(monthlyData))
		}

		// Generate forecast points for this category
		forecastPoints := make([]models.ForecastPoint, monthsAhead)
		for i := 0; i < monthsAhead; i++ {
			forecastDate := startDate.AddDate(0, i, 0)
			month := forecastDate.Format("2006-01")

			// Add some variance to each month's forecast
			varianceFactor := 1.0 + (float64(i) * 0.03) // 3% more variance per month for categories
			amount := avgCategorySpending * (1.0 + (rand() * 0.15 * varianceFactor))

			// Confidence calculation - less confidence for categories than overall
			confidence := math.Min(float64(len(monthlyData))/6.0, 0.9) // Max 90% confidence for categories
			monthConfidence := confidence * math.Pow(0.9, float64(i))

			forecastPoints[i] = models.ForecastPoint{
				Month:       month,
				Amount:      math.Round(amount*100) / 100,
				Probability: math.Round(monthConfidence*100) / 100,
			}
		}

		// Get category name from map, default to "Unknown" if not found
		categoryName := "Unknown"
		if name, exists := categoryMap[catID]; exists {
			categoryName = name
		}

		categoryForecasts = append(categoryForecasts, models.CategoryForecast{
			CategoryID:   catID,
			CategoryName: categoryName,
			Forecast:     forecastPoints,
		})
	}

	return categoryForecasts
}

// rand returns a pseudo-random number between -0.5 and 0.5
// This is a simple deterministic function for demo purposes
func rand() float64 {
	return (math.Sin(float64(time.Now().UnixNano()%1000)*0.1) * 0.5)
}
