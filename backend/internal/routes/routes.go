package routes

// the routes file

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/handlers"
	"github.com/RedShawn258/FinTrack/backend/internal/middlewares"
)

// SetupRoutes defines all endpoints in the application.
func SetupRoutes(router *gin.Engine, logger *zap.Logger, jwtSecret string) {
	// A middleware to set logger & JWT secret in context for every request
	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("jwtSecret", jwtSecret)
		c.Next()
	})

	// Public endpoints
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
		// Example: GET /api/v1/profile -> returns user profile info
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID") // from AuthMiddleware
			c.JSON(200, gin.H{
				"message": "Protected route accessed successfully",
				"userID":  userID,
			})
		})
	}
}
