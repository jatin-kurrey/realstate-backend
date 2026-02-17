package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadImages(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	files := form.File["images"]
	var urls []string

	// Ensure uploads directory exists
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	for _, file := range files {
		// Generate unique filename
		filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		path := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file %s", file.Filename)})
			return
		}

		// In a real app, use the base URL from env
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:5001"
		}
		urls = append(urls, fmt.Sprintf("%s/uploads/%s", baseURL, filename))
	}

	c.JSON(http.StatusOK, gin.H{"urls": urls})
}
