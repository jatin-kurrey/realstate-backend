package controllers

import (
	"fmt"
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"

	"github.com/gin-gonic/gin"
)

func GetRequirements(c *gin.Context) {
	var requirements []models.Requirement
	if err := config.DB.Preload("User").Where("is_active = ?", true).Find(&requirements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Anonymize if needed
	for i := range requirements {
		if requirements[i].User.PublicPreference == "Anonymized" {
			requirements[i].User.Name = "Visitor " + fmt.Sprint(requirements[i].User.ID+100)
		}
	}

	c.JSON(http.StatusOK, requirements)
}

func CreateRequirement(c *gin.Context) {
	var requirement models.Requirement
	if err := c.ShouldBindJSON(&requirement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If logged in, associate with user
	if userID, exists := c.Get("userID"); exists {
		requirement.UserID = userID.(uint)
	}
	requirement.IsActive = true    // Default to active (Direct Listing)
	requirement.IsVerified = false // Pending Approval

	if err := config.DB.Create(&requirement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create requirement"})
		return
	}

	// Auto-Reply Notification
	notification := models.Notification{
		UserID:  requirement.UserID,
		Content: "Your requirement for '" + requirement.Type + "' has been successfully posted! Property owners will contact you soon.",
		Type:    "system",
		IsRead:  false,
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusCreated, requirement)
}

func UpdateRequirement(c *gin.Context) {
	id := c.Param("id")
	var requirement models.Requirement
	if err := config.DB.First(&requirement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Requirement not found"})
		return
	}

	// Authorization
	userID, _ := c.Get("userID")
	userRole := c.GetString("role")
	uid, _ := userID.(uint)

	if requirement.UserID != uid && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	var updateData models.Requirement
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fields allowed to update
	requirement.Purpose = updateData.Purpose
	requirement.Type = updateData.Type
	requirement.MinBudget = updateData.MinBudget
	requirement.MaxBudget = updateData.MaxBudget
	requirement.Location = updateData.Location
	requirement.MinArea = updateData.MinArea
	requirement.MaxArea = updateData.MaxArea
	requirement.Description = updateData.Description
	requirement.ContactMethod = updateData.ContactMethod

	if err := config.DB.Save(&requirement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update requirement"})
		return
	}

	c.JSON(http.StatusOK, requirement)
}

func DeleteRequirement(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID") // might be nil if admin not logged in via same flow? No, AuthMiddleware sets it.
	userRole := c.GetString("role")

	var requirement models.Requirement
	if err := config.DB.First(&requirement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Requirement not found"})
		return
	}

	// Check ownership or admin
	// Note: userID from context is interface{}, need to cast to uint
	uid, ok := userID.(uint)
	if !ok {
		// If auth middleware fails to set uint, fallback or error.
		// Assuming AuthMiddleware sets it correctly as uint.
	}

	if requirement.UserID != uid && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this requirement"})
		return
	}

	if err := config.DB.Delete(&requirement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete requirement"})
		return
	}

	// Notify if Admin deleted it
	if userRole == "admin" && requirement.UserID != uid {
		content := "Your requirement for '" + requirement.Type + "' has been removed by the administrator."
		notification := models.Notification{
			UserID:  requirement.UserID,
			Content: content,
			Type:    "system",
			IsRead:  false,
		}
		config.DB.Create(&notification)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Requirement deleted successfully"})
}
