package routes

import (
	"ondrive/controllers"
	"ondrive/middleware"
	"ondrive/websocket"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, controllers *controllers.Controllers, wsHub *websocket.Hub) {
	// CORS middleware
	r.Use(middleware.CORS())

	// Logger middleware
	r.Use(middleware.Logger(utils.NewLogger()))

	// Rate limiting middleware
	r.Use(middleware.RateLimit())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "ondrive"})
	})

	// API version 1
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		SetupAuthRoutes(v1, controllers.AuthController)

		// Public routes
		SetupPublicRoutes(v1, controllers)

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired())
		{
			SetupUserRoutes(protected, controllers.UserController)
			SetupRideRoutes(protected, controllers.RideController)
			SetupDriverRoutes(protected, controllers.DriverController)
			SetupPaymentRoutes(protected, controllers.PaymentController)
			SetupFareRoutes(protected, controllers.FareController)
			SetupChatRoutes(protected, controllers.ChatController)
			SetupRatingRoutes(protected, controllers.RatingController)
			SetupCourierRoutes(protected, controllers.CourierController)
			SetupFreightRoutes(protected, controllers.FreightController)
			SetupNotificationRoutes(protected, controllers.NotificationController)
		}
	}

	// WebSocket routes
	SetupWebSocketRoutes(r, controllers.WebSocketController, wsHub)
}
