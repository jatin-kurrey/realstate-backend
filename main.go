package main

import (
	"log"
	"net/http"
	"os"
	"realstate-backend/config"
	"realstate-backend/models"
	"realstate-backend/routes"
	"realstate-backend/ws"

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
	err := config.DB.AutoMigrate(&models.User{}, &models.Property{}, &models.Requirement{}, &models.Payment{}, &models.Inquiry{}, &models.InquiryMessage{}, &models.Notification{}, &models.SiteConfig{}, &models.PageContent{}, &models.Bookmark{}, &models.ChatThread{}, &models.ChatMessage{})
	if err != nil {
		log.Fatal("Failed to migrate data: ", err)
	}

	// Seed Data
	config.SeedData()

	// Initialize WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	r := gin.Default()

	// CORS Configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://192.168.1.4:3001",
			"https://realstate-frontend-dhfn.vercel.app",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-God-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		log.Println("[HEALTH] Check received")
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	r.Static("/uploads", "./uploads")

	routes.SetupRoutes(r, hub)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for many cloud providers
	}

	// Explicitly bind to 0.0.0.0 for Render/Docker compatibility
	addr := "0.0.0.0:" + port
	log.Printf("Server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
