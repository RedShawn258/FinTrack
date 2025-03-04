package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/handlers"
)

// TestCreateBudget tests the budget creation functionality
func TestCreateBudget(t *testing.T) {
	router, _ := setup()

	router.POST("/api/v1/budgets", func(c *gin.Context) {
		c.Set("userID", uint(1))
		handlers.CreateBudget(c)
	})

	// Save original DB and restore after test
	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	t.Run("Successful Budget Creation - With Category", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		// User ID that will be set in context middleware
		userID := uint(1)
		categoryID := uint(2)

		// Prepare request
		startDate := time.Now().Format("2006-01-02")
		endDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

		reqBody := handlers.CreateBudgetRequest{
			CategoryID:  &categoryID,
			LimitAmount: 1000.0,
			StartDate:   startDate,
			EndDate:     endDate,
		}
		jsonData, _ := json.Marshal(reqBody)

		// Parse dates for SQL comparison
		parsedStart, _ := time.Parse("2006-01-02", startDate)
		parsedEnd, _ := time.Parse("2006-01-02", endDate)
		start := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, time.Local)
		end := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 0, 0, 0, 0, time.Local)

		// Mock DB operations
		// 1. Check if existing budget exists
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `budgets` WHERE")).
			WithArgs(userID, categoryID, start, end, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// 2. Insert budget
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `budgets`")).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// 3. Sum transactions for the recalcBudgetRemaining function
		sumRows := sqlmock.NewRows([]string{"total"}).AddRow(0)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(amount), 0) as total FROM `transactions`")).
			WillReturnRows(sumRows)

		// 4. Update the remaining amount
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `budgets` SET")).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/budgets", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusCreated, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response content
		assert.Contains(t, response["message"], "created successfully")

		// Make sure budget is present in the response
		_, ok := response["budget"]
		assert.True(t, ok, "Response should contain budget property")

		// Verify budget fields
		assert.NotEmpty(t, fmt.Sprintf("%v", response["budget"]), "budget should not be empty")

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Successful Budget Creation - Global Budget (No Category)", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		// User ID that will be set in context middleware
		userID := uint(1)

		// Prepare request
		startDate := time.Now().Format("2006-01-02")
		endDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

		reqBody := handlers.CreateBudgetRequest{
			CategoryID:  nil,
			LimitAmount: 5000.0,
			StartDate:   startDate,
			EndDate:     endDate,
		}
		jsonData, _ := json.Marshal(reqBody)

		// Parse dates for SQL comparison
		parsedStart, _ := time.Parse("2006-01-02", startDate)
		parsedEnd, _ := time.Parse("2006-01-02", endDate)
		start := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, time.Local)
		end := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 0, 0, 0, 0, time.Local)

		// Mock DB operations
		// 1. Check if existing budget exists
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `budgets` WHERE")).
			WithArgs(userID, nil, start, end, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// 2. Insert budget
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `budgets`")).
			WillReturnResult(sqlmock.NewResult(2, 1))
		mock.ExpectCommit()

		// 3. Sum transactions for the recalcBudgetRemaining function
		sumRows := sqlmock.NewRows([]string{"total"}).AddRow(0)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(amount), 0) as total FROM `transactions`")).
			WillReturnRows(sumRows)

		// 4. Update the remaining amount
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `budgets` SET")).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/budgets", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusCreated, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response content
		assert.Contains(t, response["message"], "created successfully")

		// Make sure budget is present in the response
		_, ok := response["budget"]
		assert.True(t, ok, "Response should contain budget property")

		// Verify budget fields
		assert.NotEmpty(t, fmt.Sprintf("%v", response["budget"]), "budget should not be empty")

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Invalid Date Format", func(t *testing.T) {
		categoryID := uint(2)

		// Prepare request with invalid date format
		reqBody := handlers.CreateBudgetRequest{
			CategoryID:  &categoryID,
			LimitAmount: 1000.0,
			StartDate:   "invalid-date",
			EndDate:     time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		}
		jsonData, _ := json.Marshal(reqBody)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/budgets", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Equal(t, "Invalid start date", response["error"])
	})

	t.Run("Category Not Found", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		// User ID that will be set in context middleware
		userID := uint(1)
		nonExistentCategoryID := uint(999)

		// Prepare request
		startDate := time.Now().Format("2006-01-02")
		endDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

		reqBody := handlers.CreateBudgetRequest{
			CategoryID:  &nonExistentCategoryID,
			LimitAmount: 1000.0,
			StartDate:   startDate,
			EndDate:     endDate,
		}
		jsonData, _ := json.Marshal(reqBody)

		// Parse dates for SQL comparison
		parsedStart, _ := time.Parse("2006-01-02", startDate)
		parsedEnd, _ := time.Parse("2006-01-02", endDate)
		start := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, time.Local)
		end := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 0, 0, 0, 0, time.Local)

		// Mock DB operations
		// 1. Check if existing budget exists
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `budgets` WHERE")).
			WithArgs(userID, nonExistentCategoryID, start, end, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// 2. Insert budget
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `budgets`")).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// 3. Sum transactions for the recalcBudgetRemaining function
		sumRows := sqlmock.NewRows([]string{"total"}).AddRow(0)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(amount), 0) as total FROM `transactions`")).
			WillReturnRows(sumRows)

		// 4. Update the remaining amount
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `budgets` SET")).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/budgets", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusCreated, w.Code, "Should create budget even with non-existent category")

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response contains success message
		assert.Contains(t, response["message"], "created successfully")
	})
}

// TestGetBudgets tests retrieving budgets
func TestGetBudgets(t *testing.T) {
	router, _ := setup()
	router.GET("/api/v1/budgets", func(c *gin.Context) {
		c.Set("userID", uint(1))
		handlers.GetBudgets(c)
	})

	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	t.Run("Successfully Get All Budgets", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		userID := uint(1)

		// Create date objects for the test
		startDate, _ := time.Parse("2006-01-02", "2025-03-01")
		endDate, _ := time.Parse("2006-01-02", "2025-03-31")

		// Mock DB operations - return two budgets
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "limit_amount", "remaining_amount", "start_date", "end_date"}).
			AddRow(1, userID, 2, 1000.0, 800.0, startDate, endDate).
			AddRow(2, userID, nil, 5000.0, 4500.0, startDate, endDate)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `budgets` WHERE")).
			WithArgs(userID).
			WillReturnRows(rows)

		// Act: Send request
		req, _ := http.NewRequest("GET", "/api/v1/budgets", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response content
		budgets, ok := response["budgets"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(budgets))

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("No Budgets Found", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		userID := uint(1)

		// Mock DB operations - empty result
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "limit_amount", "remaining_amount", "start_date", "end_date"})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `budgets` WHERE")).
			WithArgs(userID).
			WillReturnRows(rows)

		// Act: Send request
		req, _ := http.NewRequest("GET", "/api/v1/budgets", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response content
		budgets, ok := response["budgets"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 0, len(budgets), "Should return empty array when no budgets found")

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestUpdateBudget tests the UpdateBudget handler
func TestUpdateBudget(t *testing.T) {
	router, _ := setup()
	router.PUT("/api/v1/budgets/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		handlers.UpdateBudget(c)
	})

	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	t.Run("Successfully Update Budget", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		userID := uint(1)
		budgetID := "1"
		categoryID := uint(3)

		// Prepare request
		startDate := time.Now().Format("2006-01-02")
		endDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

		reqBody := handlers.CreateBudgetRequest{
			CategoryID:  &categoryID,
			LimitAmount: 1500.0,
			StartDate:   startDate,
			EndDate:     endDate,
		}
		jsonData, _ := json.Marshal(reqBody)

		// Parse dates for SQL comparison
		parsedStart, _ := time.Parse("2006-01-02", startDate)
		parsedEnd, _ := time.Parse("2006-01-02", endDate)
		start := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, time.Local)
		end := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 0, 0, 0, 0, time.Local)

		// Mock DB operations
		// 1. Find the budget
		budgetRows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "limit_amount", "remaining_amount", "start_date", "end_date"}).
			AddRow(1, userID, 2, 1000.0, 800.0, start, end)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `budgets` WHERE")).
			WithArgs(budgetID, userID, 1).
			WillReturnRows(budgetRows)

		// 2. Update the budget
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `budgets` SET")).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// 3. Sum transactions for the recalcBudgetRemaining function
		sumRows := sqlmock.NewRows([]string{"total"}).AddRow(200)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(amount), 0) as total FROM `transactions`")).
			WillReturnRows(sumRows)

		// 4. Update the remaining amount
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `budgets` SET")).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// Act: Send request
		req, _ := http.NewRequest("PUT", "/api/v1/budgets/"+budgetID, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response content
		assert.Contains(t, response["message"], "updated successfully")

		// Make sure budget is present in the response
		_, ok := response["budget"]
		assert.True(t, ok, "Response should contain budget property")

		// Verify budget fields
		assert.NotEmpty(t, fmt.Sprintf("%v", response["budget"]), "budget should not be empty")

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Budget Not Found", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		userID := uint(1)
		nonExistentBudgetID := "999"

		// Prepare request
		startDate := time.Now().Format("2006-01-02")
		endDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

		reqBody := handlers.CreateBudgetRequest{
			CategoryID:  nil, // Global budget
			LimitAmount: 2000.0,
			StartDate:   startDate,
			EndDate:     endDate,
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB operations - budget not found
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `budgets` WHERE")).
			WithArgs(nonExistentBudgetID, userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Act: Send request
		req, _ := http.NewRequest("PUT", "/api/v1/budgets/"+nonExistentBudgetID, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Equal(t, "Budget not found", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Invalid Date Format", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		userID := uint(1)
		budgetID := "1"

		// Mock DB operations - find the budget first
		startDate := time.Now()
		endDate := time.Now().AddDate(0, 1, 0)
		budgetRows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "limit_amount", "remaining_amount", "start_date", "end_date"}).
			AddRow(1, userID, 2, 1000.0, 800.0, startDate, endDate)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `budgets` WHERE")).
			WithArgs(budgetID, userID, 1).
			WillReturnRows(budgetRows)

		// Prepare request with invalid date format
		reqBody := handlers.CreateBudgetRequest{
			CategoryID:  nil,
			LimitAmount: 1500.0,
			StartDate:   "invalid-date", // Invalid date format
			EndDate:     time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		}
		jsonData, _ := json.Marshal(reqBody)

		// Act: Send request
		req, _ := http.NewRequest("PUT", "/api/v1/budgets/"+budgetID, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Equal(t, "Invalid start date", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestDeleteBudget tests deleting a budget
func TestDeleteBudget(t *testing.T) {
	router, _ := setup()
	router.DELETE("/api/v1/budgets/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		handlers.DeleteBudget(c)
	})

	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	t.Run("Successfully Delete Budget", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		userID := uint(1)
		budgetID := "1"

		// Mock DB operations - GORM soft delete
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `budgets` SET `deleted_at`=?")).
			WithArgs(sqlmock.AnyArg(), budgetID, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// Act: Send request
		req, _ := http.NewRequest("DELETE", "/api/v1/budgets/"+budgetID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response message
		assert.Equal(t, "Budget deleted successfully", response["message"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Budget Not Found", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		userID := uint(1)
		nonExistentBudgetID := "999"

		// Mock DB operations for GORM soft delete with transaction
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `budgets` SET `deleted_at`=?")).
			WithArgs(sqlmock.AnyArg(), nonExistentBudgetID, userID).
			WillReturnError(fmt.Errorf("record not found"))
		mock.ExpectRollback()

		// Act: Send request
		req, _ := http.NewRequest("DELETE", "/api/v1/budgets/"+nonExistentBudgetID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Equal(t, "Budget not found or could not be deleted", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Invalid Budget ID Format", func(t *testing.T) {
		invalidBudgetID := "invalid"

		mock, err := setupDBMock()
		require.NoError(t, err)

		// Mock DB operations - GORM will try to delete but return an error
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `budgets` SET `deleted_at`=?")).
			WithArgs(sqlmock.AnyArg(), invalidBudgetID, uint(1)).
			WillReturnError(errors.New("invalid budget ID format"))
		mock.ExpectRollback()

		// Act: Send request with invalid budget ID format
		req, _ := http.NewRequest("DELETE", "/api/v1/budgets/"+invalidBudgetID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response - the handler treats any DB error as "not found"
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Equal(t, "Budget not found or could not be deleted", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
