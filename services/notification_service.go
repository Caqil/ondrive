package services

import (
	"fmt"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"
	"ondrive/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService interface {
	// Core Notification Management
	SendNotification(notification *models.Notification) error
	SendBulkNotifications(notifications []*models.Notification) error
	GetNotifications(userID string, page, limit int) ([]*models.Notification, int64, error)
	GetNotification(id string) (*models.Notification, error)
	MarkAsRead(id, userID string) error
	MarkAllAsRead(userID string) error
	DeleteNotification(id string) error
	ClearAllNotifications(userID string) error

	// Category-specific Notifications
	GetRideNotifications(userID string, page, limit int) ([]*models.Notification, int64, error)
	GetPaymentNotifications(userID string, page, limit int) ([]*models.Notification, int64, error)
	GetPromotionalNotifications(userID string, page, limit int) ([]*models.Notification, int64, error)
	GetSystemAlerts(userID string, page, limit int) ([]*models.Notification, int64, error)

	// Push Notifications
	RegisterDevice(userID string, deviceToken *models.DeviceToken) error
	UpdateDeviceToken(userID string, oldToken, newToken string) error
	UnregisterDevice(userID, token string) error
	SendTestPush(userID, message string) error

	// Email Notifications
	GetEmailPreferences(userID string) (*models.EmailPreferences, error)
	UpdateEmailPreferences(userID string, preferences *models.EmailPreferences) error
	SubscribeNewsletter(userID, email string) error
	UnsubscribeNewsletter(userID string) error

	// SMS Notifications
	GetSMSPreferences(userID string) (*models.SMSPreferences, error)
	UpdateSMSPreferences(userID string, preferences *models.SMSPreferences) error
	VerifyPhoneForSMS(userID, phone string) error

	// In-App Notifications
	GetInAppNotifications(userID string, page, limit int) ([]*models.Notification, int64, error)
	UpdateInAppPreferences(userID string, preferences *models.InAppPreferences) error
	GetUnreadCount(userID string) (int64, error)

	// Notification Templates
	GetNotificationTemplates() ([]*models.NotificationTemplate, error)
	SendCustomNotification(userID string, notification *models.CustomNotification) error

	// Notification History
	GetNotificationHistory(userID string, days int, page, limit int) ([]*models.Notification, int64, error)
	GetDeliveryStatus(id string) (*models.Notification, error)

	// Emergency Notifications
	SendEmergencyAlert(userID string, alert *models.EmergencyAlert) error
	GetEmergencyContacts(userID string) ([]*models.EmergencyContact, error)
	SendSOSNotification(userID string, location *models.Location) error

	// Location-Based Notifications
	SetGeofenceAlert(userID string, alert *models.GeofenceAlert) error
	GetNearbyPromotions(userID string, location *models.Location) ([]*models.Promotion, error)

	// Notification Analytics
	GetNotificationAnalytics(userID string, startDate, endDate time.Time) (*models.NotificationAnalytics, error)
	GetEngagementStats(templateID string, startDate, endDate time.Time) (*models.EngagementStats, error)

	// Courier-specific notification methods
	NotifyRequestAccepted(senderID, requestID, courierID primitive.ObjectID) error
	NotifyRequestCancelled(userID, requestID primitive.ObjectID, reason string) error
	NotifyPickupConfirmed(senderID, requestID primitive.ObjectID) error
	NotifyDeliveryCompleted(senderID, requestID primitive.ObjectID) error
	NotifyRecipient(requestID primitive.ObjectID, message string, sendSMS, sendEmail, sendPush bool) error

	// Legacy SMS/Email methods for backward compatibility
	SendSMS(phone, message string) error
	SendEmail(email, subject, message string) error
	SendPushNotification(userID primitive.ObjectID, message string) error
	SendPushNotificationByContact(phone, message string) error

	// Additional methods used by existing services
	SendNotificationToUser(userID string, notification *NotificationRequest) error
	SendNotificationToMultipleUsers(userIDs []string, notification *NotificationRequest) error
	SendBulkNotification(request *BulkNotificationRequest) error
	GetNotificationByID(notificationID string) (*models.Notification, error)
	CreateTemplate(template *models.NotificationTemplate) (*models.NotificationTemplate, error)
	GetTemplate(templateID string) (*models.NotificationTemplate, error)
	GetTemplatesByCategory(category string) ([]*models.NotificationTemplate, error)
	UpdateTemplate(template *models.NotificationTemplate) (*models.NotificationTemplate, error)
	DeleteTemplate(templateID string) error
	SendTemplatedNotification(userID, templateID string, variables map[string]interface{}) error
	ScheduleNotification(notification *models.Notification, scheduledAt time.Time) error
	CancelScheduledNotification(notificationID string) error
	GetScheduledNotifications(userID string) ([]*models.Notification, error)
	SendSMSNotification(phone, message string) error
	SendSMSToUser(userID, message string) error
	SendEmailNotification(email, subject, body string) error
	SendEmailToUser(userID, subject, body string) error
	SendTemplatedEmail(userID, templateName string, variables map[string]interface{}) error
	SendInAppNotification(userID string, request *InAppNotificationRequest) error
	SendEmergencyAlert(alert *EmergencyAlertRequest) error
	SendSOSNotification(userID string, location *models.Location) error
	GetEmergencyContacts(userID string) ([]string, error)
	SendGeofenceAlert(userID string, alert *GeofenceAlertRequest) error
	GetNearbyPromotions(userID string, location *models.Location) ([]*models.Notification, error)
	ValidateNotificationRequest(request *NotificationRequest) error
	TrackNotificationClick(notificationID string) error
	ProcessNotificationQueue() error
	CleanupOldNotifications(retentionDays int) error
	RegisterDeviceToken(userID, deviceToken, platform string) error
	UnregisterDeviceToken(userID, deviceToken string) error

	// User Preferences
	GetUserPreferences(userID string) (*models.NotificationPreferences, error)
	UpdateUserPreferences(userID string, preferences *models.NotificationPreferences) error
}

type notificationService struct {
	notificationRepo repositories.NotificationRepository
	templateRepo     repositories.NotificationTemplateRepository
	preferenceRepo   repositories.NotificationPreferenceRepository
	deviceTokenRepo  repositories.DeviceTokenRepository
	userRepo         repositories.UserRepository
	wsHub            *websocket.Hub
	logger           utils.Logger
}

func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	templateRepo repositories.NotificationTemplateRepository,
	preferenceRepo repositories.NotificationPreferenceRepository,
	deviceTokenRepo repositories.DeviceTokenRepository,
	userRepo repositories.UserRepository,
	wsHub *websocket.Hub,
	logger utils.Logger,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		preferenceRepo:   preferenceRepo,
		deviceTokenRepo:  deviceTokenRepo,
		userRepo:         userRepo,
		wsHub:            wsHub,
		logger:           logger,
	}
}

// Core Notification Management

func (s *notificationService) SendNotification(notification *models.Notification) error {
	s.logger.Info().
		Str("user_id", notification.UserID.Hex()).
		Str("type", string(notification.Type)).
		Str("title", notification.Title).
		Msg("Sending notification")

	// Validate notification
	if err := s.validateNotification(notification); err != nil {
		return err
	}

	// Check user preferences
	preferences, err := s.GetUserPreferences(notification.UserID.Hex())
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to get user preferences, using defaults")
	}

	// Filter channels based on preferences
	allowedChannels := s.filterChannelsByPreferences(notification.Channels, preferences)
	notification.Channels = allowedChannels

	// Set timestamps and status
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

	// Send real-time notification via WebSocket
	if s.wsHub != nil && s.shouldSendInApp(allowedChannels) {
		s.sendRealTimeNotification(createdNotification)
	}

	// Process delivery asynchronously
	go s.processDelivery(createdNotification)

	s.logger.Info().
		Str("notification_id", createdNotification.ID.Hex()).
		Msg("Notification sent successfully")

	return nil
}

func (s *notificationService) SendBulkNotifications(notifications []*models.Notification) error {
	s.logger.Info().Int("count", len(notifications)).Msg("Sending bulk notifications")

	// Validate all notifications
	for _, notification := range notifications {
		if err := s.validateNotification(notification); err != nil {
			return fmt.Errorf("notification validation failed: %w", err)
		}
	}

	// Bulk insert
	if err := s.notificationRepo.BulkInsert(notifications); err != nil {
		return err
	}

	// Process deliveries asynchronously
	go func() {
		for _, notification := range notifications {
			s.processDelivery(notification)
		}
	}()

	return nil
}

func (s *notificationService) GetNotifications(userID string, page, limit int) ([]*models.Notification, int64, error) {
	return s.notificationRepo.GetUserNotifications(userID, page, limit)
}

func (s *notificationService) GetNotification(id string) (*models.Notification, error) {
	return s.notificationRepo.GetByID(id)
}

func (s *notificationService) MarkAsRead(id, userID string) error {
	// Mark as read in database
	if err := s.notificationRepo.MarkAsRead(id, userID); err != nil {
		return err
	}

	// Send real-time read receipt
	if s.wsHub != nil {
		s.sendReadReceipt(id, userID)
	}

	return nil
}

func (s *notificationService) MarkAllAsRead(userID string) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

func (s *notificationService) DeleteNotification(id string) error {
	return s.notificationRepo.Delete(id)
}

func (s *notificationService) ClearAllNotifications(userID string) error {
	return s.notificationRepo.ClearAllNotifications(userID)
}

// Category-specific Notifications

func (s *notificationService) GetRideNotifications(userID string, page, limit int) ([]*models.Notification, int64, error) {
	return s.notificationRepo.GetNotificationsByCategory(userID, "ride", page, limit)
}

func (s *notificationService) GetPaymentNotifications(userID string, page, limit int) ([]*models.Notification, int64, error) {
	return s.notificationRepo.GetNotificationsByCategory(userID, "payment", page, limit)
}

func (s *notificationService) GetPromotionalNotifications(userID string, page, limit int) ([]*models.Notification, int64, error) {
	return s.notificationRepo.GetNotificationsByCategory(userID, "promotional", page, limit)
}

func (s *notificationService) GetSystemAlerts(userID string, page, limit int) ([]*models.Notification, int64, error) {
	return s.notificationRepo.GetNotificationsByCategory(userID, "system", page, limit)
}

// Push Notifications

func (s *notificationService) RegisterDevice(userID string, deviceToken *models.DeviceToken) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	deviceToken.UserID = userObjectID
	deviceToken.IsActive = true
	deviceToken.LastUsed = time.Now()

	// Check if token already exists
	existingToken, err := s.deviceTokenRepo.GetByToken(deviceToken.Token)
	if err == nil && existingToken != nil {
		// Update existing token
		existingToken.DeviceName = deviceToken.DeviceName
		existingToken.Platform = deviceToken.Platform
		existingToken.LastUsed = time.Now()
		existingToken.IsActive = true
		_, err = s.deviceTokenRepo.Update(existingToken)
		return err
	}

	// Create new token
	_, err = s.deviceTokenRepo.Create(deviceToken)
	return err
}

func (s *notificationService) UpdateDeviceToken(userID string, oldToken, newToken string) error {
	// Get existing token
	existing, err := s.deviceTokenRepo.GetByToken(oldToken)
	if err != nil {
		return err
	}

	// Update with new token
	existing.Token = newToken
	existing.LastUsed = time.Now()
	existing.UpdatedAt = time.Now()

	_, err = s.deviceTokenRepo.Update(existing)
	return err
}

func (s *notificationService) UnregisterDevice(userID, token string) error {
	return s.deviceTokenRepo.DeleteByToken(token)
}

func (s *notificationService) SendTestPush(userID, message string) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		Type:     models.NotificationTypeSystem,
		Title:    "Test Push Notification",
		Body:     message,
		UserID:   userObjectID,
		Channels: []models.NotificationChannel{models.NotificationChannelPush},
		Category: "test",
		Priority: models.NotificationPriorityNormal,
	}

	return s.SendNotification(notification)
}

// Email Notifications

func (s *notificationService) GetEmailPreferences(userID string) (*models.EmailPreferences, error) {
	preferences, err := s.GetUserPreferences(userID)
	if err != nil || preferences == nil {
		return s.getDefaultEmailPreferences(), nil
	}
	return &preferences.EmailNotifications, nil
}

func (s *notificationService) UpdateEmailPreferences(userID string, emailPrefs *models.EmailPreferences) error {
	preferences, err := s.GetUserPreferences(userID)
	if err != nil {
		return err
	}

	if preferences == nil {
		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}
		preferences = s.getDefaultPreferences(userObjectID)
	}

	preferences.EmailNotifications = *emailPrefs
	return s.UpdateUserPreferences(userID, preferences)
}

func (s *notificationService) SubscribeNewsletter(userID, email string) error {
	emailPrefs, err := s.GetEmailPreferences(userID)
	if err != nil {
		return err
	}

	emailPrefs.Newsletter = true
	emailPrefs.Email = email
	return s.UpdateEmailPreferences(userID, emailPrefs)
}

func (s *notificationService) UnsubscribeNewsletter(userID string) error {
	emailPrefs, err := s.GetEmailPreferences(userID)
	if err != nil {
		return err
	}

	emailPrefs.Newsletter = false
	return s.UpdateEmailPreferences(userID, emailPrefs)
}

// SMS Notifications

func (s *notificationService) GetSMSPreferences(userID string) (*models.SMSPreferences, error) {
	preferences, err := s.GetUserPreferences(userID)
	if err != nil || preferences == nil {
		return s.getDefaultSMSPreferences(), nil
	}
	return &preferences.SMSNotifications, nil
}

func (s *notificationService) UpdateSMSPreferences(userID string, smsPrefs *models.SMSPreferences) error {
	preferences, err := s.GetUserPreferences(userID)
	if err != nil {
		return err
	}

	if preferences == nil {
		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}
		preferences = s.getDefaultPreferences(userObjectID)
	}

	preferences.SMSNotifications = *smsPrefs
	return s.UpdateUserPreferences(userID, preferences)
}

func (s *notificationService) VerifyPhoneForSMS(userID, phone string) error {
	// TODO: Implement phone verification logic
	// This would typically involve sending an OTP and verifying it
	s.logger.Info().
		Str("user_id", userID).
		Str("phone", phone).
		Msg("Phone verification requested for SMS notifications")
	return nil
}

// In-App Notifications

func (s *notificationService) GetInAppNotifications(userID string, page, limit int) ([]*models.Notification, int64, error) {
	return s.notificationRepo.GetUserNotifications(userID, page, limit)
}

func (s *notificationService) UpdateInAppPreferences(userID string, inAppPrefs *models.InAppPreferences) error {
	preferences, err := s.GetUserPreferences(userID)
	if err != nil {
		return err
	}

	if preferences == nil {
		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return err
		}
		preferences = s.getDefaultPreferences(userObjectID)
	}

	preferences.InAppNotifications = *inAppPrefs
	return s.UpdateUserPreferences(userID, preferences)
}

func (s *notificationService) GetUnreadCount(userID string) (int64, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}

// Notification Templates

func (s *notificationService) GetNotificationTemplates() ([]*models.NotificationTemplate, error) {
	return s.templateRepo.GetActive()
}

func (s *notificationService) SendCustomNotification(userID string, customNotification *models.CustomNotification) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		Type:     models.NotificationType(customNotification.Type),
		Title:    customNotification.Title,
		Body:     customNotification.Body,
		UserID:   userObjectID,
		Channels: customNotification.Channels,
		Category: customNotification.Category,
		Priority: customNotification.Priority,
		Data:     customNotification.Data,
	}

	return s.SendNotification(notification)
}

// Notification History

func (s *notificationService) GetNotificationHistory(userID string, days int, page, limit int) ([]*models.Notification, int64, error) {
	return s.notificationRepo.GetNotificationHistory(userID, days, page, limit)
}

func (s *notificationService) GetDeliveryStatus(id string) (*models.Notification, error) {
	return s.notificationRepo.GetDeliveryStatus(id)
}

// Emergency Notifications

func (s *notificationService) SendEmergencyAlert(userID string, alert *models.EmergencyAlert) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		Type:     models.NotificationTypeEmergency,
		Title:    "Emergency Alert",
		Body:     alert.Message,
		UserID:   userObjectID,
		Channels: []models.NotificationChannel{models.NotificationChannelPush, models.NotificationChannelSMS},
		Category: "emergency",
		Priority: models.NotificationPriorityHigh,
		Data: map[string]interface{}{
			"emergency_type": alert.Type,
			"location":       alert.Location,
			"contact_info":   alert.ContactInfo,
		},
	}

	return s.SendNotification(notification)
}

func (s *notificationService) GetEmergencyContacts(userID string) ([]*models.EmergencyContact, error) {
	// TODO: Implement emergency contacts retrieval
	// This would typically be stored in user profile or separate collection
	return []*models.EmergencyContact{}, nil
}

func (s *notificationService) SendSOSNotification(userID string, location *models.Location) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	notification := &models.Notification{
		Type:     models.NotificationTypeEmergency,
		Title:    "SOS Alert",
		Body:     "Emergency assistance requested",
		UserID:   userObjectID,
		Channels: []models.NotificationChannel{models.NotificationChannelPush, models.NotificationChannelSMS},
		Category: "emergency",
		Priority: models.NotificationPriorityHigh,
		Data: map[string]interface{}{
			"sos":       true,
			"location":  location,
			"timestamp": time.Now(),
		},
	}

	return s.SendNotification(notification)
}

// Location-Based Notifications

func (s *notificationService) SetGeofenceAlert(userID string, alert *models.GeofenceAlert) error {
	// TODO: Implement geofence alert setup
	// This would typically involve storing geofence rules and monitoring location
	s.logger.Info().
		Str("user_id", userID).
		Interface("alert", alert).
		Msg("Geofence alert set")
	return nil
}

func (s *notificationService) GetNearbyPromotions(userID string, location *models.Location) ([]*models.Promotion, error) {
	// TODO: Implement nearby promotions retrieval
	// This would typically query promotions based on location and user preferences
	return []*models.Promotion{}, nil
}

// Notification Analytics

func (s *notificationService) GetNotificationAnalytics(userID string, startDate, endDate time.Time) (*models.NotificationAnalytics, error) {
	return s.notificationRepo.GetAnalytics(userID, startDate, endDate)
}

func (s *notificationService) GetEngagementStats(templateID string, startDate, endDate time.Time) (*models.EngagementStats, error) {
	return s.notificationRepo.GetEngagementStats(templateID, startDate, endDate)
}

// User Preferences

func (s *notificationService) GetUserPreferences(userID string) (*models.NotificationPreferences, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	preferences, err := s.preferenceRepo.GetByUserID(userObjectID)
	if err != nil {
		return nil, err
	}

	if preferences == nil {
		// Create default preferences
		preferences = s.getDefaultPreferences(userObjectID)
		_, err = s.preferenceRepo.Create(preferences)
		if err != nil {
			return nil, err
		}
	}

	return preferences, nil
}

func (s *notificationService) UpdateUserPreferences(userID string, preferences *models.NotificationPreferences) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	preferences.UserID = userObjectID
	_, err = s.preferenceRepo.Update(preferences)
	return err
}

// Helper Methods

func (s *notificationService) validateNotification(notification *models.Notification) error {
	if notification.Title == "" {
		return fmt.Errorf("notification title is required")
	}
	if notification.Body == "" {
		return fmt.Errorf("notification body is required")
	}
	if notification.UserID.IsZero() {
		return fmt.Errorf("user ID is required")
	}
	return nil
}

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

func (s *notificationService) shouldSendInApp(channels []models.NotificationChannel) bool {
	for _, channel := range channels {
		if channel == models.NotificationChannelInApp {
			return true
		}
	}
	return false
}

func (s *notificationService) sendRealTimeNotification(notification *models.Notification) {
	if s.wsHub == nil {
		return
	}

	wsMessage := websocket.WSMessage{
		Type: "notification",
		Data: map[string]interface{}{
			"id":         notification.ID.Hex(),
			"type":       notification.Type,
			"title":      notification.Title,
			"body":       notification.Body,
			"category":   notification.Category,
			"priority":   notification.Priority,
			"data":       notification.Data,
			"created_at": notification.CreatedAt,
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
	}

	s.wsHub.SendToUser(notification.UserID.Hex(), wsMessage)
}

func (s *notificationService) sendReadReceipt(notificationID, userID string) {
	if s.wsHub == nil {
		return
	}

	wsMessage := websocket.WSMessage{
		Type: "notification_read",
		Data: map[string]interface{}{
			"notification_id": notificationID,
			"read_by":         userID,
			"read_at":         time.Now(),
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
	}

	s.wsHub.SendToUser(userID, wsMessage)
}

func (s *notificationService) processDelivery(notification *models.Notification) {
	// Process delivery through various channels
	// This would integrate with external services like Firebase, Twilio, SendGrid

	for _, channel := range notification.Channels {
		result := models.DeliveryResult{
			Channel:     channel,
			Status:      models.NotificationStatusDelivered,
			DeliveredAt: &notification.CreatedAt,
			Provider:    s.getProviderForChannel(channel),
		}

		// Simulate delivery processing
		switch channel {
		case models.NotificationChannelPush:
			s.processPushDelivery(notification, &result)
		case models.NotificationChannelEmail:
			s.processEmailDelivery(notification, &result)
		case models.NotificationChannelSMS:
			s.processSMSDelivery(notification, &result)
		case models.NotificationChannelInApp:
			s.processInAppDelivery(notification, &result)
		}

		// Update delivery result
		s.notificationRepo.UpdateDeliveryResult(notification.ID.Hex(), result)
	}

	s.logger.Info().
		Str("notification_id", notification.ID.Hex()).
		Int("channels", len(notification.Channels)).
		Msg("Notification delivery processed")
}

func (s *notificationService) processPushDelivery(notification *models.Notification, result *models.DeliveryResult) {
	// TODO: Implement push notification delivery (Firebase, etc.)
	result.Status = models.NotificationStatusDelivered
	result.Provider = "firebase"
}

func (s *notificationService) processEmailDelivery(notification *models.Notification, result *models.DeliveryResult) {
	// TODO: Implement email delivery (SendGrid, etc.)
	result.Status = models.NotificationStatusDelivered
	result.Provider = "sendgrid"
}

func (s *notificationService) processSMSDelivery(notification *models.Notification, result *models.DeliveryResult) {
	// TODO: Implement SMS delivery (Twilio, etc.)
	result.Status = models.NotificationStatusDelivered
	result.Provider = "twilio"
}

func (s *notificationService) processInAppDelivery(notification *models.Notification, result *models.DeliveryResult) {
	// In-app notifications are delivered immediately via WebSocket
	result.Status = models.NotificationStatusDelivered
	result.Provider = "websocket"
}

func (s *notificationService) getProviderForChannel(channel models.NotificationChannel) string {
	switch channel {
	case models.NotificationChannelPush:
		return "firebase"
	case models.NotificationChannelEmail:
		return "sendgrid"
	case models.NotificationChannelSMS:
		return "twilio"
	case models.NotificationChannelInApp:
		return "websocket"
	default:
		return "unknown"
	}
}

func (s *notificationService) getDefaultPreferences(userID primitive.ObjectID) *models.NotificationPreferences {
	return &models.NotificationPreferences{
		UserID: userID,
		PushNotifications: models.PushPreferences{
			Enabled:         true,
			RideUpdates:     true,
			PaymentUpdates:  true,
			Promotional:     true,
			SystemAlerts:    true,
			EmergencyAlerts: true,
		},
		EmailNotifications: models.EmailPreferences{
			Enabled:      true,
			Newsletter:   false,
			RideReceipts: true,
			Promotional:  false,
			SystemAlerts: true,
		},
		SMSNotifications: models.SMSPreferences{
			Enabled:         false,
			RideUpdates:     false,
			EmergencyAlerts: true,
		},
		InAppNotifications: models.InAppPreferences{
			Enabled:      true,
			Sound:        true,
			Vibration:    true,
			ShowPreviews: true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *notificationService) getDefaultEmailPreferences() *models.EmailPreferences {
	return &models.EmailPreferences{
		Enabled:      true,
		Newsletter:   false,
		RideReceipts: true,
		Promotional:  false,
		SystemAlerts: true,
	}
}

func (s *notificationService) getDefaultSMSPreferences() *models.SMSPreferences {
	return &models.SMSPreferences{
		Enabled:         false,
		RideUpdates:     false,
		EmergencyAlerts: true,
	}
}
