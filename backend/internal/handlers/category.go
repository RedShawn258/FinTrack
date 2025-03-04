package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/RedShawn258/FinTrack/backend/internal/db"
	"github.com/RedShawn258/FinTrack/backend/internal/models"
)

type CategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateCategory: Overwrite if (user_id, name) already exists; else create new.
func CreateCategory(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid category creation data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)

	var existing models.Category
	err := db.DB.Where("user_id = ? AND name = ?", userID, name).First(&existing).Error
	if err == nil {
		// Overwrite existing
		existing.Name = name
		if saveErr := db.DB.Save(&existing).Error; saveErr != nil {
			log.Error("Failed to update category", zap.Error(saveErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update category"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":  "Category already exists, overwriting.",
			"category": existing,
		})
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new
		newCat := models.Category{
			UserID: userID,
			Name:   name,
		}
		if createErr := db.DB.Create(&newCat).Error; createErr != nil {
			log.Error("Failed to create category", zap.Error(createErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create category"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message":  "Category created successfully",
			"category": newCat,
		})
		return
	}

	log.Error("Database error checking category", zap.Error(err))
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
}

func GetCategories(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)

	var categories []models.Category
	if err := db.DB.Where("user_id = ?", userID).Find(&categories).Error; err != nil {
		log.Error("Failed to fetch categories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func DeleteCategory(c *gin.Context) {
	logger, _ := c.Get("logger")
	log := logger.(*zap.Logger)
	userID := c.MustGet("userID").(uint)
	categoryID := c.Param("id")

	if err := db.DB.Where("id = ? AND user_id = ?", categoryID, userID).Delete(&models.Category{}).Error; err != nil {
		log.Error("Failed to delete category", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found or could not be deleted"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
