package db

import (
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

// RunMigrations performs database migrations for all models
func RunMigrations(logger *zap.Logger) error {
	if DB == nil {
		logger.Error("Database connection not initialized")
		return nil
	}

	logger.Info("Running database migrations")

	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes
	err := DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Budget{},
		&models.Transaction{},
		&models.Badge{},
		&models.UserBadge{},
		&models.UserPoints{},
	)

	if err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		return err
	}

	// Seed default badges if they don't exist
	seedDefaultBadges(logger)

	logger.Info("Database migrations completed successfully")
	return nil
}

// seedDefaultBadges adds the default badges to the database if they don't exist
func seedDefaultBadges(logger *zap.Logger) {
	var count int64
	DB.Model(&models.Badge{}).Count(&count)

	// Only seed if no badges exist
	if count == 0 {
		badges := []models.Badge{
			{
				Name:        "Budget Beginner",
				Description: "Created your first budget",
				ImageURL:    "/images/badges/budget_beginner.png",
				Category:    "budgeting",
				Threshold:   10,
			},
			{
				Name:        "Tracking Pro",
				Description: "Tracked expenses for 7 consecutive days",
				ImageURL:    "/images/badges/tracking_pro.png",
				Category:    "consistency",
				Threshold:   50,
			},
			{
				Name:        "Savings Star",
				Description: "Saved 10% of your income",
				ImageURL:    "/images/badges/savings_star.png",
				Category:    "savings",
				Threshold:   100,
			},
			{
				Name:        "Budget Master",
				Description: "Stayed under budget for 3 consecutive months",
				ImageURL:    "/images/badges/budget_master.png",
				Category:    "budgeting",
				Threshold:   150,
			},
			{
				Name:        "Finance Ninja",
				Description: "Created budgets in all essential categories",
				ImageURL:    "/images/badges/finance_ninja.png",
				Category:    "budgeting",
				Threshold:   200,
			},
			{
				Name:        "Expense Tracker",
				Description: "Logged 100 transactions",
				ImageURL:    "/images/badges/expense_tracker.png",
				Category:    "tracking",
				Threshold:   250,
			},
			{
				Name:        "Financial Wizard",
				Description: "Reached 500 total points",
				ImageURL:    "/images/badges/financial_wizard.png",
				Category:    "achievement",
				Threshold:   500,
			},
		}

		for _, badge := range badges {
			if err := DB.Create(&badge).Error; err != nil {
				logger.Error("Failed to seed badge", zap.Error(err), zap.String("badge", badge.Name))
			}
		}

		logger.Info("Seeded default badges")
	}
}
