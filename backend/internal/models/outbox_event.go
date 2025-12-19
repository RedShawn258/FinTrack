package models

import (
	"time"

	"gorm.io/gorm"
)

// OutboxEventStatus represents the status of an outbox event
type OutboxEventStatus string

const (
	OutboxStatusPending  OutboxEventStatus = "PENDING"
	OutboxStatusRetrying OutboxEventStatus = "RETRYING"
	OutboxStatusSent     OutboxEventStatus = "SENT"
	OutboxStatusFailed   OutboxEventStatus = "FAILED"
)

// OutboxEvent represents an event to be published via the Outbox pattern
// Ensures reliable event publishing without losing events or creating duplicates
type OutboxEvent struct {
	ID             uint              `gorm:"primaryKey" json:"id"`
	EventType      string            `gorm:"type:varchar(100);not null;index" json:"eventType"`
	AggregateType  string            `gorm:"type:varchar(100);not null;index" json:"aggregateType"`
	AggregateID    uint              `gorm:"not null;index" json:"aggregateId"`
	IdempotencyKey string            `gorm:"type:varchar(255);index" json:"idempotencyKey"` // Links to transfer idempotency key
	PayloadJSON    string            `gorm:"type:json;not null" json:"payloadJson"`
	Status         OutboxEventStatus `gorm:"type:varchar(20);not null;default:'PENDING';index" json:"status"`
	RetryCount     int               `gorm:"not null;default:0" json:"retryCount"`
	NextRetryAt    *time.Time        `gorm:"index" json:"nextRetryAt,omitempty"`
	LastError      string            `gorm:"type:text" json:"lastError,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"-"`
}

// TableName specifies the table name for OutboxEvent
func (OutboxEvent) TableName() string {
	return "outbox_events"
}
