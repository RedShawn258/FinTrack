package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/handlers"
)

func TestCreateCategory(t *testing.T) {
	// Initialize router and mock DB for each test
	router, _ := setup()
	mock, err := setupDBMock()
	require.NoError(t, err)

	// Set up the route to test the CreateCategory handler
	router.POST("/api/v1/categories", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("db", mock)
		handlers.CreateCategory(c)
	})

	t.Run("Successfully Create New Category", func(t *testing.T) {
		// Expect a query to check if the category already exists
		mock.ExpectQuery("SELECT \\* FROM `categories` WHERE \\(user_id = \\? AND name = \\?\\) AND `categories`\\.`deleted_at` IS NULL ORDER BY `categories`\\.`id` LIMIT \\?").
			WithArgs(uint(1), "Groceries", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Expect the category to be saved - match the exact column order GORM uses
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `categories` \\(`user_id`,`name`,`created_at`,`updated_at`,`deleted_at`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?\\)").
			WithArgs(
				uint(1),          // User ID
				"Groceries",      // Name
				sqlmock.AnyArg(), // Created at
				sqlmock.AnyArg(), // Updated at
				nil,              // Deleted at
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Create a request with valid category data
		body := map[string]interface{}{
			"name": "Groceries",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response status code
		assert.Equal(t, http.StatusCreated, w.Code)

		// Parse the response body
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert the success message
		assert.Equal(t, "Category created successfully", response["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update Existing Category", func(t *testing.T) {
		mock.ExpectationsWereMet()

		// Create fixed timestamps for testing
		now := time.Now()

		// Expect a query to check if the category already exists
		mock.ExpectQuery("SELECT \\* FROM `categories` WHERE \\(user_id = \\? AND name = \\?\\) AND `categories`\\.`deleted_at` IS NULL ORDER BY `categories`\\.`id` LIMIT \\?").
			WithArgs(uint(1), "Groceries", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "created_at", "updated_at", "deleted_at"}).
				AddRow(1, 1, "Groceries", now, now, nil))

		// Expect the category to be updated - match the exact column order GORM uses
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `categories` SET `user_id`=\\?,`name`=\\?,`created_at`=\\?,`updated_at`=\\?,`deleted_at`=\\? WHERE `categories`.`deleted_at` IS NULL AND `id` = \\?").
			WithArgs(
				uint(1),          // User ID
				"Groceries",      // Name
				now,              // Created at
				sqlmock.AnyArg(), // Updated at
				nil,              // Deleted at
				1,                // ID in WHERE clause
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		// Create a request with valid category data
		body := map[string]interface{}{
			"name": "Groceries",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response status code
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse the response body
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert the success message
		assert.Equal(t, "Category already exists, overwriting.", response["message"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid Request Data", func(t *testing.T) {
		mock.ExpectationsWereMet()

		// Create a request with invalid category data (missing name)
		body := map[string]interface{}{
			"description": "Invalid category",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response status code
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Parse the response body
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert the error message contains validation info
		assert.Contains(t, response["error"].(string), "validation")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error on Save", func(t *testing.T) {
		// Reset expectations
		mock.ExpectationsWereMet()

		// Expect a query to check if the category already exists
		mock.ExpectQuery("SELECT \\* FROM `categories` WHERE \\(user_id = \\? AND name = \\?\\) AND `categories`\\.`deleted_at` IS NULL ORDER BY `categories`\\.`id` LIMIT \\?").
			WithArgs(uint(1), "Groceries", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Expect database error on save - match the exact column order GORM uses
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `categories` \\(`user_id`,`name`,`created_at`,`updated_at`,`deleted_at`\\) VALUES \\(\\?,\\?,\\?,\\?,\\?\\)").
			WithArgs(
				uint(1),          // User ID
				"Groceries",      // Name
				sqlmock.AnyArg(), // Created at
				sqlmock.AnyArg(), // Updated at
				nil,              // Deleted at
			).WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		body := map[string]interface{}{
			"name": "Groceries",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response status code
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Parse the response body
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert the error message
		assert.Equal(t, "Could not create category", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetCategories(t *testing.T) {
	router, _ := setup()
	mock, err := setupDBMock()
	require.NoError(t, err)

	router.GET("/api/v1/categories", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("db", mock)
		handlers.GetCategories(c)
	})

	now := time.Now()

	t.Run("Successfully Get All Categories", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, "Groceries", now, now, nil).
			AddRow(2, 1, "Entertainment", now, now, nil)

		mock.ExpectQuery("SELECT \\* FROM `categories` WHERE user_id = \\? AND `categories`\\.`deleted_at` IS NULL").
			WithArgs(uint(1)).
			WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response status code
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse the response body
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert the categories array
		categories, ok := response["categories"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, categories, 2)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No Categories Found", func(t *testing.T) {
		mock.ExpectationsWereMet()

		rows := sqlmock.NewRows([]string{"id", "user_id", "name", "created_at", "updated_at", "deleted_at"})

		mock.ExpectQuery("SELECT \\* FROM `categories` WHERE user_id = \\? AND `categories`\\.`deleted_at` IS NULL").
			WithArgs(uint(1)).
			WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response status code
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse the response body
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert the categories array
		categories, ok := response["categories"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, categories, 0)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		mock.ExpectationsWereMet()

		mock.ExpectQuery("SELECT \\* FROM `categories` WHERE user_id = \\? AND `categories`\\.`deleted_at` IS NULL").
			WithArgs(uint(1)).
			WillReturnError(errors.New("database error"))

		req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response status code
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Parse the response body
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert the error message - matches the handler's message
		assert.Equal(t, "Could not fetch categories", response["error"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteCategory(t *testing.T) {
	router, _ := setup()
	mock, err := setupDBMock()
	require.NoError(t, err)

	// Set up the route to test the DeleteCategory handler
	router.DELETE("/api/v1/categories/:id", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("db", mock)
		handlers.DeleteCategory(c)
	})

	t.Run("Successfully Delete Category", func(t *testing.T) {
		categoryID := "1"
		userID := uint(1)

		// Expect transaction for the soft delete operation
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `categories` SET `deleted_at`=\\? WHERE \\(id = \\? AND user_id = \\?\\) AND `categories`\\.`deleted_at` IS NULL").
			WithArgs(sqlmock.AnyArg(), categoryID, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		// Act: Send request
		req, _ := http.NewRequest("DELETE", "/api/v1/categories/"+categoryID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check message
		assert.Equal(t, "Category deleted successfully", response["message"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Category Not Found", func(t *testing.T) {
		nonExistentCategoryID := "999"
		userID := uint(1)

		// Expect transaction that will return error
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `categories` SET `deleted_at`=\\? WHERE \\(id = \\? AND user_id = \\?\\) AND `categories`\\.`deleted_at` IS NULL").
			WithArgs(sqlmock.AnyArg(), nonExistentCategoryID, userID).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectRollback()

		// Act: Send request
		req, _ := http.NewRequest("DELETE", "/api/v1/categories/"+nonExistentCategoryID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message - matches the handler's message
		assert.Equal(t, "Category not found or could not be deleted", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Invalid Category ID Format", func(t *testing.T) {
		invalidCategoryID := "invalid"
		userID := uint(1)

		// Expect transaction that will return error
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `categories` SET `deleted_at`=\\? WHERE \\(id = \\? AND user_id = \\?\\) AND `categories`\\.`deleted_at` IS NULL").
			WithArgs(sqlmock.AnyArg(), invalidCategoryID, userID).
			WillReturnError(errors.New("invalid category ID format"))
		mock.ExpectRollback()

		// Act: Send request
		req, _ := http.NewRequest("DELETE", "/api/v1/categories/"+invalidCategoryID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message - matches the handler's message
		assert.Equal(t, "Category not found or could not be deleted", response["error"])

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
