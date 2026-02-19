package controllers

import (
	"net/http"
	"realstate-backend/config"
	"realstate-backend/models"
	"realstate-backend/ws"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	Hub *ws.Hub
}

func NewChatController(hub *ws.Hub) *ChatController {
	return &ChatController{Hub: hub}
}

func (cc *ChatController) GetMyThreads(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var threads []models.ChatThread
	if err := config.DB.Preload("Participant1").Preload("Participant2").Preload("Property").
		Where("participant1_id = ? OR participant2_id = ?", userID, userID).
		Order("updated_at desc").Find(&threads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch threads"})
		return
	}

	// Calculate unread counts
	for i := range threads {
		var count int64
		config.DB.Model(&models.ChatMessage{}).
			Where("thread_id = ? AND sender_id != ? AND is_read = ?", threads[i].ID, userID, false).
			Count(&count)
		threads[i].UnreadCount = int(count)
	}

	c.JSON(http.StatusOK, threads)
}

func (cc *ChatController) MarkThreadRead(c *gin.Context) {
	threadID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	// Mark all messages in this thread as read where user is not the sender
	if err := config.DB.Model(&models.ChatMessage{}).
		Where("thread_id = ? AND sender_id != ? AND is_read = ?", threadID, userID, false).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (cc *ChatController) GetThreadMessages(c *gin.Context) {
	threadID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	// Parse pagination parameters
	page := 1
	limit := 50
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var thread models.ChatThread
	if err := config.DB.First(&thread, threadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	if thread.Participant1ID != userID && thread.Participant2ID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get paginated messages
	var messages []models.ChatMessage
	offset := (page - 1) * limit

	query := config.DB.Preload("Sender").Preload("ReplyTo.Sender").
		Where("thread_id = ? AND is_deleted = ?", threadID, false).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	// Update message delivery status for recipient
	go func() {
		config.DB.Model(&models.ChatMessage{}).
			Where("thread_id = ? AND sender_id != ? AND status = ?",
				threadID, userID, models.MessageStatusSent).
			Updates(map[string]interface{}{
				"status":       models.MessageStatusDelivered,
				"delivered_at": time.Now(),
			})

		// Notify sender of delivery status update
		var deliveredMessages []models.ChatMessage
		config.DB.Where("thread_id = ? AND sender_id != ? AND status = ?",
			threadID, userID, models.MessageStatusDelivered).Find(&deliveredMessages)

		for _, msg := range deliveredMessages {
			cc.Hub.BroadcastToUser(msg.SenderID, gin.H{
				"type":       "MESSAGE_STATUS_UPDATE",
				"message_id": msg.ID,
				"status":     models.MessageStatusDelivered,
			})
		}
	}()

	// Get total message count for pagination
	var totalCount int64
	config.DB.Model(&models.ChatMessage{}).Where("thread_id = ? AND is_deleted = ?", threadID, false).Count(&totalCount)

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": totalCount,
			"pages": (int(totalCount) + limit - 1) / limit,
		},
	})
}

func (cc *ChatController) CreateThread(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var input struct {
		TargetUserID uint   `json:"target_user_id" binding:"required"`
		PropertyID   *uint  `json:"property_id"`
		Message      string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if userID == input.TargetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot message yourself"})
		return
	}

	// Check if thread already exists
	var thread models.ChatThread
	query := config.DB.Where("(participant1_id = ? AND participant2_id = ?) OR (participant1_id = ? AND participant2_id = ?)",
		userID, input.TargetUserID, input.TargetUserID, userID)

	if input.PropertyID != nil {
		query = query.Where("property_id = ?", input.PropertyID)
	} else {
		query = query.Where("property_id IS NULL")
	}

	if err := query.First(&thread).Error; err != nil {
		// Create new thread
		thread = models.ChatThread{
			Participant1ID: userID,
			Participant2ID: input.TargetUserID,
			PropertyID:     input.PropertyID,
			LastMessage:    input.Message,
			UpdatedAt:      time.Now(),
		}
		if err := config.DB.Create(&thread).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
			return
		}
	} else {
		// Update existing thread
		thread.LastMessage = input.Message
		thread.UpdatedAt = time.Now()
		config.DB.Save(&thread)
	}

	// Create message
	message := models.ChatMessage{
		ThreadID:  thread.ID,
		SenderID:  userID,
		Content:   input.Message,
		CreatedAt: time.Now(),
	}
	config.DB.Create(&message)

	// Preload for the response
	config.DB.Preload("Participant1").Preload("Participant2").Preload("Property").First(&thread, thread.ID)

	// Broadcast via WebSocket
	cc.Hub.BroadcastToUser(input.TargetUserID, gin.H{
		"type":    "NEW_MESSAGE",
		"thread":  thread,
		"message": message,
	})

	c.JSON(http.StatusCreated, gin.H{"thread": thread, "message": message})
}

func (cc *ChatController) SendChatMessage(c *gin.Context) {
	threadID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var input struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var thread models.ChatThread
	if err := config.DB.First(&thread, threadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	if thread.Participant1ID != userID && thread.Participant2ID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	message := models.ChatMessage{
		ThreadID:  thread.ID,
		SenderID:  userID,
		Content:   input.Content,
		CreatedAt: time.Now(),
	}

	if err := config.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Update thread
	thread.LastMessage = input.Content
	thread.UpdatedAt = message.CreatedAt
	config.DB.Save(&thread)

	// Notify other participant
	recipientID := thread.Participant2ID
	if userID == thread.Participant2ID {
		recipientID = thread.Participant1ID
	}

	// Preload thread details for the broadcast
	config.DB.Preload("Participant1").Preload("Participant2").Preload("Property").First(&thread, thread.ID)

	// Notify recipient
	cc.Hub.BroadcastToUser(recipientID, gin.H{
		"type":    "NEW_MESSAGE",
		"thread":  thread,
		"message": message,
	})

	// Notify sender's other tabs
	cc.Hub.BroadcastToUser(userID, gin.H{
		"type":    "NEW_MESSAGE",
		"thread":  thread,
		"message": message,
	})

	c.JSON(http.StatusCreated, message)
}

func (cc *ChatController) EditMessage(c *gin.Context) {
	messageID := c.Param("messageId")
	userID := c.MustGet("userID").(uint)

	var input struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var message models.ChatMessage
	if err := config.DB.First(&message, messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	if message.SenderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Can only edit your own messages"})
		return
	}

	// Check if message is older than 15 minutes (prevent editing old messages)
	if time.Since(message.CreatedAt) > 15*time.Minute {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot edit messages older than 15 minutes"})
		return
	}

	now := time.Now()
	message.Content = input.Content
	message.IsEdited = true
	message.EditedAt = &now

	if err := config.DB.Save(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit message"})
		return
	}

	// Notify all participants in the thread
	var thread models.ChatThread
	config.DB.First(&thread, message.ThreadID)

	recipientID := thread.Participant2ID
	if userID == thread.Participant2ID {
		recipientID = thread.Participant1ID
	}

	cc.Hub.BroadcastToUser(recipientID, gin.H{
		"type":       "MESSAGE_EDITED",
		"message_id": message.ID,
		"content":    message.Content,
		"edited_at":  message.EditedAt,
	})

	cc.Hub.BroadcastToUser(userID, gin.H{
		"type":       "MESSAGE_EDITED",
		"message_id": message.ID,
		"content":    message.Content,
		"edited_at":  message.EditedAt,
	})

	c.JSON(http.StatusOK, message)
}

func (cc *ChatController) DeleteMessage(c *gin.Context) {
	messageID := c.Param("messageId")
	userID := c.MustGet("userID").(uint)

	var message models.ChatMessage
	if err := config.DB.First(&message, messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	if message.SenderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Can only delete your own messages"})
		return
	}

	now := time.Now()
	message.IsDeleted = true
	message.DeletedAt = &now

	if err := config.DB.Save(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	// Notify all participants in the thread
	var thread models.ChatThread
	config.DB.First(&thread, message.ThreadID)

	recipientID := thread.Participant2ID
	if userID == thread.Participant2ID {
		recipientID = thread.Participant1ID
	}

	cc.Hub.BroadcastToUser(recipientID, gin.H{
		"type":       "MESSAGE_DELETED",
		"message_id": message.ID,
	})

	cc.Hub.BroadcastToUser(userID, gin.H{
		"type":       "MESSAGE_DELETED",
		"message_id": message.ID,
	})

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (cc *ChatController) SearchMessages(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Get user's threads first
	var threadIDs []uint
	config.DB.Model(&models.ChatThread{}).
		Where("participant1_id = ? OR participant2_id = ?", userID, userID).
		Pluck("id", &threadIDs)

	if len(threadIDs) == 0 {
		c.JSON(http.StatusOK, []models.MessageSearchResult{})
		return
	}

	// Search messages in user's threads
	var results []models.MessageSearchResult
	searchQuery := "%" + query + "%"

	err := config.DB.Table("chat_messages").
		Select(`
			chat_messages.*, 
			chat_threads.*,
			CASE 
				WHAT chat_threads.participant1_id = ? THEN chat_threads.participant2_id
				ELSE chat_threads.participant1_id
			END as other_participant_id
		`, userID).
		Joins("JOIN chat_threads ON chat_messages.thread_id = chat_threads.id").
		Where("chat_messages.thread_id IN ? AND chat_messages.content LIKE ? AND chat_messages.is_deleted = ?",
			threadIDs, searchQuery, false).
		Order("chat_messages.created_at DESC").
		Limit(50).
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (cc *ChatController) UpdateTypingStatus(c *gin.Context) {
	threadID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var input struct {
		IsTyping bool `json:"is_typing"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var thread models.ChatThread
	if err := config.DB.First(&thread, threadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	if thread.Participant1ID != userID && thread.Participant2ID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Update typing status
	if thread.Participant1ID == userID {
		thread.IsTyping1 = input.IsTyping
	} else {
		thread.IsTyping2 = input.IsTyping
	}

	if err := config.DB.Save(&thread).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update typing status"})
		return
	}

	// Notify other participant
	recipientID := thread.Participant2ID
	if userID == thread.Participant2ID {
		recipientID = thread.Participant1ID
	}

	cc.Hub.BroadcastToUser(recipientID, gin.H{
		"type":      "TYPING_STATUS",
		"thread_id": thread.ID,
		"user_id":   userID,
		"is_typing": input.IsTyping,
	})

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
