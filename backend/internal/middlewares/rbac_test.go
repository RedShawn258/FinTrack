package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupRBACTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		allowedRoles   []string
		expectedStatus int
		expectedError  string
		description    string
	}{
		{
			name:           "admin accessing admin route",
			userRole:       "admin",
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusOK,
			description:    "should allow admin to access admin route",
		},
		{
			name:           "user accessing admin route",
			userRole:       "user",
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Forbidden: insufficient permissions",
			description:    "should deny user access to admin route",
		},
		{
			name:           "user accessing user route",
			userRole:       "user",
			allowedRoles:   []string{"user"},
			expectedStatus: http.StatusOK,
			description:    "should allow user to access user route",
		},
		{
			name:           "admin accessing user route",
			userRole:       "admin",
			allowedRoles:   []string{"user"},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Forbidden: insufficient permissions",
			description:    "should deny admin access to user-only route",
		},
		{
			name:           "empty role defaults to user",
			userRole:       "",
			allowedRoles:   []string{"user"},
			expectedStatus: http.StatusOK,
			description:    "should default empty role to user",
		},
		{
			name:           "empty role accessing admin route",
			userRole:       "",
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Forbidden: insufficient permissions",
			description:    "should deny empty role access to admin route",
		},
		{
			name:           "multiple allowed roles - user matches",
			userRole:       "user",
			allowedRoles:   []string{"admin", "user"},
			expectedStatus: http.StatusOK,
			description:    "should allow access when user role matches one of allowed roles",
		},
		{
			name:           "multiple allowed roles - admin matches",
			userRole:       "admin",
			allowedRoles:   []string{"admin", "user"},
			expectedStatus: http.StatusOK,
			description:    "should allow access when admin role matches one of allowed roles",
		},
		{
			name:           "role not in allowed list",
			userRole:       "moderator",
			allowedRoles:   []string{"admin", "user"},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Forbidden: insufficient permissions",
			description:    "should deny access when role is not in allowed list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRBACTestRouter()
			logger := zap.NewNop()

			// Set up test route with RequireRole middleware
			router.Use(func(c *gin.Context) {
				c.Set("logger", logger)
			})
			router.Use(func(c *gin.Context) {
				c.Set("userRole", tt.userRole)
			})
			router.Use(RequireRole(tt.allowedRoles...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError, tt.description)
			}
		})
	}
}

func TestRequireRole_MissingRole(t *testing.T) {
	router := setupRBACTestRouter()
	logger := zap.NewNop()

	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		// Don't set userRole to simulate missing role
	})
	router.Use(RequireRole("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Forbidden: insufficient permissions")
}

func TestRequireRole_InvalidRoleType(t *testing.T) {
	router := setupRBACTestRouter()
	logger := zap.NewNop()

	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("userRole", 123) // Invalid type (not string)
	})
	router.Use(RequireRole("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
}

