package db

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
)

var DB *gorm.DB

// InitDB connects to MySQL database.
func InitDB(cfg *config.Config, zapLogger *zap.Logger) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=30s&writeTimeout=30s&maxAllowedPacket=0",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	gormConfig := &gorm.Config{
		Logger: logger.New(
			nil, // io writer
			logger.Config{
				SlowThreshold:             time.Second,   // Only log queries slower than 1 second
				LogLevel:                  logger.Silent, // Disable SQL logging
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		PrepareStmt: true, // Enable prepared statement cache
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                      dsn,
		DefaultStringSize:        256,
		DisableDatetimePrecision: true,
		DontSupportRenameIndex:   true,
	}), gormConfig)

	if err != nil {
		zapLogger.Error("Failed to connect to the database", zap.Error(err))
		return err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Error("Failed to get underlying *sql.DB", zap.Error(err))
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	zapLogger.Info("Database connection established")
	return nil
}
