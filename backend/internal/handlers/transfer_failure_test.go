//go:build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// TestTransfer_Idempotency_DuplicateRetries simulates duplicate client retries
// with the same Idempotency-Key to verify no duplicate balance changes
func TestTransfer_Idempotency_DuplicateRetries(t *testing.T) {
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

	defer func() {
		db.DB.Exec("DELETE FROM journal_lines")
		db.DB.Exec("DELETE FROM journal_entries")
		db.DB.Exec("DELETE FROM idempotency_keys")
		db.DB.Exec("DELETE FROM outbox_events")
		db.DB.Exec("DELETE FROM accounts")
	}()

	idempotencyKey := "test-duplicate-retries"

	// Simulate 5 duplicate retries
	for i := 0; i < 5; i++ {
		reqBody := map[string]interface{}{
			"fromAccountId": fromAccount.ID,
			"toAccountId":   toAccount.ID,
			"amount":        100.0,
			"description":   fmt.Sprintf("Retry attempt %d", i+1),
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// First request should succeed, subsequent should return cached response
		if i == 0 {
			assert.Equal(t, http.StatusCreated, w.Code)
		} else {
			assert.Equal(t, http.StatusOK, w.Code, "Duplicate retry should return cached response")
		}
	}

	// Verify only ONE journal entry was created
	var entryCount int64
	db.DB.Model(&models.JournalEntry{}).Where("idempotency_key = ?", idempotencyKey).Count(&entryCount)
	assert.Equal(t, int64(1), entryCount, "Should have only one journal entry despite 5 retries")

	// Verify only ONE outbox event was created
	var outboxCount int64
	db.DB.Model(&models.OutboxEvent{}).Where("idempotency_key = ?", idempotencyKey).Count(&outboxCount)
	assert.Equal(t, int64(1), outboxCount, "Should have only one outbox event despite 5 retries")

	// Verify balance changed only once
	var finalFrom, finalTo models.Account
	db.DB.First(&finalFrom, fromAccount.ID)
	db.DB.First(&finalTo, toAccount.ID)

	assert.Equal(t, 900.0, finalFrom.Balance, "Balance should reflect only one transfer")
	assert.Equal(t, 600.0, finalTo.Balance, "Balance should reflect only one transfer")
}

// TestTransfer_ConcurrentDispatchers simulates multiple dispatcher replicas
// processing outbox events concurrently to verify no duplicate publishes
func TestTransfer_ConcurrentDispatchers(t *testing.T) {
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

	defer func() {
		db.DB.Exec("DELETE FROM journal_lines")
		db.DB.Exec("DELETE FROM journal_entries")
		db.DB.Exec("DELETE FROM idempotency_keys")
		db.DB.Exec("DELETE FROM outbox_events")
		db.DB.Exec("DELETE FROM accounts")
	}()

	// Create a transfer
	reqBody := map[string]interface{}{
		"fromAccountId": fromAccount.ID,
		"toAccountId":   toAccount.ID,
		"amount":        100.0,
		"description":   "Concurrent dispatcher test",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-concurrent-dispatchers")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify outbox event exists
	var outboxEvent models.OutboxEvent
	require.NoError(t, db.DB.Where("idempotency_key = ?", "test-concurrent-dispatchers").First(&outboxEvent).Error)
	assert.Equal(t, models.OutboxStatusPending, outboxEvent.Status)

	// Simulate 3 concurrent dispatcher replicas trying to claim the same event
	var wg sync.WaitGroup
	claimedCount := 0
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Try to claim event using SKIP LOCKED
			var claimedEvent models.OutboxEvent
			err := db.DB.Transaction(func(tx *gorm.DB) error {
				// Use SKIP LOCKED to avoid blocking
				query := tx.Where("status = ? AND id = ?", models.OutboxStatusPending, outboxEvent.ID).
					Set("gorm:query_option", "FOR UPDATE SKIP LOCKED")

				if err := query.First(&claimedEvent).Error; err != nil {
					return err // Event already claimed by another worker
				}

				// Mark as RETRYING to claim it
				return tx.Model(&claimedEvent).Updates(map[string]interface{}{
					"status":     models.OutboxStatusRetrying,
					"updated_at": time.Now(),
				}).Error
			})

			if err == nil {
				mu.Lock()
				claimedCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Only ONE dispatcher should have claimed the event
	assert.Equal(t, 1, claimedCount, "Only one dispatcher should claim the event")
}

// TestTransfer_HandlerCrashAfterCommit simulates handler crash after DB commit
// but before response - verifies idempotency handles this correctly
func TestTransfer_HandlerCrashAfterCommit(t *testing.T) {
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

	defer func() {
		db.DB.Exec("DELETE FROM journal_lines")
		db.DB.Exec("DELETE FROM journal_entries")
		db.DB.Exec("DELETE FROM idempotency_keys")
		db.DB.Exec("DELETE FROM outbox_events")
		db.DB.Exec("DELETE FROM accounts")
	}()

	idempotencyKey := "test-handler-crash"

	// First request: succeeds and commits
	reqBody := map[string]interface{}{
		"fromAccountId": fromAccount.ID,
		"toAccountId":   toAccount.ID,
		"amount":        100.0,
		"description":   "First request",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req1 := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", idempotencyKey)

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Simulate handler crash: verify data is committed
	var entry models.JournalEntry
	require.NoError(t, db.DB.Where("idempotency_key = ?", idempotencyKey).First(&entry).Error)

	// Client retries with same idempotency key (simulating crash recovery)
	req2 := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", idempotencyKey)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// Should return cached response (idempotent)
	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 CreateTransferResponse
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &response2))
	assert.Equal(t, entry.ID, response2.EntryID, "Should return same entry ID")

	// Verify balance unchanged (no double deduction)
	var finalFrom models.Account
	db.DB.First(&finalFrom, fromAccount.ID)
	assert.Equal(t, 900.0, finalFrom.Balance, "Balance should reflect only one transfer")
}

// TestTransfer_ConcurrentTransfers verifies concurrent transfers don't cause
// race conditions or duplicate balance changes
func TestTransfer_ConcurrentTransfers(t *testing.T) {
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

	defer func() {
		db.DB.Exec("DELETE FROM journal_lines")
		db.DB.Exec("DELETE FROM journal_entries")
		db.DB.Exec("DELETE FROM idempotency_keys")
		db.DB.Exec("DELETE FROM outbox_events")
		db.DB.Exec("DELETE FROM accounts")
	}()

	// Simulate 10 concurrent transfers
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(transferID int) {
			defer wg.Done()

			reqBody := map[string]interface{}{
				"fromAccountId": fromAccount.ID,
				"toAccountId":   toAccount.ID,
				"amount":        10.0,
				"description":   fmt.Sprintf("Concurrent transfer %d", transferID),
			}

			bodyBytes, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/transfers", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", fmt.Sprintf("concurrent-%d", transferID))

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code == http.StatusCreated {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify all transfers succeeded
	assert.Equal(t, 10, successCount, "All concurrent transfers should succeed")

	// Verify final balance
	var finalFrom, finalTo models.Account
	db.DB.First(&finalFrom, fromAccount.ID)
	db.DB.First(&finalTo, toAccount.ID)

	expectedFromBalance := 1000.0 - (10.0 * 10) // Initial - (amount * count)
	expectedToBalance := 500.0 + (10.0 * 10)    // Initial + (amount * count)

	assert.Equal(t, expectedFromBalance, finalFrom.Balance, "Balance should reflect all transfers")
	assert.Equal(t, expectedToBalance, finalTo.Balance, "Balance should reflect all transfers")

	// Verify exactly 10 journal entries
	var entryCount int64
	db.DB.Model(&models.JournalEntry{}).Count(&entryCount)
	assert.Equal(t, int64(10), entryCount, "Should have exactly 10 journal entries")
}
