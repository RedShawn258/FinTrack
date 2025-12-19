package models

import (
	"time"

	"gorm.io/gorm"
)

// IdempotencyKey tracks processed requests to prevent duplicate processing
type IdempotencyKey struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	IdempotencyKey string         `gorm:"column:idempotency_key;type:varchar(255);uniqueIndex;not null" json:"key"`
	AccountID      uint           `gorm:"column:account_id;not null;index;type:int unsigned" json:"accountId"`
	TransactionID  *uint          `gorm:"column:transaction_id;index;type:int unsigned" json:"transactionId,omitempty"` // null if processing failed
	RequestHash    string         `gorm:"column:request_hash;type:varchar(64)" json:"requestHash"`                      // Hash of request body for validation
	Status         string         `gorm:"column:status;type:varchar(20);not null;default:'processing'" json:"status"`   // processing, completed, failed
	ResponseCode   int            `gorm:"column:response_code;type:int" json:"responseCode,omitempty"`
	ResponseBody   string         `gorm:"column:response_body;type:text" json:"responseBody,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for IdempotencyKey
func (IdempotencyKey) TableName() string {
	return "idempotency_keys"
}
