package controllers

import (
	"fmt"
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"

	"github.com/gin-gonic/gin"
)

func CreateInquiry(c *gin.Context) {
	var input struct {
		PropertyID     uint    `json:"property_id" binding:"required"`
		InitialMessage string  `json:"initial_message" binding:"required"`
		ExpectedDate   string  `json:"expected_date"`
		Budget         float64 `json:"budget"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(uint)

	// Fetch property to get OwnerID
	var property models.Property
	if err := config.DB.First(&property, input.PropertyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	inquiry := models.Inquiry{
		PropertyID:     input.PropertyID,
		SeekerID:       userID,
		OwnerID:        property.OwnerID,
		InitialMessage: input.InitialMessage,
		ExpectedDate:   input.ExpectedDate,
		Budget:         input.Budget,
		Status:         "Open",
	}

	if err := config.DB.Create(&inquiry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create inquiry"})
		return
	}

	// Create Notification for the owner
	notification := models.Notification{
		UserID:  property.OwnerID,
		Content: "You have a new inquiry for " + property.Title,
		Type:    "inquiry",
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusCreated, inquiry)
}

func GetMyInquiries(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	role := c.MustGet("userRole").(string)

	var inquiries []models.Inquiry
	query := config.DB.Preload("Property").Preload("Seeker").Preload("Owner")

	if role == "owner" {
		query = query.Where("owner_id = ?", userID)
	} else {
		query = query.Where("seeker_id = ?", userID)
	}

	if err := query.Order("updated_at desc").Find(&inquiries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Anonymize names based on preference
	for i := range inquiries {
		// If I am the owner, anonymize the seeker if they want
		if role == "owner" {
			if inquiries[i].Seeker.PublicPreference == "Anonymized" {
				inquiries[i].Seeker.Name = "Requester " + fmt.Sprint(inquiries[i].Seeker.ID+500)
			}
		} else {
			// If I am the seeker, anonymize the owner if they want
			if inquiries[i].Owner.PublicPreference == "Anonymized" {
				inquiries[i].Owner.Name = "Owner " + fmt.Sprint(inquiries[i].Owner.ID+200)
			}
		}
	}

	c.JSON(http.StatusOK, inquiries)
}

func GetInquiryDetail(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var inquiry models.Inquiry
	if err := config.DB.Preload("Property").Preload("Seeker").Preload("Owner").Preload("Messages.Sender").First(&inquiry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inquiry not found"})
		return
	}

	// Security check
	if inquiry.SeekerID != userID && inquiry.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Anonymize thread participants
	if inquiry.Seeker.PublicPreference == "Anonymized" {
		inquiry.Seeker.Name = "Requester " + fmt.Sprint(inquiry.Seeker.ID+500)
	}
	if inquiry.Owner.PublicPreference == "Anonymized" {
		inquiry.Owner.Name = "Owner " + fmt.Sprint(inquiry.Owner.ID+200)
	}

	for i := range inquiry.Messages {
		if inquiry.Messages[i].SenderID == inquiry.SeekerID {
			inquiry.Messages[i].Sender.Name = inquiry.Seeker.Name
		} else {
			inquiry.Messages[i].Sender.Name = inquiry.Owner.Name
		}
	}

	c.JSON(http.StatusOK, inquiry)
}

func SendMessage(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var input struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var inquiry models.Inquiry
	if err := config.DB.First(&inquiry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inquiry not found"})
		return
	}

	// Check if participant
	if inquiry.SeekerID != userID && inquiry.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	message := models.InquiryMessage{
		InquiryID: inquiry.ID,
		SenderID:  userID,
		Message:   input.Message,
	}

	if err := config.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Update inquiry timestamp
	config.DB.Model(&inquiry).Update("updated_at", message.CreatedAt)

	// Notify the other party
	recipientID := inquiry.OwnerID
	if userID == inquiry.OwnerID {
		recipientID = inquiry.SeekerID
	}

	notification := models.Notification{
		UserID:  recipientID,
		Content: "New message in inquiry thread",
		Type:    "inquiry",
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusCreated, message)
}

func UpdateInquiryStatus(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var inquiry models.Inquiry
	if err := config.DB.First(&inquiry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inquiry not found"})
		return
	}

	// Only owner or seeker involved can change status?
	// Usually admin or the concerned parties.
	if inquiry.SeekerID != userID && inquiry.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	inquiry.Status = input.Status
	config.DB.Save(&inquiry)

	c.JSON(http.StatusOK, inquiry)
}
