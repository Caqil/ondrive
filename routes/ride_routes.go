package routes

import (
	"ondrive/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRideRoutes(rg *gin.RouterGroup, controller *controllers.RideController) {
	rides := rg.Group("/rides")
	{
		// Ride Management
		rides.POST("/", controller.CreateRide)
		rides.GET("/", controller.GetUserRides)
		rides.GET("/:id", controller.GetRide)
		rides.PUT("/:id", controller.UpdateRide)
		rides.DELETE("/:id", controller.CancelRide)

		// Ride Actions
		rides.POST("/:id/accept", controller.AcceptRide)
		rides.POST("/:id/start", controller.StartRide)
		rides.POST("/:id/complete", controller.CompleteRide)
		rides.POST("/:id/cancel", controller.CancelRide)

		// Ride Tracking
		rides.GET("/:id/track", controller.TrackRide)
		rides.POST("/:id/location", controller.UpdateRideLocation)
		rides.GET("/:id/route", controller.GetRideRoute)
		rides.GET("/:id/eta", controller.GetRideETA)

		// Share Ride
		rides.POST("/:id/share", controller.ShareRide)
		rides.GET("/shared/:share_code", controller.GetSharedRide)

		// Ride Scheduling
		rides.POST("/schedule", controller.ScheduleRide)
		rides.GET("/scheduled", controller.GetScheduledRides)
		rides.PUT("/scheduled/:id", controller.UpdateScheduledRide)
		rides.DELETE("/scheduled/:id", controller.CancelScheduledRide)

		// Driver Matching & Nearby
		rides.GET("/nearby-drivers", controller.FindNearbyDrivers)
		rides.POST("/:id/request-driver", controller.RequestSpecificDriver)
		rides.GET("/:id/driver-location", controller.GetDriverLocation)

		// Fare Estimation & Negotiation
		rides.POST("/estimate-fare", controller.EstimateFare)
		rides.POST("/:id/negotiate-fare", controller.NegotiateFare)
		rides.POST("/:id/accept-fare", controller.AcceptFare)
		rides.POST("/:id/counter-offer", controller.CounterOffer)
		rides.GET("/:id/fare-history", controller.GetFareHistory)

		// Ride Preferences
		rides.POST("/:id/preferences", controller.SetRidePreferences)
		rides.GET("/:id/preferences", controller.GetRidePreferences)

		// Special Requirements
		rides.POST("/:id/requirements", controller.SetSpecialRequirements)
		rides.GET("/:id/requirements", controller.GetSpecialRequirements)

		// Ride Reports & Issues
		rides.POST("/:id/report", controller.ReportRideIssue)
		rides.GET("/:id/reports", controller.GetRideReports)

		// Repeat Rides
		rides.POST("/:id/repeat", controller.RepeatRide)
		rides.GET("/frequent-routes", controller.GetFrequentRoutes)
	}
}
