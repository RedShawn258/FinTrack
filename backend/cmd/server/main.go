package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
	"github.com/RedShawn258/FinTrack/backend/internal/routes"
)

func main() {
	// 1. Load .env (if present in project root)
	// This will set environment variables for the current process
	// so config.LoadConfig() can pick them up.
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or failed to load")
	}

	// 2. Load configuration from environment
	// (Reads environment variables, falling back to defaults if not set)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 3. Initialize Logger
	// Switch between production and development modes based on ENV variable
	var logger *zap.Logger
	if cfg.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync() // flushes buffer, if any
	logger.Info("Starting FinTrack server", zap.String("environment", cfg.Env))

	// 4. Connect to Database (MySQL or PostgreSQL)
	// db.InitDB reads cfg.DBType, cfg.DBHost, cfg.DBPort, etc.
	if err := db.InitDB(cfg, logger); err != nil {
		logger.Fatal("Database initialization failed", zap.Error(err))
	}

	// 5. Run GORM Auto-Migrate for the User model (and any future models)
	if err := db.DB.AutoMigrate(&models.User{}); err != nil {
		logger.Fatal("Failed to auto-migrate User model", zap.Error(err))
	}
	logger.Info("Database migration successful")

	// 6. Set up Gin router
	r := gin.Default()

	// 7. Define routes (public & protected)
	routes.SetupRoutes(r, logger, cfg.JWTSecret)

	// 8. Run the server
	addr := ":" + cfg.ServerPort
	logger.Info("Server listening", zap.String("port", cfg.ServerPort))

	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to run server", zap.Error(err))
	}
}
