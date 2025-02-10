package db

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
)

var DB *gorm.DB

// InitDB connects to MySQL database.
func InitDB(cfg *config.Config, logger *zap.Logger) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to the database", zap.Error(err))
		return err
	}

	DB = db
	logger.Info("Database connection established")
	return nil
}
