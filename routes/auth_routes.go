package routes

import (
	"ondrive/controllers"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(rg *gin.RouterGroup, controller *controllers.AuthController) {
	auth := rg.Group("/auth")
	{
		// Registration & Login
		auth.POST("/register", controller.Register)
		auth.POST("/login", controller.Login)
		auth.POST("/logout", controller.Logout)

		// Phone & Email Verification
		auth.POST("/send-otp", controller.SendOTP)
		auth.POST("/verify-otp", controller.VerifyOTP)
		auth.POST("/send-email-verification", controller.SendEmailVerification)
		auth.POST("/verify-email", controller.VerifyEmail)

		// Password Management
		auth.POST("/forgot-password", controller.ForgotPassword)
		auth.POST("/reset-password", controller.ResetPassword)
		auth.POST("/change-password", controller.ChangePassword)

		// Token Management
		auth.POST("/refresh-token", controller.RefreshToken)
		auth.POST("/revoke-token", controller.RevokeToken)

		// Social Authentication
		auth.POST("/google", controller.GoogleAuth)
		auth.POST("/facebook", controller.FacebookAuth)
		auth.POST("/apple", controller.AppleAuth)
	}
}
