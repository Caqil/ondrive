package routes

import (
	"ondrive/controllers"

	"github.com/gin-gonic/gin"
)

func SetupNotificationRoutes(rg *gin.RouterGroup, controller *controllers.NotificationController) {
	notifications := rg.Group("/notifications")
	{
		// Notification Management
		notifications.GET("/", controller.GetNotifications)
		notifications.GET("/:id", controller.GetNotification)
		notifications.PUT("/:id/read", controller.MarkAsRead)
		notifications.PUT("/mark-all-read", controller.MarkAllAsRead)
		notifications.DELETE("/:id", controller.DeleteNotification)
		notifications.DELETE("/clear-all", controller.ClearAllNotifications)

		// Notification Categories
		notifications.GET("/ride-updates", controller.GetRideNotifications)
		notifications.GET("/payment-updates", controller.GetPaymentNotifications)
		notifications.GET("/promotional", controller.GetPromotionalNotifications)
		notifications.GET("/system-alerts", controller.GetSystemAlerts)

		// Push Notifications
		notifications.POST("/register-device", controller.RegisterDevice)
		notifications.PUT("/update-device", controller.UpdateDeviceToken)
		notifications.DELETE("/unregister-device", controller.UnregisterDevice)
		notifications.POST("/test-push", controller.SendTestPush)

		// Email Notifications
		notifications.GET("/email-preferences", controller.GetEmailPreferences)
		notifications.PUT("/email-preferences", controller.UpdateEmailPreferences)
		notifications.POST("/subscribe-newsletter", controller.SubscribeNewsletter)
		notifications.POST("/unsubscribe-newsletter", controller.UnsubscribeNewsletter)

		// SMS Notifications
		notifications.GET("/sms-preferences", controller.GetSMSPreferences)
		notifications.PUT("/sms-preferences", controller.UpdateSMSPreferences)
		notifications.POST("/verify-phone-for-sms", controller.VerifyPhoneForSMS)

		// In-App Notifications
		notifications.GET("/in-app", controller.GetInAppNotifications)
		notifications.PUT("/in-app-preferences", controller.UpdateInAppPreferences)
		notifications.GET("/unread-count", controller.GetUnreadCount)

		// Notification Templates
		notifications.GET("/templates", controller.GetNotificationTemplates)
		notifications.POST("/custom-notification", controller.SendCustomNotification)

		// Notification History
		notifications.GET("/history", controller.GetNotificationHistory)
		notifications.GET("/delivery-status/:id", controller.GetDeliveryStatus)

		// Emergency Notifications
		notifications.POST("/emergency-alert", controller.SendEmergencyAlert)
		notifications.GET("/emergency-contacts", controller.GetEmergencyContacts)
		notifications.POST("/sos-notification", controller.SendSOSNotification)

		// Location-Based Notifications
		notifications.POST("/geofence-alert", controller.SetGeofenceAlert)
		notifications.GET("/nearby-promotions", controller.GetNearbyPromotions)

		// Notification Analytics
		notifications.GET("/analytics", controller.GetNotificationAnalytics)
		notifications.GET("/engagement-stats", controller.GetEngagementStats)
	}
}
