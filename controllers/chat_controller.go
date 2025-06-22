package controllers

import (
	"net/http"
	"strconv"
	"time"

	"ondrive/middleware"
	"ondrive/models"
	"ondrive/repositories"
	"ondrive/services"
	"ondrive/utils"
	"ondrive/websocket"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatController struct {
	chatService         services.ChatService
	conversationRepo    repositories.ConversationRepository
	messageRepo         repositories.MessageRepository
	fileUploadService   services.UploadService
	notificationService services.NotificationService
	userRepo            repositories.UserRepository
	wsHub               *websocket.Hub
	logger              utils.Logger
}

type CreateConversationRequest struct {
	Type           models.ConversationType `json:"type" validate:"required"`
	Title          string                  `json:"title"`
	ParticipantIDs []string                `json:"participant_ids" validate:"required,min=1"`
	RideID         string                  `json:"ride_id,omitempty"`
}

type SendMessageRequest struct {
	Type         models.MessageType  `json:"type" validate:"required"`
	Content      string              `json:"content"`
	ReplyToID    string              `json:"reply_to_id,omitempty"`
	QuickReplies []models.QuickReply `json:"quick_replies,omitempty"`
}

type SendLocationMessageRequest struct {
	Latitude       float64 `json:"latitude" validate:"required"`
	Longitude      float64 `json:"longitude" validate:"required"`
	Address        string  `json:"address"`
	PlaceName      string  `json:"place_name"`
	IsLiveLocation bool    `json:"is_live_location"`
	LiveDuration   int     `json:"live_duration"` // minutes
}

type UpdateChatSettingsRequest struct {
	MessageNotifications bool `json:"message_notifications"`
	SoundEnabled         bool `json:"sound_enabled"`
	VibrationEnabled     bool `json:"vibration_enabled"`
	ReadReceipts         bool `json:"read_receipts"`
	TypingIndicator      bool `json:"typing_indicator"`
	AutoDownloadMedia    bool `json:"auto_download_media"`
	ShowLastSeen         bool `json:"show_last_seen"`
	ShowOnlineStatus     bool `json:"show_online_status"`
}

type BlockUserRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Reason string `json:"reason"`
}

type CreateQuickReplyRequest struct {
	Text     string `json:"text" validate:"required"`
	Category string `json:"category"`
}

func NewChatController(
	chatService services.ChatService,
	conversationRepo repositories.ConversationRepository,
	messageRepo repositories.MessageRepository,
	userRepo repositories.UserRepository,
	fileUploadService services.UploadService,
	notificationService services.NotificationService,
	wsHub *websocket.Hub,
	logger utils.Logger,
) *ChatController {
	return &ChatController{
		chatService:         chatService,
		conversationRepo:    conversationRepo,
		messageRepo:         messageRepo,
		userRepo:            userRepo,
		fileUploadService:   fileUploadService,
		notificationService: notificationService,
		wsHub:               wsHub,
		logger:              logger,
	}
}

// Chat Management

func (cc *ChatController) GetConversations(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	// Get pagination params
	params := utils.GetPaginationParams(c)

	conversations, total, err := cc.conversationRepo.GetUserConversations(userID, params.Page, params.Limit)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get conversations")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Conversations retrieved successfully", conversations, meta)
}

func (cc *ChatController) GetConversation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	conversationID := c.Param("id")
	if conversationID == "" {
		utils.BadRequestResponse(c, "Conversation ID is required")
		return
	}

	conversation, err := cc.conversationRepo.GetByID(conversationID)
	if err != nil {
		cc.logger.Error().Err(err).Str("conversation_id", conversationID).Msg("Failed to get conversation")
		utils.NotFoundResponse(c, "Conversation")
		return
	}

	// Check if user is participant
	if !cc.isUserParticipant(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Conversation retrieved successfully", conversation)
}

func (cc *ChatController) CreateConversation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate participants
	participants := make([]models.ConversationParticipant, 0)
	for _, participantID := range req.ParticipantIDs {
		user, err := cc.userRepo.GetByID(participantID)
		if err != nil {
			utils.BadRequestResponse(c, "Invalid participant ID: "+participantID)
			return
		}

		participants = append(participants, models.ConversationParticipant{
			UserID:      user.ID,
			Role:        "member",
			JoinedAt:    time.Now(),
			UnreadCount: 0,
		})
	}

	// Add creator as owner
	creatorObjectID, _ := primitive.ObjectIDFromHex(userID)
	participants = append(participants, models.ConversationParticipant{
		UserID:      creatorObjectID,
		Role:        "owner",
		JoinedAt:    time.Now(),
		UnreadCount: 0,
	})

	conversation := &models.Conversation{
		Type:         req.Type,
		Title:        req.Title,
		Participants: participants,
		CreatedBy:    creatorObjectID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.RideID != "" {
		rideObjectID, err := primitive.ObjectIDFromHex(req.RideID)
		if err != nil {
			utils.BadRequestResponse(c, "Invalid ride ID")
			return
		}
		conversation.RideID = &rideObjectID
	}

	createdConversation, err := cc.conversationRepo.Create(conversation)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create conversation")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Conversation created successfully", createdConversation)
}

func (cc *ChatController) DeleteConversation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	conversationID := c.Param("id")
	conversation, err := cc.conversationRepo.GetByID(conversationID)
	if err != nil {
		utils.NotFoundResponse(c, "Conversation")
		return
	}

	// Check if user is owner or admin
	if !cc.canManageConversation(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	err = cc.conversationRepo.SoftDelete(conversationID)
	if err != nil {
		cc.logger.Error().Err(err).Str("conversation_id", conversationID).Msg("Failed to delete conversation")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.DeletedResponse(c, "Conversation deleted successfully")
}

// Messages

func (cc *ChatController) GetMessages(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	conversationID := c.Param("id")
	conversation, err := cc.conversationRepo.GetByID(conversationID)
	if err != nil {
		utils.NotFoundResponse(c, "Conversation")
		return
	}

	if !cc.isUserParticipant(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	messages, total, err := cc.messageRepo.GetConversationMessages(conversationID, params.Page, params.Limit)
	if err != nil {
		cc.logger.Error().Err(err).Str("conversation_id", conversationID).Msg("Failed to get messages")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Messages retrieved successfully", messages, meta)
}

func (cc *ChatController) SendMessage(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	user, _ := middleware.GetUserFromContext(c)
	conversationID := c.Param("id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	conversation, err := cc.conversationRepo.GetByID(conversationID)
	if err != nil {
		utils.NotFoundResponse(c, "Conversation")
		return
	}

	if !cc.isUserParticipant(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	senderObjectID, _ := primitive.ObjectIDFromHex(userID)
	conversationObjectID, _ := primitive.ObjectIDFromHex(conversationID)

	message := &models.Message{
		ConversationID: conversationObjectID,
		SenderID:       senderObjectID,
		SenderRole:     string(user.Role),
		Type:           req.Type,
		Content:        req.Content,
		QuickReplies:   req.QuickReplies,
		Status:         models.MessageStatusSent,
		Platform:       c.GetHeader("X-Platform"),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if req.ReplyToID != "" {
		replyObjectID, err := primitive.ObjectIDFromHex(req.ReplyToID)
		if err == nil {
			message.ReplyToID = &replyObjectID
		}
	}

	createdMessage, err := cc.messageRepo.Create(message)
	if err != nil {
		cc.logger.Error().Err(err).Str("conversation_id", conversationID).Msg("Failed to send message")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Update conversation last message
	err = cc.conversationRepo.UpdateLastMessage(conversationID, createdMessage)
	if err != nil {
		cc.logger.Warn().Err(err).Msg("Failed to update conversation last message")
	}

	// Send WebSocket message to conversation participants
	go cc.broadcastMessage(conversation, createdMessage, userID)

	// Send notifications to other participants
	go cc.sendMessageNotifications(conversation, createdMessage, user)

	utils.CreatedResponse(c, "Message sent successfully", createdMessage)
}

func (cc *ChatController) UpdateMessage(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	messageID := c.Param("id")
	message, err := cc.messageRepo.GetByID(messageID)
	if err != nil {
		utils.NotFoundResponse(c, "Message")
		return
	}

	// Check if user is the sender
	if message.SenderID.Hex() != userID {
		utils.ForbiddenResponse(c)
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Store original content for edit history
	message.OriginalContent = message.Content
	message.Content = req.Content
	message.IsEdited = true
	message.EditedAt = &time.Time{}
	*message.EditedAt = time.Now()
	message.UpdatedAt = time.Now()

	updatedMessage, err := cc.messageRepo.Update(message)
	if err != nil {
		cc.logger.Error().Err(err).Str("message_id", messageID).Msg("Failed to update message")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Message updated successfully", updatedMessage)
}

func (cc *ChatController) DeleteMessage(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	messageID := c.Param("id")
	message, err := cc.messageRepo.GetByID(messageID)
	if err != nil {
		utils.NotFoundResponse(c, "Message")
		return
	}

	// Check if user is the sender or admin
	if message.SenderID.Hex() != userID {
		utils.ForbiddenResponse(c)
		return
	}

	err = cc.messageRepo.SoftDelete(messageID, userID)
	if err != nil {
		cc.logger.Error().Err(err).Str("message_id", messageID).Msg("Failed to delete message")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.DeletedResponse(c, "Message deleted successfully")
}

// Message Status

func (cc *ChatController) MarkAsRead(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	messageID := c.Param("id")
	message, err := cc.messageRepo.GetByID(messageID)
	if err != nil {
		utils.NotFoundResponse(c, "Message")
		return
	}

	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	err = cc.messageRepo.MarkAsRead(messageID, userObjectID)
	if err != nil {
		cc.logger.Error().Err(err).Str("message_id", messageID).Msg("Failed to mark message as read")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Send read receipt via WebSocket
	go cc.sendReadReceipt(message, userID)

	utils.SuccessResponse(c, http.StatusOK, "Message marked as read", nil)
}

func (cc *ChatController) MarkAsDelivered(c *gin.Context) {
	_, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	messageID := c.Param("id")
	err := cc.messageRepo.MarkAsDelivered(messageID)
	if err != nil {
		cc.logger.Error().Err(err).Str("message_id", messageID).Msg("Failed to mark message as delivered")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Message marked as delivered", nil)
}

func (cc *ChatController) GetUnreadCount(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	conversationID := c.Param("id")
	count, err := cc.messageRepo.GetUnreadCount(conversationID, userID)
	if err != nil {
		cc.logger.Error().Err(err).Str("conversation_id", conversationID).Msg("Failed to get unread count")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Unread count retrieved successfully", gin.H{
		"conversation_id": conversationID,
		"unread_count":    count,
	})
}

// Media Messages

func (cc *ChatController) SendImageMessage(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	conversationID := c.Param("id")
	conversation, err := cc.conversationRepo.GetByID(conversationID)
	if err != nil {
		utils.NotFoundResponse(c, "Conversation")
		return
	}

	if !cc.isUserParticipant(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		utils.BadRequestResponse(c, "Image file is required")
		return
	}

	// Upload image
	uploadResult, err := cc.fileUploadService.UploadChatImage(file, "chat/images")
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to upload image")
		utils.BadRequestResponse(c, "Failed to upload image")
		return
	}

	// Create message
	message := cc.createMediaMessage(conversationID, userID, models.MessageTypeImage, uploadResult.URL)
	message.ThumbnailURL = uploadResult.

	createdMessage, err := cc.messageRepo.Create(message)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to create image message")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Broadcast message
	go cc.broadcastMessage(conversation, createdMessage, userID)

	utils.CreatedResponse(c, "Image message sent successfully", createdMessage)
}

func (cc *ChatController) SendLocationMessage(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	conversationID := c.Param("id")
	conversation, err := cc.conversationRepo.GetByID(conversationID)
	if err != nil {
		utils.NotFoundResponse(c, "Conversation")
		return
	}

	if !cc.isUserParticipant(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	var req SendLocationMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Create location message
	message := cc.createMediaMessage(conversationID, userID, models.MessageTypeLocation, "")

	location := &models.MessageLocation{
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Address:        req.Address,
		PlaceName:      req.PlaceName,
		IsLiveLocation: req.IsLiveLocation,
	}

	if req.IsLiveLocation && req.LiveDuration > 0 {
		liveUntil := time.Now().Add(time.Duration(req.LiveDuration) * time.Minute)
		location.LiveUntil = &liveUntil
	}

	message.Location = location

	createdMessage, err := cc.messageRepo.Create(message)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to create location message")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Broadcast message
	go cc.broadcastMessage(conversation, createdMessage, userID)

	utils.CreatedResponse(c, "Location message sent successfully", createdMessage)
}

func (cc *ChatController) SendAudioMessage(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	conversationID := c.Param("id")
	conversation, err := cc.conversationRepo.GetByID(conversationID)
	if err != nil {
		utils.NotFoundResponse(c, "Conversation")
		return
	}

	if !cc.isUserParticipant(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	file, err := c.FormFile("audio")
	if err != nil {
		utils.BadRequestResponse(c, "Audio file is required")
		return
	}

	// Get duration from form data
	durationStr := c.PostForm("duration")
	duration, _ := strconv.Atoi(durationStr)

	// Upload audio
	uploadResult, err := cc.fileUploadService.UploadChatAudio(file, "chat/audio")
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to upload audio")
		utils.BadRequestResponse(c, "Failed to upload audio")
		return
	}

	// Create message
	message := cc.createMediaMessage(conversationID, userID, models.MessageTypeAudio, uploadResult.URL)
	message.Duration = duration
	message.FileSize = uploadResult.Size

	createdMessage, err := cc.messageRepo.Create(message)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to create audio message")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Broadcast message
	go cc.broadcastMessage(conversation, createdMessage, userID)

	utils.CreatedResponse(c, "Audio message sent successfully", createdMessage)
}

// Quick Replies & Templates

func (cc *ChatController) GetQuickReplies(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	quickReplies, err := cc.chatService.GetUserQuickReplies(userID)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get quick replies")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Quick replies retrieved successfully", quickReplies)
}

func (cc *ChatController) CreateQuickReply(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req CreateQuickReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	quickReply := &models.QuickReplyTemplate{
		UserID:    userObjectID,
		Text:      req.Text,
		Category:  req.Category,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdQuickReply, err := cc.chatService.CreateQuickReply(quickReply)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create quick reply")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Quick reply created successfully", createdQuickReply)
}

func (cc *ChatController) GetMessageTemplates(c *gin.Context) {
	templates, err := cc.chatService.GetMessageTemplates()
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to get message templates")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Message templates retrieved successfully", templates)
}

// Chat Settings

func (cc *ChatController) GetChatSettings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	settings, err := cc.chatService.GetChatSettings(userID)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get chat settings")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Chat settings retrieved successfully", settings)
}

func (cc *ChatController) UpdateChatSettings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateChatSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	settings, err := cc.chatService.UpdateChatSettings(userID, &req)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update chat settings")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Chat settings updated successfully", settings)
}

func (cc *ChatController) BlockUser(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := cc.chatService.BlockUser(userID, req.UserID, req.Reason)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Str("blocked_user_id", req.UserID).Msg("Failed to block user")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User blocked successfully", nil)
}

func (cc *ChatController) UnblockUser(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := cc.chatService.UnblockUser(userID, req.UserID)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Str("unblocked_user_id", req.UserID).Msg("Failed to unblock user")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User unblocked successfully", nil)
}

func (cc *ChatController) GetBlockedUsers(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	blockedUsers, err := cc.chatService.GetBlockedUsers(userID)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get blocked users")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Blocked users retrieved successfully", blockedUsers)
}

// In-Ride Chat

func (cc *ChatController) GetRideChat(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("ride_id")
	conversation, err := cc.conversationRepo.GetByRideID(rideID)
	if err != nil {
		// Create new ride conversation if doesn't exist
		conversation, err = cc.createRideConversation(rideID, userID)
		if err != nil {
			cc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to create ride conversation")
			utils.InternalServerErrorResponse(c)
			return
		}
	}

	if !cc.isUserParticipant(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	// Get messages
	messages, _, err := cc.messageRepo.GetConversationMessages(conversation.ID.Hex(), 1, 50)
	if err != nil {
		cc.logger.Error().Err(err).Str("conversation_id", conversation.ID.Hex()).Msg("Failed to get ride chat messages")
		utils.InternalServerErrorResponse(c)
		return
	}

	response := gin.H{
		"conversation": conversation,
		"messages":     messages,
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride chat retrieved successfully", response)
}

func (cc *ChatController) SendRideMessage(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("ride_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	conversation, err := cc.conversationRepo.GetByRideID(rideID)
	if err != nil {
		utils.NotFoundResponse(c, "Ride conversation")
		return
	}

	if !cc.isUserParticipant(conversation, userID) {
		utils.ForbiddenResponse(c)
		return
	}

	// Reuse SendMessage logic by setting conversation ID in context
	c.Params = append(c.Params, gin.Param{Key: "id", Value: conversation.ID.Hex()})
	cc.SendMessage(c)
}

// Emergency Chat

func (cc *ChatController) StartEmergencyChat(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	user, _ := middleware.GetUserFromContext(c)

	// Create emergency conversation
	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	conversation := &models.Conversation{
		Type:      models.ConversationTypeEmergency,
		Title:     "Emergency Support",
		CreatedBy: userObjectID,
		Participants: []models.ConversationParticipant{
			{
				UserID:      userObjectID,
				Role:        "member",
				JoinedAt:    time.Now(),
				UnreadCount: 0,
			},
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdConversation, err := cc.conversationRepo.Create(conversation)
	if err != nil {
		cc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create emergency conversation")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Send initial system message
	systemMessage := &models.Message{
		ConversationID: createdConversation.ID,
		SenderID:       userObjectID,
		Type:           models.MessageTypeSystem,
		Content:        "Emergency chat started. Support will be with you shortly.",
		Status:         models.MessageStatusSent,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = cc.messageRepo.Create(systemMessage)
	if err != nil {
		cc.logger.Warn().Err(err).Msg("Failed to create emergency system message")
	}

	// Notify emergency response team
	go cc.notifyEmergencyTeam(user, createdConversation)

	utils.CreatedResponse(c, "Emergency chat started successfully", createdConversation)
}

func (cc *ChatController) GetEmergencyChatContacts(c *gin.Context) {
	_, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	user, _ := middleware.GetUserFromContext(c)
	contacts := user.EmergencyContacts

	utils.SuccessResponse(c, http.StatusOK, "Emergency contacts retrieved successfully", contacts)
}

// Helper methods

func (cc *ChatController) isUserParticipant(conversation *models.Conversation, userID string) bool {
	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	for _, participant := range conversation.Participants {
		if participant.UserID == userObjectID && participant.LeftAt == nil {
			return true
		}
	}
	return false
}

func (cc *ChatController) canManageConversation(conversation *models.Conversation, userID string) bool {
	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	for _, participant := range conversation.Participants {
		if participant.UserID == userObjectID && (participant.Role == "owner" || participant.Role == "admin") {
			return true
		}
	}
	return false
}

func (cc *ChatController) createMediaMessage(conversationID, userID string, messageType models.MessageType, mediaURL string) *models.Message {
	senderObjectID, _ := primitive.ObjectIDFromHex(userID)
	conversationObjectID, _ := primitive.ObjectIDFromHex(conversationID)

	return &models.Message{
		ConversationID: conversationObjectID,
		SenderID:       senderObjectID,
		Type:           messageType,
		MediaURL:       mediaURL,
		Status:         models.MessageStatusSent,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (cc *ChatController) createRideConversation(rideID, userID string) (*models.Conversation, error) {
	// TODO: Get ride details and participants from ride service
	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	rideObjectID, _ := primitive.ObjectIDFromHex(rideID)

	conversation := &models.Conversation{
		Type:      models.ConversationTypeRide,
		Title:     "Ride Chat",
		CreatedBy: userObjectID,
		RideID:    &rideObjectID,
		Participants: []models.ConversationParticipant{
			{
				UserID:      userObjectID,
				Role:        "member",
				JoinedAt:    time.Now(),
				UnreadCount: 0,
			},
		},
		IsActive:       true,
		AutoCloseAfter: 24 * time.Hour, // Auto-close after 24 hours
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return cc.conversationRepo.Create(conversation)
}

func (cc *ChatController) broadcastMessage(conversation *models.Conversation, message *models.Message, senderID string) {
	if cc.wsHub == nil {
		return
	}

	wsMessage := websocket.NewChatMessageEvent(
		conversation.ID.Hex(),
		message.SenderID.Hex(),
		"", // sender name - should be populated from user data
		message.Content,
		message.Type,
	)

	cc.wsHub.SendToChat(conversation.ID.Hex(), wsMessage, senderID)
}

func (cc *ChatController) sendReadReceipt(message *models.Message, userID string) {
	if cc.wsHub == nil {
		return
	}

	wsMessage := websocket.WSMessage{
		Type: "chat_message_read",
		Data: websocket.MessageReadData{
			ConversationID: message.ConversationID.Hex(),
			MessageID:      message.ID.Hex(),
			ReadBy:         userID,
			ReadAt:         time.Now(),
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
	}

	cc.wsHub.SendToChat(message.ConversationID.Hex(), wsMessage, userID)
}

func (cc *ChatController) sendMessageNotifications(conversation *models.Conversation, message *models.Message, sender *models.User) {
	// TODO: Implement notification sending to conversation participants
	// This should integrate with the notification service
}

func (cc *ChatController) notifyEmergencyTeam(user *models.User, conversation *models.Conversation) {
	// TODO: Implement emergency team notification
	// This should send high-priority notifications to emergency response team
	cc.logger.Warn().
		Str("user_id", user.ID.Hex()).
		Str("conversation_id", conversation.ID.Hex()).
		Msg("Emergency chat started - notifying response team")
}
