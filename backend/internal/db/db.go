package db

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
)

var DB *gorm.DB

// InitDB connects to the database based on config.DBType.
// Currently supports MySQL or PostgreSQL.
func InitDB(cfg *config.Config, logger *zap.Logger) error {
	var dsn string
	var dialector gorm.Dialector

	if cfg.DBType == "mysql" {
		// Example DSN: username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		dialector = mysql.Open(dsn)
	} else if cfg.DBType == "postgres" {
		// Example DSN: host=localhost user=postgres password=secret dbname=github.com/RedShawn258/FinTrack port=5432 sslmode=disable
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort)
		dialector = postgres.Open(dsn)
	} else {
		logger.Error("Unsupported DB type in config", zap.String("dbType", cfg.DBType))
		return fmt.Errorf("unsupported db type: %s", cfg.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to the database", zap.Error(err))
		return err
	}

	DB = db

	logger.Info("Database connection established", zap.String("dbType", cfg.DBType))
	return nil
}
