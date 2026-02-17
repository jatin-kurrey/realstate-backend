package controllers

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func ToggleUserBan(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// For now, simulate ban by soft deleting
	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ban user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Unscoped().Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to permanently delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted permanently"})
}

func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Model(&models.User{}).Where("id = ?", id).Update("role", input.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

func TogglePropertyVerification(c *gin.Context) {
	id := c.Param("id")
	var property models.Property
	if err := config.DB.First(&property, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	property.IsVerified = !property.IsVerified
	if err := config.DB.Save(&property).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update property"})
		return
	}

	// Create Notification
	status := "Rejected"
	if property.IsVerified {
		status = "Approved"
	}
	content := "Your property '" + property.Title + "' has been " + status + "."
	notification := models.Notification{
		UserID:  property.OwnerID,
		Content: content,
		Type:    "system",
		IsRead:  false,
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusOK, property)
}

func TogglePropertyFeatured(c *gin.Context) {
	id := c.Param("id")
	var property models.Property
	if err := config.DB.First(&property, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	property.IsFeatured = !property.IsFeatured
	config.DB.Save(&property)
	c.JSON(http.StatusOK, property)
}

func TogglePropertyActive(c *gin.Context) {
	id := c.Param("id")
	var property models.Property
	if err := config.DB.First(&property, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	property.IsActive = !property.IsActive
	config.DB.Save(&property)

	// Create Notification
	status := "Deactivated"
	if property.IsActive {
		status = "Activated"
	}
	content := "Your property '" + property.Title + "' is now " + status + "."
	notification := models.Notification{
		UserID:  property.OwnerID,
		Content: content,
		Type:    "system",
		IsRead:  false,
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusOK, property)
}

func GetStats(c *gin.Context) {
	var userCount, propertyCount, requirementCount int64
	var totalRevenue float64

	config.DB.Model(&models.User{}).Count(&userCount)
	config.DB.Model(&models.Property{}).Count(&propertyCount)
	config.DB.Model(&models.Requirement{}).Count(&requirementCount)

	// Calculate revenue from Success payments
	config.DB.Model(&models.Payment{}).Where("status = ?", "Success").Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)

	c.JSON(http.StatusOK, gin.H{
		"users":        userCount,
		"properties":   propertyCount,
		"requirements": requirementCount,
		"revenue":      totalRevenue,
	})
}

func ToggleRequirementVerification(c *gin.Context) {
	id := c.Param("id")
	var requirement models.Requirement
	if err := config.DB.First(&requirement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Requirement not found"})
		return
	}

	requirement.IsVerified = !requirement.IsVerified
	if err := config.DB.Save(&requirement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update requirement"})
		return
	}

	// Create Notification
	status := "Rejected"
	if requirement.IsVerified {
		status = "Approved"
	}
	content := "Your requirement for '" + requirement.Type + "' in " + requirement.Location + " has been " + status + "."
	notification := models.Notification{
		UserID:  requirement.UserID,
		Content: content,
		Type:    "system",
		IsRead:  false,
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusOK, requirement)
}

func ToggleRequirementActive(c *gin.Context) {
	id := c.Param("id")
	var requirement models.Requirement
	if err := config.DB.First(&requirement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Requirement not found"})
		return
	}

	requirement.IsActive = !requirement.IsActive
	if err := config.DB.Save(&requirement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update requirement"})
		return
	}

	// Create Notification
	status := "Deactivated"
	if requirement.IsActive {
		status = "Activated"
	}
	content := "Your requirement for '" + requirement.Type + "' in " + requirement.Location + " is now " + status + "."
	notification := models.Notification{
		UserID:  requirement.UserID,
		Content: content,
		Type:    "system",
		IsRead:  false,
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusOK, requirement)
}

func GetAllProperties(c *gin.Context) {
	var properties []models.Property
	if err := config.DB.Preload("Owner").Order("id desc").Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, properties)
}

func GetAllRequirements(c *gin.Context) {
	var requirements []models.Requirement
	if err := config.DB.Preload("User").Order("id desc").Find(&requirements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, requirements)
}

func ExportData(c *gin.Context) {
	var users []models.User
	var properties []models.Property
	var requirements []models.Requirement
	var payments []models.Payment

	// Fetch all data
	config.DB.Find(&users)
	config.DB.Preload("Owner").Find(&properties)
	config.DB.Preload("User").Find(&requirements)
	config.DB.Preload("User").Find(&payments)

	// Create Zip Writer
	c.Header("Content-Disposition", "attachment; filename=rjg_system_export.zip")
	c.Header("Content-Type", "application/zip")

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	// Helper to write CSV to zip
	writeCSV := func(filename string, header []string, rows [][]string) {
		f, err := zipWriter.Create(filename)
		if err != nil {
			return
		}
		w := csv.NewWriter(f)
		w.Write(header)
		w.WriteAll(rows)
		w.Flush()
	}

	// 1. Users CSV
	var userRows [][]string
	for _, u := range users {
		userRows = append(userRows, []string{
			strconv.Itoa(int(u.ID)), u.Name, u.Email, u.Phone, u.Role, u.CompanyName,
			u.CreatedAt.Format("2006-01-02"),
		})
	}
	writeCSV("users.csv", []string{"ID", "Name", "Email", "Phone", "Role", "Company", "Joined Date"}, userRows)

	// 2. Properties CSV
	var propRows [][]string
	for _, p := range properties {
		ownerName := "Unknown"
		if p.Owner.Name != "" {
			ownerName = p.Owner.Name
		}
		propRows = append(propRows, []string{
			strconv.Itoa(int(p.ID)), p.Title, p.Type, p.Status,
			strconv.FormatFloat(p.Price, 'f', 2, 64),
			p.Location, ownerName,
			strconv.FormatBool(p.IsVerified), strconv.FormatBool(p.IsActive),
			p.CreatedAt.Format("2006-01-02"),
		})
	}
	writeCSV("properties.csv", []string{"ID", "Title", "Type", "Status", "Price", "Location", "Owner", "Verified", "Active", "Date"}, propRows)

	// 3. Requirements CSV
	var reqRows [][]string
	for _, r := range requirements {
		userName := "Visitor"
		if r.User.Name != "" {
			userName = r.User.Name
		}
		reqRows = append(reqRows, []string{
			strconv.Itoa(int(r.ID)), r.Type, r.Purpose,
			fmt.Sprintf("%v-%v", r.MinBudget, r.MaxBudget),
			r.Location, userName,
			strconv.FormatBool(r.IsVerified),
			r.CreatedAt.Format("2006-01-02"),
		})
	}
	writeCSV("requirements.csv", []string{"ID", "Type", "Purpose", "Budget Range", "Location", "User", "Verified", "Date"}, reqRows)

	// 4. Payments
	var payRows [][]string
	for _, pay := range payments {
		userName := "Unknown"
		if pay.User.Name != "" {
			userName = pay.User.Name
		}
		payRows = append(payRows, []string{
			strconv.Itoa(int(pay.ID)), userName,
			strconv.FormatFloat(pay.Amount, 'f', 2, 64),
			pay.Status, pay.Plan,
			pay.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	writeCSV("payments.csv", []string{"ID", "User", "Amount", "Status", "Plan", "Date"}, payRows)

	// 5. Summary Text
	f, _ := zipWriter.Create("summary.txt")
	summary := fmt.Sprintf("System Summary Report\nGenerated: %s\n\nTotal Users: %d\nTotal Properties: %d\nTotal Requirements: %d\nTotal Payments: %d\n",
		time.Now().Format(time.RFC1123), len(users), len(properties), len(requirements), len(payments))
	f.Write([]byte(summary))
}

func SendUserMessage(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	notification := models.Notification{
		UserID:  user.ID,
		Content: "Message from Admin: " + input.Content,
		Type:    "system",
		IsRead:  false,
	}
	config.DB.Create(&notification)

	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully"})
}
