package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/handlers"
)

// Setup runs before each test
func setup() (*gin.Engine, *zap.Logger) {

	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger, _ := zap.NewDevelopment()

	// Set up middleware to include logger and JWT secret
	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("jwtSecret", "test_jwt_secret_for_unit_tests")
		c.Next()
	})

	return router, logger
}

// setupDBMock creates a new sqlmock database connection
func setupDBMock() (sqlmock.Sqlmock, error) {
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

// TestRegisterHandler tests the user registration functionality
func TestRegisterHandler(t *testing.T) {
	router, _ := setup()
	router.POST("/api/v1/auth/register", handlers.RegisterHandler)

	// Save original DB and restore after test
	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	t.Run("Successful Registration", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		reqBody := handlers.RegistrationRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB operations
		// 1. Check for existing user
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE")).
			WillReturnError(gorm.ErrRecordNotFound)

		// 2. Insert new user (Create operation)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
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
		assert.Equal(t, "User registration successful", response["message"])
		user, ok := response["user"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(1), user["id"])
		assert.Equal(t, "testuser", user["username"])
		assert.Equal(t, "test@example.com", user["email"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Duplicate Username", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		reqBody := handlers.RegistrationRequest{
			Username: "existinguser",
			Email:    "existing@example.com",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB to return an existing user
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(1, "existinguser", "existing@example.com", "hashedpassword")
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE")).
			WillReturnRows(rows)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusConflict, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Equal(t, "Username or email already in use", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Invalid Request Data", func(t *testing.T) {

		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Response should contain an error
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}

// TestLoginHandler tests the login functionality
func TestLoginHandler(t *testing.T) {
	router, _ := setup()
	router.POST("/api/v1/auth/login", handlers.LoginHandler)

	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	t.Run("Successful Login with Username", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		// Create hashed password for test user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		// Prepare request
		reqBody := handlers.LoginRequest{
			Identifier: "testuser",
			Password:   "password123",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB to return a user for the query
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(1, "testuser", "test@example.com", string(hashedPassword))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE")).
			WillReturnRows(rows)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
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
		assert.Equal(t, "Login successful", response["message"])
		assert.NotEmpty(t, response["token"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Successful Login with Email", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		// Create hashed password for test user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		// Prepare request
		reqBody := handlers.LoginRequest{
			Identifier: "test@example.com",
			Password:   "password123",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB to return a user for the query
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(1, "testuser", "test@example.com", string(hashedPassword))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE")).
			WillReturnRows(rows)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusOK, w.Code)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		// Prepare request
		reqBody := handlers.LoginRequest{
			Identifier: "nonexistentuser",
			Password:   "password123",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB to return no user
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE")).
			WillReturnError(gorm.ErrRecordNotFound)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Equal(t, "Invalid credentials", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		// Create hashed password for test user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		require.NoError(t, err)

		// Prepare request with wrong password
		reqBody := handlers.LoginRequest{
			Identifier: "testuser",
			Password:   "wrongpassword",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB to return a user for the query
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(1, "testuser", "test@example.com", string(hashedPassword))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE")).
			WillReturnRows(rows)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Equal(t, "Invalid credentials", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestResetPasswordHandler tests the password reset functionality
func TestResetPasswordHandler(t *testing.T) {
	router, _ := setup()
	router.POST("/api/v1/auth/reset-password", handlers.ResetPasswordHandler)

	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	t.Run("Successful Password Reset", func(t *testing.T) {
		// Setup mock DB
		mock, err := setupDBMock()
		require.NoError(t, err)

		// Create hashed password for test user
		oldHashedPassword, err := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
		require.NoError(t, err)

		// Prepare request
		reqBody := handlers.ResetPasswordRequest{
			Identifier:      "testuser",
			NewPassword:     "newpassword123",
			ConfirmPassword: "newpassword123",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB operations
		// 1. Find user
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(1, "testuser", "test@example.com", string(oldHashedPassword))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE")).
			WillReturnRows(rows)

		// 2. Update user password
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET")).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(jsonData))
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
		assert.Equal(t, "Password reset successful", response["message"])

		// Verify that all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Password Mismatch", func(t *testing.T) {
		// Prepare request with mismatched passwords
		reqBody := handlers.ResetPasswordRequest{
			Identifier:      "testuser",
			NewPassword:     "newpassword123",
			ConfirmPassword: "differentpassword",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert: Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Passwords do not match", response["error"])
	})

	t.Run("User Not Found", func(t *testing.T) {
		mock, err := setupDBMock()
		require.NoError(t, err)

		// Prepare request
		reqBody := handlers.ResetPasswordRequest{
			Identifier:      "nonexistentuser",
			NewPassword:     "newpassword123",
			ConfirmPassword: "newpassword123",
		}
		jsonData, _ := json.Marshal(reqBody)

		// Mock DB to return no user
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE")).
			WillReturnError(gorm.ErrRecordNotFound)

		// Act: Send request
		req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(jsonData))
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
		assert.Equal(t, "User not found", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
