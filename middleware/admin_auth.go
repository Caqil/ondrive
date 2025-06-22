package middleware

import (
	"net/http"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"github.com/gin-gonic/gin"
)

type AdminAuthMiddleware struct {
	jwtManager *utils.JWTManager
	userRepo   repositories.UserRepository
	logger     utils.Logger
}

func NewAdminAuthMiddleware(jwtManager *utils.JWTManager, userRepo repositories.UserRepository, logger utils.Logger) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		jwtManager: jwtManager,
		userRepo:   userRepo,
		logger:     logger,
	}
}

// AdminAuthRequired validates admin authentication for admin panel
func (m *AdminAuthMiddleware) AdminAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for session-based authentication first (for admin panel UI)
		if m.checkSessionAuth(c) {
			c.Next()
			return
		}

		// Fallback to JWT-based authentication (for admin API)
		if m.checkJWTAuth(c) {
			c.Next()
			return
		}

		// No valid authentication found
		m.handleAuthFailure(c)
	}
}

// checkSessionAuth checks for session-based authentication
func (m *AdminAuthMiddleware) checkSessionAuth(c *gin.Context) bool {
	// Get session token from cookie
	sessionToken, err := c.Cookie("admin_session")
	if err != nil {
		return false
	}

	if sessionToken == "" {
		return false
	}

	// Validate session token (you might want to use Redis for session storage)
	claims, err := m.jwtManager.ValidateToken(sessionToken)
	if err != nil {
		m.logger.Debug().Err(err).Msg("Invalid admin session token")
		return false
	}

	// Verify admin user
	user, err := m.userRepo.GetByID(claims.UserID)
	if err != nil {
		m.logger.Error().Err(err).Str("user_id", claims.UserID).Msg("Admin user not found")
		return false
	}

	// Check if user is admin
	if user.Role != models.RoleAdmin {
		m.logger.Warn().
			Str("user_id", claims.UserID).
			Str("role", string(user.Role)).
			Msg("Non-admin user attempted admin access via session")
		return false
	}

	// Check account status
	if !user.IsActive || user.IsDeleted || user.IsSuspended {
		m.logger.Warn().
			Str("user_id", claims.UserID).
			Bool("is_active", user.IsActive).
			Bool("is_deleted", user.IsDeleted).
			Bool("is_suspended", user.IsSuspended).
			Msg("Inactive admin attempted access")
		return false
	}

	// Set admin context
	m.setAdminContext(c, user, claims)
	return true
}

// checkJWTAuth checks for JWT-based authentication
func (m *AdminAuthMiddleware) checkJWTAuth(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	token, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		m.logger.Debug().Err(err).Msg("Invalid admin authorization header")
		return false
	}

	claims, err := m.jwtManager.ValidateToken(token)
	if err != nil {
		m.logger.Debug().Err(err).Msg("Invalid admin JWT token")
		return false
	}

	// Verify admin user
	user, err := m.userRepo.GetByID(claims.UserID)
	if err != nil {
		m.logger.Error().Err(err).Str("user_id", claims.UserID).Msg("Admin user not found via JWT")
		return false
	}

	// Check if user is admin
	if user.Role != models.RoleAdmin {
		m.logger.Warn().
			Str("user_id", claims.UserID).
			Str("role", string(user.Role)).
			Msg("Non-admin user attempted admin access via JWT")
		return false
	}

	// Check account status
	if !user.IsActive || user.IsDeleted || user.IsSuspended {
		return false
	}

	// Set admin context
	m.setAdminContext(c, user, claims)
	return true
}

// setAdminContext sets admin user context
func (m *AdminAuthMiddleware) setAdminContext(c *gin.Context, user *models.User, claims *utils.JWTClaims) {
	c.Set("admin_id", user.ID.Hex())
	c.Set("admin_email", user.Email)
	c.Set("admin_name", user.Profile.FullName)
	c.Set("admin_user", user)
	c.Set("jwt_claims", claims)
	c.Set("is_admin", true)
}

// handleAuthFailure handles authentication failure
func (m *AdminAuthMiddleware) handleAuthFailure(c *gin.Context) {
	// Check if it's an AJAX/API request
	if c.GetHeader("X-Requested-With") == "XMLHttpRequest" ||
		c.GetHeader("Content-Type") == "application/json" ||
		c.Request.URL.Path[:11] == "/admin/api/" {
		// Return JSON error for API requests
		utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "ADMIN_AUTH_REQUIRED", "Admin authentication required", "")
		c.Abort()
		return
	}

	// Redirect to admin login for browser requests
	c.Redirect(http.StatusFound, "/admin/login")
	c.Abort()
}

// AdminLoginRequired middleware for admin login page (redirects if already logged in)
func (m *AdminAuthMiddleware) AdminLoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if admin is already authenticated
		if m.checkSessionAuth(c) || m.checkJWTAuth(c) {
			// Redirect to dashboard if already logged in
			c.Redirect(http.StatusFound, "/admin/dashboard")
			c.Abort()
			return
		}
		c.Next()
	}
}

// SetAdminSession creates an admin session
func (m *AdminAuthMiddleware) SetAdminSession(c *gin.Context, user *models.User) error {
	// Generate session token (longer-lived than API tokens)
	sessionToken, _, err := m.jwtManager.GenerateAdminToken(user.ID.Hex(), user.Email)
	if err != nil {
		return err
	}

	// Set secure HTTP-only cookie
	c.SetCookie(
		"admin_session",                 // name
		sessionToken,                    // value
		int((24 * time.Hour).Seconds()), // maxAge (24 hours)
		"/admin",                        // path
		"",                              // domain
		false,                           // secure (set to true in production with HTTPS)
		true,                            // httpOnly
	)

	m.logger.Info().
		Str("admin_id", user.ID.Hex()).
		Str("admin_email", user.Email).
		Msg("Admin session created")

	return nil
}

// ClearAdminSession clears the admin session
func (m *AdminAuthMiddleware) ClearAdminSession(c *gin.Context) {
	c.SetCookie(
		"admin_session",
		"",
		-1, // maxAge (delete cookie)
		"/admin",
		"",
		false,
		true,
	)

	adminID := c.GetString("admin_id")
	if adminID != "" {
		m.logger.Info().Str("admin_id", adminID).Msg("Admin session cleared")
	}
}

// SuperAdminOnly restricts access to super admin only
func (m *AdminAuthMiddleware) SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("admin_user")
		if !exists {
			utils.ErrorResponseWithCode(c, http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "Super admin access required", "")
			c.Abort()
			return
		}

		adminUser := user.(*models.User)

		// Check if user has super admin privileges (you might want to add this field to user model)
		// For now, we'll check if user email matches super admin email from config
		// In production, you should have a proper role/permission system
		if adminUser.Email != "admin@indrive.com" { // This should come from config
			m.logger.Warn().
				Str("admin_id", adminUser.ID.Hex()).
				Str("admin_email", adminUser.Email).
				Msg("Non-super-admin attempted super admin access")

			utils.ErrorResponseWithCode(c, http.StatusForbidden, "SUPER_ADMIN_REQUIRED", "Super admin access required", "")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetAdminFromContext extracts admin user from gin context
func GetAdminFromContext(c *gin.Context) (*models.User, bool) {
	admin, exists := c.Get("admin_user")
	if !exists {
		return nil, false
	}
	u, ok := admin.(*models.User)
	return u, ok
}

// GetAdminIDFromContext extracts admin ID from gin context
func GetAdminIDFromContext(c *gin.Context) (string, bool) {
	adminID := c.GetString("admin_id")
	if adminID == "" {
		return "", false
	}
	return adminID, true
}
