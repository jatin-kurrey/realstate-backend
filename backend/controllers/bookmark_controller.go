package controllers

import (
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"

	"github.com/gin-gonic/gin"
)

// GetBookmarks returns all bookmarks for the authenticated user
func GetBookmarks(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var bookmarks []models.Bookmark
	if err := config.DB.Where("user_id = ?", userID).Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookmarks"})
		return
	}
	c.JSON(http.StatusOK, bookmarks)
}

// ToggleBookmark adds or removes a bookmark
func ToggleBookmark(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var input struct {
		TargetID uint   `json:"target_id" binding:"required"`
		Type     string `json:"type" binding:"required"` // 'property' or 'requirement'
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var bookmark models.Bookmark
	err := config.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, input.TargetID, input.Type).First(&bookmark).Error

	if err == nil {
		// Already exists, so delete it
		if err := config.DB.Delete(&bookmark).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove bookmark"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Bookmark removed", "isBookmarked": false})
	} else {
		// Doesn't exist, so create it
		newBookmark := models.Bookmark{
			UserID:   userID,
			TargetID: input.TargetID,
			Type:     input.Type,
		}
		if err := config.DB.Create(&newBookmark).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add bookmark"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Bookmark added", "isBookmarked": true})
	}
}

// IsBookmarked checks if a specific target is bookmarked by the user
func IsBookmarked(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	targetIDStr := c.Query("target_id")
	targetType := c.Query("type")

	if targetIDStr == "" || targetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_id and type are required"})
		return
	}

	var bookmark models.Bookmark
	err := config.DB.Where("user_id = ? AND target_id = ? AND type = ?", userID, targetIDStr, targetType).First(&bookmark).Error

	if err == nil {
		c.JSON(http.StatusOK, gin.H{"isBookmarked": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"isBookmarked": false})
	}
}
