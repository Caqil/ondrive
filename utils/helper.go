package utils

import (
	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext extracts user ID from gin context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID := c.GetString("user_id")
	if userID == "" {
		return "", false
	}
	return userID, true
}

// GetUserFromContext extracts user from gin context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	u, ok := user.(*models.User)
	return u, ok
}

// GetUserRoleFromContext extracts user role from gin context
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	userRole := c.GetString("user_role")
	if userRole == "" {
		return "", false
	}
	return userRole, true
}

// GetJWTClaimsFromContext extracts JWT claims from gin context
func GetJWTClaimsFromContext(c *gin.Context) (interface{}, bool) {
	claims, exists := c.Get("jwt_claims")
	return claims, exists
}
