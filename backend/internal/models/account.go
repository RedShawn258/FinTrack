package models

import (
	"time"

	"gorm.io/gorm"
)

// Account represents a financial account with a balance
type Account struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Balance   float64        `gorm:"type:decimal(15,2);not null;default:0" json:"balance"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Account
func (Account) TableName() string {
	return "accounts"
}
