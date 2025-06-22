package repositories

import (
	"context"
	"fmt"
	"time"

	"ondrive/models"
	"ondrive/services"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	GetByID(userID string) (*models.User, error)
	GetByPhone(phone string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(userID string) error
	List(filter UserFilter) ([]*models.User, int64, error)
	UpdateLastLogin(userID string) error
	UpdateProfile(userID string, profile *models.UserProfile) error
	UpdateSettings(userID string, settings *models.UserSettings) error
	UpdateVerificationStatus(userID string, verification *models.UserVerification) error
	AddEmergencyContact(userID string, contact *models.EmergencyContact) error
	RemoveEmergencyContact(userID, contactID string) error
	AddFavoritePlace(userID string, place *models.FavoritePlace) error
	RemoveFavoritePlace(userID, placeID string) error
	GetUserStats(userID string) (*models.UserStats, error)
	UpdateUserStats(userID string, stats *models.UserStats) error
}

// TokenRepository interface for auth token operations
type TokenRepository interface {
	SavePasswordResetToken(token *services.PasswordResetToken) error
	GetPasswordResetToken(token string) (*services.PasswordResetToken, error)
	DeletePasswordResetToken(token string) error
	SaveEmailVerificationToken(token *services.EmailVerificationToken) error
	GetEmailVerificationToken(token string) (*services.EmailVerificationToken, error)
	DeleteEmailVerificationToken(token string) error
	BlacklistToken(token string) error
	IsTokenBlacklisted(token string) bool
	CleanupExpiredTokens() error
}

// UserFilter for user listing and searching
type UserFilter struct {
	Role          *models.UserRole `json:"role,omitempty"`
	IsActive      *bool            `json:"is_active,omitempty"`
	IsVerified    *bool            `json:"is_verified,omitempty"`
	IsSuspended   *bool            `json:"is_suspended,omitempty"`
	CreatedAfter  *time.Time       `json:"created_after,omitempty"`
	CreatedBefore *time.Time       `json:"created_before,omitempty"`
	Search        string           `json:"search,omitempty"` // Search by name, phone, email
	City          string           `json:"city,omitempty"`
	Country       string           `json:"country,omitempty"`
	Page          int              `json:"page,omitempty"`
	Limit         int              `json:"limit,omitempty"`
	SortBy        string           `json:"sort_by,omitempty"`
	SortOrder     int              `json:"sort_order,omitempty"` // 1 for asc, -1 for desc
}

// Implementations

type userRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

type tokenRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *mongo.Database, logger utils.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *mongo.Database, logger utils.Logger) TokenRepository {
	return &tokenRepository{
		db:     db,
		logger: logger,
	}
}

// UserRepository Implementation

func (r *userRepository) Create(user *models.User) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", user.ID.Hex()).Msg("Failed to create user")
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("user with this phone or email already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info().Str("user_id", user.ID.Hex()).Msg("User created successfully")
	return nil
}

func (r *userRepository) GetByID(userID string) (*models.User, error) {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objectID, "is_deleted": false}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByPhone(phone string) (*models.User, error) {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	normalizedPhone := utils.NormalizePhoneNumber(phone)

	var user models.User
	err := collection.FindOne(ctx, bson.M{
		"phone":      normalizedPhone,
		"is_deleted": false,
	}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("phone", normalizedPhone).Msg("Failed to get user by phone")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{
		"email":      email,
		"is_deleted": false,
	}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("email", email).Msg("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(
		ctx,
		bson.M{"_id": user.ID},
		user,
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", user.ID.Hex()).Msg("Failed to update user")
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepository) Delete(userID string) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	// Soft delete by setting is_deleted flag
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"is_deleted": true,
				"deleted_at": time.Now(),
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *userRepository) List(filter UserFilter) ([]*models.User, int64, error) {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build filter
	mongoFilter := bson.M{"is_deleted": false}

	if filter.Role != nil {
		mongoFilter["role"] = *filter.Role
	}
	if filter.IsActive != nil {
		mongoFilter["is_active"] = *filter.IsActive
	}
	if filter.IsVerified != nil {
		mongoFilter["is_verified"] = *filter.IsVerified
	}
	if filter.IsSuspended != nil {
		mongoFilter["is_suspended"] = *filter.IsSuspended
	}
	if filter.CreatedAfter != nil {
		mongoFilter["created_at"] = bson.M{"$gte": *filter.CreatedAfter}
	}
	if filter.CreatedBefore != nil {
		if mongoFilter["created_at"] != nil {
			mongoFilter["created_at"].(bson.M)["$lte"] = *filter.CreatedBefore
		} else {
			mongoFilter["created_at"] = bson.M{"$lte": *filter.CreatedBefore}
		}
	}
	if filter.City != "" {
		mongoFilter["profile.city"] = filter.City
	}
	if filter.Country != "" {
		mongoFilter["profile.country"] = filter.Country
	}
	if filter.Search != "" {
		mongoFilter["$or"] = []bson.M{
			{"profile.full_name": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"phone": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"email": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}

	// Count total documents
	total, err := collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count users")
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Build options
	findOptions := options.Find()

	// Pagination
	if filter.Page > 0 && filter.Limit > 0 {
		skip := (filter.Page - 1) * filter.Limit
		findOptions.SetSkip(int64(skip))
		findOptions.SetLimit(int64(filter.Limit))
	} else if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
	}

	// Sorting
	if filter.SortBy != "" {
		sortOrder := 1
		if filter.SortOrder == -1 {
			sortOrder = -1
		}
		findOptions.SetSort(bson.M{filter.SortBy: sortOrder})
	} else {
		findOptions.SetSort(bson.M{"created_at": -1}) // Default sort by creation date desc
	}

	// Find users
	cursor, err := collection.Find(ctx, mongoFilter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to find users")
		return nil, 0, fmt.Errorf("failed to find users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			r.logger.Error().Err(err).Msg("Failed to decode user")
			continue
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error().Err(err).Msg("Cursor error while listing users")
		return nil, 0, fmt.Errorf("cursor error: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) UpdateLastLogin(userID string) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	now := time.Now()
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"last_login_at": now,
				"updated_at":    now,
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update last login")
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateProfile(userID string, profile *models.UserProfile) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"profile":    profile,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user profile")
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateSettings(userID string, settings *models.UserSettings) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"settings":   settings,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user settings")
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateVerificationStatus(userID string, verification *models.UserVerification) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"verification": verification,
				"updated_at":   time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update verification status")
		return fmt.Errorf("failed to update verification status: %w", err)
	}

	return nil
}

func (r *userRepository) AddEmergencyContact(userID string, contact *models.EmergencyContact) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	contact.ID = primitive.NewObjectID()
	contact.CreatedAt = time.Now()

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$push": bson.M{
				"emergency_contacts": contact,
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add emergency contact")
		return fmt.Errorf("failed to add emergency contact: %w", err)
	}

	return nil
}

func (r *userRepository) RemoveEmergencyContact(userID, contactID string) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	contactObjectID, err := primitive.ObjectIDFromHex(contactID)
	if err != nil {
		return fmt.Errorf("invalid contact ID format")
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": userObjectID},
		bson.M{
			"$pull": bson.M{
				"emergency_contacts": bson.M{"_id": contactObjectID},
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("contact_id", contactID).Msg("Failed to remove emergency contact")
		return fmt.Errorf("failed to remove emergency contact: %w", err)
	}

	return nil
}

func (r *userRepository) AddFavoritePlace(userID string, place *models.FavoritePlace) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	place.ID = primitive.NewObjectID()
	place.CreatedAt = time.Now()
	place.UpdatedAt = time.Now()

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$push": bson.M{
				"favorite_places": place,
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add favorite place")
		return fmt.Errorf("failed to add favorite place: %w", err)
	}

	return nil
}

func (r *userRepository) RemoveFavoritePlace(userID, placeID string) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	placeObjectID, err := primitive.ObjectIDFromHex(placeID)
	if err != nil {
		return fmt.Errorf("invalid place ID format")
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": userObjectID},
		bson.M{
			"$pull": bson.M{
				"favorite_places": bson.M{"_id": placeObjectID},
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("place_id", placeID).Msg("Failed to remove favorite place")
		return fmt.Errorf("failed to remove favorite place: %w", err)
	}

	return nil
}

func (r *userRepository) GetUserStats(userID string) (*models.UserStats, error) {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	var result struct {
		Stats models.UserStats `bson:"stats"`
	}

	err = collection.FindOne(
		ctx,
		bson.M{"_id": objectID},
		options.FindOne().SetProjection(bson.M{"stats": 1}),
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user stats")
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &result.Stats, nil
}

func (r *userRepository) UpdateUserStats(userID string, stats *models.UserStats) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"stats":      stats,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user stats")
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	return nil
}

// TokenRepository Implementation

func (r *tokenRepository) SavePasswordResetToken(token *services.PasswordResetToken) error {
	collection := r.db.Collection("password_reset_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Remove any existing tokens for this user
	_, err := collection.DeleteMany(ctx, bson.M{"user_id": token.UserID})
	if err != nil {
		r.logger.Warn().Err(err).Str("user_id", token.UserID).Msg("Failed to cleanup old password reset tokens")
	}

	// Insert new token
	_, err = collection.InsertOne(ctx, token)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", token.UserID).Msg("Failed to save password reset token")
		return fmt.Errorf("failed to save password reset token: %w", err)
	}

	// Create TTL index if it doesn't exist
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	collection.Indexes().CreateOne(ctx, indexModel)

	return nil
}

func (r *tokenRepository) GetPasswordResetToken(token string) (*services.PasswordResetToken, error) {
	collection := r.db.Collection("password_reset_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resetToken services.PasswordResetToken
	err := collection.FindOne(ctx, bson.M{"token": token}).Decode(&resetToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("reset token not found")
		}
		r.logger.Error().Err(err).Str("token", utils.MaskSensitiveData(token, '*', 4)).Msg("Failed to get password reset token")
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}

	return &resetToken, nil
}

func (r *tokenRepository) DeletePasswordResetToken(token string) error {
	collection := r.db.Collection("password_reset_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"token": token})
	if err != nil {
		r.logger.Error().Err(err).Str("token", utils.MaskSensitiveData(token, '*', 4)).Msg("Failed to delete password reset token")
		return fmt.Errorf("failed to delete password reset token: %w", err)
	}

	return nil
}

func (r *tokenRepository) SaveEmailVerificationToken(token *services.EmailVerificationToken) error {
	collection := r.db.Collection("email_verification_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Remove any existing tokens for this user
	_, err := collection.DeleteMany(ctx, bson.M{"user_id": token.UserID})
	if err != nil {
		r.logger.Warn().Err(err).Str("user_id", token.UserID).Msg("Failed to cleanup old email verification tokens")
	}

	// Insert new token
	_, err = collection.InsertOne(ctx, token)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", token.UserID).Msg("Failed to save email verification token")
		return fmt.Errorf("failed to save email verification token: %w", err)
	}

	// Create TTL index if it doesn't exist
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	collection.Indexes().CreateOne(ctx, indexModel)

	return nil
}

func (r *tokenRepository) GetEmailVerificationToken(token string) (*services.EmailVerificationToken, error) {
	collection := r.db.Collection("email_verification_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var verificationToken services.EmailVerificationToken
	err := collection.FindOne(ctx, bson.M{"token": token}).Decode(&verificationToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("verification token not found")
		}
		r.logger.Error().Err(err).Str("token", utils.MaskSensitiveData(token, '*', 4)).Msg("Failed to get email verification token")
		return nil, fmt.Errorf("failed to get email verification token: %w", err)
	}

	return &verificationToken, nil
}

func (r *tokenRepository) DeleteEmailVerificationToken(token string) error {
	collection := r.db.Collection("email_verification_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"token": token})
	if err != nil {
		r.logger.Error().Err(err).Str("token", utils.MaskSensitiveData(token, '*', 4)).Msg("Failed to delete email verification token")
		return fmt.Errorf("failed to delete email verification token: %w", err)
	}

	return nil
}

func (r *tokenRepository) BlacklistToken(token string) error {
	collection := r.db.Collection("blacklisted_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blacklistEntry := bson.M{
		"token":      token,
		"created_at": time.Now(),
		"expires_at": time.Now().Add(24 * time.Hour), // Keep blacklisted for 24 hours
	}

	_, err := collection.InsertOne(ctx, blacklistEntry)
	if err != nil {
		r.logger.Error().Err(err).Str("token", utils.MaskSensitiveData(token, '*', 4)).Msg("Failed to blacklist token")
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	// Create TTL index if it doesn't exist
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	collection.Indexes().CreateOne(ctx, indexModel)

	return nil
}

func (r *tokenRepository) IsTokenBlacklisted(token string) bool {
	collection := r.db.Collection("blacklisted_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := collection.CountDocuments(ctx, bson.M{"token": token})
	if err != nil {
		r.logger.Error().Err(err).Str("token", utils.MaskSensitiveData(token, '*', 4)).Msg("Failed to check if token is blacklisted")
		return false
	}

	return count > 0
}

func (r *tokenRepository) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()

	// Cleanup expired password reset tokens
	passwordResetCollection := r.db.Collection("password_reset_tokens")
	result1, err := passwordResetCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": now}})
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup expired password reset tokens")
	} else if result1.DeletedCount > 0 {
		r.logger.Info().Int64("count", result1.DeletedCount).Msg("Cleaned up expired password reset tokens")
	}

	// Cleanup expired email verification tokens
	emailVerificationCollection := r.db.Collection("email_verification_tokens")
	result2, err := emailVerificationCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": now}})
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup expired email verification tokens")
	} else if result2.DeletedCount > 0 {
		r.logger.Info().Int64("count", result2.DeletedCount).Msg("Cleaned up expired email verification tokens")
	}

	// Cleanup expired blacklisted tokens
	blacklistedTokensCollection := r.db.Collection("blacklisted_tokens")
	result3, err := blacklistedTokensCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": now}})
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup expired blacklisted tokens")
	} else if result3.DeletedCount > 0 {
		r.logger.Info().Int64("count", result3.DeletedCount).Msg("Cleaned up expired blacklisted tokens")
	}

	return nil
}
