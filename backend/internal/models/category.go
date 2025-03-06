package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index;type:int unsigned"`
	Name      string `gorm:"size:50;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
