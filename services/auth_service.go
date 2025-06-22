package services

import (
	"fmt"
	"strings"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService interface {
	Register(req *RegisterUserRequest) (*models.User, error)
	Login(identifier, password string) (*models.User, *utils.TokenPair, error)
	VerifyPhone(phone, code string) error
	VerifyEmail(userID, token string) error
	RequestPasswordReset(identifier string) (*PasswordResetToken, error)
	ResetPassword(token, newPassword string) error
	ChangePassword(userID, currentPassword, newPassword string) error
	RefreshTokens(refreshToken string) (*utils.TokenPair, error)
	RevokeToken(token string) error
	ValidateUser(userID string) (*models.User, error)
	GenerateEmailVerificationToken(userID string) (*EmailVerificationToken, error)
	ValidateEmailVerificationToken(token string) (*models.User, error)
	CheckAccountStatus(user *models.User) error
}

type authService struct {
	userRepo         repositories.UserRepository
	tokenRepo        repositories.TokenRepository
	jwtManager       *utils.JWTManager
	otpManager       *utils.OTPManager
	logger           utils.Logger
	passwordResetTTL time.Duration
	emailVerifyTTL   time.Duration
}

type RegisterUserRequest struct {
	Phone       string
	Email       string
	Password    string
	FirstName   string
	LastName    string
	Role        models.UserRole
	Language    string
	Coordinates []float64
}

type PasswordResetToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type EmailVerificationToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAuthService(
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepository,
	jwtManager *utils.JWTManager,
	otpManager *utils.OTPManager,
	logger utils.Logger,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		jwtManager:       jwtManager,
		otpManager:       otpManager,
		logger:           logger,
		passwordResetTTL: 1 * time.Hour,  // 1 hour for password reset
		emailVerifyTTL:   24 * time.Hour, // 24 hours for email verification
	}
}

func (s *authService) Register(req *RegisterUserRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByPhone(req.Phone)
	if existingUser != nil {
		return nil, fmt.Errorf("phone number already registered")
	}

	if req.Email != "" {
		existingUser, _ = s.userRepo.GetByEmail(req.Email)
		if existingUser != nil {
			return nil, fmt.Errorf("email already registered")
		}
	}

	// Validate input
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password")
		return nil, fmt.Errorf("failed to process password")
	}

	// Create user location if coordinates provided
	var userLocation *models.Location
	if len(req.Coordinates) == 2 {
		userLocation = &models.Location{
			Type:        "Point",
			Coordinates: req.Coordinates,
			UpdatedAt:   time.Now(),
		}
	}

	// Create user
	user := &models.User{
		ID:           primitive.NewObjectID(),
		Phone:        utils.NormalizePhoneNumber(req.Phone),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		Profile: models.UserProfile{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			FullName:  req.FirstName + " " + req.LastName,
			Language:  req.Language,
		},
		Location: userLocation,
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
			ShareLocation:            true,
			LocationAccuracy:         "high",
			SaveLocationHistory:      true,
			AutoPayEnabled:           false,
			SavePaymentMethods:       true,
		},
		Verification: models.UserVerification{
			PhoneVerified:    false,
			EmailVerified:    false,
			IdentityVerified: false,
			Documents:        []models.VerificationDoc{},
		},
		EmergencyContacts: []models.EmergencyContact{},
		FavoritePlaces:    []models.FavoritePlace{},
		DeviceTokens:      []models.DeviceToken{},
		IsActive:          true,
		IsDeleted:         false,
		IsSuspended:       false,
		IsVerified:        false,
		Stats: models.UserStats{
			TotalRides:          0,
			CompletedRides:      0,
			CancelledRides:      0,
			AverageRating:       0.0,
			TotalRatings:        0,
			TotalSpent:          0.0,
			TotalSaved:          0.0,
			CancellationRate:    0.0,
			OnTimeRate:          0.0,
			FavoriteServiceType: "",
			JoinedDaysAgo:       0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Initialize driver info if role is driver
	if req.Role == models.RoleDriver {
		user.DriverInfo = &models.DriverInfo{
			Status:      models.DriverStatusOffline,
			IsOnline:    false,
			IsAvailable: false,
			Rating: models.DriverRating{
				Average:     0.0,
				TotalRides:  0,
				TotalRating: 0.0,
			},
			Earnings: models.DriverEarnings{
				TotalEarnings:      0.0,
				NetEarnings:        0.0,
				GrossEarnings:      0.0,
				Today:              0.0,
				ThisWeek:           0.0,
				ThisMonth:          0.0,
				LastMonth:          0.0,
				CommissionRate:     0.20, // 20% default commission
				PayoutFrequency:    "weekly",
				AvgEarningsPerRide: 0.0,
				AvgEarningsPerHour: 0.0,
			},
			Performance: models.DriverPerformance{
				AcceptanceRate:   0.0,
				CompletionRate:   0.0,
				CancellationRate: 0.0,
				OnTimeRate:       0.0,
				CustomerRating:   0.0,
			},
			WorkingHours: models.WorkingHours{
				IsFlexible:            true,
				MaxHoursPerDay:        12,
				MaxHoursPerWeek:       60,
				MinHoursBetweenShifts: 8,
				BreakDuration:         30,
				MaxContinuousHours:    8,
				CurrentlyWorking:      false,
				UpdatedAt:             time.Now(),
			},
			Preferences: models.DriverPreferences{
				AutoAcceptRides:          false,
				MaxDistance:              50.0, // 50km
				MinFareAmount:            5.0,
				AcceptCashPayments:       true,
				AcceptCardPayments:       true,
				AcceptPets:               false,
				AcceptLuggage:            true,
				AcceptSmoking:            false,
				AcceptChildren:           true,
				AcceptFoodDelivery:       true,
				MaxPassengers:            4,
				RequirePhoneVerification: true,
				RideRequestNotifications: true,
				PaymentNotifications:     true,
				PromoNotifications:       true,
				AvoidTolls:               false,
				AvoidHighways:            false,
				PreferFastestRoute:       true,
				WaitingTime:              5, // 5 minutes
				FlexiblePickup:           true,
				FlexibleDropoff:          true,
				UpdatedAt:                time.Now(),
			},
			ServiceAreas: []models.ServiceArea{},
			LastActive:   time.Now(),
			JoinedAt:     time.Now(),
		}
	}

	// Save user
	err = s.userRepo.Create(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user account")
	}

	s.logger.Info().
		Str("user_id", user.ID.Hex()).
		Str("phone", user.Phone).
		Str("role", string(user.Role)).
		Msg("User registered successfully")

	return user, nil
}

func (s *authService) Login(identifier, password string) (*models.User, *utils.TokenPair, error) {
	// Get user by phone or email
	var user *models.User
	var err error

	if utils.IsValidEmail(identifier) {
		user, err = s.userRepo.GetByEmail(identifier)
	} else {
		normalizedPhone := utils.NormalizePhoneNumber(identifier)
		user, err = s.userRepo.GetByPhone(normalizedPhone)
	}

	if err != nil || user == nil {
		s.logger.Warn().Str("identifier", identifier).Msg("Login attempt with invalid identifier")
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		s.logger.Warn().Str("user_id", user.ID.Hex()).Msg("Login attempt with wrong password")
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Check account status
	if err := s.CheckAccountStatus(user); err != nil {
		return nil, nil, err
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(
		user.ID.Hex(),
		string(user.Role),
		user.Email,
		user.Phone,
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate tokens")
		return nil, nil, fmt.Errorf("failed to generate authentication tokens")
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	s.userRepo.Update(user)

	s.logger.Info().
		Str("user_id", user.ID.Hex()).
		Str("phone", user.Phone).
		Str("role", string(user.Role)).
		Msg("User logged in successfully")

	return user, tokens, nil
}

func (s *authService) VerifyPhone(phone, code string) error {
	normalizedPhone := utils.NormalizePhoneNumber(phone)

	// Verify OTP
	err := s.otpManager.ValidateOTP(normalizedPhone, code)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("phone", normalizedPhone).
			Msg("Phone verification failed")
		return fmt.Errorf("invalid verification code")
	}

	// Update user phone verification status
	user, err := s.userRepo.GetByPhone(normalizedPhone)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	user.Verification.PhoneVerified = true
	now := time.Now()
	user.Verification.VerifiedAt = &now
	user.UpdatedAt = now

	// Check if user is fully verified
	if user.Verification.PhoneVerified && (user.Email == "" || user.Verification.EmailVerified) {
		user.IsVerified = true
	}

	err = s.userRepo.Update(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to update user verification status")
		return fmt.Errorf("failed to update verification status")
	}

	s.logger.Info().Str("phone", normalizedPhone).Msg("Phone verified successfully")
	return nil
}

func (s *authService) VerifyEmail(userID, token string) error {
	// Validate email verification token
	user, err := s.ValidateEmailVerificationToken(token)
	if err != nil {
		return err
	}

	if user.ID.Hex() != userID {
		return fmt.Errorf("invalid verification token")
	}

	// Update email verification status
	user.Verification.EmailVerified = true
	now := time.Now()
	user.Verification.VerifiedAt = &now
	user.UpdatedAt = now

	// Check if user is fully verified
	if user.Verification.PhoneVerified && user.Verification.EmailVerified {
		user.IsVerified = true
	}

	err = s.userRepo.Update(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to update email verification status")
		return fmt.Errorf("failed to update verification status")
	}

	// Remove used token
	s.tokenRepo.DeleteEmailVerificationToken(token)

	s.logger.Info().Str("user_id", userID).Msg("Email verified successfully")
	return nil
}

func (s *authService) RequestPasswordReset(identifier string) (*PasswordResetToken, error) {
	// Get user by phone or email
	var user *models.User
	var err error

	if utils.IsValidEmail(identifier) {
		user, err = s.userRepo.GetByEmail(identifier)
	} else {
		normalizedPhone := utils.NormalizePhoneNumber(identifier)
		user, err = s.userRepo.GetByPhone(normalizedPhone)
	}

	if err != nil || user == nil {
		// Don't reveal if user exists to prevent enumeration
		s.logger.Warn().Str("identifier", identifier).Msg("Password reset requested for non-existent user")
		return nil, fmt.Errorf("if the account exists, reset instructions have been sent")
	}

	// Generate reset token
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate password reset token")
		return nil, fmt.Errorf("failed to generate reset token")
	}

	resetToken := &PasswordResetToken{
		Token:     token,
		UserID:    user.ID.Hex(),
		ExpiresAt: time.Now().Add(s.passwordResetTTL),
		CreatedAt: time.Now(),
	}

	// Save reset token
	err = s.tokenRepo.SavePasswordResetToken(resetToken)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to save password reset token")
		return nil, fmt.Errorf("failed to save reset token")
	}

	s.logger.Info().
		Str("user_id", user.ID.Hex()).
		Str("identifier", identifier).
		Msg("Password reset token generated")

	return resetToken, nil
}

func (s *authService) ResetPassword(token, newPassword string) error {
	// Validate token
	resetToken, err := s.tokenRepo.GetPasswordResetToken(token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	if time.Now().After(resetToken.ExpiresAt) {
		s.tokenRepo.DeletePasswordResetToken(token)
		return fmt.Errorf("reset token has expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(resetToken.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Validate password strength
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash new password")
		return fmt.Errorf("failed to process new password")
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()
	err = s.userRepo.Update(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to update password")
		return fmt.Errorf("failed to update password")
	}

	// Remove used token
	s.tokenRepo.DeletePasswordResetToken(token)

	s.logger.Info().Str("user_id", user.ID.Hex()).Msg("Password reset successfully")
	return nil
}

func (s *authService) ChangePassword(userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	if !utils.CheckPasswordHash(currentPassword, user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Check if new password is different from current
	if utils.CheckPasswordHash(newPassword, user.PasswordHash) {
		return fmt.Errorf("new password must be different from current password")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash new password")
		return fmt.Errorf("failed to process new password")
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()
	err = s.userRepo.Update(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to update password")
		return fmt.Errorf("failed to update password")
	}

	s.logger.Info().Str("user_id", userID).Msg("Password changed successfully")
	return nil
}

func (s *authService) RefreshTokens(refreshToken string) (*utils.TokenPair, error) {
	tokens, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Token refresh failed")
		return nil, fmt.Errorf("invalid refresh token")
	}

	return tokens, nil
}

func (s *authService) RevokeToken(token string) error {
	// Add token to blacklist
	err := s.tokenRepo.BlacklistToken(token)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to blacklist token")
		return fmt.Errorf("failed to revoke token")
	}

	return nil
}

func (s *authService) ValidateUser(userID string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := s.CheckAccountStatus(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) GenerateEmailVerificationToken(userID string) (*EmailVerificationToken, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Email == "" {
		return nil, fmt.Errorf("no email address found")
	}

	if user.Verification.EmailVerified {
		return nil, fmt.Errorf("email already verified")
	}

	// Generate verification token
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate email verification token")
		return nil, fmt.Errorf("failed to generate verification token")
	}

	verificationToken := &EmailVerificationToken{
		Token:     token,
		UserID:    userID,
		Email:     user.Email,
		ExpiresAt: time.Now().Add(s.emailVerifyTTL),
		CreatedAt: time.Now(),
	}

	// Save verification token
	err = s.tokenRepo.SaveEmailVerificationToken(verificationToken)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to save email verification token")
		return nil, fmt.Errorf("failed to save verification token")
	}

	s.logger.Info().
		Str("user_id", userID).
		Str("email", user.Email).
		Msg("Email verification token generated")

	return verificationToken, nil
}

func (s *authService) ValidateEmailVerificationToken(token string) (*models.User, error) {
	verificationToken, err := s.tokenRepo.GetEmailVerificationToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired verification token")
	}

	if time.Now().After(verificationToken.ExpiresAt) {
		s.tokenRepo.DeleteEmailVerificationToken(token)
		return nil, fmt.Errorf("verification token has expired")
	}

	user, err := s.userRepo.GetByID(verificationToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *authService) CheckAccountStatus(user *models.User) error {
	if !user.IsActive {
		return fmt.Errorf("account is inactive")
	}

	if user.IsSuspended {
		return fmt.Errorf("account is suspended")
	}

	if user.IsDeleted {
		return fmt.Errorf("account has been deleted")
	}

	return nil
}

// Helper methods

func (s *authService) validateRegisterRequest(req *RegisterUserRequest) error {
	if req.Phone == "" {
		return fmt.Errorf("phone number is required")
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}

	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}

	if req.Role != models.RolePassenger && req.Role != models.RoleDriver {
		return fmt.Errorf("role must be either 'passenger' or 'driver'")
	}

	if req.Email != "" && !utils.IsValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate coordinates if provided
	if len(req.Coordinates) > 0 {
		if len(req.Coordinates) != 2 {
			return fmt.Errorf("coordinates must contain exactly 2 values [longitude, latitude]")
		}
		if !utils.ValidateCoordinates(req.Coordinates[1], req.Coordinates[0]) {
			return fmt.Errorf("invalid coordinates")
		}
	}

	return nil
}

func (s *authService) validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Add more password validation rules as needed
	// - Must contain uppercase
	// - Must contain lowercase
	// - Must contain numbers
	// - Must contain special characters

	return nil
}

// Add email validation utility if not exists
func IsValidEmail(email string) bool {
	// Simple email validation - in production use a proper email validation library
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
