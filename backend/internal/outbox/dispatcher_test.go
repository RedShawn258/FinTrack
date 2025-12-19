//go:build integration

package outbox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

func setupDispatcherTest(t *testing.T) (*Dispatcher, *httptest.Server, func()) {
	logger := zap.NewNop()

	// Create a test HTTP server to receive webhook events
	receivedEvents := make([]map[string]interface{}, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event map[string]interface{}
		json.NewDecoder(r.Body).Decode(&event)
		receivedEvents = append(receivedEvents, event)
		w.WriteHeader(http.StatusOK)
	}))

	cfg := &config.Config{
		OutboxEnabled:      true,
		OutboxPollInterval: 1 * time.Second,
		OutboxMaxRetries:   3,
		EventPublisherType: "webhook",
		WebhookURL:         server.URL,
	}

	dispatcher := NewDispatcher(cfg, logger)

	cleanup := func() {
		server.Close()
		db.DB.Exec("DELETE FROM outbox_events")
	}

	return dispatcher, server, cleanup
}

func TestDispatcher_Integration_PublishEvent(t *testing.T) {
	dispatcher, server, cleanup := setupDispatcherTest(t)
	defer cleanup()

	// Create a test outbox event
	event := models.OutboxEvent{
		EventType:      "TransferCreated",
		AggregateType:  "Transfer",
		AggregateID:    123,
		IdempotencyKey: "test-event-1",
		PayloadJSON:    `{"eventId":123,"amount":100.0}`,
		Status:         models.OutboxStatusPending,
		RetryCount:     0,
	}

	require.NoError(t, db.DB.Create(&event).Error)

	// Process the event
	ctx := context.Background()
	dispatcher.processPendingEvents(ctx)

	// Give dispatcher time to process
	time.Sleep(100 * time.Millisecond)

	// Verify event was marked as SENT
	var updatedEvent models.OutboxEvent
	require.NoError(t, db.DB.First(&updatedEvent, event.ID).Error)
	assert.Equal(t, models.OutboxStatusSent, updatedEvent.Status)

	// Verify webhook was called
	// Note: In a real test, you'd check the receivedEvents slice
	// For now, we verify the status change
}

func TestDispatcher_Integration_RetryOnFailure(t *testing.T) {
	logger := zap.NewNop()

	// Create a test HTTP server that fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{
		OutboxEnabled:      true,
		OutboxPollInterval: 1 * time.Second,
		OutboxMaxRetries:   2,
		EventPublisherType: "webhook",
		WebhookURL:         server.URL,
	}

	dispatcher := NewDispatcher(cfg, logger)

	// Create a test outbox event
	event := models.OutboxEvent{
		EventType:      "TransferCreated",
		AggregateType:  "Transfer",
		AggregateID:    456,
		IdempotencyKey: "test-event-2",
		PayloadJSON:    `{"eventId":456,"amount":200.0}`,
		Status:         models.OutboxStatusPending,
		RetryCount:     0,
	}

	require.NoError(t, db.DB.Create(&event).Error)

	// Process the event (will fail)
	ctx := context.Background()
	dispatcher.processPendingEvents(ctx)

	time.Sleep(100 * time.Millisecond)

	// Verify event was marked as RETRYING with retry count incremented
	var updatedEvent models.OutboxEvent
	require.NoError(t, db.DB.First(&updatedEvent, event.ID).Error)
	assert.Equal(t, models.OutboxStatusRetrying, updatedEvent.Status)
	assert.Equal(t, 1, updatedEvent.RetryCount)
	assert.NotNil(t, updatedEvent.NextRetryAt)
}
