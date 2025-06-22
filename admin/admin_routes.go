package admin

import (
	"ondrive/admin/admin_controllers"
	"ondrive/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(r *gin.Engine, controllers *admin_controllers.AdminControllers) {
	// Serve static files
	r.Static("/static", "./admin/static")
	r.LoadHTMLGlob("admin/templates/**/*")

	// Admin authentication (public)
	admin := r.Group("/admin")
	{
		// Auth routes (no authentication required)
		admin.GET("/", controllers.AuthController.LoginPage)
		admin.GET("/login", controllers.AuthController.LoginPage)
		admin.POST("/login", controllers.AuthController.Login)
		admin.GET("/forgot-password", controllers.AuthController.ForgotPasswordPage)
		admin.POST("/forgot-password", controllers.AuthController.ForgotPassword)
		admin.GET("/reset-password", controllers.AuthController.ResetPasswordPage)
		admin.POST("/reset-password", controllers.AuthController.ResetPassword)
	}

	// Protected admin routes
	adminProtected := r.Group("/admin")
	adminProtected.Use(middleware.AdminAuthRequired())
	{
		// Dashboard
		adminProtected.GET("/dashboard", controllers.DashboardController.Index)
		adminProtected.GET("/dashboard/stats", controllers.DashboardController.GetStats)
		adminProtected.GET("/dashboard/analytics", controllers.DashboardController.Analytics)
		adminProtected.GET("/dashboard/real-time", controllers.DashboardController.RealTimeData)

		// User Management
		users := adminProtected.Group("/users")
		{
			users.GET("/", controllers.UserController.Index)
			users.GET("/passengers", controllers.UserController.Passengers)
			users.GET("/drivers", controllers.UserController.Drivers)
			users.GET("/:id", controllers.UserController.Show)
			users.POST("/:id/verify", controllers.UserController.VerifyUser)
			users.POST("/:id/reject", controllers.UserController.RejectUser)
			users.POST("/:id/suspend", controllers.UserController.SuspendUser)
			users.POST("/:id/activate", controllers.UserController.ActivateUser)
			users.DELETE("/:id", controllers.UserController.DeleteUser)
			users.GET("/:id/rides", controllers.UserController.GetUserRides)
			users.GET("/:id/payments", controllers.UserController.GetUserPayments)
			users.POST("/:id/send-notification", controllers.UserController.SendNotification)
		}

		// Driver Management
		drivers := adminProtected.Group("/drivers")
		{
			drivers.GET("/", controllers.DriverController.Index)
			drivers.GET("/pending-verification", controllers.DriverController.PendingVerification)
			drivers.GET("/online", controllers.DriverController.OnlineDrivers)
			drivers.GET("/:id", controllers.DriverController.Show)
			drivers.POST("/:id/approve", controllers.DriverController.ApproveDriver)
			drivers.POST("/:id/reject", controllers.DriverController.RejectDriver)
			drivers.GET("/:id/documents", controllers.DriverController.ViewDocuments)
			drivers.POST("/:id/documents/approve", controllers.DriverController.ApproveDocuments)
			drivers.POST("/:id/documents/reject", controllers.DriverController.RejectDocuments)
			drivers.GET("/:id/earnings", controllers.DriverController.ViewEarnings)
			drivers.POST("/:id/payout", controllers.DriverController.ProcessPayout)
			drivers.GET("/:id/vehicle", controllers.DriverController.ViewVehicle)
			drivers.POST("/:id/vehicle/approve", controllers.DriverController.ApproveVehicle)
		}

		// Ride Management
		rides := adminProtected.Group("/rides")
		{
			rides.GET("/", controllers.RideController.Index)
			rides.GET("/active", controllers.RideController.ActiveRides)
			rides.GET("/completed", controllers.RideController.CompletedRides)
			rides.GET("/cancelled", controllers.RideController.CancelledRides)
			rides.GET("/:id", controllers.RideController.Show)
			rides.GET("/:id/tracking", controllers.RideController.TrackRide)
			rides.POST("/:id/cancel", controllers.RideController.CancelRide)
			rides.GET("/:id/chat", controllers.RideController.ViewChat)
			rides.GET("/reports", controllers.RideController.Reports)
			rides.GET("/analytics", controllers.RideController.Analytics)
			rides.POST("/:id/refund", controllers.RideController.ProcessRefund)
		}

		// Payment Management
		payments := adminProtected.Group("/payments")
		{
			payments.GET("/", controllers.PaymentController.Index)
			payments.GET("/transactions", controllers.PaymentController.Transactions)
			payments.GET("/refunds", controllers.PaymentController.Refunds)
			payments.GET("/disputes", controllers.PaymentController.Disputes)
			payments.GET("/:id", controllers.PaymentController.Show)
			payments.POST("/:id/refund", controllers.PaymentController.ProcessRefund)
			payments.POST("/:id/resolve-dispute", controllers.PaymentController.ResolveDispute)
			payments.GET("/analytics", controllers.PaymentController.Analytics)
			payments.GET("/revenue-report", controllers.PaymentController.RevenueReport)
			payments.GET("/commission-report", controllers.PaymentController.CommissionReport)
		}

		// Fare Management
		fares := adminProtected.Group("/fares")
		{
			fares.GET("/", controllers.FareController.Index)
			fares.GET("/settings", controllers.FareController.Settings)
			fares.POST("/settings", controllers.FareController.UpdateSettings)
			fares.GET("/base-rates", controllers.FareController.BaseRates)
			fares.POST("/base-rates", controllers.FareController.UpdateBaseRates)
			fares.GET("/commission-rates", controllers.FareController.CommissionRates)
			fares.POST("/commission-rates", controllers.FareController.UpdateCommissionRates)
			fares.GET("/surge-pricing", controllers.FareController.SurgePricing)
			fares.POST("/surge-pricing", controllers.FareController.UpdateSurgePricing)
			fares.GET("/analytics", controllers.FareController.Analytics)
		}

		// Support & Issues
		support := adminProtected.Group("/support")
		{
			support.GET("/", controllers.SupportController.Index)
			support.GET("/tickets", controllers.SupportController.Tickets)
			support.GET("/tickets/:id", controllers.SupportController.ShowTicket)
			support.POST("/tickets/:id/reply", controllers.SupportController.ReplyToTicket)
			support.POST("/tickets/:id/close", controllers.SupportController.CloseTicket)
			support.GET("/reports", controllers.SupportController.Reports)
			support.GET("/reports/:id", controllers.SupportController.ShowReport)
			support.POST("/reports/:id/resolve", controllers.SupportController.ResolveReport)
			support.GET("/faq", controllers.SupportController.FAQ)
			support.POST("/faq", controllers.SupportController.AddFAQ)
			support.PUT("/faq/:id", controllers.SupportController.UpdateFAQ)
			support.DELETE("/faq/:id", controllers.SupportController.DeleteFAQ)
		}

		// Settings & Configuration
		settings := adminProtected.Group("/settings")
		{
			settings.GET("/", controllers.SettingsController.Index)
			settings.GET("/general", controllers.SettingsController.General)
			settings.POST("/general", controllers.SettingsController.UpdateGeneral)
			settings.GET("/notifications", controllers.SettingsController.Notifications)
			settings.POST("/notifications", controllers.SettingsController.UpdateNotifications)
			settings.GET("/integrations", controllers.SettingsController.Integrations)
			settings.POST("/integrations", controllers.SettingsController.UpdateIntegrations)
			settings.GET("/api-keys", controllers.SettingsController.APIKeys)
			settings.POST("/api-keys", controllers.SettingsController.UpdateAPIKeys)
			settings.GET("/email-templates", controllers.SettingsController.EmailTemplates)
			settings.POST("/email-templates", controllers.SettingsController.UpdateEmailTemplates)
			settings.GET("/backup", controllers.SettingsController.Backup)
			settings.POST("/backup/create", controllers.SettingsController.CreateBackup)
		}

		// Analytics & Reports
		analytics := adminProtected.Group("/analytics")
		{
			analytics.GET("/", controllers.AnalyticsController.Index)
			analytics.GET("/revenue", controllers.AnalyticsController.Revenue)
			analytics.GET("/users", controllers.AnalyticsController.Users)
			analytics.GET("/rides", controllers.AnalyticsController.Rides)
			analytics.GET("/drivers", controllers.AnalyticsController.Drivers)
			analytics.GET("/geographic", controllers.AnalyticsController.Geographic)
			analytics.GET("/performance", controllers.AnalyticsController.Performance)
			analytics.GET("/export/:type", controllers.AnalyticsController.ExportReport)
			analytics.GET("/custom-report", controllers.AnalyticsController.CustomReport)
			analytics.POST("/custom-report", controllers.AnalyticsController.GenerateCustomReport)
		}

		// Promotions & Marketing
		promotions := adminProtected.Group("/promotions")
		{
			promotions.GET("/", controllers.PromotionController.Index)
			promotions.GET("/create", controllers.PromotionController.CreatePage)
			promotions.POST("/", controllers.PromotionController.Create)
			promotions.GET("/:id", controllers.PromotionController.Show)
			promotions.GET("/:id/edit", controllers.PromotionController.EditPage)
			promotions.PUT("/:id", controllers.PromotionController.Update)
			promotions.DELETE("/:id", controllers.PromotionController.Delete)
			promotions.POST("/:id/activate", controllers.PromotionController.Activate)
			promotions.POST("/:id/deactivate", controllers.PromotionController.Deactivate)
			promotions.GET("/:id/analytics", controllers.PromotionController.Analytics)
		}

		// Notifications & Communications
		communications := adminProtected.Group("/communications")
		{
			communications.GET("/", controllers.CommunicationController.Index)
			communications.GET("/send-notification", controllers.CommunicationController.SendNotificationPage)
			communications.POST("/send-notification", controllers.CommunicationController.SendNotification)
			communications.GET("/send-email", controllers.CommunicationController.SendEmailPage)
			communications.POST("/send-email", controllers.CommunicationController.SendEmail)
			communications.GET("/send-sms", controllers.CommunicationController.SendSMSPage)
			communications.POST("/send-sms", controllers.CommunicationController.SendSMS)
			communications.GET("/templates", controllers.CommunicationController.Templates)
			communications.POST("/templates", controllers.CommunicationController.CreateTemplate)
			communications.GET("/history", controllers.CommunicationController.History)
		}

		// System Management
		system := adminProtected.Group("/system")
		{
			system.GET("/", controllers.SystemController.Index)
			system.GET("/logs", controllers.SystemController.Logs)
			system.GET("/health", controllers.SystemController.Health)
			system.GET("/performance", controllers.SystemController.Performance)
			system.POST("/clear-cache", controllers.SystemController.ClearCache)
			system.POST("/restart-service", controllers.SystemController.RestartService)
			system.GET("/database", controllers.SystemController.Database)
			system.POST("/database/optimize", controllers.SystemController.OptimizeDatabase)
			system.GET("/websocket-stats", controllers.SystemController.WebSocketStats)
		}

		// Admin User Management
		adminUsers := adminProtected.Group("/admin-users")
		{
			adminUsers.GET("/", controllers.AdminUserController.Index)
			adminUsers.GET("/create", controllers.AdminUserController.CreatePage)
			adminUsers.POST("/", controllers.AdminUserController.Create)
			adminUsers.GET("/:id", controllers.AdminUserController.Show)
			adminUsers.GET("/:id/edit", controllers.AdminUserController.EditPage)
			adminUsers.PUT("/:id", controllers.AdminUserController.Update)
			adminUsers.DELETE("/:id", controllers.AdminUserController.Delete)
			adminUsers.GET("/profile", controllers.AdminUserController.Profile)
			adminUsers.PUT("/profile", controllers.AdminUserController.UpdateProfile)
			adminUsers.POST("/change-password", controllers.AdminUserController.ChangePassword)
		}

		// Live Tracking & Monitoring
		tracking := adminProtected.Group("/tracking")
		{
			tracking.GET("/", controllers.TrackingController.Index)
			tracking.GET("/live-map", controllers.TrackingController.LiveMap)
			tracking.GET("/drivers-online", controllers.TrackingController.OnlineDrivers)
			tracking.GET("/active-rides", controllers.TrackingController.ActiveRides)
			tracking.GET("/ride/:id", controllers.TrackingController.TrackSpecificRide)
			tracking.GET("/heatmap", controllers.TrackingController.Heatmap)
			tracking.GET("/demand-analysis", controllers.TrackingController.DemandAnalysis)
		}

		// API Management
		api := adminProtected.Group("/api-management")
		{
			api.GET("/", controllers.APIController.Index)
			api.GET("/endpoints", controllers.APIController.Endpoints)
			api.GET("/rate-limits", controllers.APIController.RateLimits)
			api.POST("/rate-limits", controllers.APIController.UpdateRateLimits)
			api.GET("/api-keys", controllers.APIController.APIKeys)
			api.POST("/api-keys", controllers.APIController.GenerateAPIKey)
			api.DELETE("/api-keys/:id", controllers.APIController.RevokeAPIKey)
			api.GET("/webhooks", controllers.APIController.Webhooks)
			api.POST("/webhooks", controllers.APIController.CreateWebhook)
		}

		// Logout
		adminProtected.POST("/logout", controllers.AuthController.Logout)
	}

	// API endpoints for admin dashboard (AJAX calls)
	adminAPI := r.Group("/admin/api")
	adminAPI.Use(middleware.AdminAuthRequired())
	{
		adminAPI.GET("/dashboard/stats", controllers.DashboardController.GetDashboardStats)
		adminAPI.GET("/real-time/rides", controllers.DashboardController.GetRealTimeRides)
		adminAPI.GET("/real-time/drivers", controllers.DashboardController.GetRealTimeDrivers)
		adminAPI.GET("/notifications/recent", controllers.DashboardController.GetRecentNotifications)
		adminAPI.GET("/alerts/system", controllers.DashboardController.GetSystemAlerts)
	}
}
