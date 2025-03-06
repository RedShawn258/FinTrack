package models

import (
	"time"

	"gorm.io/gorm"
)

// Budget represents a user's budget, now storing only the date portion for StartDate and EndDate.
type Budget struct {
	ID              uint      `gorm:"primaryKey"`
	UserID          uint      `gorm:"not null;index;type:int unsigned"`
	CategoryID      *uint     `gorm:"index;type:int unsigned"` // null => global
	LimitAmount     float64   `gorm:"type:decimal(10,2);not null" json:"LimitAmount"`
	RemainingAmount float64   `gorm:"type:decimal(10,2);default:0.00" json:"RemainingAmount"`
	StartDate       time.Time `gorm:"type:date;not null;index" json:"StartDate"` // Only store the date, no time
	EndDate         time.Time `gorm:"type:date;not null;index" json:"EndDate"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
