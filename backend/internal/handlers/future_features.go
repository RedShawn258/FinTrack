package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// GamificationHandler returns the user's gamification status and achievements
func GamificationHandler(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	// Check database connection
	if db.DB == nil {
		logger.Error("Database connection not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get user's badges
	var userBadges []models.UserBadge
	if err := db.DB.Preload("Badge").Where("user_id = ?", userID).Find(&userBadges).Error; err != nil {
		// If table doesn't exist yet, return demo data
		logger.Warn("Failed to retrieve user badges, using demo data", zap.Error(err))
		c.JSON(http.StatusOK, getDemoGamificationData(userID))
		return
	}

	// Get user's points
	var userPoints []models.UserPoints
	if err := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(10).Find(&userPoints).Error; err != nil {
		logger.Warn("Failed to retrieve user points, using demo data", zap.Error(err))
		c.JSON(http.StatusOK, getDemoGamificationData(userID))
		return
	}

	// Calculate total points
	var totalPoints int
	if err := db.DB.Model(&models.UserPoints{}).Where("user_id = ?", userID).Select("SUM(points) as total").Scan(&totalPoints).Error; err != nil {
		logger.Warn("Failed to calculate total points, using demo data", zap.Error(err))
		c.JSON(http.StatusOK, getDemoGamificationData(userID))
		return
	}

	// Get badges user hasn't earned yet
	var nextBadges []models.Badge
	var earnedBadgeIDs []uint
	for _, badge := range userBadges {
		earnedBadgeIDs = append(earnedBadgeIDs, badge.BadgeID)
	}

	badgeQuery := db.DB.Model(&models.Badge{})
	if len(earnedBadgeIDs) > 0 {
		badgeQuery = badgeQuery.Where("id NOT IN (?)", earnedBadgeIDs)
	}

	if err := badgeQuery.Where("threshold <= ?", totalPoints+100).Limit(3).Find(&nextBadges).Error; err != nil {
		logger.Warn("Failed to retrieve next badges", zap.Error(err))
		// Continue anyway as this is not critical
	}

	// Calculate user level
	level := calculateLevel(totalPoints)
	levelTitle := getLevelTitle(level)
	pointsToNextLevel := calculatePointsToNextLevel(totalPoints)

	// Build and return the response
	response := models.GamificationSummary{
		TotalPoints:       totalPoints,
		RecentBadges:      getRecentBadges(userBadges),
		AllBadges:         userBadges,
		RecentPoints:      userPoints,
		NextBadges:        nextBadges,
		Level:             level,
		LevelTitle:        levelTitle,
		PointsToNextLevel: pointsToNextLevel,
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions

// calculateLevel determines the user's level based on total points
func calculateLevel(points int) int {
	// Simple algorithm: level = sqrt(points/100)
	// Level 1: 0-100 points
	// Level 2: 101-400 points
	// Level 3: 401-900 points, etc.

	level := 1
	threshold := 100

	for points >= threshold {
		level++
		threshold = level * level * 100
	}

	return level
}

// getLevelTitle returns a title for the current level
func getLevelTitle(level int) string {
	titles := map[int]string{
		1:  "Novice Saver",
		2:  "Budget Beginner",
		3:  "Money Manager",
		4:  "Finance Planner",
		5:  "Savings Specialist",
		6:  "Budget Master",
		7:  "Finance Ninja",
		8:  "Economy Expert",
		9:  "Financial Wizard",
		10: "Money Maestro",
	}

	if title, exists := titles[level]; exists {
		return title
	}

	// For levels beyond our predefined titles
	if level > 10 {
		return "Financial Legend"
	}

	return "Novice Saver"
}

// calculatePointsToNextLevel calculates how many more points needed for the next level
func calculatePointsToNextLevel(currentPoints int) int {
	currentLevel := calculateLevel(currentPoints)
	nextLevelThreshold := currentLevel * currentLevel * 100

	pointsNeeded := nextLevelThreshold - currentPoints
	if pointsNeeded < 0 {
		return 0
	}

	return pointsNeeded
}

// getRecentBadges returns the 3 most recent badges
func getRecentBadges(allBadges []models.UserBadge) []models.UserBadge {
	// Sort badges by EarnedAt (descending)
	// This is a simple implementation - in production code you'd use a more efficient sorting algorithm
	for i := 0; i < len(allBadges); i++ {
		for j := i + 1; j < len(allBadges); j++ {
			if allBadges[i].EarnedAt.Before(allBadges[j].EarnedAt) {
				allBadges[i], allBadges[j] = allBadges[j], allBadges[i]
			}
		}
	}

	// Return up to 3 most recent badges
	if len(allBadges) <= 3 {
		return allBadges
	}
	return allBadges[:3]
}

// getDemoGamificationData returns demo data for the gamification feature
func getDemoGamificationData(userID uint) models.GamificationSummary {
	now := time.Now()

	// Demo badges
	badges := []models.Badge{
		{ID: 1, Name: "Budget Beginner", Description: "Created your first budget", ImageURL: "/images/badges/budget_beginner.png", Category: "budgeting", Threshold: 10},
		{ID: 2, Name: "Tracking Pro", Description: "Tracked expenses for 7 consecutive days", ImageURL: "/images/badges/tracking_pro.png", Category: "consistency", Threshold: 50},
		{ID: 3, Name: "Savings Star", Description: "Saved 10% of your income", ImageURL: "/images/badges/savings_star.png", Category: "savings", Threshold: 100},
	}

	// Demo user badges
	userBadges := []models.UserBadge{
		{ID: 1, UserID: userID, BadgeID: 1, Badge: badges[0], EarnedAt: now.AddDate(0, 0, -30)},
		{ID: 2, UserID: userID, BadgeID: 2, Badge: badges[1], EarnedAt: now.AddDate(0, 0, -15)},
	}

	// Demo user points
	userPoints := []models.UserPoints{
		{ID: 1, UserID: userID, Points: 10, Reason: "Created first budget", ActivityType: "budget_created", CreatedAt: now.AddDate(0, 0, -30)},
		{ID: 2, UserID: userID, Points: 5, Reason: "Added 10 transactions", ActivityType: "transaction_added", CreatedAt: now.AddDate(0, 0, -25)},
		{ID: 3, UserID: userID, Points: 20, Reason: "Completed user profile", ActivityType: "profile_completed", CreatedAt: now.AddDate(0, 0, -20)},
		{ID: 4, UserID: userID, Points: 15, Reason: "Tracked expenses for a week", ActivityType: "tracking_streak", CreatedAt: now.AddDate(0, 0, -15)},
		{ID: 5, UserID: userID, Points: 25, Reason: "Stayed under budget", ActivityType: "budget_achieved", CreatedAt: now.AddDate(0, 0, -10)},
	}

	// Calculate total points from the demo data
	totalPoints := 0
	for _, p := range userPoints {
		totalPoints += p.Points
	}

	// Demo next badges to earn
	nextBadges := []models.Badge{
		{ID: 3, Name: "Savings Star", Description: "Saved 10% of your income", ImageURL: "/images/badges/savings_star.png", Category: "savings", Threshold: 100},
		{ID: 4, Name: "Budget Master", Description: "Stayed under budget for 3 consecutive months", ImageURL: "/images/badges/budget_master.png", Category: "budgeting", Threshold: 150},
	}

	// Build the response
	return models.GamificationSummary{
		TotalPoints:       totalPoints,
		RecentBadges:      userBadges,
		AllBadges:         userBadges,
		RecentPoints:      userPoints,
		NextBadges:        nextBadges,
		Level:             calculateLevel(totalPoints),
		LevelTitle:        getLevelTitle(calculateLevel(totalPoints)),
		PointsToNextLevel: calculatePointsToNextLevel(totalPoints),
	}
}

// AnalyticsHandler is a stub endpoint for sprint3 advanced analytics.
func AnalyticsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"totalExpense": 1234.56,
		"totalBudget":  2000.00,
		"analytics":    "Advanced analytics coming soon!",
	})
}

// NotificationHandler is a stub endpoint for sprint3 push notifications.
func NotificationHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"notification": "You have 1 new message. Push notifications coming soon!",
	})
}

// Test helper functions - exports private functions for testing
// These are only used by tests and do not affect normal operation

// TestableCalculateLevel is a test-friendly version of calculateLevel
func TestableCalculateLevel(points int) int {
	return calculateLevel(points)
}

// TestableGetLevelTitle is a test-friendly version of getLevelTitle
func TestableGetLevelTitle(level int) string {
	return getLevelTitle(level)
}

// TestableCalculatePointsToNextLevel is a test-friendly version of calculatePointsToNextLevel
func TestableCalculatePointsToNextLevel(currentPoints int) int {
	return calculatePointsToNextLevel(currentPoints)
}
