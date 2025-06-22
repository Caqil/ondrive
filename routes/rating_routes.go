package routes

import (
	"ondrive/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRatingRoutes(rg *gin.RouterGroup, controller *controllers.RatingController) {
	ratings := rg.Group("/ratings")
	{
		// Rating Management
		ratings.POST("/", controller.CreateRating)
		ratings.GET("/:id", controller.GetRating)
		ratings.PUT("/:id", controller.UpdateRating)
		ratings.DELETE("/:id", controller.DeleteRating)

		// Ride Ratings
		ratings.POST("/rides/:ride_id", controller.RateRide)
		ratings.GET("/rides/:ride_id", controller.GetRideRating)
		ratings.GET("/rides/:ride_id/mutual", controller.GetMutualRating)

		// User Ratings
		ratings.GET("/users/:user_id", controller.GetUserRatings)
		ratings.GET("/users/:user_id/summary", controller.GetRatingSummary)
		ratings.GET("/users/:user_id/average", controller.GetAverageRating)

		// Driver Ratings
		ratings.GET("/drivers/:driver_id", controller.GetDriverRatings)
		ratings.GET("/drivers/:driver_id/stats", controller.GetDriverRatingStats)
		ratings.GET("/drivers/:driver_id/breakdown", controller.GetRatingBreakdown)

		// Rating History
		ratings.GET("/history", controller.GetRatingHistory)
		ratings.GET("/given", controller.GetGivenRatings)
		ratings.GET("/received", controller.GetReceivedRatings)

		// Rating Analytics
		ratings.GET("/analytics", controller.GetRatingAnalytics)
		ratings.GET("/trends", controller.GetRatingTrends)
		ratings.GET("/comparison", controller.GetRatingComparison)

		// Rating Categories
		ratings.GET("/categories", controller.GetRatingCategories)
		ratings.GET("/criteria", controller.GetRatingCriteria)

		// Feedback & Reviews
		ratings.POST("/:id/feedback", controller.AddFeedback)
		ratings.GET("/:id/feedback", controller.GetFeedback)
		ratings.POST("/:id/report", controller.ReportRating)

		// Rating Incentives
		ratings.GET("/rewards", controller.GetRatingRewards)
		ratings.POST("/claim-reward", controller.ClaimRatingReward)
	}
}
