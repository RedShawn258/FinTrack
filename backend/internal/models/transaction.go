package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	ID              uint      `gorm:"primaryKey"`
	UserID          uint      `gorm:"not null;index;type:int unsigned"`
	CategoryID      *uint     `gorm:"index;type:int unsigned"` // null => uncategorized
	Amount          float64   `gorm:"type:decimal(10,2);not null"`
	Description     string    `gorm:"type:text"`
	TransactionDate time.Time `gorm:"type:date;not null;index"` // store only date if you want day-level precision
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
