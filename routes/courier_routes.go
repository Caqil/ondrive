package routes

import (
	"ondrive/controllers"

	"github.com/gin-gonic/gin"
)

func SetupCourierRoutes(rg *gin.RouterGroup, controller *controllers.CourierController) {
	courier := rg.Group("/courier")
	{
		// Courier Service Management
		courier.POST("/", controller.CreateCourierRequest)
		courier.GET("/", controller.GetCourierRequests)
		courier.GET("/:id", controller.GetCourierRequest)
		courier.PUT("/:id", controller.UpdateCourierRequest)
		courier.DELETE("/:id", controller.CancelCourierRequest)

		// Package Management
		courier.POST("/:id/package-details", controller.SetPackageDetails)
		courier.GET("/:id/package-details", controller.GetPackageDetails)
		courier.PUT("/:id/package-details", controller.UpdatePackageDetails)
		courier.POST("/:id/package-photos", controller.UploadPackagePhotos)
		courier.GET("/:id/package-photos", controller.GetPackagePhotos)

		// Courier Tracking
		courier.GET("/:id/track", controller.TrackCourierDelivery)
		courier.POST("/:id/pickup-confirmation", controller.ConfirmPickup)
		courier.POST("/:id/delivery-confirmation", controller.ConfirmDelivery)
		courier.GET("/:id/delivery-proof", controller.GetDeliveryProof)
		courier.POST("/:id/delivery-proof", controller.UploadDeliveryProof)

		// Recipient Management
		courier.POST("/:id/recipient", controller.SetRecipientDetails)
		courier.GET("/:id/recipient", controller.GetRecipientDetails)
		courier.PUT("/:id/recipient", controller.UpdateRecipientDetails)
		courier.POST("/:id/notify-recipient", controller.NotifyRecipient)

		// Delivery Options
		courier.GET("/delivery-options", controller.GetDeliveryOptions)
		courier.POST("/:id/schedule-delivery", controller.ScheduleDelivery)
		courier.POST("/:id/reschedule-delivery", controller.RescheduleDelivery)
		courier.GET("/:id/delivery-time-slots", controller.GetDeliveryTimeSlots)

		// Special Services
		courier.GET("/fragile-handling", controller.GetFragileHandlingOptions)
		courier.POST("/:id/insurance", controller.AddInsurance)
		courier.GET("/:id/insurance", controller.GetInsuranceDetails)
		courier.POST("/:id/signature-required", controller.RequireSignature)

		// Courier Pricing
		courier.POST("/estimate-price", controller.EstimateCourierPrice)
		courier.GET("/pricing-calculator", controller.GetPricingCalculator)
		courier.GET("/weight-limits", controller.GetWeightLimits)
		courier.GET("/size-limits", controller.GetSizeLimits)

		// Courier History
		courier.GET("/history", controller.GetCourierHistory)
		courier.GET("/sent", controller.GetSentPackages)
		courier.GET("/received", controller.GetReceivedPackages)

		// Courier Support
		courier.POST("/:id/report-issue", controller.ReportCourierIssue)
		courier.GET("/:id/support", controller.GetCourierSupport)
		courier.POST("/:id/claim", controller.FileClaim)

		// Address Book
		courier.GET("/address-book", controller.GetAddressBook)
		courier.POST("/address-book", controller.AddAddress)
		courier.PUT("/address-book/:id", controller.UpdateAddress)
		courier.DELETE("/address-book/:id", controller.DeleteAddress)
	}
}
