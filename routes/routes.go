package routes

import (
	"realstate-backend/controllers"
	"realstate-backend/middleware"
	"realstate-backend/ws"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, hub *ws.Hub) {
	api := r.Group("/api")
	{
		// WebSocket Route
		api.GET("/ws", middleware.AuthMiddleware(), func(c *gin.Context) {
			userID := c.MustGet("userID").(uint)
			ws.ServeWs(hub, c.Writer, c.Request, userID)
		})

		api.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Welcome to RJG Property API", "version": "1.0"})
		})
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/admin-login", controllers.AdminLogin)
			auth.POST("/google", controllers.GoogleLogin)
		}

		// User routes
		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", controllers.GetProfile)
			user.PUT("/profile", controllers.UpdateProfile)
			user.DELETE("/deactivate", controllers.DeactivateAccount)
		}

		// Property routes
		properties := api.Group("/properties")
		{
			properties.GET("", controllers.GetProperties)
			properties.GET("/:id", controllers.GetProperty)

			// Protected routes
			protected := properties.Group("")
			protected.Use(middleware.AuthMiddleware())
			{
				protected.GET("/me", controllers.GetMyProperties)
				protected.POST("", controllers.CreateProperty)
				protected.PUT("/:id", controllers.UpdateProperty)
				protected.DELETE("/:id", controllers.DeleteProperty)
			}

		}

		// Requirement routes
		requirements := api.Group("/requirements")
		{
			requirements.GET("", controllers.GetRequirements)                                 // Public
			requirements.GET("/:id", controllers.GetRequirement)                              // Public
			requirements.POST("", middleware.AuthMiddleware(), controllers.CreateRequirement) // Protected
			requirements.PUT("/:id", middleware.AuthMiddleware(), controllers.UpdateRequirement)
			requirements.DELETE("/:id", middleware.AuthMiddleware(), controllers.DeleteRequirement)

			adminOnly := requirements.Group("")
			adminOnly.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
			{
				// adminOnly.DELETE("/:id", controllers.DeleteRequirement) // Moved up
			}
		}

		// Payment routes
		payments := api.Group("/payments")
		payments.Use(middleware.AuthMiddleware())
		{
			payments.GET("/me", controllers.GetMyPayments)
			payments.POST("", controllers.CreatePayment)
			payments.POST("/process", controllers.ProcessListingPayment)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
		{
			admin.GET("/users", controllers.GetUsers)
			admin.PATCH("/users/:id/role", controllers.UpdateUserRole)
			admin.PATCH("/users/:id/badge", controllers.UpdateUserBadge)
			admin.DELETE("/users/:id", controllers.DeleteUser)
			admin.PATCH("/users/:id/toggle-ban", controllers.ToggleUserBan)
			admin.POST("/users/:id/message", controllers.SendUserMessage)
			admin.GET("/payments", controllers.GetPayments)
			admin.GET("/stats", controllers.GetStats)
			admin.GET("/export", controllers.ExportData)
			admin.GET("/properties", controllers.GetAllProperties)
			admin.GET("/requirements", controllers.GetAllRequirements)
			admin.PATCH("/properties/:id/verify", controllers.TogglePropertyVerification)
			admin.PATCH("/properties/:id/feature", controllers.TogglePropertyFeatured)
			admin.PATCH("/properties/:id/toggle-active", controllers.TogglePropertyActive)
			admin.PATCH("/requirements/:id/verify", controllers.ToggleRequirementVerification)
			admin.PATCH("/requirements/:id/toggle-active", controllers.ToggleRequirementActive)
		}

		// Inquiry routes
		inquiries := api.Group("/inquiries")
		inquiries.Use(middleware.AuthMiddleware())
		{
			inquiries.POST("", controllers.CreateInquiry)
			inquiries.GET("/me", controllers.GetMyInquiries)
			inquiries.GET("/:id", controllers.GetInquiryDetail)
			inquiries.POST("/:id/messages", controllers.SendMessage)
			inquiries.PATCH("/:id/status", controllers.UpdateInquiryStatus)
		}

		// Notification routes
		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware())
		{
			notifications.GET("", controllers.GetNotifications)
			notifications.PATCH("/:id/read", controllers.MarkNotificationRead)
			notifications.POST("/read-all", controllers.MarkAllNotificationsRead)
		}

		// Upload route
		api.POST("/upload", middleware.AuthMiddleware(), controllers.UploadImages)

		// CMS Routes
		api.GET("/config", controllers.GetSiteConfig)
		api.POST("/cms/upload", middleware.AuthMiddleware(), middleware.AdminOnly(), controllers.UploadImage)

		adminCms := api.Group("/admin/cms")
		adminCms.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
		{
			adminCms.PUT("/config", controllers.UpdateSiteConfig)
		}
		// Bookmark routes
		bookmarks := api.Group("/bookmarks")
		bookmarks.Use(middleware.AuthMiddleware())
		{
			bookmarks.GET("", controllers.GetBookmarks)
			bookmarks.POST("/toggle", controllers.ToggleBookmark)
			bookmarks.GET("/check", controllers.IsBookmarked)
		}

		// Chat routes
		chatController := controllers.NewChatController(hub)
		chat := api.Group("/chat")
		chat.Use(middleware.AuthMiddleware())
		{
			chat.GET("/threads", chatController.GetMyThreads)
			chat.GET("/threads/:id", chatController.GetThreadMessages)
			chat.POST("/threads", chatController.CreateThread)
			chat.POST("/threads/:id/messages", chatController.SendChatMessage)
			chat.POST("/threads/:id/read", chatController.MarkThreadRead)
			chat.PUT("/messages/:messageId", chatController.EditMessage)
			chat.DELETE("/messages/:messageId", chatController.DeleteMessage)
			chat.POST("/threads/:id/typing", chatController.UpdateTypingStatus)
			chat.GET("/search", chatController.SearchMessages)
		}

		// God Mode Route - Critical System Access
		god := api.Group("/god")
		{
			god.PUT("/config", func(c *gin.Context) {
				key := c.GetHeader("X-God-Key")
				// Hardcoded key to match frontend - serving as emergency backdoor
				if key != "RJG_GOD_ACCESS_2024" {
					c.JSON(403, gin.H{"error": "Unauthorized God Access"})
					c.Abort()
					return
				}
				controllers.UpdateSiteConfig(c)
			})
		}
	}
}
