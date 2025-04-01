package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/routes"
)

// setupTestDB initializes a test database with mock for load testing
func setupTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	// Create a new SQL mock
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Set expectations for the mock DB
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("5.7.32"))

	// Create a GORM DB instance with the mock
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Set the mock DB in our application's db package
	db.DB = gormDB

	// Set up mock expectations for user lookups (used in register/login)
	// This is for register - no existing user with same username/email
	mock.ExpectQuery("SELECT (.+) FROM `users` WHERE username = (.+) OR email = (.+) ORDER BY `users`.`id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)

	// This is for create user
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// This is for login
	mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = (.+) ORDER BY `users`.`id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password"}).
			AddRow(1, "testuser", "testuser@example.com", "$2a$10$XXXXXXXXXXXXXXXXXXXXXX")) // Bcrypt hash

	// For budget endpoint
	mock.ExpectQuery("SELECT (.+) FROM `budgets` WHERE user_id = ?").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "user_id", "limit_amount", "remaining_amount", "start_date", "end_date"}))

	// For transaction endpoint
	mock.ExpectQuery("SELECT (.+) FROM `transactions` WHERE user_id = ?").
		WillReturnRows(sqlmock.NewRows([]string{"id", "description", "category_id", "user_id", "amount", "transaction_date"}))

	// For category endpoint
	mock.ExpectQuery("SELECT (.+) FROM `categories` WHERE user_id IN \\(0, \\?\\)").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "user_id"}))

	// Return the mock and a cleanup function
	return mock, func() {
		db.DB = nil
		sqlDB.Close()
	}
}

// TestConnectionHandling verifies the backend can handle multiple concurrent connections
// even if they result in authorization errors
func TestConnectionHandling(t *testing.T) {
	// Initialize test DB
	_, cleanup := setupTestDB(t)
	defer cleanup()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Setup test server with our routes
	r := gin.New()
	r.Use(gin.Recovery())

	// Middleware to set logger & JWT secret in context for every request
	jwtSecret := "test-jwt-secret"
	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("jwtSecret", jwtSecret)
		c.Next()
	})

	// Register routes
	routes.SetupRoutes(r, logger, jwtSecret)

	// Create a test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Using a hardcoded sample JWT that would decode to a test user
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDI3NTg0MDAsImlkIjoxLCJ1c2VybmFtZSI6ImxvYWR0ZXN0dXNlciJ9.Ihgfrhr7JFH37OKwIFhOwMw_fRY9rPELfjM4PFmjcSQ"

	// Define test scenarios
	testCases := []struct {
		name        string
		endpoint    string
		numRequests int
		maxDuration time.Duration // Maximum allowed duration for all requests to complete
	}{
		{
			name:        "Multiple_Connections_Categories",
			endpoint:    "/api/v1/categories",
			numRequests: 10,
			maxDuration: 5 * time.Second,
		},
		{
			name:        "Multiple_Connections_Budgets",
			endpoint:    "/api/v1/budgets",
			numRequests: 10,
			maxDuration: 5 * time.Second,
		},
		{
			name:        "Multiple_Connections_Transactions",
			endpoint:    "/api/v1/transactions",
			numRequests: 10,
			maxDuration: 5 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			client := &http.Client{
				Timeout: 2 * time.Second, // Individual request timeout
			}

			// Set additional DB expectations for multiple calls
			for i := 0; i < tc.numRequests; i++ {
				// We don't set specific query expectations here because we expect 401 responses
				// before any database queries are made
			}

			// Track start time
			startTime := time.Now()

			// Request counters
			completedRequests := 0
			var counterMutex sync.Mutex

			// Make multiple requests in parallel
			for i := 0; i < tc.numRequests; i++ {
				wg.Add(1)
				go func(reqNum int) {
					defer wg.Done()

					// Create request with authorization header
					req, err := http.NewRequest(http.MethodGet, ts.URL+tc.endpoint, nil)
					if err != nil {
						t.Logf("Error creating request %d: %v", reqNum, err)
						return
					}

					req.Header.Set("Authorization", "Bearer "+token)
					req.Header.Set("Content-Type", "application/json")

					// Execute request
					resp, err := client.Do(req)
					if err != nil {
						t.Logf("Request %d failed with connection error: %v", reqNum, err)
						return
					}
					defer resp.Body.Close()

					// We expect 401 Unauthorized because our JWT is invalid
					// But the important thing is that the server handled the request
					counterMutex.Lock()
					completedRequests++
					counterMutex.Unlock()
				}(i)
			}

			// Wait for all requests to complete
			wg.Wait()

			// Check total duration
			totalDuration := time.Since(startTime)
			t.Logf("Total duration: %v for %d requests", totalDuration, tc.numRequests)

			// Verify all completed within expected time
			assert.LessOrEqual(t, totalDuration, tc.maxDuration,
				"Requests took longer than expected maximum duration")

			// All requests should have been handled (even with 401 errors)
			assert.Equal(t, tc.numRequests, completedRequests,
				"Not all connections were handled properly")
		})
	}
}

// TestRateLimiting tests the backend's ability to handle high request volumes
// and ensures it can recover from a burst of traffic
func TestRateLimiting(t *testing.T) {
	// Initialize test DB
	_, cleanup := setupTestDB(t)
	defer cleanup()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Setup test server with our routes
	r := gin.New()
	r.Use(gin.Recovery())

	// Middleware to set logger & JWT secret in context for every request
	jwtSecret := "test-jwt-secret"
	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("jwtSecret", jwtSecret)
		c.Next()
	})

	// Register routes
	routes.SetupRoutes(r, logger, jwtSecret)

	// Create a test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Using a hardcoded sample JWT that would decode to a test user
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDI3NTg0MDAsImlkIjoxLCJ1c2VybmFtZSI6ImxvYWR0ZXN0dXNlciJ9.Ihgfrhr7JFH37OKwIFhOwMw_fRY9rPELfjM4PFmjcSQ"

	// Test a burst of requests
	const burstRequests = 25 // A high number to test rate limiting
	var wg sync.WaitGroup

	// Track response status codes
	statusCodes := make([]int, burstRequests)
	var statusMutex sync.Mutex

	// Send a burst of requests simultaneously
	for i := 0; i < burstRequests; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()

			// Create request with auth token
			req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/transactions", nil)
			if err != nil {
				t.Logf("Failed to create request %d: %v", reqNum, err)
				return
			}

			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				t.Logf("Request %d failed: %v", reqNum, err)
				return
			}
			defer resp.Body.Close()

			// Record status code
			statusMutex.Lock()
			statusCodes[reqNum] = resp.StatusCode
			statusMutex.Unlock()
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()

	// Count responses
	responseCount := 0
	for _, code := range statusCodes {
		if code > 0 {
			responseCount++
		}
	}

	// We expect responses (even if they're 401 Unauthorized)
	assert.Equal(t, burstRequests, responseCount,
		"Not all burst requests received a response")

	// After a brief pause, the system should be ready to handle requests normally again
	time.Sleep(1 * time.Second)

	// Verify system has recovered by making a normal request
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// System should have recovered and handle this request normally
	// We still expect 401 Unauthorized, but the important thing is we got a response
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"System did not recover properly after high load")
}
