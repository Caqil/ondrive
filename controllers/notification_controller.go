package controllers

import (
	"net/http"
	"strconv"
	"time"

	"ondrive/middleware"
	"ondrive/models"
	"ondrive/services"
	"ondrive/utils"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	notificationService services.NotificationService
	logger              utils.Logger
}

// Request DTOs
type SendCustomNotificationRequest struct {
	Type     string                       `json:"type" validate:"required"`
	Title    string                       `json:"title" validate:"required"`
	Body     string                       `json:"body" validate:"required"`
	Category string                       `json:"category"`
	Priority models.NotificationPriority  `json:"priority"`
	Channels []models.NotificationChannel `json:"channels"`
	Data     map[string]interface{}       `json:"data,omitempty"`
}

type RegisterDeviceRequest struct {
	Token      string `json:"token" validate:"required"`
	DeviceName string `json:"device_name"`
	Platform   string `json:"platform" validate:"required"`
}

type UpdateDeviceTokenRequest struct {
	OldToken string `json:"old_token" validate:"required"`
	NewToken string `json:"new_token" validate:"required"`
}

type UnregisterDeviceRequest struct {
	Token string `json:"token" validate:"required"`
}

type SendTestPushRequest struct {
	Message string `json:"message" validate:"required"`
}

type UpdateEmailPreferencesRequest struct {
	Enabled      bool   `json:"enabled"`
	Email        string `json:"email"`
	Newsletter   bool   `json:"newsletter"`
	RideReceipts bool   `json:"ride_receipts"`
	Promotional  bool   `json:"promotional"`
	SystemAlerts bool   `json:"system_alerts"`
}

type SubscribeNewsletterRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdateSMSPreferencesRequest struct {
	Enabled         bool   `json:"enabled"`
	Phone           string `json:"phone"`
	RideUpdates     bool   `json:"ride_updates"`
	EmergencyAlerts bool   `json:"emergency_alerts"`
}

type VerifyPhoneRequest struct {
	Phone string `json:"phone" validate:"required"`
}

type UpdateInAppPreferencesRequest struct {
	Enabled      bool `json:"enabled"`
	Sound        bool `json:"sound"`
	Vibration    bool `json:"vibration"`
	ShowPreviews bool `json:"show_previews"`
}

type SendEmergencyAlertRequest struct {
	Type        string            `json:"type" validate:"required"`
	Message     string            `json:"message" validate:"required"`
	Location    *models.Location  `json:"location"`
	ContactInfo map[string]string `json:"contact_info"`
}

type SendSOSRequest struct {
	Location *models.Location `json:"location" validate:"required"`
}

type SetGeofenceAlertRequest struct {
	Name      string          `json:"name" validate:"required"`
	Center    models.Location `json:"center" validate:"required"`
	Radius    float64         `json:"radius" validate:"required,min=1"`
	AlertType string          `json:"alert_type" validate:"required"`
	Message   string          `json:"message"`
	IsActive  bool            `json:"is_active"`
}

type GetNearbyPromotionsRequest struct {
	Location models.Location `json:"location" validate:"required"`
	Radius   float64         `json:"radius"`
}

func NewNotificationController(
	notificationService services.NotificationService,
	logger utils.Logger,
) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		logger:              logger,
	}
}

// Notification Management

func (nc *NotificationController) GetNotifications(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	// Get pagination params
	params := utils.GetPaginationParams(c)

	notifications, total, err := nc.notificationService.GetNotifications(userID, params.Page, params.Limit)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get notifications")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Notifications retrieved successfully", notifications, meta)
}

func (nc *NotificationController) GetNotification(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		utils.BadRequestResponse(c, "Notification ID is required")
		return
	}

	notification, err := nc.notificationService.GetNotification(notificationID)
	if err != nil {
		nc.logger.Error().Err(err).Str("notification_id", notificationID).Msg("Failed to get notification")
		utils.NotFoundResponse(c, "Notification")
		return
	}

	// Check if user owns this notification
	if notification.UserID.Hex() != userID {
		utils.ForbiddenResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification retrieved successfully", notification)
}

func (nc *NotificationController) MarkAsRead(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		utils.BadRequestResponse(c, "Notification ID is required")
		return
	}

	err := nc.notificationService.MarkAsRead(notificationID, userID)
	if err != nil {
		nc.logger.Error().Err(err).
			Str("notification_id", notificationID).
			Str("user_id", userID).
			Msg("Failed to mark notification as read")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification marked as read", nil)
}

func (nc *NotificationController) MarkAllAsRead(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err := nc.notificationService.MarkAllAsRead(userID)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to mark all notifications as read")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "All notifications marked as read", nil)
}

func (nc *NotificationController) DeleteNotification(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		utils.BadRequestResponse(c, "Notification ID is required")
		return
	}

	// Verify ownership before deletion
	notification, err := nc.notificationService.GetNotification(notificationID)
	if err != nil {
		utils.NotFoundResponse(c, "Notification")
		return
	}

	if notification.UserID.Hex() != userID {
		utils.ForbiddenResponse(c)
		return
	}

	err = nc.notificationService.DeleteNotification(notificationID)
	if err != nil {
		nc.logger.Error().Err(err).
			Str("notification_id", notificationID).
			Str("user_id", userID).
			Msg("Failed to delete notification")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification deleted successfully", nil)
}

func (nc *NotificationController) ClearAllNotifications(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err := nc.notificationService.ClearAllNotifications(userID)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to clear all notifications")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "All notifications cleared successfully", nil)
}

// Notification Categories

func (nc *NotificationController) GetRideNotifications(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	notifications, total, err := nc.notificationService.GetRideNotifications(userID, params.Page, params.Limit)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get ride notifications")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Ride notifications retrieved successfully", notifications, meta)
}

func (nc *NotificationController) GetPaymentNotifications(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	notifications, total, err := nc.notificationService.GetPaymentNotifications(userID, params.Page, params.Limit)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get payment notifications")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Payment notifications retrieved successfully", notifications, meta)
}

func (nc *NotificationController) GetPromotionalNotifications(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	notifications, total, err := nc.notificationService.GetPromotionalNotifications(userID, params.Page, params.Limit)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get promotional notifications")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Promotional notifications retrieved successfully", notifications, meta)
}

func (nc *NotificationController) GetSystemAlerts(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	notifications, total, err := nc.notificationService.GetSystemAlerts(userID, params.Page, params.Limit)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get system alerts")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "System alerts retrieved successfully", notifications, meta)
}

// Push Notifications

func (nc *NotificationController) RegisterDevice(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	deviceToken := &models.DeviceToken{
		Token:    req.Token,
		DeviceID: req.DeviceName,
		Platform: req.Platform,
	}

	err := nc.notificationService.RegisterDevice(userID, deviceToken)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to register device")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Device registered successfully", nil)
}

func (nc *NotificationController) UpdateDeviceToken(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := nc.notificationService.UpdateDeviceToken(userID, req.OldToken, req.NewToken)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update device token")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Device token updated successfully", nil)
}

func (nc *NotificationController) UnregisterDevice(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UnregisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := nc.notificationService.UnregisterDevice(userID, req.Token)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to unregister device")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Device unregistered successfully", nil)
}

func (nc *NotificationController) SendTestPush(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req SendTestPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := nc.notificationService.SendTestPush(userID, req.Message)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to send test push")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Test push notification sent", nil)
}

// Email Notifications

func (nc *NotificationController) GetEmailPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	preferences, err := nc.notificationService.GetEmailPreferences(userID)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get email preferences")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Email preferences retrieved successfully", preferences)
}

func (nc *NotificationController) UpdateEmailPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateEmailPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	preferences := &models.EmailPreferences{
		Enabled:      req.Enabled,
		Email:        req.Email,
		Newsletter:   req.Newsletter,
		RideReceipts: req.RideReceipts,
		Promotional:  req.Promotional,
		SystemAlerts: req.SystemAlerts,
	}

	err := nc.notificationService.UpdateEmailPreferences(userID, preferences)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update email preferences")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Email preferences updated successfully", preferences)
}

func (nc *NotificationController) SubscribeNewsletter(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req SubscribeNewsletterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := nc.notificationService.SubscribeNewsletter(userID, req.Email)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to subscribe to newsletter")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successfully subscribed to newsletter", nil)
}

func (nc *NotificationController) UnsubscribeNewsletter(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err := nc.notificationService.UnsubscribeNewsletter(userID)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to unsubscribe from newsletter")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successfully unsubscribed from newsletter", nil)
}

// SMS Notifications

func (nc *NotificationController) GetSMSPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	preferences, err := nc.notificationService.GetSMSPreferences(userID)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get SMS preferences")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "SMS preferences retrieved successfully", preferences)
}

func (nc *NotificationController) UpdateSMSPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateSMSPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	preferences := &models.SMSPreferences{
		Enabled:         req.Enabled,
		Phone:           req.Phone,
		RideUpdates:     req.RideUpdates,
		EmergencyAlerts: req.EmergencyAlerts,
	}

	err := nc.notificationService.UpdateSMSPreferences(userID, preferences)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update SMS preferences")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "SMS preferences updated successfully", preferences)
}

func (nc *NotificationController) VerifyPhoneForSMS(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req VerifyPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := nc.notificationService.VerifyPhoneForSMS(userID, req.Phone)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to verify phone for SMS")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Phone verification initiated", nil)
}

// In-App Notifications

func (nc *NotificationController) GetInAppNotifications(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	notifications, total, err := nc.notificationService.GetInAppNotifications(userID, params.Page, params.Limit)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get in-app notifications")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "In-app notifications retrieved successfully", notifications, meta)
}

func (nc *NotificationController) UpdateInAppPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateInAppPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	preferences := &models.InAppPreferences{
		Enabled:      req.Enabled,
		Sound:        req.Sound,
		Vibration:    req.Vibration,
		ShowPreviews: req.ShowPreviews,
	}

	err := nc.notificationService.UpdateInAppPreferences(userID, preferences)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update in-app preferences")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "In-app preferences updated successfully", preferences)
}

func (nc *NotificationController) GetUnreadCount(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	count, err := nc.notificationService.GetUnreadCount(userID)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get unread count")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Unread count retrieved successfully", map[string]interface{}{
		"unread_count": count,
	})
}

// Notification Templates

func (nc *NotificationController) GetNotificationTemplates(c *gin.Context) {
	templates, err := nc.notificationService.GetNotificationTemplates()
	if err != nil {
		nc.logger.Error().Err(err).Msg("Failed to get notification templates")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification templates retrieved successfully", templates)
}

func (nc *NotificationController) SendCustomNotification(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req SendCustomNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	customNotification := &models.CustomNotification{
		Type:     req.Type,
		Title:    req.Title,
		Body:     req.Body,
		Category: req.Category,
		Priority: req.Priority,
		Channels: req.Channels,
		Data:     req.Data,
	}

	err := nc.notificationService.SendCustomNotification(userID, customNotification)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to send custom notification")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Custom notification sent successfully", nil)
}

// Notification History

func (nc *NotificationController) GetNotificationHistory(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	// Get days parameter (default 30)
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 30
	}

	params := utils.GetPaginationParams(c)
	notifications, total, err := nc.notificationService.GetNotificationHistory(userID, days, params.Page, params.Limit)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get notification history")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Notification history retrieved successfully", notifications, meta)
}

func (nc *NotificationController) GetDeliveryStatus(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		utils.BadRequestResponse(c, "Notification ID is required")
		return
	}

	notification, err := nc.notificationService.GetDeliveryStatus(notificationID)
	if err != nil {
		nc.logger.Error().Err(err).Str("notification_id", notificationID).Msg("Failed to get delivery status")
		utils.NotFoundResponse(c, "Notification")
		return
	}

	// Check ownership
	if notification.UserID.Hex() != userID {
		utils.ForbiddenResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery status retrieved successfully", map[string]interface{}{
		"notification_id":  notification.ID.Hex(),
		"status":           notification.Status,
		"delivery_results": notification.DeliveryResults,
		"created_at":       notification.CreatedAt,
		"updated_at":       notification.UpdatedAt,
	})
}

// Emergency Notifications

func (nc *NotificationController) SendEmergencyAlert(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req SendEmergencyAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	alert := &models.EmergencyAlert{
		Type:        req.Type,
		Message:     req.Message,
		Location:    req.Location,
		ContactInfo: req.ContactInfo,
	}

	err := nc.notificationService.SendEmergencyAlert(userID, alert)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to send emergency alert")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Emergency alert sent successfully", nil)
}

func (nc *NotificationController) GetEmergencyContacts(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	contacts, err := nc.notificationService.GetEmergencyContacts(userID)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get emergency contacts")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Emergency contacts retrieved successfully", contacts)
}

func (nc *NotificationController) SendSOSNotification(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req SendSOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := nc.notificationService.SendSOSNotification(userID, req.Location)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to send SOS notification")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "SOS notification sent successfully", nil)
}

// Location-Based Notifications

func (nc *NotificationController) SetGeofenceAlert(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req SetGeofenceAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	alert := &models.GeofenceAlert{
		Name:      req.Name,
		Center:    req.Center,
		Radius:    req.Radius,
		AlertType: req.AlertType,
		Message:   req.Message,
		IsActive:  req.IsActive,
	}

	err := nc.notificationService.SetGeofenceAlert(userID, alert)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to set geofence alert")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Geofence alert set successfully", nil)
}

func (nc *NotificationController) GetNearbyPromotions(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req GetNearbyPromotionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	promotions, err := nc.notificationService.GetNearbyPromotions(userID, &req.Location)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get nearby promotions")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Nearby promotions retrieved successfully", promotions)
}

// Notification Analytics

func (nc *NotificationController) GetNotificationAnalytics(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	// Get date range parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	analytics, err := nc.notificationService.GetNotificationAnalytics(userID, startDate, endDate)
	if err != nil {
		nc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get notification analytics")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification analytics retrieved successfully", analytics)
}

func (nc *NotificationController) GetEngagementStats(c *gin.Context) {
	templateID := c.Query("template_id")
	if templateID == "" {
		utils.BadRequestResponse(c, "Template ID is required")
		return
	}

	// Get date range parameters
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	stats, err := nc.notificationService.GetEngagementStats(templateID, startDate, endDate)
	if err != nil {
		nc.logger.Error().Err(err).Str("template_id", templateID).Msg("Failed to get engagement stats")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Engagement stats retrieved successfully", stats)
}
