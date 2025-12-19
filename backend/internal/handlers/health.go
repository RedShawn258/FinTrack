package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// HealthHandler returns a simple health check response
// This endpoint is used by Kubernetes liveness and readiness probes
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: "fintrack-api",
		Version: "1.0.0",
	})
}

// ReadinessHandler checks if the service is ready to serve traffic
// Checks database connectivity with strict timeout
func ReadinessHandler(c *gin.Context) {
	logger, exists := c.Get("logger")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logger not found"})
		return
	}
	log := logger.(*zap.Logger)

	// If DATABASE_URL is not set, consider it optional and return ready with warning
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("DB_HOST") == "" {
		log.Warn("No database configured; returning ready (database may be optional)")
		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"service": "fintrack-api",
			"warning": "database not configured",
		})
		return
	}

	// Check database connectivity
	if db.DB == nil {
		log.Warn("Database connection not initialized")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database not initialized",
		})
		return
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Error("Failed to get underlying database connection", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database connection error",
		})
		return
	}

	// Ping the database with strict timeout (1.5 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	// Use PingContext for timeout-aware ping
	pingDone := make(chan error, 1)
	go func() {
		pingDone <- sqlDB.PingContext(ctx)
	}()

	select {
	case err := <-pingDone:
		if err != nil {
			log.Error("Database ping failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "database ping failed",
			})
			return
		}
	case <-ctx.Done():
		log.Error("Database ping timeout", zap.Error(ctx.Err()))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database ping timeout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "fintrack-api",
	})
}
