package models

import (
	"time"

	"gorm.io/gorm"
)

// JournalEntry represents a double-entry accounting journal entry
// All entries must satisfy: SUM(debits) == SUM(credits)
type JournalEntry struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	IdempotencyKey string         `gorm:"column:idempotency_key;type:varchar(255);uniqueIndex;not null" json:"idempotencyKey"`
	Description    string         `gorm:"type:text" json:"description"`
	Metadata       string         `gorm:"type:json" json:"metadata,omitempty"` // Optional JSON metadata
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationship
	Lines []JournalLine `gorm:"foreignKey:EntryID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`
}

// TableName specifies the table name for JournalEntry
func (JournalEntry) TableName() string {
	return "journal_entries"
}
