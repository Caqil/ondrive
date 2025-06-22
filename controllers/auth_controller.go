package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"indrive-backend/models"
	"indrive-backend/repositories"
	"indrive-backend/services"
	"indrive-backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthController struct {
	userRepo    repositories.UserRepository
	authService services.AuthService
	jwtManager  *utils.JWTManager
	otpManager  *utils.OTPManager
	logger      utils.Logger
}

func NewAuthController(
	userRepo repositories.UserRepository,
	authService services.AuthService,
	jwtManager *utils.JWTManager,
	otpManager *utils.OTPManager,
	logger utils.Logger,
) *AuthController {
	return &AuthController{
		userRepo:    userRepo,
		authService: authService,
		jwtManager:  jwtManager,
		otpManager:  otpManager,
		logger:      logger,
	}
}

// Register handles user registration
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn().Err(err).Msg("Invalid registration request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Validate request
	if err := ac.validateRegisterRequest(req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Check if user already exists
	existingUser, _ := ac.userRepo.GetByPhone(req.Phone)
	if existingUser != nil {
		utils.ConflictResponse(c, "Phone number already registered")
		return
	}

	if req.Email != "" {
		existingUser, _ = ac.userRepo.GetByEmail(req.Email)
		if existingUser != nil {
			utils.ConflictResponse(c, "Email already registered")
			return
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		ac.logger.Error().Err(err).Msg("Failed to hash password")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Create user
	user := &models.User{
		ID:           primitive.NewObjectID(),
		Phone:        req.Phone,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         models.UserRole(req.Role),
		Profile: models.UserProfile{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			FullName:  req.FirstName + " " + req.LastName,
			Language:  req.Language,
		},
		Settings: models.UserSettings{
			NotificationsEnabled:     true,
			PushNotifications:        true,
			EmailNotifications:       true,
			SMSNotifications:         true,
			RideNotifications:        true,
			PromotionalNotifications: false,
			Language:                 req.Language,
			Currency:                 "USD",
			Theme:                    "light",
		},
		Verification: models.UserVerification{
			PhoneVerified:    false,
			EmailVerified:    false,
			IdentityVerified: false,
		},
		IsActive:  true,
		IsDeleted: false,
		Stats: models.UserStats{
			TotalRides:     0,
			CompletedRides: 0,
			CancelledRides: 0,
			AverageRating:  0.0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user
	err = ac.userRepo.Create(user)
	if err != nil {
		ac.logger.Error().Err(err).Msg("Failed to create user")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Send OTP for phone verification
	otpData, err := ac.otpManager.GenerateOTP(req.Phone)
	if err != nil {
		ac.logger.Error().Err(err).Str("phone", req.Phone).Msg("Failed to generate OTP")
		// Don't fail registration if OTP fails
	} else {
		// Send OTP via SMS (integrate with SMS service)
		ac.logger.Info().Str("phone", req.Phone).Str("otp", utils.MaskOTPForLogging(otpData.Code)).Msg("OTP generated for new user")
	}

	// Generate tokens
	tokens, err := ac.jwtManager.GenerateTokenPair(
		user.ID.Hex(),
		string(user.Role),
		user.Email,
		user.Phone,
	)
	if err != nil {
		ac.logger.Error().Err(err).Msg("Failed to generate tokens")
		utils.InternalServerErrorResponse(c)
		return
	}

	ac.logger.Info().
		Str("user_id", user.ID.Hex()).
		Str("phone", user.Phone).
		Str("role", string(user.Role)).
		Msg("User registered successfully")

	response := RegisterResponse{
		User:        ac.sanitizeUser(user),
		Tokens:      tokens,
		RequiresOTP: true,
		OTPSent:     otpData != nil,
		Message:     "Registration successful. Please verify your phone number.",
	}

	utils.CreatedResponse(c, "User registered successfully", response)
}

// Login handles user authentication
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Get user by phone or email
	var user *models.User
	var err error

	if strings.Contains(req.Identifier, "@") {
		user, err = ac.userRepo.GetByEmail(req.Identifier)
	} else {
		user, err = ac.userRepo.GetByPhone(req.Identifier)
	}

	if err != nil || user == nil {
		ac.logger.Warn().Str("identifier", req.Identifier).Msg("Login attempt with invalid credentials")
		utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid phone/email or password", "")
		return
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		ac.logger.Warn().Str("user_id", user.ID.Hex()).Msg("Login attempt with wrong password")
		utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid phone/email or password", "")
		return
	}

	// Check account status
	if !user.IsActive {
		utils.ErrorResponseWithCode(c, http.StatusForbidden, "ACCOUNT_INACTIVE", "Account is inactive", "")
		return
	}

	if user.IsSuspended {
		utils.ErrorResponseWithCode(c, http.StatusForbidden, "ACCOUNT_SUSPENDED", "Account is suspended", "")
		return
	}

	if user.IsDeleted {
		utils.ErrorResponseWithCode(c, http.StatusForbidden, "ACCOUNT_DELETED", "Account has been deleted", "")
		return
	}

	// Generate tokens
	tokens, err := ac.jwtManager.GenerateTokenPair(
		user.ID.Hex(),
		string(user.Role),
		user.Email,
		user.Phone,
	)
	if err != nil {
		ac.logger.Error().Err(err).Msg("Failed to generate tokens")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	ac.userRepo.Update(user)

	ac.logger.Info().
		Str("user_id", user.ID.Hex()).
		Str("phone", user.Phone).
		Str("role", string(user.Role)).
		Msg("User logged in successfully")

	response := LoginResponse{
		User:                      ac.sanitizeUser(user),
		Tokens:                    tokens,
		RequiresPhoneVerification: !user.Verification.PhoneVerified,
		RequiresEmailVerification: !user.Verification.EmailVerified && user.Email != "",
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// Logout handles user logout
func (ac *AuthController) Logout(c *gin.Context) {
	userID := c.GetString("user_id")

	// In a real implementation, you would blacklist the token
	// For now, we'll just log the logout
	ac.logger.Info().Str("user_id", userID).Msg("User logged out")

	utils.SuccessResponse(c, http.StatusOK, "Logged out successfully", nil)
}

// SendOTP sends OTP for phone verification
func (ac *AuthController) SendOTP(c *gin.Context) {
	var req SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Normalize phone number
	normalizedPhone := utils.NormalizePhoneNumber(req.Phone)

	// Generate OTP
	otpData, err := ac.otpManager.GenerateOTP(normalizedPhone)
	if err != nil {
		ac.logger.Error().Err(err).Str("phone", normalizedPhone).Msg("Failed to generate OTP")
		utils.ErrorResponseWithCode(c, http.StatusTooManyRequests, "OTP_GENERATION_FAILED", err.Error(), "")
		return
	}

	// In production, integrate with SMS service to send OTP
	ac.logger.Info().
		Str("phone", normalizedPhone).
		Str("otp", utils.MaskOTPForLogging(otpData.Code)).
		Msg("OTP generated and sent")

	response := SendOTPResponse{
		Phone:     normalizedPhone,
		ExpiresAt: otpData.ExpiresAt,
		Message:   "OTP sent successfully",
	}

	utils.SuccessResponse(c, http.StatusOK, "OTP sent successfully", response)
}

// VerifyOTP verifies the OTP code
func (ac *AuthController) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Normalize phone number
	normalizedPhone := utils.NormalizePhoneNumber(req.Phone)

	// Verify OTP
	err := ac.otpManager.ValidateOTP(normalizedPhone, req.Code)
	if err != nil {
		ac.logger.Warn().
			Err(err).
			Str("phone", normalizedPhone).
			Str("code", utils.MaskOTPForLogging(req.Code)).
			Msg("OTP verification failed")
		utils.ErrorResponseWithCode(c, http.StatusBadRequest, "INVALID_OTP", err.Error(), "")
		return
	}

	// Update user phone verification status
	user, err := ac.userRepo.GetByPhone(normalizedPhone)
	if err == nil && user != nil {
		user.Verification.PhoneVerified = true
		now := time.Now()
		user.Verification.VerifiedAt = &now
		user.UpdatedAt = now
		ac.userRepo.Update(user)
	}

	ac.logger.Info().Str("phone", normalizedPhone).Msg("Phone verified successfully")

	utils.SuccessResponse(c, http.StatusOK, "Phone verified successfully", gin.H{
		"phone_verified": true,
		"verified_at":    time.Now(),
	})
}

// SendEmailVerification sends email verification
func (ac *AuthController) SendEmailVerification(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := ac.userRepo.GetByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User")
		return
	}

	if user.Email == "" {
		utils.BadRequestResponse(c, "No email address found")
		return
	}

	if user.Verification.EmailVerified {
		utils.BadRequestResponse(c, "Email already verified")
		return
	}

	// Generate verification token
	verificationToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		ac.logger.Error().Err(err).Msg("Failed to generate email verification token")
		utils.InternalServerErrorResponse(c)
		return
	}

	// In production, save token and send email
	ac.logger.Info().
		Str("user_id", userID).
		Str("email", user.Email).
		Str("token", utils.MaskSensitiveData(verificationToken, '*', 4)).
		Msg("Email verification sent")

	utils.SuccessResponse(c, http.StatusOK, "Email verification sent", gin.H{
		"email": utils.MaskEmail(user.Email),
	})
}

// VerifyEmail verifies email with token
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// In production, verify token against stored tokens
	// For now, we'll assume token is valid if it's not empty
	if req.Token == "" {
		utils.BadRequestResponse(c, "Invalid verification token")
		return
	}

	userID := c.GetString("user_id")
	user, err := ac.userRepo.GetByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User")
		return
	}

	// Update email verification status
	user.Verification.EmailVerified = true
	now := time.Now()
	user.Verification.VerifiedAt = &now
	user.UpdatedAt = now
	err = ac.userRepo.Update(user)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	ac.logger.Info().Str("user_id", userID).Msg("Email verified successfully")

	utils.SuccessResponse(c, http.StatusOK, "Email verified successfully", gin.H{
		"email_verified": true,
		"verified_at":    now,
	})
}

// ForgotPassword initiates password reset
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Get user by phone or email
	var user *models.User
	var err error

	if strings.Contains(req.Identifier, "@") {
		user, err = ac.userRepo.GetByEmail(req.Identifier)
	} else {
		normalizedPhone := utils.NormalizePhoneNumber(req.Identifier)
		user, err = ac.userRepo.GetByPhone(normalizedPhone)
	}

	// Always return success to prevent user enumeration
	if err != nil || user == nil {
		ac.logger.Warn().Str("identifier", req.Identifier).Msg("Password reset requested for non-existent user")
		utils.SuccessResponse(c, http.StatusOK, "If the account exists, password reset instructions have been sent", nil)
		return
	}

	// Generate reset token
	resetToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		ac.logger.Error().Err(err).Msg("Failed to generate password reset token")
		utils.InternalServerErrorResponse(c)
		return
	}

	// In production, save token with expiration and send email/SMS
	ac.logger.Info().
		Str("user_id", user.ID.Hex()).
		Str("identifier", req.Identifier).
		Str("token", utils.MaskSensitiveData(resetToken, '*', 4)).
		Msg("Password reset token generated")

	utils.SuccessResponse(c, http.StatusOK, "If the account exists, password reset instructions have been sent", nil)
}

// ResetPassword resets password with token
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// In production, verify token against stored tokens
	if req.Token == "" {
		utils.BadRequestResponse(c, "Invalid reset token")
		return
	}

	// Validate password strength
	if len(req.NewPassword) < 8 {
		utils.ValidationErrorResponse(c, map[string]string{"password": "Password must be at least 8 characters long"})
		return
	}

	// Hash new password
	_, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		ac.logger.Error().Err(err).Msg("Failed to hash new password")
		utils.InternalServerErrorResponse(c)
		return
	}

	// In production, find user by token and update password
	// For now, we'll use a placeholder approach
	ac.logger.Info().Str("token", utils.MaskSensitiveData(req.Token, '*', 4)).Msg("Password reset completed")

	utils.SuccessResponse(c, http.StatusOK, "Password reset successfully", nil)
}

// ChangePassword changes password for authenticated user
func (ac *AuthController) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID := c.GetString("user_id")
	user, err := ac.userRepo.GetByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User")
		return
	}

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		utils.ErrorResponseWithCode(c, http.StatusBadRequest, "INVALID_PASSWORD", "Current password is incorrect", "")
		return
	}

	// Validate new password
	if len(req.NewPassword) < 8 {
		utils.ValidationErrorResponse(c, map[string]string{"password": "Password must be at least 8 characters long"})
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		ac.logger.Error().Err(err).Msg("Failed to hash new password")
		utils.InternalServerErrorResponse(c)
		return
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()
	err = ac.userRepo.Update(user)
	if err != nil {
		utils.InternalServerErrorResponse(c)
		return
	}

	ac.logger.Info().Str("user_id", userID).Msg("Password changed successfully")

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}

// RefreshToken refreshes the access token
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Refresh tokens
	tokens, err := ac.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		ac.logger.Warn().Err(err).Msg("Token refresh failed")
		utils.ErrorResponseWithCode(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid refresh token", "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", tokens)
}

// RevokeToken revokes/blacklists a token
func (ac *AuthController) RevokeToken(c *gin.Context) {
	var req RevokeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// In production, add token to blacklist
	ac.logger.Info().
		Str("user_id", c.GetString("user_id")).
		Str("token", utils.MaskSensitiveData(req.Token, '*', 4)).
		Msg("Token revoked")

	utils.SuccessResponse(c, http.StatusOK, "Token revoked successfully", nil)
}

// GoogleAuth handles Google OAuth authentication
func (ac *AuthController) GoogleAuth(c *gin.Context) {
	var req SocialAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// In production, verify Google token and get user info
	// For now, return not implemented
	utils.ErrorResponseWithCode(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Google authentication not implemented", "")
}

// FacebookAuth handles Facebook OAuth authentication
func (ac *AuthController) FacebookAuth(c *gin.Context) {
	var req SocialAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// In production, verify Facebook token and get user info
	// For now, return not implemented
	utils.ErrorResponseWithCode(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Facebook authentication not implemented", "")
}

// AppleAuth handles Apple OAuth authentication
func (ac *AuthController) AppleAuth(c *gin.Context) {
	var req SocialAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// In production, verify Apple token and get user info
	// For now, return not implemented
	utils.ErrorResponseWithCode(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Apple authentication not implemented", "")
}

// Helper methods

func (ac *AuthController) validateRegisterRequest(req RegisterRequest) error {
	if req.Phone == "" {
		return fmt.Errorf("phone number is required")
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}

	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}

	if req.Role == "" {
		return fmt.Errorf("role is required")
	}

	if req.Role != "passenger" && req.Role != "driver" {
		return fmt.Errorf("role must be either 'passenger' or 'driver'")
	}

	// Validate coordinates if provided
	if len(req.Coordinates) > 0 && len(req.Coordinates) != 2 {
		return fmt.Errorf("coordinates must contain exactly 2 values [longitude, latitude]")
	}

	if len(req.Coordinates) == 2 {
		if !utils.ValidateCoordinates(req.Coordinates[1], req.Coordinates[0]) {
			return fmt.Errorf("invalid coordinates")
		}
	}

	return nil
}

func (ac *AuthController) sanitizeUser(user *models.User) *SanitizedUser {
	return &SanitizedUser{
		ID:           user.ID,
		Phone:        utils.MaskPhoneNumber(user.Phone),
		Email:        utils.MaskEmail(user.Email),
		Role:         user.Role,
		Profile:      user.Profile,
		Verification: user.Verification,
		IsActive:     user.IsActive,
		IsVerified:   user.IsVerified,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// Request/Response Types

type RegisterRequest struct {
	Phone       string    `json:"phone" binding:"required"`
	Email       string    `json:"email"`
	Password    string    `json:"password" binding:"required"`
	FirstName   string    `json:"first_name" binding:"required"`
	LastName    string    `json:"last_name" binding:"required"`
	Role        string    `json:"role" binding:"required"`
	Language    string    `json:"language"`
	Coordinates []float64 `json:"coordinates"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // phone or email
	Password   string `json:"password" binding:"required"`
}

type SendOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Identifier string `json:"identifier" binding:"required"` // phone or email
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RevokeTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type SocialAuthRequest struct {
	Token string `json:"token" binding:"required"`
	Type  string `json:"type"`
}

type RegisterResponse struct {
	User        *SanitizedUser   `json:"user"`
	Tokens      *utils.TokenPair `json:"tokens"`
	RequiresOTP bool             `json:"requires_otp"`
	OTPSent     bool             `json:"otp_sent"`
	Message     string           `json:"message"`
}

type LoginResponse struct {
	User                      *SanitizedUser   `json:"user"`
	Tokens                    *utils.TokenPair `json:"tokens"`
	RequiresPhoneVerification bool             `json:"requires_phone_verification"`
	RequiresEmailVerification bool             `json:"requires_email_verification"`
}

type SendOTPResponse struct {
	Phone     string    `json:"phone"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

type SanitizedUser struct {
	ID           primitive.ObjectID      `json:"id"`
	Phone        string                  `json:"phone"`
	Email        string                  `json:"email"`
	Role         models.UserRole         `json:"role"`
	Profile      models.UserProfile      `json:"profile"`
	Verification models.UserVerification `json:"verification"`
	IsActive     bool                    `json:"is_active"`
	IsVerified   bool                    `json:"is_verified"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}
