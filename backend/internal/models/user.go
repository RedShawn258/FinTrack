package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID               uint   `gorm:"primaryKey;type:int unsigned"`
	Username         string `gorm:"uniqueIndex;size:50;not null"`
	Email            string `gorm:"uniqueIndex;size:100;not null"`
	PasswordHash     string `gorm:"not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	ResetToken       string         `gorm:"column:reset_token"`
	ResetTokenExpiry time.Time      `gorm:"column:reset_token_expiry"`

	// Profile extension fields
	FirstName            string `gorm:"size:50"`
	LastName             string `gorm:"size:50"`
	ProfileImage         string `gorm:"size:255"`
	PhoneNumber          string `gorm:"size:20"`
	Currency             string `gorm:"size:10;default:'USD'"`
	NotificationsEnabled bool   `gorm:"default:true"`
	Theme                string `gorm:"size:20;default:'light'"`
}

// ProfileResponse represents the public-facing profile data
type ProfileResponse struct {
	Username             string `json:"username"`
	Email                string `json:"email"`
	FirstName            string `json:"firstName,omitempty"`
	LastName             string `json:"lastName,omitempty"`
	ProfileImage         string `json:"profileImage,omitempty"`
	PhoneNumber          string `json:"phoneNumber,omitempty"`
	Currency             string `json:"currency"`
	NotificationsEnabled bool   `json:"notificationsEnabled"`
	Theme                string `json:"theme"`
}
