package controllers

import (
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

func GetPayments(c *gin.Context) {
	var payments []models.Payment
	if err := config.DB.Preload("User").Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payments)
}

func GetMyPayments(c *gin.Context) {
	userID := c.GetUint("userID")
	var payments []models.Payment
	if err := config.DB.Where("user_id = ?", userID).Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payments)
}

func CreatePayment(c *gin.Context) {
	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	payment.UserID = userID

	if err := config.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

func ProcessListingPayment(c *gin.Context) {
	var input struct {
		PropertyID uint    `json:"property_id" binding:"required"`
		Amount     float64 `json:"amount" binding:"required"`
		Plan       string  `json:"plan"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")

	// Simulate success
	payment := models.Payment{
		UserID: userID,
		Amount: input.Amount,
		Status: "Success",
	}

	if err := config.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
		return
	}

	// Update property status
	expiry := time.Now().AddDate(0, 0, 30)
	if err := config.DB.Model(&models.Property{}).Where("id = ?", input.PropertyID).Updates(map[string]interface{}{
		"is_featured": true,
		"expiry_date": expiry,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update property status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment successful!", "payment": payment, "expiry": expiry})
}
