package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/handlers"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

func setupSimpleForecastTest() (*httptest.ResponseRecorder, *gin.Context) {
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

// Tests that validate the API behavior without relying on specific DB implementation details
func TestForecastExpensesHandlerBasic(t *testing.T) {
	w, c := setupSimpleForecastTest()

	// Set up request with months ahead = 3
	reqBody := models.ForecastRequest{
		MonthsAhead: 3,
	}
	reqJSON, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/forecast/expenses", bytes.NewBuffer(reqJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set a nil DB to force the handler to return an internal server error
	originalDB := db.DB
	db.DB = nil
	defer func() { db.DB = originalDB }()

	// Execute the handler
	handlers.ForecastExpensesHandler(c)

	// When DB is nil, we should get a 500 error
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check the error message
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])
}

func TestForecastExpensesHandler_InvalidRequest(t *testing.T) {
	// Test cases for invalid requests
	testCases := []struct {
		name           string
		monthsAhead    int
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "MonthsAhead too low",
			monthsAhead:    0,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MonthsAhead must be between 1 and 12",
		},
		{
			name:           "MonthsAhead too high",
			monthsAhead:    13,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MonthsAhead must be between 1 and 12",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new setup for each test case
			w, c := setupSimpleForecastTest()

			// Set up request
			reqBody := map[string]interface{}{
				"monthsAhead": tc.monthsAhead,
			}
			reqJSON, _ := json.Marshal(reqBody)
			c.Request, _ = http.NewRequest("POST", "/api/v1/forecast/expenses", bytes.NewBuffer(reqJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			// Execute the handler
			handlers.ForecastExpensesHandler(c)

			// Check status code
			assert.Equal(t, tc.expectedStatus, w.Code, fmt.Sprintf("Test case %s failed", tc.name))

			// Verify error response
			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"], tc.expectedError)
		})
	}
}
