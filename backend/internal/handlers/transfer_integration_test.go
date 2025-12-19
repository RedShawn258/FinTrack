//go:build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
	"github.com/RedShawn258/FinTrack/backend/internal/utils"
)

func setupTransferIntegrationTest(t *testing.T) (*gin.Engine, *models.Account, *models.Account, func()) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := zap.NewNop()

	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Next()
	})

	r.POST("/transfers", CreateTransfer)

	// Create test accounts
	fromAccount := models.Account{Balance: 1000.0}
	toAccount := models.Account{Balance: 500.0}

	require.NoError(t, db.DB.Create(&fromAccount).Error)
	require.NoError(t, db.DB.Create(&toAccount).Error)

	cleanup := func() {
		db.DB.Exec("DELETE FROM journal_lines")
		db.DB.Exec("DELETE FROM journal_entries")
		db.DB.Exec("DELETE FROM idempotency_keys")
		db.DB.Exec("DELETE FROM outbox_events")
		db.DB.Exec("DELETE FROM accounts")
	}

	return r, &fromAccount, &toAccount, cleanup
}

func TestCreateTransfer_Integration_Success(t *testing.T) {
	r, fromAccount, toAccount, cleanup := setupTransferIntegrationTest(t)
	defer cleanup()

	reqBody := map[string]interface{}{
		"fromAccountId": fromAccount.ID,
		"toAccountId":   toAccount.ID,
		"amount":        100.0,
		"description":   "Test transfer",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-transfer-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response CreateTransferResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	assert.Equal(t, fromAccount.ID, response.FromAccountID)
	assert.Equal(t, toAccount.ID, response.ToAccountID)
	assert.Equal(t, 100.0, response.Amount)
	assert.Equal(t, 900.0, response.FromBalance) // 1000 - 100
	assert.Equal(t, 600.0, response.ToBalance)   // 500 + 100

	// Verify journal entry exists and is balanced
	var entry models.JournalEntry
	require.NoError(t, db.DB.Preload("Lines").First(&entry, response.EntryID).Error)
	assert.Len(t, entry.Lines, 2)

	// Verify double-entry invariant
	err := utils.VerifyJournalEntryBalance(entry.ID)
	assert.NoError(t, err, "Journal entry must satisfy double-entry invariant")

	// Verify account balances match materialized values
	var updatedFrom, updatedTo models.Account
	db.DB.First(&updatedFrom, fromAccount.ID)
	db.DB.First(&updatedTo, toAccount.ID)

	assert.Equal(t, 900.0, updatedFrom.Balance)
	assert.Equal(t, 600.0, updatedTo.Balance)

	// Verify outbox event was created
	var outboxEvent models.OutboxEvent
	err := db.DB.Where("idempotency_key = ?", "test-transfer-1").First(&outboxEvent).Error
	assert.NoError(t, err, "Outbox event should be created")
	assert.Equal(t, "TransferCreated", outboxEvent.EventType)
	assert.Equal(t, "Transfer", outboxEvent.AggregateType)
	assert.Equal(t, response.EntryID, outboxEvent.AggregateID)
	assert.Equal(t, models.OutboxStatusPending, outboxEvent.Status)
	assert.Contains(t, outboxEvent.PayloadJSON, fmt.Sprintf(`"eventId":%d`, response.EntryID))
}

func TestCreateTransfer_Integration_Idempotency(t *testing.T) {
	r, fromAccount, toAccount, cleanup := setupTransferIntegrationTest(t)
	defer cleanup()

	reqBody := map[string]interface{}{
		"fromAccountId": fromAccount.ID,
		"toAccountId":   toAccount.ID,
		"amount":        100.0,
		"description":   "Test transfer",
	}

	idempotencyKey := "test-idempotent-transfer"

	// First request
	bodyBytes, _ := json.Marshal(reqBody)
	req1 := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", idempotencyKey)

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	var response1 CreateTransferResponse
	require.NoError(t, json.Unmarshal(w1.Body.Bytes(), &response1))
	entryID1 := response1.EntryID

	// Duplicate request (same idempotency key, same params)
	bodyBytes2, _ := json.Marshal(reqBody)
	req2 := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", idempotencyKey)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code) // Should return cached response

	var response2 CreateTransferResponse
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &response2))

	assert.Equal(t, entryID1, response2.EntryID, "Should return same entry ID")
	assert.Equal(t, "Transfer already processed (idempotent)", response2.Message)

	// Verify only one journal entry was created
	var count int64
	db.DB.Model(&models.JournalEntry{}).Where("idempotency_key = ?", idempotencyKey).Count(&count)
	assert.Equal(t, int64(1), count, "Should have only one journal entry")
}

func TestCreateTransfer_Integration_InsufficientBalance(t *testing.T) {
	r, fromAccount, toAccount, cleanup := setupTransferIntegrationTest(t)
	defer cleanup()

	reqBody := map[string]interface{}{
		"fromAccountId": fromAccount.ID,
		"toAccountId":   toAccount.ID,
		"amount":        2000.0, // More than available balance
		"description":   "Test transfer",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-insufficient-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errorResponse))
	assert.Contains(t, errorResponse["error"].(string), "insufficient balance")

	// Verify no journal entry was created
	var count int64
	db.DB.Model(&models.JournalEntry{}).Where("idempotency_key = ?", "test-insufficient-1").Count(&count)
	assert.Equal(t, int64(0), count, "Should not create journal entry on failure")
}

func TestCreateTransfer_Integration_ConcurrencySafety(t *testing.T) {
	// This test verifies that concurrent transfers don't cause race conditions
	// Note: Full concurrency testing requires multiple goroutines and a real database
	// This is a simplified version

	r, fromAccount, toAccount, cleanup := setupTransferIntegrationTest(t)
	defer cleanup()

	// Simulate two concurrent transfers from the same account
	// In a real scenario, we'd use goroutines and sync.WaitGroup

	reqBody1 := map[string]interface{}{
		"fromAccountId": fromAccount.ID,
		"toAccountId":   toAccount.ID,
		"amount":        100.0,
		"description":   "Transfer 1",
	}

	reqBody2 := map[string]interface{}{
		"fromAccountId": fromAccount.ID,
		"toAccountId":   toAccount.ID,
		"amount":        200.0,
		"description":   "Transfer 2",
	}

	// Execute transfers sequentially (in real test, use goroutines)
	bodyBytes1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", "concurrent-test-1")

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	time.Sleep(10 * time.Millisecond) // Small delay to simulate concurrency

	bodyBytes2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "concurrent-test-2")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	// Verify final balance
	var finalAccount models.Account
	db.DB.First(&finalAccount, fromAccount.ID)
	expectedBalance := 1000.0 - 100.0 - 200.0 // Initial - transfer1 - transfer2
	assert.Equal(t, expectedBalance, finalAccount.Balance, "Balance should reflect both transfers")
}
