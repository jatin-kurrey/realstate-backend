package controllers

import (
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"

	"github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var input struct {
		Name               string `json:"name"`
		Phone              string `json:"phone"`
		PublicPreference   string `json:"public_preference"`
		ContactPreference  string `json:"contact_preference"`
		EmailNotifications bool   `json:"email_notifications"`
		InAppNotifications bool   `json:"in_app_notifications"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	updates := models.User{
		Name:               input.Name,
		Phone:              input.Phone,
		PublicPreference:   input.PublicPreference,
		ContactPreference:  input.ContactPreference,
		EmailNotifications: input.EmailNotifications,
		InAppNotifications: input.InAppNotifications,
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func DeactivateAccount(c *gin.Context) {
	userID, _ := c.Get("userID")

	// GORM's Delete on a model with DeletedAt will perform a soft-delete
	if err := config.DB.Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deactivated successfully"})
}
