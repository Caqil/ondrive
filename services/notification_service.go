package services

import (
	"fmt"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationService interface defines notification operations
type NotificationService interface {
	// Core Notification Methods
	SendNotification(notification *models.Notification) error
	SendNotificationToUser(userID string, notification *NotificationRequest) error
	SendNotificationToMultipleUsers(userIDs []string, notification *NotificationRequest) error
	SendBulkNotification(request *BulkNotificationRequest) error

	// Notification Management
	GetUserNotifications(userID string, page, limit int) ([]*models.Notification, int64, error)
	GetNotificationByID(notificationID string) (*models.Notification, error)
	MarkAsRead(notificationID, userID string) error
	MarkAllAsRead(userID string) error
	DeleteNotification(notificationID, userID string) error
	GetUnreadCount(userID string) (int64, error)

	// Template Management
	CreateTemplate(template *models.NotificationTemplate) (*models.NotificationTemplate, error)
	GetTemplate(templateID string) (*models.NotificationTemplate, error)
	GetTemplatesByCategory(category string) ([]*models.NotificationTemplate, error)
	UpdateTemplate(template *models.NotificationTemplate) (*models.NotificationTemplate, error)
	DeleteTemplate(templateID string) error
	SendTemplatedNotification(userID, templateID string, variables map[string]interface{}) error

	// Preference Management
	GetUserPreferences(userID string) (*models.NotificationPreferences, error)
	UpdateUserPreferences(userID string, preferences *models.NotificationPreferences) error
	CreateDefaultPreferences(userID string) (*models.NotificationPreferences, error)

	// Scheduled Notifications
	ScheduleNotification(notification *models.Notification, scheduledAt time.Time) error
	CancelScheduledNotification(notificationID string) error
	GetScheduledNotifications(userID string) ([]*models.Notification, error)

	// Push Notifications
	SendPushNotification(userID string, title, body string, data map[string]interface{}) error
	RegisterDeviceToken(userID, deviceToken, platform string) error
	UnregisterDeviceToken(userID, deviceToken string) error

	// SMS Notifications
	SendSMSNotification(phone, message string) error
	SendSMSToUser(userID, message string) error

	// Email Notifications
	SendEmailNotification(email, subject, body string) error
	SendEmailToUser(userID, subject, body string) error
	SendTemplatedEmail(userID, templateName string, variables map[string]interface{}) error

	// In-App Notifications
	SendInAppNotification(userID string, notification *InAppNotificationRequest) error
	GetInAppNotifications(userID string, page, limit int) ([]*models.Notification, error)

	// Emergency Notifications
	SendEmergencyAlert(alert *EmergencyAlertRequest) error
	SendSOSNotification(userID string, location *models.Location) error
	GetEmergencyContacts(userID string) ([]string, error)

	// Location-Based Notifications
	SendGeofenceAlert(userID string, alert *GeofenceAlertRequest) error
	GetNearbyPromotions(userID string, location *models.Location) ([]*models.Notification, error)

	// Analytics
	GetNotificationAnalytics(userID string, startDate, endDate time.Time) (*NotificationAnalytics, error)
	GetEngagementStats(templateID string) (*EngagementStats, error)
	TrackNotificationClick(notificationID string) error

	// Utility Methods
	ValidateNotificationRequest(request *NotificationRequest) error
	ProcessNotificationQueue() error
	CleanupOldNotifications(retentionDays int) error
}

// Request/Response structs
type NotificationRequest struct {
	Title       string                       `json:"title" validate:"required"`
	Body        string                       `json:"body" validate:"required"`
	Type        models.NotificationType      `json:"type"`
	Priority    models.NotificationPriority  `json:"priority"`
	Category    string                       `json:"category"`
	Channels    []models.NotificationChannel `json:"channels"`
	ActionURL   string                       `json:"action_url"`
	ActionText  string                       `json:"action_text"`
	ImageURL    string                       `json:"image_url"`
	Data        map[string]interface{}       `json:"data"`
	ScheduledAt *time.Time                   `json:"scheduled_at,omitempty"`
}

type BulkNotificationRequest struct {
	UserIDs      []string             `json:"user_ids"`
	Notification *NotificationRequest `json:"notification"`
	BatchSize    int                  `json:"batch_size"`
}

type InAppNotificationRequest struct {
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	ActionURL string                 `json:"action_url"`
}

type EmergencyAlertRequest struct {
	UserIDs    []string         `json:"user_ids"`
	Title      string           `json:"title"`
	Message    string           `json:"message"`
	AlertLevel string           `json:"alert_level"` // low, medium, high, critical
	Location   *models.Location `json:"location,omitempty"`
}

type GeofenceAlertRequest struct {
	Title     string  `json:"title"`
	Message   string  `json:"message"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius"` // in meters
}

type NotificationAnalytics struct {
	TotalSent      int64                   `json:"total_sent"`
	TotalDelivered int64                   `json:"total_delivered"`
	TotalRead      int64                   `json:"total_read"`
	TotalClicked   int64                   `json:"total_clicked"`
	DeliveryRate   float64                 `json:"delivery_rate"`
	ReadRate       float64                 `json:"read_rate"`
	ClickRate      float64                 `json:"click_rate"`
	ChannelStats   map[string]ChannelStats `json:"channel_stats"`
}

type ChannelStats struct {
	Sent      int64   `json:"sent"`
	Delivered int64   `json:"delivered"`
	Failed    int64   `json:"failed"`
	Rate      float64 `json:"rate"`
}

type EngagementStats struct {
	TemplateID   string  `json:"template_id"`
	TotalSent    int64   `json:"total_sent"`
	TotalRead    int64   `json:"total_read"`
	TotalClicked int64   `json:"total_clicked"`
	ReadRate     float64 `json:"read_rate"`
	ClickRate    float64 `json:"click_rate"`
}

// notificationService implements NotificationService interface
type notificationService struct {
	notificationRepo repositories.NotificationRepository
	templateRepo     repositories.NotificationTemplateRepository
	preferenceRepo   repositories.NotificationPreferenceRepository
	userRepo         repositories.UserRepository
	logger           utils.Logger
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	templateRepo repositories.NotificationTemplateRepository,
	preferenceRepo repositories.NotificationPreferenceRepository,
	userRepo repositories.UserRepository,
	logger utils.Logger,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		preferenceRepo:   preferenceRepo,
		userRepo:         userRepo,
		logger:           logger,
	}
}

// SendNotification sends a notification
func (s *notificationService) SendNotification(notification *models.Notification) error {
	s.logger.Info().
		Str("recipient_id", notification.RecipientID.Hex()).
		Str("type", string(notification.Type)).
		Str("title", notification.Title).
		Msg("Sending notification")

	// Validate notification
	if err := s.validateNotification(notification); err != nil {
		return err
	}

	// Check user preferences
	preferences, err := s.GetUserPreferences(notification.RecipientID.Hex())
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to get user preferences, using defaults")
	}

	// Filter channels based on preferences
	allowedChannels := s.filterChannelsByPreferences(notification.Channels, preferences)
	notification.Channels = allowedChannels

	// Set timestamps
	now := time.Now()
	notification.CreatedAt = now
	notification.UpdatedAt = now
	notification.Status = models.NotificationStatusPending

	// Save notification
	createdNotification, err := s.notificationRepo.Create(notification)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create notification")
		return err
	}

	// Process delivery asynchronously
	go s.processDelivery(createdNotification)

	s.logger.Info().
		Str("notification_id", createdNotification.ID.Hex()).
		Msg("Notification created successfully")

	return nil
}

// SendNotificationToUser sends a notification to a specific user
func (s *notificationService) SendNotificationToUser(userID string, request *NotificationRequest) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	notification := &models.Notification{
		ID:          primitive.NewObjectID(),
		RecipientID: userObjectID,
		Title:       request.Title,
		Body:        request.Body,
		Type:        request.Type,
		Priority:    request.Priority,
		Category:    request.Category,
		Channels:    request.Channels,
		ActionURL:   request.ActionURL,
		ActionText:  request.ActionText,
		ImageURL:    request.ImageURL,
		ActionData:  request.Data,
	}

	if request.ScheduledAt != nil {
		notification.ScheduledAt = request.ScheduledAt
	}

	return s.SendNotification(notification)
}

// SendNotificationToMultipleUsers sends notification to multiple users
func (s *notificationService) SendNotificationToMultipleUsers(userIDs []string, request *NotificationRequest) error {
	for _, userID := range userIDs {
		if err := s.SendNotificationToUser(userID, request); err != nil {
			s.logger.Error().
				Err(err).
				Str("user_id", userID).
				Msg("Failed to send notification to user")
		}
	}
	return nil
}

// SendBulkNotification sends bulk notifications
func (s *notificationService) SendBulkNotification(request *BulkNotificationRequest) error {
	s.logger.Info().
		Int("user_count", len(request.UserIDs)).
		Msg("Sending bulk notification")

	batchSize := request.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	// Process in batches
	for i := 0; i < len(request.UserIDs); i += batchSize {
		end := i + batchSize
		if end > len(request.UserIDs) {
			end = len(request.UserIDs)
		}

		batch := request.UserIDs[i:end]
		go s.SendNotificationToMultipleUsers(batch, request.Notification)
	}

	return nil
}

// GetUserNotifications retrieves user notifications with pagination
func (s *notificationService) GetUserNotifications(userID string, page, limit int) ([]*models.Notification, int64, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	return s.notificationRepo.GetByRecipientID(userObjectID, page, limit)
}

// GetNotificationByID retrieves a notification by ID
func (s *notificationService) GetNotificationByID(notificationID string) (*models.Notification, error) {
	return s.notificationRepo.GetByID(notificationID)
}

// MarkAsRead marks a notification as read
func (s *notificationService) MarkAsRead(notificationID, userID string) error {
	notification, err := s.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}

	if notification.RecipientID.Hex() != userID {
		return fmt.Errorf("notification does not belong to user")
	}

	now := time.Now()
	notification.IsRead = true
	notification.ReadAt = &now
	notification.UpdatedAt = now

	return s.notificationRepo.Update(notification)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *notificationService) MarkAllAsRead(userID string) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	return s.notificationRepo.MarkAllAsRead(userObjectID)
}

// DeleteNotification deletes a notification
func (s *notificationService) DeleteNotification(notificationID, userID string) error {
	notification, err := s.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}

	if notification.RecipientID.Hex() != userID {
		return fmt.Errorf("notification does not belong to user")
	}

	return s.notificationRepo.SoftDelete(notificationID)
}

// GetUnreadCount gets unread notification count for user
func (s *notificationService) GetUnreadCount(userID string) (int64, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %w", err)
	}

	return s.notificationRepo.GetUnreadCount(userObjectID)
}

// Template Management Methods

// CreateTemplate creates a new notification template
func (s *notificationService) CreateTemplate(template *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.IsActive = true

	return s.templateRepo.Create(template)
}

// GetTemplate retrieves a template by ID
func (s *notificationService) GetTemplate(templateID string) (*models.NotificationTemplate, error) {
	return s.templateRepo.GetByID(templateID)
}

// GetTemplatesByCategory retrieves templates by category
func (s *notificationService) GetTemplatesByCategory(category string) ([]*models.NotificationTemplate, error) {
	return s.templateRepo.GetByCategory(category)
}

// UpdateTemplate updates a notification template
func (s *notificationService) UpdateTemplate(template *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	template.UpdatedAt = time.Now()
	return s.templateRepo.Update(template)
}

// DeleteTemplate deletes a notification template
func (s *notificationService) DeleteTemplate(templateID string) error {
	return s.templateRepo.Delete(templateID)
}

// SendTemplatedNotification sends notification using template
func (s *notificationService) SendTemplatedNotification(userID, templateID string, variables map[string]interface{}) error {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return err
	}

	// Process template with variables
	title := s.processTemplate(template.TitleTemplate, variables)
	body := s.processTemplate(template.BodyTemplate, variables)

	request := &NotificationRequest{
		Title:    title,
		Body:     body,
		Type:     template.Type,
		Priority: template.DefaultPriority,
		Category: template.Category,
		Channels: template.DefaultChannels,
		ImageURL: template.DefaultImageURL,
	}

	return s.SendNotificationToUser(userID, request)
}

// Preference Management Methods

// GetUserPreferences retrieves user notification preferences
func (s *notificationService) GetUserPreferences(userID string) (*models.NotificationPreferences, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	preferences, err := s.preferenceRepo.GetByUserID(userObjectID)
	if err != nil {
		// Create default preferences if not found
		return s.CreateDefaultPreferences(userID)
	}

	return preferences, nil
}

// UpdateUserPreferences updates user notification preferences
func (s *notificationService) UpdateUserPreferences(userID string, preferences *models.NotificationPreferences) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	preferences.UserID = userObjectID
	preferences.UpdatedAt = time.Now()

	return s.preferenceRepo.Update(preferences)
}

// CreateDefaultPreferences creates default preferences for user
func (s *notificationService) CreateDefaultPreferences(userID string) (*models.NotificationPreferences, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	preferences := &models.NotificationPreferences{
		ID:                  primitive.NewObjectID(),
		UserID:              userObjectID,
		GlobalNotifications: true,
		QuietHoursEnabled:   false,
		QuietHoursStart:     "22:00",
		QuietHoursEnd:       "08:00",
		TimeZone:            "UTC",
		// Set default channel preferences
		PushNotifications: models.ChannelPreference{
			Enabled: true,
			Types:   []string{"ride", "payment", "system"},
		},
		SMSNotifications: models.ChannelPreference{
			Enabled: true,
			Types:   []string{"emergency", "verification"},
		},
		EmailNotifications: models.ChannelPreference{
			Enabled: true,
			Types:   []string{"payment", "system", "promotional"},
		},
		InAppNotifications: models.ChannelPreference{
			Enabled: true,
			Types:   []string{"all"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.preferenceRepo.Create(preferences)
}

// Utility methods

// validateNotification validates notification data
func (s *notificationService) validateNotification(notification *models.Notification) error {
	if notification.Title == "" {
		return fmt.Errorf("notification title is required")
	}
	if notification.Body == "" {
		return fmt.Errorf("notification body is required")
	}
	if notification.RecipientID.IsZero() {
		return fmt.Errorf("recipient ID is required")
	}
	return nil
}

// filterChannelsByPreferences filters channels based on user preferences
func (s *notificationService) filterChannelsByPreferences(channels []models.NotificationChannel, preferences *models.NotificationPreferences) []models.NotificationChannel {
	if preferences == nil {
		return channels
	}

	filtered := make([]models.NotificationChannel, 0)
	for _, channel := range channels {
		switch channel {
		case models.NotificationChannelPush:
			if preferences.PushNotifications.Enabled {
				filtered = append(filtered, channel)
			}
		case models.NotificationChannelSMS:
			if preferences.SMSNotifications.Enabled {
				filtered = append(filtered, channel)
			}
		case models.NotificationChannelEmail:
			if preferences.EmailNotifications.Enabled {
				filtered = append(filtered, channel)
			}
		case models.NotificationChannelInApp:
			if preferences.InAppNotifications.Enabled {
				filtered = append(filtered, channel)
			}
		}
	}

	return filtered
}

// processTemplate processes template with variables
func (s *notificationService) processTemplate(template string, variables map[string]interface{}) string {
	// Simple template processing - in production, use a proper template engine
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = fmt.Sprintf(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// processDelivery processes notification delivery
func (s *notificationService) processDelivery(notification *models.Notification) {
	// Process delivery through various channels
	// This would integrate with external services like Firebase, Twilio, SendGrid

	s.logger.Info().
		Str("notification_id", notification.ID.Hex()).
		Msg("Processing notification delivery")

	// Update status to sent
	notification.Status = models.NotificationStatusSent
	now := time.Now()
	notification.SentAt = &now
	notification.UpdatedAt = now

	s.notificationRepo.Update(notification)
}

// Placeholder implementations for remaining methods
func (s *notificationService) ValidateNotificationRequest(request *NotificationRequest) error {
	if request.Title == "" {
		return fmt.Errorf("title is required")
	}
	if request.Body == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

func (s *notificationService) ScheduleNotification(notification *models.Notification, scheduledAt time.Time) error {
	notification.ScheduledAt = &scheduledAt
	notification.Status = models.NotificationStatusScheduled
	_, err := s.notificationRepo.Create(notification)
	return err
}

func (s *notificationService) CancelScheduledNotification(notificationID string) error {
	notification, err := s.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}
	notification.Status = models.NotificationStatusCancelled
	return s.notificationRepo.Update(notification)
}

func (s *notificationService) GetScheduledNotifications(userID string) ([]*models.Notification, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return s.notificationRepo.GetScheduledByRecipientID(userObjectID)
}

func (s *notificationService) SendPushNotification(userID string, title, body string, data map[string]interface{}) error {
	request := &NotificationRequest{
		Title:    title,
		Body:     body,
		Type:     models.NotificationTypePush,
		Priority: models.NotificationPriorityMedium,
		Channels: []models.NotificationChannel{models.NotificationChannelPush},
		Data:     data,
	}
	return s.SendNotificationToUser(userID, request)
}

func (s *notificationService) RegisterDeviceToken(userID, deviceToken, platform string) error {
	// Implementation would store device token for push notifications
	return nil
}

func (s *notificationService) UnregisterDeviceToken(userID, deviceToken string) error {
	// Implementation would remove device token
	return nil
}

func (s *notificationService) SendSMSNotification(phone, message string) error {
	// Implementation would use SMS service like Twilio
	return nil
}

func (s *notificationService) SendSMSToUser(userID, message string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	return s.SendSMSNotification(user.Phone, message)
}

func (s *notificationService) SendEmailNotification(email, subject, body string) error {
	// Implementation would use email service like SendGrid
	return nil
}

func (s *notificationService) SendEmailToUser(userID, subject, body string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	return s.SendEmailNotification(user.Email, subject, body)
}

func (s *notificationService) SendTemplatedEmail(userID, templateName string, variables map[string]interface{}) error {
	// Implementation would use email templates
	return nil
}

func (s *notificationService) SendInAppNotification(userID string, request *InAppNotificationRequest) error {
	notificationRequest := &NotificationRequest{
		Title:     request.Title,
		Body:      request.Body,
		Type:      models.NotificationTypeInApp,
		Priority:  models.NotificationPriorityLow,
		Channels:  []models.NotificationChannel{models.NotificationChannelInApp},
		ActionURL: request.ActionURL,
		Data:      request.Data,
	}
	return s.SendNotificationToUser(userID, notificationRequest)
}

func (s *notificationService) GetInAppNotifications(userID string, page, limit int) ([]*models.Notification, error) {
	return s.GetUserNotifications(userID, page, limit)
}

func (s *notificationService) SendEmergencyAlert(alert *EmergencyAlertRequest) error {
	request := &NotificationRequest{
		Title:    alert.Title,
		Body:     alert.Message,
		Type:     models.NotificationTypeEmergency,
		Priority: models.NotificationPriorityCritical,
		Channels: []models.NotificationChannel{
			models.NotificationChannelPush,
			models.NotificationChannelSMS,
		},
	}
	return s.SendNotificationToMultipleUsers(alert.UserIDs, request)
}

func (s *notificationService) SendSOSNotification(userID string, location *models.Location) error {
	// Implementation would send SOS to emergency contacts
	return nil
}

func (s *notificationService) GetEmergencyContacts(userID string) ([]string, error) {
	// Implementation would return emergency contact user IDs
	return []string{}, nil
}

func (s *notificationService) SendGeofenceAlert(userID string, alert *GeofenceAlertRequest) error {
	request := &NotificationRequest{
		Title:    alert.Title,
		Body:     alert.Message,
		Type:     models.NotificationTypeLocation,
		Priority: models.NotificationPriorityMedium,
		Channels: []models.NotificationChannel{models.NotificationChannelPush},
	}
	return s.SendNotificationToUser(userID, request)
}

func (s *notificationService) GetNearbyPromotions(userID string, location *models.Location) ([]*models.Notification, error) {
	// Implementation would find location-based promotions
	return []*models.Notification{}, nil
}

func (s *notificationService) GetNotificationAnalytics(userID string, startDate, endDate time.Time) (*NotificationAnalytics, error) {
	// Implementation would calculate analytics
	return &NotificationAnalytics{}, nil
}

func (s *notificationService) GetEngagementStats(templateID string) (*EngagementStats, error) {
	// Implementation would calculate engagement stats
	return &EngagementStats{}, nil
}

func (s *notificationService) TrackNotificationClick(notificationID string) error {
	notification, err := s.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}

	notification.ClickCount++
	now := time.Now()
	notification.ClickedAt = &now
	notification.UpdatedAt = now

	return s.notificationRepo.Update(notification)
}

func (s *notificationService) ProcessNotificationQueue() error {
	// Implementation would process queued notifications
	return nil
}

func (s *notificationService) CleanupOldNotifications(retentionDays int) error {
	// Implementation would clean up old notifications
	return nil
}
