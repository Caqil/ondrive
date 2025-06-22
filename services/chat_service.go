package services

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatService interface for chat-related business logic
type ChatService interface {
	// Quick Replies
	GetUserQuickReplies(userID string) ([]*models.QuickReplyTemplate, error)
	CreateQuickReply(quickReply *models.QuickReplyTemplate) (*models.QuickReplyTemplate, error)
	UpdateQuickReply(quickReply *models.QuickReplyTemplate) (*models.QuickReplyTemplate, error)
	DeleteQuickReply(id string) error
	IncrementQuickReplyUsage(id string) error

	// Message Templates
	GetMessageTemplates() ([]*models.MessageTemplate, error)
	GetMessageTemplatesByCategory(category string) ([]*models.MessageTemplate, error)
	CreateMessageTemplate(template *models.MessageTemplate) (*models.MessageTemplate, error)
	UpdateMessageTemplate(template *models.MessageTemplate) (*models.MessageTemplate, error)

	// Chat Settings
	GetChatSettings(userID string) (*models.ChatSettings, error)
	UpdateChatSettings(userID string, settings interface{}) (*models.ChatSettings, error)
	CreateDefaultChatSettings(userID string) (*models.ChatSettings, error)

	// User Blocking
	BlockUser(blockerID, blockedID, reason string) error
	UnblockUser(blockerID, blockedID string) error
	IsUserBlocked(blockerID, blockedID string) (bool, error)
	GetBlockedUsers(userID string) ([]primitive.ObjectID, error)

	// Conversation Utilities
	CanUserAccessConversation(userID, conversationID string) (bool, error)
	GetConversationStats(conversationID string) (*ConversationStats, error)
	CleanupOldMessages(conversationID string, retentionDays int) error
	ArchiveInactiveConversations(inactiveDays int) error

	// Message Processing
	ProcessMessage(message *models.Message) error
	ValidateMessageContent(content string, messageType models.MessageType) error
	FilterProfanity(content string) string
	DetectSpam(message *models.Message, userID string) (bool, error)

	// Additional helper methods
	GetConversationParticipants(conversationID string) ([]models.ConversationParticipant, error)
	AddConversationParticipant(conversationID string, participant models.ConversationParticipant) error
	RemoveConversationParticipant(conversationID, userID string) error
	UpdateUserOnlineStatus(userID string, isOnline bool) error
	GetUnreadMessagesCount(userID string) (int64, error)
}

// ConversationStats represents conversation statistics
type ConversationStats struct {
	TotalMessages       int64                        `json:"total_messages"`
	MessagesByType      map[models.MessageType]int64 `json:"messages_by_type"`
	ParticipantStats    map[string]*ParticipantStats `json:"participant_stats"`
	ActivePeriods       []TimePeriod                 `json:"active_periods"`
	LastActivity        time.Time                    `json:"last_activity"`
	AverageResponseTime time.Duration                `json:"average_response_time"`
}

// ParticipantStats represents individual participant statistics
type ParticipantStats struct {
	UserID       string    `json:"user_id"`
	MessageCount int64     `json:"message_count"`
	LastMessage  time.Time `json:"last_message"`
	JoinedAt     time.Time `json:"joined_at"`
	IsActive     bool      `json:"is_active"`
}

// TimePeriod represents a time period with message count
type TimePeriod struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Messages  int64     `json:"messages"`
}

// chatService implementation
type chatService struct {
	conversationRepo    repositories.ConversationRepository
	messageRepo         repositories.MessageRepository
	quickReplyRepo      repositories.QuickReplyRepository
	messageTemplateRepo repositories.MessageTemplateRepository
	chatSettingsRepo    repositories.ChatSettingsRepository
	userRepo            repositories.UserRepository
	logger              utils.Logger
}

// NewChatService creates a new chat service instance
func NewChatService(
	conversationRepo repositories.ConversationRepository,
	messageRepo repositories.MessageRepository,
	quickReplyRepo repositories.QuickReplyRepository,
	messageTemplateRepo repositories.MessageTemplateRepository,
	chatSettingsRepo repositories.ChatSettingsRepository,
	userRepo repositories.UserRepository,
	logger utils.Logger,
) ChatService {
	return &chatService{
		conversationRepo:    conversationRepo,
		messageRepo:         messageRepo,
		quickReplyRepo:      quickReplyRepo,
		messageTemplateRepo: messageTemplateRepo,
		chatSettingsRepo:    chatSettingsRepo,
		userRepo:            userRepo,
		logger:              utils.ServiceLogger("chat"),
	}
}

// Quick Replies Implementation

func (s *chatService) GetUserQuickReplies(userID string) ([]*models.QuickReplyTemplate, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.quickReplyRepo.GetByUserID(userObjectID)
}

func (s *chatService) CreateQuickReply(quickReply *models.QuickReplyTemplate) (*models.QuickReplyTemplate, error) {
	// Validate quick reply
	if quickReply.Text == "" {
		return nil, errors.New("quick reply text is required")
	}

	if len(quickReply.Text) > 200 {
		return nil, errors.New("quick reply text too long (max 200 characters)")
	}

	quickReply.CreatedAt = time.Now()
	quickReply.UpdatedAt = time.Now()
	quickReply.IsActive = true

	return s.quickReplyRepo.Create(quickReply)
}

func (s *chatService) UpdateQuickReply(quickReply *models.QuickReplyTemplate) (*models.QuickReplyTemplate, error) {
	quickReply.UpdatedAt = time.Now()
	return s.quickReplyRepo.Update(quickReply)
}

func (s *chatService) DeleteQuickReply(id string) error {
	return s.quickReplyRepo.SoftDelete(id)
}

func (s *chatService) IncrementQuickReplyUsage(id string) error {
	return s.quickReplyRepo.IncrementUsage(id)
}

// Message Templates Implementation

func (s *chatService) GetMessageTemplates() ([]*models.MessageTemplate, error) {
	return s.messageTemplateRepo.GetActive()
}

func (s *chatService) GetMessageTemplatesByCategory(category string) ([]*models.MessageTemplate, error) {
	return s.messageTemplateRepo.GetByCategory(category)
}

func (s *chatService) CreateMessageTemplate(template *models.MessageTemplate) (*models.MessageTemplate, error) {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.IsActive = true

	return s.messageTemplateRepo.Create(template)
}

func (s *chatService) UpdateMessageTemplate(template *models.MessageTemplate) (*models.MessageTemplate, error) {
	template.UpdatedAt = time.Now()
	return s.messageTemplateRepo.Update(template)
}

// Chat Settings Implementation

func (s *chatService) GetChatSettings(userID string) (*models.ChatSettings, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	settings, err := s.chatSettingsRepo.GetByUserID(userObjectID)
	if err != nil {
		// Create default settings if not found
		return s.CreateDefaultChatSettings(userID)
	}

	return settings, nil
}

func (s *chatService) UpdateChatSettings(userID string, settingsUpdate interface{}) (*models.ChatSettings, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Get existing settings
	existingSettings, err := s.chatSettingsRepo.GetByUserID(userObjectID)
	if err != nil {
		// Create default if not found
		existingSettings, err = s.CreateDefaultChatSettings(userID)
		if err != nil {
			return nil, err
		}
	}

	// Update fields based on the provided settings
	// Type assertion to map for flexible updates
	if updateMap, ok := settingsUpdate.(map[string]interface{}); ok {
		s.updateSettingsFromMap(existingSettings, updateMap)
	}

	existingSettings.UpdatedAt = time.Now()

	return s.chatSettingsRepo.Update(existingSettings)
}

func (s *chatService) updateSettingsFromMap(settings *models.ChatSettings, updateMap map[string]interface{}) {
	if val, exists := updateMap["is_online"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.IsOnline = boolVal
		}
	}
	if val, exists := updateMap["show_last_seen"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.ShowLastSeen = boolVal
		}
	}
	if val, exists := updateMap["show_online_status"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.ShowOnlineStatus = boolVal
		}
	}
	if val, exists := updateMap["read_receipts"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.ReadReceipts = boolVal
		}
	}
	if val, exists := updateMap["typing_indicator"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.TypingIndicator = boolVal
		}
	}
	if val, exists := updateMap["auto_download_media"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.AutoDownloadMedia = boolVal
		}
	}
	if val, exists := updateMap["message_notifications"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.MessageNotifications = boolVal
		}
	}
	if val, exists := updateMap["sound_enabled"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.SoundEnabled = boolVal
		}
	}
	if val, exists := updateMap["vibration_enabled"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.VibrationEnabled = boolVal
		}
	}
	if val, exists := updateMap["allow_messages_from"]; exists {
		if strVal, ok := val.(string); ok {
			settings.AllowMessagesFrom = strVal
		}
	}
	if val, exists := updateMap["block_unknown_senders"]; exists {
		if boolVal, ok := val.(bool); ok {
			settings.BlockUnknownSenders = boolVal
		}
	}
}

func (s *chatService) CreateDefaultChatSettings(userID string) (*models.ChatSettings, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	settings := &models.ChatSettings{
		UserID:               userObjectID,
		IsOnline:             true,
		ShowLastSeen:         true,
		ShowOnlineStatus:     true,
		ReadReceipts:         true,
		TypingIndicator:      true,
		AutoDownloadMedia:    true,
		MessageNotifications: true,
		SoundEnabled:         true,
		VibrationEnabled:     true,
		AllowMessagesFrom:    "everyone",
		BlockUnknownSenders:  false,
		BlockedUsers:         []primitive.ObjectID{},
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	return s.chatSettingsRepo.Create(settings)
}

// User Blocking Implementation

func (s *chatService) BlockUser(blockerID, blockedID, reason string) error {
	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}

	blockerObjectID, err := primitive.ObjectIDFromHex(blockerID)
	if err != nil {
		return err
	}

	blockedObjectID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		return err
	}

	// Check if user exists
	_, err = s.userRepo.GetByID(blockedID)
	if err != nil {
		return errors.New("user to block not found")
	}

	// Add to blocked users list
	return s.chatSettingsRepo.AddBlockedUser(blockerObjectID, blockedObjectID)
}

func (s *chatService) UnblockUser(blockerID, blockedID string) error {
	blockerObjectID, err := primitive.ObjectIDFromHex(blockerID)
	if err != nil {
		return err
	}

	blockedObjectID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		return err
	}

	return s.chatSettingsRepo.RemoveBlockedUser(blockerObjectID, blockedObjectID)
}

func (s *chatService) IsUserBlocked(blockerID, blockedID string) (bool, error) {
	blockerObjectID, err := primitive.ObjectIDFromHex(blockerID)
	if err != nil {
		return false, err
	}

	blockedObjectID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		return false, err
	}

	return s.chatSettingsRepo.IsUserBlocked(blockerObjectID, blockedObjectID)
}

func (s *chatService) GetBlockedUsers(userID string) ([]primitive.ObjectID, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	settings, err := s.chatSettingsRepo.GetByUserID(userObjectID)
	if err != nil {
		return []primitive.ObjectID{}, nil
	}

	return settings.BlockedUsers, nil
}

// Conversation Utilities Implementation

func (s *chatService) CanUserAccessConversation(userID, conversationID string) (bool, error) {
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return false, err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	for _, participant := range conversation.Participants {
		if participant.UserID == userObjectID && participant.LeftAt == nil {
			return true, nil
		}
	}

	return false, nil
}

func (s *chatService) GetConversationStats(conversationID string) (*ConversationStats, error) {
	// This is a placeholder implementation
	// You would need to aggregate data from the messages collection
	stats := &ConversationStats{
		TotalMessages:       0,
		MessagesByType:      make(map[models.MessageType]int64),
		ParticipantStats:    make(map[string]*ParticipantStats),
		ActivePeriods:       []TimePeriod{},
		LastActivity:        time.Now(),
		AverageResponseTime: 0,
	}

	// TODO: Implement actual aggregation logic
	s.logger.Warn().Str("conversation_id", conversationID).Msg("GetConversationStats not fully implemented")

	return stats, nil
}

func (s *chatService) CleanupOldMessages(conversationID string, retentionDays int) error {
	// This would soft delete old messages
	s.logger.Info().
		Str("conversation_id", conversationID).
		Int("retention_days", retentionDays).
		Msg("CleanupOldMessages called - implementation needed")

	// TODO: Implement cleanup logic
	return nil
}

func (s *chatService) ArchiveInactiveConversations(inactiveDays int) error {
	// This would archive conversations that haven't had activity
	s.logger.Info().
		Int("inactive_days", inactiveDays).
		Msg("ArchiveInactiveConversations called - implementation needed")

	// TODO: Implement archiving logic
	return nil
}

// Message Processing Implementation

func (s *chatService) ProcessMessage(message *models.Message) error {
	// Validate message content
	if err := s.ValidateMessageContent(message.Content, message.Type); err != nil {
		return err
	}

	// Filter profanity
	message.Content = s.FilterProfanity(message.Content)

	// Detect spam
	isSpam, err := s.DetectSpam(message, message.SenderID.Hex())
	if err != nil {
		s.logger.Warn().Err(err).Msg("Error detecting spam")
	}

	if isSpam {
		return errors.New("message detected as spam")
	}

	return nil
}

func (s *chatService) ValidateMessageContent(content string, messageType models.MessageType) error {
	switch messageType {
	case models.MessageTypeText:
		if len(content) == 0 {
			return errors.New("text message content cannot be empty")
		}
		if len(content) > 4096 {
			return errors.New("text message too long (max 4096 characters)")
		}
	case models.MessageTypeSystem:
		if len(content) > 1024 {
			return errors.New("system message too long (max 1024 characters)")
		}
	}

	return nil
}

func (s *chatService) FilterProfanity(content string) string {
	// Basic profanity filter - replace with more sophisticated implementation
	profanityWords := []string{
		"badword1", "badword2", "badword3", // Add actual profanity words
	}

	filteredContent := content
	for _, word := range profanityWords {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
		filteredContent = re.ReplaceAllString(filteredContent, strings.Repeat("*", len(word)))
	}

	return filteredContent
}

func (s *chatService) DetectSpam(message *models.Message, userID string) (bool, error) {
	// Basic spam detection - replace with more sophisticated implementation

	// Check for excessive caps
	if len(message.Content) > 10 {
		upperCount := 0
		for _, char := range message.Content {
			if char >= 'A' && char <= 'Z' {
				upperCount++
			}
		}
		if float64(upperCount)/float64(len(message.Content)) > 0.8 {
			return true, nil
		}
	}

	// Check for repeated characters
	re := regexp.MustCompile(`(.)\1{4,}`)
	if re.MatchString(message.Content) {
		return true, nil
	}

	// Check for excessive punctuation
	re = regexp.MustCompile(`[!?]{4,}`)
	if re.MatchString(message.Content) {
		return true, nil
	}

	return false, nil
}

// Additional Helper Methods

func (s *chatService) GetConversationParticipants(conversationID string) ([]models.ConversationParticipant, error) {
	return s.conversationRepo.GetConversationParticipants(conversationID)
}

func (s *chatService) AddConversationParticipant(conversationID string, participant models.ConversationParticipant) error {
	return s.conversationRepo.AddParticipant(conversationID, participant)
}

func (s *chatService) RemoveConversationParticipant(conversationID, userID string) error {
	return s.conversationRepo.RemoveParticipant(conversationID, userID)
}

func (s *chatService) UpdateUserOnlineStatus(userID string, isOnline bool) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	return s.chatSettingsRepo.UpdateOnlineStatus(userObjectID, isOnline)
}

func (s *chatService) GetUnreadMessagesCount(userID string) (int64, error) {
	// Get all active conversations for user
	conversations, err := s.conversationRepo.GetActiveConversationsByUser(userID)
	if err != nil {
		return 0, err
	}

	var totalUnread int64
	for _, conversation := range conversations {
		count, err := s.messageRepo.GetUnreadCount(conversation.ID.Hex(), userID)
		if err != nil {
			s.logger.Warn().Err(err).Str("conversation_id", conversation.ID.Hex()).Msg("Failed to get unread count")
			continue
		}
		totalUnread += count
	}

	return totalUnread, nil
}

// Additional utility methods for chat operations

func (s *chatService) GetMessageTemplateByID(templateID string) (*models.MessageTemplate, error) {
	return s.messageTemplateRepo.GetByID(templateID)
}

func (s *chatService) GetQuickReplyByID(quickReplyID string) (*models.QuickReplyTemplate, error) {
	return s.quickReplyRepo.GetByID(quickReplyID)
}

func (s *chatService) GetQuickRepliesByCategory(userID, category string) ([]*models.QuickReplyTemplate, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.quickReplyRepo.GetByCategory(userObjectID, category)
}

func (s *chatService) IncrementTemplateUsage(templateID string) error {
	return s.messageTemplateRepo.IncrementUsage(templateID)
}

func (s *chatService) GetSystemMessageTemplates() ([]*models.MessageTemplate, error) {
	return s.messageTemplateRepo.GetSystemTemplates()
}

// Emergency chat helpers

func (s *chatService) CreateEmergencyConversation(userID string) (*models.Conversation, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

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

	return s.conversationRepo.Create(conversation)
}

func (s *chatService) IsEmergencyConversation(conversationID string) (bool, error) {
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return false, err
	}

	return conversation.Type == models.ConversationTypeEmergency, nil
}

// Bulk operations

func (s *chatService) BulkDeleteQuickReplies(userID string, quickReplyIDs []string) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	for _, id := range quickReplyIDs {
		// Verify ownership before deletion
		quickReply, err := s.quickReplyRepo.GetByID(id)
		if err != nil {
			continue
		}

		if quickReply.UserID != userObjectID {
			continue
		}

		if err := s.quickReplyRepo.SoftDelete(id); err != nil {
			s.logger.Warn().Err(err).Str("quick_reply_id", id).Msg("Failed to delete quick reply")
		}
	}

	return nil
}

func (s *chatService) BulkArchiveConversations(userID string, conversationIDs []string) error {
	for _, conversationID := range conversationIDs {
		// Verify user can access conversation
		canAccess, err := s.CanUserAccessConversation(userID, conversationID)
		if err != nil || !canAccess {
			continue
		}

		if err := s.conversationRepo.ArchiveConversation(conversationID); err != nil {
			s.logger.Warn().Err(err).Str("conversation_id", conversationID).Msg("Failed to archive conversation")
		}
	}

	return nil
}
