package models

import (
	"time"

	"gorm.io/gorm"
)

// Direction represents debit or credit in double-entry accounting
type Direction string

const (
	DirectionDebit  Direction = "debit"
	DirectionCredit Direction = "credit"
)

// JournalLine represents a single line in a journal entry
// Each entry must have balanced debits and credits
type JournalLine struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	EntryID   uint           `gorm:"not null;index;type:int unsigned" json:"entryId"`
	AccountID uint           `gorm:"not null;index;type:int unsigned" json:"accountId"`
	Direction Direction      `gorm:"type:enum('debit','credit');not null" json:"direction"`
	Amount    float64        `gorm:"type:decimal(15,2);not null" json:"amount"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Entry   JournalEntry `gorm:"foreignKey:EntryID" json:"entry,omitempty"`
	Account Account      `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

// TableName specifies the table name for JournalLine
func (JournalLine) TableName() string {
	return "journal_lines"
}
