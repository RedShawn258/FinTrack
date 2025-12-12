package models

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents a refresh token stored in the database
// Tokens are hashed before storage for security
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index;type:int unsigned"`
	TokenHash string    `gorm:"size:255;not null;uniqueIndex"` // SHA-256 hash of the token
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// Token rotation tracking - ID of the token that replaced this one
	// Used to detect token reuse attacks
	ReplacedBy *uint `gorm:"index"`
	
	// Relationship
	User User `gorm:"foreignKey:UserID"`
}

