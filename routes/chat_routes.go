package routes

import (
	"indrive-backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupChatRoutes(rg *gin.RouterGroup, controller *controllers.ChatController) {
	chat := rg.Group("/chat")
	{
		// Chat Management
		chat.GET("/conversations", controller.GetConversations)
		chat.GET("/conversations/:id", controller.GetConversation)
		chat.POST("/conversations", controller.CreateConversation)
		chat.DELETE("/conversations/:id", controller.DeleteConversation)

		// Messages
		chat.GET("/conversations/:id/messages", controller.GetMessages)
		chat.POST("/conversations/:id/messages", controller.SendMessage)
		chat.PUT("/messages/:id", controller.UpdateMessage)
		chat.DELETE("/messages/:id", controller.DeleteMessage)

		// Message Status
		chat.POST("/messages/:id/read", controller.MarkAsRead)
		chat.POST("/messages/:id/delivered", controller.MarkAsDelivered)
		chat.GET("/conversations/:id/unread-count", controller.GetUnreadCount)

		// Media Messages
		chat.POST("/conversations/:id/send-image", controller.SendImageMessage)
		chat.POST("/conversations/:id/send-location", controller.SendLocationMessage)
		chat.POST("/conversations/:id/send-audio", controller.SendAudioMessage)

		// Quick Replies & Templates
		chat.GET("/quick-replies", controller.GetQuickReplies)
		chat.POST("/quick-replies", controller.CreateQuickReply)
		chat.GET("/message-templates", controller.GetMessageTemplates)

		// Chat Settings
		chat.GET("/settings", controller.GetChatSettings)
		chat.PUT("/settings", controller.UpdateChatSettings)
		chat.POST("/block-user", controller.BlockUser)
		chat.POST("/unblock-user", controller.UnblockUser)
		chat.GET("/blocked-users", controller.GetBlockedUsers)

		// In-Ride Chat (specific to ride)
		chat.GET("/rides/:ride_id/chat", controller.GetRideChat)
		chat.POST("/rides/:ride_id/messages", controller.SendRideMessage)

		// Emergency Chat
		chat.POST("/emergency-chat", controller.StartEmergencyChat)
		chat.GET("/emergency-contacts", controller.GetEmergencyChatContacts)
	}
}
