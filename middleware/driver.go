package middleware

import (
	"net/http"

	"ondrive/models"
	"ondrive/utils"

	"github.com/gin-gonic/gin"
)

// Package-level middleware functions for easier usage in routes

// DriverOnly restricts access to drivers only - package level function
func DriverOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != string(models.RoleDriver) {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "DRIVER_ONLY", "This endpoint is only accessible to drivers", "")
			c.Abort()
			return
		}
		c.Next()
	}
}

// PassengerOnly restricts access to passengers only - package level function
func PassengerOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != string(models.RolePassenger) {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "PASSENGER_ONLY", "This endpoint is only accessible to passengers", "")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthRequired validates JWT token and sets user context - package level function
func AuthRequired() gin.HandlerFunc {
	// This would need to be implemented with proper dependencies
	// For now, return a placeholder that checks if user context is set
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// VerifiedOnly restricts access to verified users only - package level function
func VerifiedOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}

		u := user.(*models.User)
		if !u.IsVerified || !u.Verification.IdentityVerified {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "VERIFICATION_REQUIRED", "Account verification is required", "")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminOnly restricts access to admin users only - package level function
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != string(models.RoleAdmin) {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "ADMIN_ONLY", "This endpoint is only accessible to administrators", "")
			c.Abort()
			return
		}
		c.Next()
	}
}

// OnlineDriverOnly restricts access to online drivers only
func OnlineDriverOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is a driver
		userRole := c.GetString("user_role")
		if userRole != string(models.RoleDriver) {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "DRIVER_ONLY", "This endpoint is only accessible to drivers", "")
			c.Abort()
			return
		}

		// Check if driver is online (this would require additional context or database check)
		user, exists := c.Get("user")
		if !exists {
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}

		u := user.(*models.User)
		if u.DriverInfo == nil || !u.DriverInfo.IsOnline {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "DRIVER_OFFLINE", "Driver must be online to access this endpoint", "")
			c.Abort()
			return
		}

		c.Next()
	}
}

// AvailableDriverOnly restricts access to available drivers only
func AvailableDriverOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is a driver
		userRole := c.GetString("user_role")
		if userRole != string(models.RoleDriver) {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "DRIVER_ONLY", "This endpoint is only accessible to drivers", "")
			c.Abort()
			return
		}

		// Check if driver is available
		user, exists := c.Get("user")
		if !exists {
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}

		u := user.(*models.User)
		if u.DriverInfo == nil || !u.DriverInfo.IsOnline || !u.DriverInfo.IsAvailable {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "DRIVER_UNAVAILABLE", "Driver must be online and available to access this endpoint", "")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper functions for context extraction
