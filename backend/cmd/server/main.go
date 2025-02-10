package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/config"
	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
	"github.com/RedShawn258/FinTrack/backend/internal/routes"
)

func main() {
	// Load .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or failed to load")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Logger
	var logger *zap.Logger
	if cfg.Env == "production" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()
	logger.Info("Starting FinTrack server", zap.String("environment", cfg.Env))

	// Connect to Database
	if err := db.InitDB(cfg, logger); err != nil {
		logger.Fatal("Database initialization failed", zap.Error(err))
	}

	// Run Auto-Migrate for User model
	if err := db.DB.AutoMigrate(&models.User{}); err != nil {
		logger.Fatal("Failed to auto-migrate User model", zap.Error(err))
	}
	logger.Info("Database migration successful")

	// Set up Gin router
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Allow frontend origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Middleware to set logger & JWT secret in context for every request
	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Set("jwtSecret", cfg.JWTSecret)
		c.Next()
	})

	// Define routes
	routes.SetupRoutes(r, logger, cfg.JWTSecret)

	// Run the server
	addr := ":" + cfg.ServerPort
	logger.Info("Server listening", zap.String("port", cfg.ServerPort))

	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to run server", zap.Error(err))
	}
}
