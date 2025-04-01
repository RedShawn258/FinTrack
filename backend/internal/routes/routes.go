package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/handlers"
	"github.com/RedShawn258/FinTrack/backend/internal/middlewares"
)

func SetupRoutes(router *gin.Engine, logger *zap.Logger, jwtSecret string) {
	// Middleware to set logger & JWT secret in context for every request
	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("jwtSecret", jwtSecret)
		c.Next()
	})

	// Public endpoints for authentication
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", handlers.RegisterHandler)
		auth.POST("/login", handlers.LoginHandler)
		auth.POST("/reset-password", handlers.ResetPasswordHandler)
	}

	// Protected endpoints (JWT required)
	protected := router.Group("/api/v1")
	protected.Use(middlewares.AuthMiddleware())
	{
		// Profile example
		protected.GET("/profile", func(c *gin.Context) {
			userID := c.MustGet("userID")
			c.JSON(200, gin.H{
				"message": "Protected route accessed successfully",
				"userID":  userID,
			})
		})

		// Budget endpoints
		protected.POST("/budgets", handlers.CreateBudget)
		protected.GET("/budgets", handlers.GetBudgets)
		protected.PUT("/budgets/:id", handlers.UpdateBudget)
		protected.DELETE("/budgets/:id", handlers.DeleteBudget)

		// Category endpoints
		protected.POST("/categories", handlers.CreateCategory)
		protected.GET("/categories", handlers.GetCategories)
		protected.DELETE("/categories/:id", handlers.DeleteCategory)

		// Transaction endpoints
		protected.POST("/transactions", handlers.CreateTransaction)
		protected.GET("/transactions", handlers.GetTransactions)
		protected.PUT("/transactions/:id", handlers.UpdateTransaction)
		protected.DELETE("/transactions/:id", handlers.DeleteTransaction)

		// Future feature endpoints (stub implementations)
		protected.GET("/features/gamification", handlers.GamificationHandler)
		protected.GET("/features/analytics", handlers.AnalyticsHandler)
		protected.GET("/features/notifications", handlers.NotificationHandler)
	}
}
