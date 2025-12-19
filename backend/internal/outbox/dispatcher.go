package outbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// Dispatcher handles publishing outbox events reliably
type Dispatcher struct {
	config     *config.Config
	logger     *zap.Logger
	httpClient *http.Client
	stopChan   chan struct{}
	metrics    *Metrics
}

// NewDispatcher creates a new outbox dispatcher
func NewDispatcher(cfg *config.Config, logger *zap.Logger) *Dispatcher {
	return &Dispatcher{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopChan: make(chan struct{}),
		metrics:  NewMetrics(),
	}
}

// Start begins polling and dispatching events
func (d *Dispatcher) Start(ctx context.Context) {
	if !d.config.OutboxEnabled {
		d.logger.Info("Outbox pattern disabled, skipping dispatcher")
		return
	}

	d.logger.Info("Starting outbox dispatcher",
		zap.String("publisher", d.config.EventPublisherType),
		zap.Duration("pollInterval", d.config.OutboxPollInterval),
		zap.Int("maxRetries", d.config.OutboxMaxRetries),
	)

	ticker := time.NewTicker(d.config.OutboxPollInterval)
	defer ticker.Stop()

	// Initial poll
	d.processPendingEvents(ctx)

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("Outbox dispatcher stopping")
			return
		case <-d.stopChan:
			d.logger.Info("Outbox dispatcher stopped")
			return
		case <-ticker.C:
			d.processPendingEvents(ctx)
		}
	}
}

// Stop gracefully stops the dispatcher
func (d *Dispatcher) Stop() {
	close(d.stopChan)
}

// processPendingEvents polls for pending events and publishes them
func (d *Dispatcher) processPendingEvents(ctx context.Context) {
	// Use SELECT ... FOR UPDATE SKIP LOCKED to safely claim events in concurrent workers
	// MySQL 8.0+ supports SKIP LOCKED
	var events []models.OutboxEvent

	// Claim up to 10 events at a time
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Use SKIP LOCKED to avoid blocking on rows locked by other workers
		// If SKIP LOCKED not supported, fall back to regular FOR UPDATE
		now := time.Now()
		query := tx.Where("status IN ? AND (next_retry_at IS NULL OR next_retry_at <= ?)",
			[]string{string(models.OutboxStatusPending), string(models.OutboxStatusRetrying)},
			now).
			Order("created_at ASC").
			Limit(10)

		// Try SKIP LOCKED first (MySQL 8.0+)
		// If not supported, will fall back to regular locking
		queryWithLock := query.Set("gorm:query_option", "FOR UPDATE SKIP LOCKED")
		if err := queryWithLock.Find(&events).Error; err != nil {
			// Fallback: try without SKIP LOCKED
			queryWithLock = query.Set("gorm:query_option", "FOR UPDATE")
			if err := queryWithLock.Find(&events).Error; err != nil {
				return err
			}
		}

		// Mark events as RETRYING to prevent other workers from claiming them
		if len(events) > 0 {
			eventIDs := make([]uint, len(events))
			for i, e := range events {
				eventIDs[i] = e.ID
			}
			if err := tx.Model(&models.OutboxEvent{}).
				Where("id IN ?", eventIDs).
				Updates(map[string]interface{}{
					"status":     models.OutboxStatusRetrying,
					"updated_at": now,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		d.logger.Error("Failed to claim pending events", zap.Error(err))
		return
	}

	if len(events) == 0 {
		return // No events to process
	}

	d.logger.Info("Claimed events for processing", zap.Int("count", len(events)))
	d.metrics.PendingTotal.Add(float64(len(events)))

	// Process each event
	for _, event := range events {
		d.processEvent(ctx, event)
	}
}

// processEvent publishes a single event
func (d *Dispatcher) processEvent(ctx context.Context, event models.OutboxEvent) {
	startTime := time.Now()
	d.logger.Info("Processing outbox event",
		zap.Uint("eventId", event.ID),
		zap.String("eventType", event.EventType),
		zap.String("aggregateId", fmt.Sprintf("%d", event.AggregateID)),
		zap.Int("retryCount", event.RetryCount),
	)

	var err error
	switch d.config.EventPublisherType {
	case "pubsub":
		err = d.publishToPubSub(ctx, event)
	case "webhook":
		err = d.publishToWebhook(ctx, event)
	default:
		err = fmt.Errorf("unknown event publisher type: %s", d.config.EventPublisherType)
	}

	latency := time.Since(startTime).Seconds()
	d.metrics.PublishLatency.Observe(latency)

	if err != nil {
		d.logger.Error("Failed to publish event",
			zap.Uint("eventId", event.ID),
			zap.Error(err),
			zap.Int("retryCount", event.RetryCount),
		)

		d.handlePublishFailure(event, err)
		return
	}

	// Mark as SENT
	if err := db.DB.Model(&event).Updates(map[string]interface{}{
		"status":     models.OutboxStatusSent,
		"updated_at": time.Now(),
	}).Error; err != nil {
		d.logger.Error("Failed to mark event as sent", zap.Uint("eventId", event.ID), zap.Error(err))
		return
	}

	d.logger.Info("Event published successfully",
		zap.Uint("eventId", event.ID),
		zap.Float64("latencySeconds", latency),
	)

	d.metrics.SentTotal.Inc()
}

// publishToWebhook publishes event to HTTP webhook endpoint
func (d *Dispatcher) publishToWebhook(ctx context.Context, event models.OutboxEvent) error {
	// Parse payload JSON
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(event.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("failed to parse event payload: %w", err)
	}

	// Add deterministic event ID for deduplication
	payload["eventId"] = event.ID
	payload["idempotencyKey"] = event.IdempotencyKey

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.config.WebhookURL,
		bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Id", fmt.Sprintf("%d", event.ID))
	req.Header.Set("X-Event-Type", event.EventType)
	req.Header.Set("X-Idempotency-Key", event.IdempotencyKey)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}

// publishToPubSub publishes event to Google Pub/Sub
// Note: Requires google.golang.org/api/pubsub/v1 and proper GCP credentials
func (d *Dispatcher) publishToPubSub(ctx context.Context, event models.OutboxEvent) error {
	// For now, return an error indicating Pub/Sub not fully implemented
	// In production, this would use the Pub/Sub client library
	return fmt.Errorf("Pub/Sub publisher not implemented yet - use webhook for now")
}

// handlePublishFailure handles failed publish attempts with exponential backoff
func (d *Dispatcher) handlePublishFailure(event models.OutboxEvent, err error) {
	newRetryCount := event.RetryCount + 1

	if newRetryCount > d.config.OutboxMaxRetries {
		// Mark as FAILED after max retries
		if updateErr := db.DB.Model(&event).Updates(map[string]interface{}{
			"status":      models.OutboxStatusFailed,
			"retry_count": newRetryCount,
			"last_error":  err.Error(),
			"updated_at":  time.Now(),
		}).Error; updateErr != nil {
			d.logger.Error("Failed to mark event as failed", zap.Uint("eventId", event.ID), zap.Error(updateErr))
			return
		}

		d.logger.Warn("Event marked as FAILED after max retries",
			zap.Uint("eventId", event.ID),
			zap.Int("retryCount", newRetryCount),
		)

		d.metrics.FailedTotal.Inc()
		return
	}

	// Calculate exponential backoff: 2^retryCount seconds, max 1 hour
	backoffSeconds := math.Min(math.Pow(2, float64(newRetryCount)), 3600)
	nextRetryAt := time.Now().Add(time.Duration(backoffSeconds) * time.Second)

	if updateErr := db.DB.Model(&event).Updates(map[string]interface{}{
		"status":        models.OutboxStatusRetrying,
		"retry_count":   newRetryCount,
		"next_retry_at": nextRetryAt,
		"last_error":    err.Error(),
		"updated_at":    time.Now(),
	}).Error; updateErr != nil {
		d.logger.Error("Failed to update retry info", zap.Uint("eventId", event.ID), zap.Error(updateErr))
		return
	}

	d.logger.Info("Event scheduled for retry",
		zap.Uint("eventId", event.ID),
		zap.Int("retryCount", newRetryCount),
		zap.Time("nextRetryAt", nextRetryAt),
	)

	d.metrics.RetriesTotal.Inc()
}
