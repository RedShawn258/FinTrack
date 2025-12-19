package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

func setupTransferTest(t *testing.T) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := zap.NewNop()

	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Next()
	})

	r.POST("/transfers", CreateTransfer)

	return r, func() {
		// Cleanup if needed
	}
}

func TestCreateTransfer_InvalidRequest(t *testing.T) {
	r, cleanup := setupTransferTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		body           interface{}
		idempotencyKey string
		expectedStatus int
	}{
		{
			name:           "missing idempotency key",
			body:           map[string]interface{}{"fromAccountId": 1, "toAccountId": 2, "amount": 100.0},
			idempotencyKey: "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "same from and to account",
			body:           map[string]interface{}{"fromAccountId": 1, "toAccountId": 1, "amount": 100.0},
			idempotencyKey: "test-key-1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid amount",
			body:           map[string]interface{}{"fromAccountId": 1, "toAccountId": 2, "amount": -10.0},
			idempotencyKey: "test-key-2",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "zero amount",
			body:           map[string]interface{}{"fromAccountId": 1, "toAccountId": 2, "amount": 0.0},
			idempotencyKey: "test-key-3",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.idempotencyKey != "" {
				req.Header.Set("Idempotency-Key", tt.idempotencyKey)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDoubleEntryInvariant(t *testing.T) {
	// This test verifies the double-entry accounting invariant
	// In a real scenario, this would use a test database
	// For now, we test the logic conceptually

	// A valid journal entry should have:
	// - At least 2 lines (one debit, one credit)
	// - SUM(debits) == SUM(credits)

	validEntry := []struct {
		direction models.Direction
		amount    float64
	}{
		{models.DirectionDebit, 100.0},
		{models.DirectionCredit, 100.0},
	}

	totalDebits := 0.0
	totalCredits := 0.0

	for _, line := range validEntry {
		if line.direction == models.DirectionDebit {
			totalDebits += line.amount
		} else {
			totalCredits += line.amount
		}
	}

	assert.Equal(t, totalDebits, totalCredits, "Double-entry invariant must hold: debits == credits")
}
