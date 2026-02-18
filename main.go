package main

import (
	"log"
	"net/http"
	"os"
	"realstate-backend/config"
	"realstate-backend/models"
	"realstate-backend/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config.ConnectDB()

	// Auto Migration
	err := config.DB.AutoMigrate(&models.User{}, &models.Property{}, &models.Requirement{}, &models.Payment{}, &models.Inquiry{}, &models.InquiryMessage{}, &models.Notification{}, &models.SiteConfig{}, &models.PageContent{}, &models.Bookmark{})
	if err != nil {
		log.Fatal("Failed to migrate data: ", err)
	}

	// Seed Data
	config.SeedData()

	r := gin.Default()

	// CORS Configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-God-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	r.Static("/uploads", "./uploads")

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
