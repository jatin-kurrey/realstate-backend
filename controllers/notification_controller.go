package controllers

import (
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"

	"github.com/gin-gonic/gin"
)

func GetNotifications(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var notifications []models.Notification
	if err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)

	if err := config.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func MarkAllNotificationsRead(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	if err := config.DB.Model(&models.Notification{}).Where("user_id = ?", userID).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
