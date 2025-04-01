package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey;type:int unsigned"`
	Username     string `gorm:"uniqueIndex;size:50;not null"`
	Email        string `gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
