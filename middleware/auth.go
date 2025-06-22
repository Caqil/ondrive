package middleware

import (
	"net/http"

	"indrive-backend/models"
	"indrive-backend/repositories"
	"indrive-backend/utils"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtManager *utils.JWTManager
	userRepo   repositories.UserRepository
	logger     utils.Logger
}

func NewAuthMiddleware(jwtManager *utils.JWTManager, userRepo repositories.UserRepository, logger utils.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		userRepo:   userRepo,
		logger:     logger,
	}
}

// AuthRequired validates JWT token and sets user context
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn().Str("path", c.Request.URL.Path).Msg("Missing authorization header")
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}

		// Extract Bearer token
		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			m.logger.Warn().Err(err).Str("path", c.Request.URL.Path).Msg("Invalid authorization header format")
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}

		// Validate token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			m.logger.Warn().Err(err).Str("path", c.Request.URL.Path).Msg("Invalid JWT token")
			utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", "")
			c.Abort()
			return
		}

		// Verify user exists and is active
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil {
			m.logger.Error().Err(err).Str("user_id", claims.UserID).Msg("User not found in database")
			utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "USER_NOT_FOUND", "User not found", "")
			c.Abort()
			return
		}

		// Check if user account is active
		if !user.IsActive {
			m.logger.Warn().Str("user_id", claims.UserID).Msg("Inactive user attempted access")
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "ACCOUNT_INACTIVE", "Account is inactive", "")
			c.Abort()
			return
		}

		// Check if user account is suspended
		if user.IsSuspended {
			m.logger.Warn().Str("user_id", claims.UserID).Msg("Suspended user attempted access")
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "ACCOUNT_SUSPENDED", "Account is suspended", "")
			c.Abort()
			return
		}

		// Check if user account is deleted
		if user.IsDeleted {
			m.logger.Warn().Str("user_id", claims.UserID).Msg("Deleted user attempted access")
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "ACCOUNT_DELETED", "Account has been deleted", "")
			c.Abort()
			return
		}

		// Set user context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_email", claims.Email)
		c.Set("user_phone", claims.Phone)
		c.Set("user", user)
		c.Set("jwt_claims", claims)

		// Log successful authentication
		m.logger.Debug().
			Str("user_id", claims.UserID).
			Str("role", claims.Role).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Msg("User authenticated successfully")

		c.Next()
	}
}

// OptionalAuth validates JWT token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Extract and validate token if provided
		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Invalid token format, continue without authentication
			c.Next()
			return
		}

		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Try to get user
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil || !user.IsActive || user.IsDeleted {
			// User not found or inactive, continue without authentication
			c.Next()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_email", claims.Email)
		c.Set("user_phone", claims.Phone)
		c.Set("user", user)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// DriverOnly restricts access to drivers only
func (m *AuthMiddleware) DriverOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != string(models.RoleDriver) {
			m.logger.Warn().
				Str("user_id", c.GetString("user_id")).
				Str("user_role", userRole).
				Str("path", c.Request.URL.Path).
				Msg("Non-driver attempted to access driver-only endpoint")

			utils.ErrorResponseWithCode(c, http.StatusForbidden, "DRIVER_ONLY", "This endpoint is only accessible to drivers", "")
			c.Abort()
			return
		}
		c.Next()
	}
}

// PassengerOnly restricts access to passengers only
func (m *AuthMiddleware) PassengerOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != string(models.RolePassenger) {
			m.logger.Warn().
				Str("user_id", c.GetString("user_id")).
				Str("user_role", userRole).
				Str("path", c.Request.URL.Path).
				Msg("Non-passenger attempted to access passenger-only endpoint")

			utils.ErrorResponseWithCode(c, http.StatusForbidden, "PASSENGER_ONLY", "This endpoint is only accessible to passengers", "")
			c.Abort()
			return
		}
		c.Next()
	}
}

// VerifiedOnly restricts access to verified users only
func (m *AuthMiddleware) VerifiedOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}

		u := user.(*models.User)
		if !u.IsVerified || !u.Verification.IdentityVerified {
			m.logger.Warn().
				Str("user_id", u.ID.Hex()).
				Bool("is_verified", u.IsVerified).
				Bool("identity_verified", u.Verification.IdentityVerified).
				Str("path", c.Request.URL.Path).
				Msg("Unverified user attempted to access verified-only endpoint")

			utils.ErrorResponseWithCode(c, http.StatusForbidden, "VERIFICATION_REQUIRED", "Account verification is required", "")
			c.Abort()
			return
		}
		c.Next()
	}
}

// WSAuthRequired validates JWT token for WebSocket connections
func (m *AuthMiddleware) WSAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check token in query parameter for WebSocket
		token := c.Query("token")
		if token == "" {
			// Fallback to Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				var err error
				token, err = utils.ExtractTokenFromHeader(authHeader)
				if err != nil {
					utils.UnauthorizedResponse(c)
					c.Abort()
					return
				}
			} else {
				utils.UnauthorizedResponse(c)
				c.Abort()
				return
			}
		}

		// Validate token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			m.logger.Warn().Err(err).Str("path", c.Request.URL.Path).Msg("Invalid WebSocket JWT token")
			utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "INVALID_WS_TOKEN", "Invalid WebSocket token", "")
			c.Abort()
			return
		}

		// Verify user exists and is active
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil || !user.IsActive || user.IsDeleted {
			utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "WS_USER_INVALID", "WebSocket user invalid", "")
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_email", claims.Email)
		c.Set("user_phone", claims.Phone)
		c.Set("user", user)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RefreshTokenRequired validates refresh token
func (m *AuthMiddleware) RefreshTokenRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}

		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			utils.UnauthorizedResponse(c)
			c.Abort()
			return
		}

		// For refresh tokens, we allow expired access tokens
		claims, err := m.jwtManager.GetTokenClaims(token)
		if err != nil {
			utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid refresh token", "")
			c.Abort()
			return
		}

		// Set claims for refresh token processing
		c.Set("jwt_claims", claims)
		c.Next()
	}
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

// GetUserIDFromContext extracts user ID from gin context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID := c.GetString("user_id")
	if userID == "" {
		return "", false
	}
	return userID, true
}

// GetUserRoleFromContext extracts user role from gin context
func GetUserRoleFromContext(c *gin.Context) (models.UserRole, bool) {
	role := c.GetString("user_role")
	if role == "" {
		return "", false
	}
	return models.UserRole(role), true
}
