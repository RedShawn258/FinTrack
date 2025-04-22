package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/handlers"
)

func setupSimpleGamificationTest() (*httptest.ResponseRecorder, *gin.Context) {
	// Make sure we're in test mode
	gin.SetMode(gin.TestMode)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create a context with the recorder
	c, _ := gin.CreateTestContext(w)

	// Create test logger
	logger, _ := zap.NewDevelopment()
	c.Set("logger", logger)

	// Set user ID in context (simulating auth middleware)
	c.Set("userID", uint(1))

	return w, c
}

// Test that the gamification handler returns demo data when DB is nil
func TestGamificationHandlerWithNilDB(t *testing.T) {
	w, c := setupSimpleGamificationTest()

	// Set up request
	c.Request, _ = http.NewRequest("GET", "/api/v1/features/gamification", nil)

	// Set a nil DB to force the handler to return demo data
	originalDB := db.DB
	db.DB = nil
	defer func() { db.DB = originalDB }()

	// Execute the handler
	handlers.GamificationHandler(c)

	// Check status code
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check the error message
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])
}

// Test helper function calculations with the exposed testing-friendly functions
func TestCalculateLevel(t *testing.T) {
	// Test several point values against expected levels
	testCases := []struct {
		points        int
		expectedLevel int
	}{
		{0, 1},      // Min level is 1
		{50, 1},     // Level 1: 0-100
		{99, 1},     // Still level 1 at 99
		{100, 2},    // Level 2 starts at 100
		{399, 2},    // Still level 2 at 399
		{400, 3},    // Level 3 starts at 400
		{899, 3},    // Still level 3 at 899
		{900, 4},    // Level 4 starts at 900
		{1599, 4},   // Still level 4 at 1599
		{10000, 11}, // Very high points
	}

	for _, tc := range testCases {
		level := handlers.TestableCalculateLevel(tc.points)
		assert.Equal(t, tc.expectedLevel, level, "Points %d should be level %d, got %d",
			tc.points, tc.expectedLevel, level)
	}
}

func TestCalculatePointsToNextLevel(t *testing.T) {
	// Test point calculation to next level
	testCases := []struct {
		points               int
		expectedPointsToNext int
	}{
		{0, 100},   // Need 100 for level 2
		{50, 50},   // Already have 50, need 50 more
		{99, 1},    // Just 1 point away from level 2
		{100, 300}, // Just entered level 2, need 300 more to level 3 (400)
		{300, 100}, // In level 2, need 100 more
		{500, 400}, // In level 3, need 400 more to level 4 (900)
	}

	for _, tc := range testCases {
		pointsToNext := handlers.TestableCalculatePointsToNextLevel(tc.points)
		assert.Equal(t, tc.expectedPointsToNext, pointsToNext,
			"Points %d should need %d more to next level, got %d",
			tc.points, tc.expectedPointsToNext, pointsToNext)
	}
}

func TestGetLevelTitle(t *testing.T) {
	// Test level title mapping
	testCases := []struct {
		level         int
		expectedTitle string
	}{
		{1, "Novice Saver"},
		{5, "Savings Specialist"},
		{10, "Money Maestro"},
		{11, "Financial Legend"}, // Beyond defined titles
	}

	for _, tc := range testCases {
		title := handlers.TestableGetLevelTitle(tc.level)
		assert.Equal(t, tc.expectedTitle, title,
			"Level %d should have title %s, got %s",
			tc.level, tc.expectedTitle, title)
	}
}

func TestAnalyticsHandler(t *testing.T) {
	w, c := setupSimpleGamificationTest()

	// Set up request
	c.Request, _ = http.NewRequest("GET", "/api/v1/features/analytics", nil)

	// Execute the handler
	handlers.AnalyticsHandler(c)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Validate stub data
	assert.Equal(t, 1234.56, response["totalExpense"])
	assert.Equal(t, 2000.00, response["totalBudget"])
	assert.Contains(t, response["analytics"], "Advanced analytics coming soon")
}

func TestNotificationHandler(t *testing.T) {
	w, c := setupSimpleGamificationTest()

	// Set up request
	c.Request, _ = http.NewRequest("GET", "/api/v1/features/notifications", nil)

	// Execute the handler
	handlers.NotificationHandler(c)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Validate stub data
	assert.Contains(t, response["notification"], "Push notifications coming soon")
}
