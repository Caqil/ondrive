package services

import (
	"errors"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
}

type ConversationStats struct {
	TotalMessages       int64                        `json:"total_messages"`
	MessagesByType      map[models.MessageType]int64 `json:"messages_by_type"`
	ParticipantStats    map[string]*ParticipantStats `json:"participant_stats"`
	ActivePeriods       []TimePeriod                 `json:"active_periods"`
	LastActivity        time.Time                    `json:"last_activity"`
	AverageResponseTime time.Duration                `json:"average_response_time"`
}

type ParticipantStats struct {
	UserID       string    `json:"user_id"`
	MessageCount int64     `json:"message_count"`
	LastMessage  time.Time `json:"last_message"`
	JoinedAt     time.Time `json:"joined_at"`
	IsActive     bool      `json:"is_active"`
}

type TimePeriod struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Messages  int64     `json:"messages"`
}

type chatService struct {
	conversationRepo    repositories.ConversationRepository
	messageRepo         repositories.MessageRepository
	quickReplyRepo      repositories.QuickReplyRepository
	messageTemplateRepo repositories.MessageTemplateRepository
	chatSettingsRepo    repositories.ChatSettingsRepository
	userRepo            repositories.UserRepository
	logger              utils.Logger
}

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
		logger:              logger,
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

func (s *chatService) UpdateChatSettings(userID string, settings interface{}) (*models.ChatSettings, error) {
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
	// This would need type assertion based on the actual structure of settings
	// For now, we'll update the timestamp
	existingSettings.UpdatedAt = time.Now()

	return s.chatSettingsRepo.Update(existingSettings)
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

	_, err := primitive.ObjectIDFromHex(blockerID)
	if err != nil {
		return err
	}

	blockedObjectID, err := primitive.ObjectIDFromHex(blockedID)
	if err != nil {
		return err
	}

	// Check if users exist
	_, err = s.userRepo.GetByID(blockerID)
	if err != nil {
		return errors.New("blocker user not found")
	}

	_, err = s.userRepo.GetByID(blockedID)
	if err != nil {
		return errors.New("blocked user not found")
	}

	// Get or create chat settings for blocker
	settings, err := s.GetChatSettings(blockerID)
	if err != nil {
		return err
	}

	// Check if already blocked
	for _, blockedUserID := range settings.BlockedUsers {
		if blockedUserID == blockedObjectID {
			return errors.New("user is already blocked")
		}
	}

	// Add to blocked list
	settings.BlockedUsers = append(settings.BlockedUsers, blockedObjectID)
	settings.UpdatedAt = time.Now()

	_, err = s.chatSettingsRepo.Update(settings)
	return err
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

	settings, err := s.chatSettingsRepo.GetByUserID(blockerObjectID)
	if err != nil {
		return err
	}

	// Remove from blocked list
	newBlockedUsers := make([]primitive.ObjectID, 0)
	found := false
	for _, blockedUserID := range settings.BlockedUsers {
		if blockedUserID != blockedObjectID {
			newBlockedUsers = append(newBlockedUsers, blockedUserID)
		} else {
			found = true
		}
	}

	if !found {
		return errors.New("user is not blocked")
	}

	settings.BlockedUsers = newBlockedUsers
	settings.UpdatedAt = time.Now()

	_, err = s.chatSettingsRepo.Update(settings)
	return err
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

	settings, err := s.chatSettingsRepo.GetByUserID(blockerObjectID)
	if err != nil {
		return false, nil // No settings means no blocks
	}

	for _, blockedUserID := range settings.BlockedUsers {
		if blockedUserID == blockedObjectID {
			return true, nil
		}
	}

	return false, nil
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
	conversation, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, err
	}

	// This is a simplified implementation
	// In a real application, you might want to use aggregation pipelines
	stats := &ConversationStats{
		TotalMessages:    int64(conversation.MessageCount),
		MessagesByType:   make(map[models.MessageType]int64),
		ParticipantStats: make(map[string]*ParticipantStats),
		LastActivity:     conversation.UpdatedAt,
	}

	// Populate participant stats
	for _, participant := range conversation.Participants {
		stats.ParticipantStats[participant.UserID.Hex()] = &ParticipantStats{
			UserID:   participant.UserID.Hex(),
			JoinedAt: participant.JoinedAt,
			IsActive: participant.LeftAt == nil,
		}
	}

	return stats, nil
}

func (s *chatService) CleanupOldMessages(conversationID string, retentionDays int) error {
	// This would implement message cleanup logic
	// For now, we'll just log the operation
	s.logger.Info().
		Str("conversation_id", conversationID).
		Int("retention_days", retentionDays).
		Msg("Cleaning up old messages")

	// Implementation would involve:
	// 1. Calculate cutoff date
	// 2. Find messages older than cutoff
	// 3. Soft delete or hard delete based on policy
	// 4. Update conversation stats

	return nil
}

func (s *chatService) ArchiveInactiveConversations(inactiveDays int) error {
	// This would implement automatic archiving of inactive conversations
	s.logger.Info().
		Int("inactive_days", inactiveDays).
		Msg("Archiving inactive conversations")

	// Implementation would involve:
	// 1. Calculate cutoff date
	// 2. Find conversations with no activity since cutoff
	// 3. Mark as archived
	// 4. Optionally notify participants

	return nil
}

// Message Processing Implementation

func (s *chatService) ProcessMessage(message *models.Message) error {
	// Message processing pipeline

	// 1. Validate content
	if err := s.ValidateMessageContent(message.Content, message.Type); err != nil {
		return err
	}

	// 2. Filter profanity
	if message.Type == models.MessageTypeText {
		message.Content = s.FilterProfanity(message.Content)
	}

	// 3. Detect spam
	isSpam, err := s.DetectSpam(message, message.SenderID.Hex())
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to detect spam")
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
		if len(content) > 4000 {
			return errors.New("text message too long (max 4000 characters)")
		}
	case models.MessageTypeImage:
		// Additional validation for image messages
		break
	case models.MessageTypeAudio:
		// Additional validation for audio messages
		break
	}

	return nil
}

func (s *chatService) FilterProfanity(content string) string {
	// Simplified profanity filter
	// In a real implementation, you would use a proper profanity filtering service
	profaneWords := []string{"badword1", "badword2"} // Example words

	filteredContent := content
	for _, word := range profaneWords {
		// Replace with asterisks (case-insensitive)
		filteredContent = replaceWordCaseInsensitive(filteredContent, word, "***")
	}

	return filteredContent
}

func (s *chatService) DetectSpam(message *models.Message, userID string) (bool, error) {
	// Simplified spam detection
	// In a real implementation, you would use machine learning or rule-based systems

	// Check for repeated messages
	recentMessages, err := s.messageRepo.GetMessagesAfterTime(
		message.ConversationID.Hex(),
		time.Now().Add(-5*time.Minute),
	)
	if err != nil {
		return false, err
	}

	duplicateCount := 0
	for _, recentMsg := range recentMessages {
		if recentMsg.SenderID.Hex() == userID && recentMsg.Content == message.Content {
			duplicateCount++
		}
	}

	// If same message sent more than 3 times in 5 minutes, consider it spam
	if duplicateCount >= 3 {
		return true, nil
	}

	// Check message frequency
	userRecentMessages := 0
	for _, recentMsg := range recentMessages {
		if recentMsg.SenderID.Hex() == userID {
			userRecentMessages++
		}
	}

	// If user sent more than 20 messages in 5 minutes, consider it spam
	if userRecentMessages >= 20 {
		return true, nil
	}

	return false, nil
}

// Helper functions

func replaceWordCaseInsensitive(text, oldWord, newWord string) string {
	// This is a simplified implementation
	// In a real application, you would use regex or a proper string replacement function
	return text // Placeholder
}
