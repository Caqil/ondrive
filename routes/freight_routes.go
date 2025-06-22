package routes

import (
	"indrive-backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupFreightRoutes(rg *gin.RouterGroup, controller *controllers.FreightController) {
	freight := rg.Group("/freight")
	{
		// Freight Service Management
		freight.POST("/", controller.CreateFreightRequest)
		freight.GET("/", controller.GetFreightRequests)
		freight.GET("/:id", controller.GetFreightRequest)
		freight.PUT("/:id", controller.UpdateFreightRequest)
		freight.DELETE("/:id", controller.CancelFreightRequest)

		// Cargo Management
		freight.POST("/:id/cargo-details", controller.SetCargoDetails)
		freight.GET("/:id/cargo-details", controller.GetCargoDetails)
		freight.PUT("/:id/cargo-details", controller.UpdateCargoDetails)
		freight.POST("/:id/cargo-photos", controller.UploadCargoPhotos)
		freight.GET("/:id/cargo-photos", controller.GetCargoPhotos)

		// Vehicle Requirements
		freight.GET("/vehicle-types", controller.GetFreightVehicleTypes)
		freight.POST("/:id/vehicle-requirements", controller.SetVehicleRequirements)
		freight.GET("/:id/vehicle-requirements", controller.GetVehicleRequirements)
		freight.GET("/vehicle-availability", controller.CheckVehicleAvailability)

		// Loading & Unloading
		freight.GET("/loading-options", controller.GetLoadingOptions)
		freight.POST("/:id/loading-assistance", controller.RequestLoadingAssistance)
		freight.POST("/:id/unloading-assistance", controller.RequestUnloadingAssistance)
		freight.GET("/:id/equipment-needed", controller.GetEquipmentNeeded)

		// Freight Tracking
		freight.GET("/:id/track", controller.TrackFreightDelivery)
		freight.POST("/:id/loading-confirmation", controller.ConfirmLoading)
		freight.POST("/:id/delivery-confirmation", controller.ConfirmFreightDelivery)
		freight.GET("/:id/milestone-updates", controller.GetMilestoneUpdates)
		freight.POST("/:id/milestone-updates", controller.AddMilestoneUpdate)

		// Route & Distance
		freight.POST("/calculate-route", controller.CalculateFreightRoute)
		freight.GET("/:id/route-optimization", controller.OptimizeRoute)
		freight.GET("/:id/distance-calculation", controller.CalculateDistance)
		freight.GET("/:id/fuel-estimation", controller.EstimateFuelCost)

		// Freight Pricing
		freight.POST("/estimate-price", controller.EstimateFreightPrice)
		freight.GET("/pricing-factors", controller.GetPricingFactors)
		freight.GET("/weight-pricing", controller.GetWeightBasedPricing)
		freight.GET("/distance-pricing", controller.GetDistanceBasedPricing)

		// Documentation
		freight.POST("/:id/documents", controller.UploadFreightDocuments)
		freight.GET("/:id/documents", controller.GetFreightDocuments)
		freight.POST("/:id/customs-declaration", controller.SubmitCustomsDeclaration)
		freight.GET("/:id/shipping-manifest", controller.GetShippingManifest)

		// Insurance & Safety
		freight.GET("/insurance-options", controller.GetInsuranceOptions)
		freight.POST("/:id/insurance", controller.AddFreightInsurance)
		freight.GET("/:id/safety-guidelines", controller.GetSafetyGuidelines)
		freight.POST("/:id/safety-check", controller.PerformSafetyCheck)

		// Multi-Stop Delivery
		freight.POST("/:id/stops", controller.AddDeliveryStop)
		freight.GET("/:id/stops", controller.GetDeliveryStops)
		freight.PUT("/:id/stops/:stop_id", controller.UpdateDeliveryStop)
		freight.DELETE("/:id/stops/:stop_id", controller.RemoveDeliveryStop)
		freight.POST("/:id/optimize-stops", controller.OptimizeDeliveryStops)

		// Freight History & Analytics
		freight.GET("/history", controller.GetFreightHistory)
		freight.GET("/analytics", controller.GetFreightAnalytics)
		freight.GET("/cost-breakdown", controller.GetCostBreakdown)

		// Special Services
		freight.GET("/temperature-controlled", controller.GetTemperatureControlledOptions)
		freight.GET("/hazardous-materials", controller.GetHazardousMaterialsGuidelines)
		freight.GET("/oversized-cargo", controller.GetOversizedCargoOptions)
	}
}
