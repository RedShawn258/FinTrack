package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GamificationHandler is a stub endpoint for sprint3 gamification features.
// It returns dummy gamification data.
func GamificationHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Gamification features under development",
		"badges":  []string{"Novice Saver", "Budget Beginner", "Spending Strategist"},
	})
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
