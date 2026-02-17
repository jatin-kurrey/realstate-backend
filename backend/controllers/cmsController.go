package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"realstate-backend/config"
	"realstate-backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSiteConfig returns all public site configurations
func GetSiteConfig(c *gin.Context) {
	var configs []models.SiteConfig
	if err := config.DB.Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configurations"})
		return
	}

	// Convert to map for easier frontend consumption
	configMap := make(map[string]string)
	for _, conf := range configs {
		configMap[conf.Key] = conf.Value
	}

	c.JSON(http.StatusOK, configMap)
}

// UpdateSiteConfig updates multiple configurations (Admin only)
func UpdateSiteConfig(c *gin.Context) {
	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := config.DB.Begin()
	for key, value := range input {
		var cfg models.SiteConfig
		if err := tx.Where("key = ?", key).First(&cfg).Error; err != nil {
			// Create new if not exists
			if createErr := tx.Create(&models.SiteConfig{Key: key, Value: value, Group: "general", Type: "text"}).Error; createErr != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create " + key})
				return
			}
		} else {
			// Update existing
			if updateErr := tx.Model(&cfg).Update("value", value).Error; updateErr != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update " + key})
				return
			}
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

// UploadImage handles image upload locally
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Create uploads directory if not exists
	uploadPath := "./uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.Mkdir(uploadPath, os.ModePerm)
	}

	// Secure filename (timestamp + original name)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	dst := filepath.Join(uploadPath, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return public URL
	publicURL := fmt.Sprintf("/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": publicURL})
}
