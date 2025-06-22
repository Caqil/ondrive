package routes

import (
	"ondrive/controllers"
	"ondrive/middleware"
	"ondrive/websocket"

	"github.com/gin-gonic/gin"
)

func SetupWebSocketRoutes(r *gin.Engine, controller *controllers.WebSocketController, wsHub *websocket.Hub) {
	// WebSocket connection endpoint
	ws := r.Group("/ws")
	ws.Use(middleware.WSAuthRequired()) // WebSocket authentication
	{
		// Main WebSocket connection
		ws.GET("/connect", controller.HandleConnection)

		// Specific WebSocket channels
		ws.GET("/rides/:ride_id", controller.HandleRideConnection)
		ws.GET("/drivers/location", controller.HandleDriverLocationConnection)
		ws.GET("/fare-negotiation/:ride_id", controller.HandleFareNegotiationConnection)
		ws.GET("/chat/:conversation_id", controller.HandleChatConnection)

		// Admin WebSocket connections
		ws.GET("/admin/dashboard", controller.HandleAdminDashboardConnection)
		ws.GET("/admin/live-tracking", controller.HandleAdminTrackingConnection)
	}

	// WebSocket API endpoints (for sending messages via REST)
	wsAPI := r.Group("/api/v1/websocket")
	wsAPI.Use(middleware.AuthRequired())
	{
		// Send messages via REST API
		wsAPI.POST("/send-to-user", controller.SendToUser)
		wsAPI.POST("/send-to-ride", controller.SendToRide)
		wsAPI.POST("/broadcast", controller.BroadcastMessage)

		// Connection management
		wsAPI.GET("/connections", controller.GetActiveConnections)
		wsAPI.GET("/users/:user_id/online", controller.CheckUserOnline)
		wsAPI.POST("/disconnect-user", controller.DisconnectUser)

		// WebSocket statistics
		wsAPI.GET("/stats", controller.GetWebSocketStats)
		wsAPI.GET("/health", controller.GetWebSocketHealth)
	}
}
