package routes

import (
	"ondrive/controllers"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(rg *gin.RouterGroup, controller *controllers.UserController) {
	users := rg.Group("/users")
	{
		// Profile Management
		users.GET("/profile", controller.GetProfile)
		users.PUT("/profile", controller.UpdateProfile)
		users.DELETE("/profile", controller.DeleteProfile)

		// Avatar & Documents
		users.POST("/avatar", controller.UploadAvatar)
		users.DELETE("/avatar", controller.DeleteAvatar)
		users.POST("/documents", controller.UploadDocument)
		users.GET("/documents", controller.GetDocuments)
		users.DELETE("/documents/:id", controller.DeleteDocument)

		// Location Management
		users.PUT("/location", controller.UpdateLocation)
		users.GET("/location", controller.GetLocation)

		// Settings
		users.GET("/settings", controller.GetSettings)
		users.PUT("/settings", controller.UpdateSettings)
		users.PUT("/privacy-settings", controller.UpdatePrivacySettings)
		users.PUT("/notification-settings", controller.UpdateNotificationSettings)

		// Emergency Contacts
		users.GET("/emergency-contacts", controller.GetEmergencyContacts)
		users.POST("/emergency-contacts", controller.AddEmergencyContact)
		users.PUT("/emergency-contacts/:id", controller.UpdateEmergencyContact)
		users.DELETE("/emergency-contacts/:id", controller.DeleteEmergencyContact)

		// Favorite Places
		users.GET("/favorite-places", controller.GetFavoritePlaces)
		users.POST("/favorite-places", controller.AddFavoritePlace)
		users.PUT("/favorite-places/:id", controller.UpdateFavoritePlace)
		users.DELETE("/favorite-places/:id", controller.DeleteFavoritePlace)

		// User Statistics
		users.GET("/stats", controller.GetUserStats)
		users.GET("/ride-history", controller.GetRideHistory)
		users.GET("/payment-history", controller.GetPaymentHistory)
	}
}
