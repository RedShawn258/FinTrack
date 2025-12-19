package models

import (
	"time"

	"gorm.io/gorm"
)

// TransactionType represents the type of ledger transaction
type TransactionType string

const (
	TransactionTypeCredit TransactionType = "credit" // Money added to account
	TransactionTypeDebit  TransactionType = "debit"  // Money removed from account
)

// LedgerTransaction represents a credit or debit transaction on an account
type LedgerTransaction struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	AccountID      uint            `gorm:"not null;index;type:int unsigned" json:"accountId"`
	Type           TransactionType `gorm:"type:enum('credit','debit');not null" json:"type"`
	Amount         float64         `gorm:"type:decimal(15,2);not null" json:"amount"`
	Description    string          `gorm:"type:text" json:"description"`
	IdempotencyKey string          `gorm:"type:varchar(255);uniqueIndex;not null" json:"idempotencyKey"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationship
	Account Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

// TableName specifies the table name for LedgerTransaction
func (LedgerTransaction) TableName() string {
	return "ledger_transactions"
}
