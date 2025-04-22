package models

import (
	"time"
)

// Badge represents an achievement badge that users can earn
type Badge struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Description string `gorm:"size:255;not null" json:"description"`
	ImageURL    string `gorm:"size:255" json:"imageUrl"`
	Category    string `gorm:"size:50;not null" json:"category"` // e.g., "savings", "budgeting", "consistency"
	Threshold   int    `gorm:"not null" json:"threshold"`        // Minimum points or condition needed to earn
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserBadge represents badges earned by a user
type UserBadge struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	BadgeID   uint      `gorm:"index;not null" json:"badgeId"`
	Badge     Badge     `gorm:"foreignKey:BadgeID" json:"badge"`
	EarnedAt  time.Time `json:"earnedAt"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserPoints tracks point history for a user
type UserPoints struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userId"`
	Points       int       `gorm:"not null" json:"points"`
	Reason       string    `gorm:"size:255;not null" json:"reason"`
	ActivityType string    `gorm:"size:50;not null" json:"activityType"` // e.g., "budget_created", "transaction_added"
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time
}

// GamificationSummary is the user's gamification status response
type GamificationSummary struct {
	TotalPoints       int          `json:"totalPoints"`
	RecentBadges      []UserBadge  `json:"recentBadges"`
	AllBadges         []UserBadge  `json:"allBadges"`
	RecentPoints      []UserPoints `json:"recentPoints"`
	NextBadges        []Badge      `json:"nextBadges"`
	Level             int          `json:"level"`
	LevelTitle        string       `json:"levelTitle"`
	PointsToNextLevel int          `json:"pointsToNextLevel"`
}
