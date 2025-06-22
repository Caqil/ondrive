package routes

import (
	"ondrive/controllers"
	"ondrive/middleware"

	"github.com/gin-gonic/gin"
)

func SetupDriverRoutes(rg *gin.RouterGroup, controller *controllers.DriverController) {
	drivers := rg.Group("/drivers")
	drivers.Use(middleware.DriverOnly()) // Only drivers can access these routes
	{
		// Driver Profile
		drivers.GET("/profile", controller.GetDriverProfile)
		drivers.PUT("/profile", controller.UpdateDriverProfile)
		drivers.POST("/documents", controller.UploadDriverDocuments)
		drivers.GET("/documents", controller.GetDriverDocuments)

		// Vehicle Management
		drivers.POST("/vehicle", controller.AddVehicle)
		drivers.GET("/vehicle", controller.GetVehicle)
		drivers.PUT("/vehicle", controller.UpdateVehicle)
		drivers.POST("/vehicle/photos", controller.UploadVehiclePhotos)
		drivers.DELETE("/vehicle/photos/:id", controller.DeleteVehiclePhoto)

		// Driver Status & Availability
		drivers.POST("/online", controller.GoOnline)
		drivers.POST("/offline", controller.GoOffline)
		drivers.GET("/status", controller.GetDriverStatus)
		drivers.PUT("/availability", controller.UpdateAvailability)

		// Ride Management for Drivers
		drivers.GET("/available-rides", controller.GetAvailableRides)
		drivers.GET("/current-ride", controller.GetCurrentRide)
		drivers.GET("/ride-requests", controller.GetRideRequests)
		drivers.POST("/rides/:id/accept", controller.AcceptRideRequest)
		drivers.POST("/rides/:id/decline", controller.DeclineRideRequest)

		// Location & Navigation
		drivers.PUT("/location", controller.UpdateDriverLocation)
		drivers.GET("/location", controller.GetDriverLocation)
		drivers.GET("/navigation/:ride_id", controller.GetNavigation)

		// Earnings & Finance
		drivers.GET("/earnings", controller.GetEarnings)
		drivers.GET("/earnings/daily", controller.GetDailyEarnings)
		drivers.GET("/earnings/weekly", controller.GetWeeklyEarnings)
		drivers.GET("/earnings/monthly", controller.GetMonthlyEarnings)
		drivers.GET("/payouts", controller.GetPayouts)
		drivers.POST("/payouts/request", controller.RequestPayout)

		// Driver Statistics
		drivers.GET("/stats", controller.GetDriverStats)
		drivers.GET("/performance", controller.GetPerformanceMetrics)
		drivers.GET("/rating-summary", controller.GetRatingSummary)

		// Working Hours & Schedule
		drivers.GET("/working-hours", controller.GetWorkingHours)
		drivers.PUT("/working-hours", controller.UpdateWorkingHours)
		drivers.GET("/schedule", controller.GetDriverSchedule)
		drivers.POST("/schedule/break", controller.TakeBreak)
		drivers.POST("/schedule/resume", controller.ResumeWork)

		// Driver Preferences
		drivers.GET("/preferences", controller.GetDriverPreferences)
		drivers.PUT("/preferences", controller.UpdateDriverPreferences)
		drivers.PUT("/service-areas", controller.UpdateServiceAreas)
		drivers.PUT("/service-types", controller.UpdateServiceTypes)

		// Documents & Verification
		drivers.POST("/license", controller.UploadDriverLicense)
		drivers.POST("/insurance", controller.UploadInsurance)
		drivers.POST("/registration", controller.UploadVehicleRegistration)
		drivers.GET("/verification-status", controller.GetVerificationStatus)

		// Driver Support
		drivers.POST("/support-ticket", controller.CreateSupportTicket)
		drivers.GET("/support-tickets", controller.GetSupportTickets)
		drivers.PUT("/support-tickets/:id", controller.UpdateSupportTicket)
	}
}
