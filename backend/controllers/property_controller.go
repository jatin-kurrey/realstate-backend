package controllers

import (
	"fmt"
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"

	"github.com/gin-gonic/gin"
)

func GetProperties(c *gin.Context) {
	var properties []models.Property
	if err := config.DB.Preload("Owner").Where("is_active = ?", true).Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Anonymize if needed
	for i := range properties {
		if properties[i].Owner.PublicPreference == "Anonymized" {
			properties[i].Owner.Name = "Member " + fmt.Sprint(properties[i].Owner.ID+200)
		}
	}

	c.JSON(http.StatusOK, properties)
}

func GetMyProperties(c *gin.Context) {
	userID, _ := c.Get("userID")
	var properties []models.Property
	if err := config.DB.Where("owner_id = ?", userID).Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, properties)
}

func GetProperty(c *gin.Context) {
	id := c.Param("id")
	var property models.Property
	if err := config.DB.Preload("Owner").First(&property, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	// Anonymize if needed
	if property.Owner.PublicPreference == "Anonymized" {
		property.Owner.Name = "Member " + fmt.Sprint(property.Owner.ID+200)
	}

	c.JSON(http.StatusOK, property)
}

func CreateProperty(c *gin.Context) {
	var property models.Property
	if err := c.ShouldBindJSON(&property); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	property.OwnerID = userID
	property.IsActive = true    // Default to active (Direct Listing)
	property.IsVerified = false // Pending Approval

	if err := config.DB.Create(&property).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create property"})
		return
	}

	// Auto-Reply Notification
	notification := models.Notification{
		UserID:  property.OwnerID,
		Content: "Your property '" + property.Title + "' has been successfully listed! It is now visible to all users.",
		Type:    "system",
		IsRead:  false,
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusCreated, property)
}

func UpdateProperty(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("userID")
	userRole := c.GetString("role")

	var property models.Property
	if err := config.DB.First(&property, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	// Check ownership or admin
	if property.OwnerID != userID && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this property"})
		return
	}

	var input models.Property
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force re-verification on edit
	input.IsVerified = false
	input.OwnerID = property.OwnerID // Prevent changing owner

	config.DB.Model(&property).Updates(input)
	// Ensure IsVerified is explicitly set to false in DB (Updates might skip zero values if struct)
	// But bool false is zero value. GORM Updates using struct ignores zero values.
	// So IsVerified=false might be ignored if it was true.
	// We must use map or explicit Select.
	config.DB.Model(&property).Select("IsVerified").Updates(map[string]interface{}{"is_verified": false})

	c.JSON(http.StatusOK, property)
}

func DeleteProperty(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("userID")
	userRole := c.GetString("role")

	var property models.Property
	if err := config.DB.First(&property, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	if property.OwnerID != userID && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this property"})
		return
	}

	if err := config.DB.Delete(&property).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete property"})
		return
	}

	// Notify if Admin deleted it (Reject) and it wasn't the owner
	if userRole == "admin" && property.OwnerID != userID {
		content := "Your property '" + property.Title + "' has been removed by the administrator."
		notification := models.Notification{
			UserID:  property.OwnerID,
			Content: content,
			Type:    "system",
			IsRead:  false,
		}
		config.DB.Create(&notification)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Property deleted successfully"})
}
