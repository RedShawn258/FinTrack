package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequireRole returns a middleware that checks if the user's role is in the allowed roles list
// Must be used after AuthMiddleware which sets userRole in context
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger, _ := c.Get("logger")
		log := logger.(*zap.Logger)

		// Get user role from context (set by AuthMiddleware)
		userRoleInterface, exists := c.Get("userRole")
		if !exists {
			log.Warn("User role not found in context")
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}

		userRole, ok := userRoleInterface.(string)
		if !ok {
			log.Error("Invalid user role type in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// Default to "user" if role is empty
		if userRole == "" {
			userRole = "user"
		}

		// Check if user's role is in the allowed roles list
		allowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			log.Warn("Access denied: insufficient role",
				zap.String("userRole", userRole),
				zap.Strings("allowedRoles", allowedRoles))
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: insufficient permissions",
				"requiredRoles": allowedRoles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

