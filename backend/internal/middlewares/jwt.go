package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/services"
)

// AuthMiddleware verifies the JWT in the Authorization header.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger, _ := c.Get("logger")
		log := logger.(*zap.Logger)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn("Invalid Authorization format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization format"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// Get auth service from context to validate access token
		authService, exists := c.Get("authService")
		if !exists {
			log.Error("Auth service not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		service := authService.(*services.AuthService)

		// Validate access token using the service
		claims, err := service.ValidateAccessToken(tokenStr)
		if err != nil {
			log.Warn("Invalid or expired access token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Token is valid; set the user ID and role in context
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}
