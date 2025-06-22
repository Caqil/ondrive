package routes

import (
	"ondrive/controllers"

	"github.com/gin-gonic/gin"
)

func SetupPublicRoutes(rg *gin.RouterGroup, controllers *controllers.Controllers) {
	public := rg.Group("/public")
	{
		// Service Information
		public.GET("/service-areas", controllers.PublicController.GetServiceAreas)
		public.GET("/vehicle-types", controllers.PublicController.GetVehicleTypes)
		public.GET("/service-types", controllers.PublicController.GetServiceTypes)
		public.GET("/supported-cities", controllers.PublicController.GetSupportedCities)

		// Pricing Information
		public.GET("/base-fares", controllers.PublicController.GetBaseFares)
		public.GET("/fare-calculator", controllers.PublicController.GetFareCalculator)
		public.POST("/estimate-fare", controllers.PublicController.EstimatePublicFare)

		// App Information
		public.GET("/app-config", controllers.PublicController.GetAppConfig)
		public.GET("/terms-of-service", controllers.PublicController.GetTermsOfService)
		public.GET("/privacy-policy", controllers.PublicController.GetPrivacyPolicy)
		public.GET("/support-info", controllers.PublicController.GetSupportInfo)

		// Driver Information
		public.GET("/driver-requirements", controllers.PublicController.GetDriverRequirements)
		public.GET("/driver-documents", controllers.PublicController.GetRequiredDocuments)
		public.GET("/vehicle-requirements", controllers.PublicController.GetVehicleRequirements)

		// Contact & Support
		public.POST("/contact", controllers.PublicController.SubmitContactForm)
		public.GET("/faq", controllers.PublicController.GetFAQ)
		public.GET("/help-center", controllers.PublicController.GetHelpCenter)

		// Promotions & Announcements
		public.GET("/promotions", controllers.PublicController.GetActivePromotions)
		public.GET("/announcements", controllers.PublicController.GetAnnouncements)
		public.GET("/news", controllers.PublicController.GetNews)

		// App Downloads
		public.GET("/download-links", controllers.PublicController.GetDownloadLinks)
		public.GET("/version-info", controllers.PublicController.GetVersionInfo)

		// Statistics (Public)
		public.GET("/city-stats", controllers.PublicController.GetCityStats)
		public.GET("/service-stats", controllers.PublicController.GetServiceStats)
	}

	// File uploads (public endpoints)
	uploads := rg.Group("/uploads")
	{
		uploads.POST("/documents", controllers.UploadController.UploadDocument)
		uploads.POST("/images", controllers.UploadController.UploadImage)
		uploads.GET("/files/:filename", controllers.UploadController.ServeFile)
	}

	// Webhooks (external services)
	webhooks := rg.Group("/webhooks")
	{
		webhooks.POST("/stripe", controllers.PaymentController.StripeWebhook)
		webhooks.POST("/paypal", controllers.PaymentController.PayPalWebhook)
		webhooks.POST("/firebase", controllers.NotificationController.FirebaseWebhook)
		webhooks.POST("/twilio", controllers.NotificationController.TwilioWebhook)
	}
}
