package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// transactionSetup runs before each test
func transactionSetup() (*gin.Engine, *zap.Logger) {
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()

	// Create a test logger
	logger, _ := zap.NewDevelopment()

	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("jwtSecret", "test_jwt_secret_for_unit_tests")
		c.Next()
	})

	return router, logger
}

// transactionSetupDBMock creates a new sqlmock database connection
func transactionSetupDBMock() (sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	dialector := mysql.New(mysql.Config{
		DSN:                       "sqlmock_db_0",
		DriverName:                "mysql",
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Replace the global DB instance with our mocked DB
	db.DB = gormDB

	return mock, nil
}

// TestCreateTransaction tests the transaction creation handler
func TestCreateTransaction(t *testing.T) {
	router, _ := transactionSetup()

	originalDB := db.DB
	defer func() { db.DB = originalDB }()
	mock, err := transactionSetupDBMock()
	require.NoError(t, err)

	router.POST("/transactions", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("db", db.DB)
		handlers.CreateTransaction(c)
	})

	loc, _ := time.LoadLocation("Local")
	testTime := time.Date(2023, time.January, 1, 0, 0, 0, 0, loc)

	t.Run("Successfully_Create_Transaction", func(t *testing.T) {
		// Setup mock expectations
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `transactions` \\(`user_id`,`category_id`,`amount`,`description`,`transaction_date`,`created_at`,`updated_at`,`deleted_at`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?,\\?,\\?,\\?\\)").
			WithArgs(uint(1), uint(1), 100.5, "Grocery shopping", testTime, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Prepare request body
		reqBody := `{
			"categoryId": 1,
			"amount": 100.5,
			"description": "Grocery shopping",
			"transactionDate": "2023-01-01"
		}`

		// Perform request
		req, _ := http.NewRequest("POST", "/transactions", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction created successfully", response["message"])
		assert.NotNil(t, response["transaction"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Successfully_Create_Uncategorized_Transaction", func(t *testing.T) {
		// Setup mock expectations for uncategorized transaction
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `transactions` \\(`user_id`,`category_id`,`amount`,`description`,`transaction_date`,`created_at`,`updated_at`,`deleted_at`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?,\\?,\\?,\\?\\)").
			WithArgs(uint(1), nil, 100.5, "Grocery shopping", testTime, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Prepare request body without categoryId
		reqBody := `{
			"amount": 100.5,
			"description": "Grocery shopping",
			"transactionDate": "2023-01-01"
		}`

		// Perform request
		req, _ := http.NewRequest("POST", "/transactions", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction created successfully", response["message"])
		assert.NotNil(t, response["transaction"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid_Request_Data", func(t *testing.T) {
		// Invalid request (missing required field 'amount')
		reqBody := `{
			"categoryId": 1,
			"description": "Invalid transaction",
			"transactionDate": "2023-01-01"
		}`

		// Perform request
		req, _ := http.NewRequest("POST", "/transactions", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check that the error message contains information about the validation error
		assert.Contains(t, response["error"].(string), "Amount")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid_Date_Format", func(t *testing.T) {
		reqBody := `{
			"categoryId": 1,
			"amount": 100.5,
			"description": "Invalid date",
			"transactionDate": "invalid-date"
		}`

		// Perform request
		req, _ := http.NewRequest("POST", "/transactions", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid transaction date format", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database_Error_On_Create", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `transactions` \\(`user_id`,`category_id`,`amount`,`description`,`transaction_date`,`created_at`,`updated_at`,`deleted_at`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?,\\?,\\?,\\?\\)").
			WithArgs(uint(1), uint(1), 100.5, "Grocery shopping", testTime, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		// Valid request body
		reqBody := `{
			"categoryId": 1,
			"amount": 100.5,
			"description": "Grocery shopping",
			"transactionDate": "2023-01-01"
		}`

		// Perform request
		req, _ := http.NewRequest("POST", "/transactions", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Could not create transaction", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetTransactions tests the transaction retrieval handler
func TestGetTransactions(t *testing.T) {
	router, _ := transactionSetup()

	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	mock, err := transactionSetupDBMock()
	require.NoError(t, err)

	router.GET("/transactions", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("db", db.DB)
		handlers.GetTransactions(c)
	})

	loc, _ := time.LoadLocation("Local")
	testTime := time.Date(2023, time.January, 1, 0, 0, 0, 0, loc)

	t.Run("Successfully_Get_All_Transactions", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "amount", "description", "transaction_date", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, 100.50, "Grocery shopping", testTime, testTime, testTime, nil)

		// Expect the query to get all transactions for user 1
		mock.ExpectQuery("^SELECT \\* FROM `transactions` WHERE user_id = \\? AND `transactions`.`deleted_at` IS NULL$").
			WithArgs(uint(1)).
			WillReturnRows(rows)

		// Perform request
		req, _ := http.NewRequest("GET", "/transactions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check the transactions array
		transactions, ok := response["transactions"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 1, len(transactions))

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No_Transactions_Found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "amount", "description", "transaction_date", "created_at", "updated_at", "deleted_at"})

		// Expect the query to get all transactions for user 1 but return empty
		mock.ExpectQuery("^SELECT \\* FROM `transactions` WHERE user_id = \\? AND `transactions`.`deleted_at` IS NULL$").
			WithArgs(uint(1)).
			WillReturnRows(rows)

		// Perform request
		req, _ := http.NewRequest("GET", "/transactions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check that transactions array is empty
		transactions, ok := response["transactions"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 0, len(transactions))

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database_Error", func(t *testing.T) {
		// Setup mock to return an error
		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE user_id = \\? AND `transactions`.`deleted_at` IS NULL").
			WithArgs(uint(1)).
			WillReturnError(errors.New("database error"))

		// Perform request
		req, _ := http.NewRequest("GET", "/transactions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Could not fetch transactions", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestUpdateTransaction tests the transaction update handler
func TestUpdateTransaction(t *testing.T) {
	router, _ := transactionSetup()

	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	mock, err := transactionSetupDBMock()
	require.NoError(t, err)
	router.PUT("/transactions/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("db", db.DB)
		handlers.UpdateTransaction(c)
	})

	loc, _ := time.LoadLocation("Local")
	testTime := time.Date(2023, time.January, 1, 0, 0, 0, 0, loc)

	t.Run("Successfully_Update_Transaction", func(t *testing.T) {
		// Mock finding the existing transaction
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "amount", "description", "transaction_date", "created_at", "updated_at"}).
			AddRow(1, 1, 1, 100.50, "Grocery shopping", testTime, testTime, testTime)

		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE \\(id = \\? AND user_id = \\?\\) AND `transactions`.`deleted_at` IS NULL ORDER BY `transactions`.`id` LIMIT \\?").
			WithArgs("1", uint(1), 1).
			WillReturnRows(rows)

		// Mock the update operation
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `transactions` SET").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Prepare request body for update
		reqBody := `{
			"categoryId": 2,
			"amount": 150.75,
			"description": "Updated grocery shopping",
			"transactionDate": "2023-01-01"
		}`

		// Perform request
		req, _ := http.NewRequest("PUT", "/transactions/1", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction updated successfully", response["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid_ID_Parameter", func(t *testing.T) {
		// Perform request with non-numeric ID
		req, _ := http.NewRequest("PUT", "/transactions/invalid", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction not found", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Transaction_Not_Found", func(t *testing.T) {
		// Mock no transaction found
		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE \\(id = \\? AND user_id = \\?\\) AND `transactions`.`deleted_at` IS NULL ORDER BY `transactions`.`id` LIMIT \\?").
			WithArgs("999", uint(1), 1).
			WillReturnRows(sqlmock.NewRows([]string{}))

		// Prepare request body for update
		reqBody := `{
			"categoryId": 2,
			"amount": 150.75,
			"description": "Updated transaction",
			"transactionDate": "2023-01-01"
		}`

		// Perform request
		req, _ := http.NewRequest("PUT", "/transactions/999", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction not found", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid_Request_Data", func(t *testing.T) {
		// Invalid request (missing required field 'amount')
		reqBody := `{
			"categoryId": 2,
			"description": "Invalid update",
			"transactionDate": "2023-01-01"
		}`

		// Mock finding the existing transaction
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "amount", "description", "transaction_date", "created_at", "updated_at"}).
			AddRow(1, 1, 1, 100.50, "Grocery shopping", testTime, testTime, testTime)

		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE \\(id = \\? AND user_id = \\?\\) AND `transactions`.`deleted_at` IS NULL ORDER BY `transactions`.`id` LIMIT \\?").
			WithArgs("1", uint(1), 1).
			WillReturnRows(rows)

		// Perform request
		req, _ := http.NewRequest("PUT", "/transactions/1", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// Check that the error message contains information about the validation error
		assert.Contains(t, response["error"].(string), "Amount")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database_Error_On_Update", func(t *testing.T) {
		// Mock finding the existing transaction
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "amount", "description", "transaction_date", "created_at", "updated_at"}).
			AddRow(1, 1, 1, 100.50, "Grocery shopping", testTime, testTime, testTime)

		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE \\(id = \\? AND user_id = \\?\\) AND `transactions`.`deleted_at` IS NULL ORDER BY `transactions`.`id` LIMIT \\?").
			WithArgs("1", uint(1), 1).
			WillReturnRows(rows)

		// Mock the update operation with error
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `transactions` SET").
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		// Prepare request body for update
		reqBody := `{
			"categoryId": 2,
			"amount": 150.75,
			"description": "Updated grocery shopping",
			"transactionDate": "2023-01-01"
		}`

		// Perform request
		req, _ := http.NewRequest("PUT", "/transactions/1", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Could not update transaction", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestDeleteTransaction tests the transaction deletion handler
func TestDeleteTransaction(t *testing.T) {
	router, _ := transactionSetup()

	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	mock, err := transactionSetupDBMock()
	require.NoError(t, err)

	router.DELETE("/transactions/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("db", db.DB)
		handlers.DeleteTransaction(c)
	})

	loc, _ := time.LoadLocation("Local")
	testTime := time.Date(2023, time.January, 1, 0, 0, 0, 0, loc)

	t.Run("Successfully_Delete_Transaction", func(t *testing.T) {
		// Mock finding the existing transaction
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "amount", "description", "transaction_date", "created_at", "updated_at"}).
			AddRow(1, 1, 1, 100.50, "Grocery shopping", testTime, testTime, testTime)

		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE \\(id = \\? AND user_id = \\?\\) AND `transactions`.`deleted_at` IS NULL ORDER BY `transactions`.`id` LIMIT \\?").
			WithArgs("1", uint(1), 1).
			WillReturnRows(rows)

		// Mock the delete operation
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `transactions` SET").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Perform request
		req, _ := http.NewRequest("DELETE", "/transactions/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction deleted successfully", response["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid_ID_Parameter", func(t *testing.T) {
		// Mock no transaction found for invalid ID
		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE \\(id = \\? AND user_id = \\?\\) AND `transactions`.`deleted_at` IS NULL ORDER BY `transactions`.`id` LIMIT \\?").
			WithArgs("invalid", uint(1), 1).
			WillReturnError(errors.New("invalid ID"))

		// Perform request with non-numeric ID
		req, _ := http.NewRequest("DELETE", "/transactions/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction not found", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Transaction_Not_Found", func(t *testing.T) {
		// Mock no transaction found
		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE \\(id = \\? AND user_id = \\?\\) AND `transactions`.`deleted_at` IS NULL ORDER BY `transactions`.`id` LIMIT \\?").
			WithArgs("999", uint(1), 1).
			WillReturnRows(sqlmock.NewRows([]string{}))

		// Perform request
		req, _ := http.NewRequest("DELETE", "/transactions/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction not found", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database_Error_On_Delete", func(t *testing.T) {
		// Mock finding the existing transaction
		rows := sqlmock.NewRows([]string{"id", "user_id", "category_id", "amount", "description", "transaction_date", "created_at", "updated_at"}).
			AddRow(1, 1, 1, 100.50, "Grocery shopping", testTime, testTime, testTime)

		mock.ExpectQuery("SELECT \\* FROM `transactions` WHERE \\(id = \\? AND user_id = \\?\\) AND `transactions`.`deleted_at` IS NULL ORDER BY `transactions`.`id` LIMIT \\?").
			WithArgs("1", uint(1), 1).
			WillReturnRows(rows)

		// Mock the delete operation with error
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `transactions` SET").
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		// Perform request
		req, _ := http.NewRequest("DELETE", "/transactions/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Could not delete transaction", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
