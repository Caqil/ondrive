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
	// Basic CRUD operations
	Create(user *models.User) error
	GetByID(userID string) (*models.User, error)
	GetByPhone(phone string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(userID string) error
	List(filter UserFilter) ([]*models.User, int64, error)

	// Profile and settings management
	UpdateLastLogin(userID string) error
	UpdateProfile(userID string, profile *models.UserProfile) error
	UpdateSettings(userID string, settings *models.UserSettings) error
	UpdateVerificationStatus(userID string, verification *models.UserVerification) error

	// Location management
	UpdateLocation(userID string, location *models.Location) error

	// Document management
	AddVerificationDocument(userID string, doc *models.VerificationDoc) error
	RemoveVerificationDocument(userID, documentID string) error

	// Emergency contacts management
	AddEmergencyContact(userID string, contact *models.EmergencyContact) error
	RemoveEmergencyContact(userID, contactID string) error
	UpdateEmergencyContacts(userID string, contacts []models.EmergencyContact) error

	// Favorite places management
	AddFavoritePlace(userID string, place *models.FavoritePlace) error
	RemoveFavoritePlace(userID, placeID string) error
	UpdateFavoritePlaces(userID string, places []models.FavoritePlace) error

	// Statistics and analytics
	GetUserStats(userID string) (*models.UserStats, error)
	UpdateUserStats(userID string, stats *models.UserStats) error

	// User management
	SoftDelete(userID, reason string) error

	// Advanced queries (optional)
	GetUsersByLocation(lat, lng, radiusKm float64, userType models.UserRole, limit int) ([]*models.User, error)
	GetActiveDriversInArea(lat, lng, radiusKm float64) ([]*models.User, error)
	UpdateDriverOnlineStatus(userID string, isOnline bool) error
	BulkUpdateUserStats(updates []models.UserStatsBulkUpdate) error
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

// Basic CRUD operations

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
		mongoFilter["profile.city"] = bson.M{"$regex": filter.City, "$options": "i"}
	}
	if filter.Country != "" {
		mongoFilter["profile.country"] = bson.M{"$regex": filter.Country, "$options": "i"}
	}
	if filter.Search != "" {
		mongoFilter["$or"] = []bson.M{
			{"profile.first_name": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"profile.last_name": bson.M{"$regex": filter.Search, "$options": "i"}},
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
	opts := options.Find()

	// Sorting
	sortBy := "created_at"
	sortOrder := -1
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	if filter.SortOrder != 0 {
		sortOrder = filter.SortOrder
	}
	opts.SetSort(bson.D{{Key: sortBy, Value: sortOrder}})

	// Pagination
	if filter.Page > 0 && filter.Limit > 0 {
		skip := (filter.Page - 1) * filter.Limit
		opts.SetSkip(int64(skip))
		opts.SetLimit(int64(filter.Limit))
	}

	// Execute query
	cursor, err := collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to list users")
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err = cursor.All(ctx, &users); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode users")
		return nil, 0, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, total, nil
}

// Profile and settings management

func (r *userRepository) UpdateLastLogin(userID string) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
				"last_login_at": &now,
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

// Location management

func (r *userRepository) UpdateLocation(userID string, location *models.Location) error {
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
				"location":   location,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user location")
		return fmt.Errorf("failed to update user location: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Msg("User location updated successfully")
	return nil
}

// Document management

func (r *userRepository) AddVerificationDocument(userID string, doc *models.VerificationDoc) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	doc.ID = primitive.NewObjectID()
	doc.UploadedAt = time.Now()

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$push": bson.M{
				"verification.documents": doc,
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add verification document")
		return fmt.Errorf("failed to add verification document: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Str("document_type", doc.Type).Msg("Verification document added successfully")
	return nil
}

func (r *userRepository) RemoveVerificationDocument(userID, documentID string) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	documentObjectID, err := primitive.ObjectIDFromHex(documentID)
	if err != nil {
		return fmt.Errorf("invalid document ID format")
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": userObjectID},
		bson.M{
			"$pull": bson.M{
				"verification.documents": bson.M{"_id": documentObjectID},
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("document_id", documentID).Msg("Failed to remove verification document")
		return fmt.Errorf("failed to remove verification document: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Str("document_id", documentID).Msg("Verification document removed successfully")
	return nil
}

// Emergency contacts management

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

func (r *userRepository) UpdateEmergencyContacts(userID string, contacts []models.EmergencyContact) error {
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
				"emergency_contacts": contacts,
				"updated_at":         time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update emergency contacts")
		return fmt.Errorf("failed to update emergency contacts: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Msg("Emergency contacts updated successfully")
	return nil
}

// Favorite places management

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

func (r *userRepository) UpdateFavoritePlaces(userID string, places []models.FavoritePlace) error {
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
				"favorite_places": places,
				"updated_at":      time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update favorite places")
		return fmt.Errorf("failed to update favorite places: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Msg("Favorite places updated successfully")
	return nil
}

// Statistics and analytics

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

// User management

func (r *userRepository) SoftDelete(userID, reason string) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
				"is_deleted":      true,
				"is_active":       false,
				"deleted_at":      &now,
				"deletion_reason": reason,
				"updated_at":      now,
			},
		},
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to soft delete user")
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Str("reason", reason).Msg("User soft deleted successfully")
	return nil
}

// Advanced queries

func (r *userRepository) GetUsersByLocation(lat, lng, radiusKm float64, userType models.UserRole, limit int) ([]*models.User, error) {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build query with geospatial search
	query := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lng, lat},
				},
				"$maxDistance": radiusKm * 1000, // Convert km to meters
			},
		},
		"is_active":  true,
		"is_deleted": false,
		"role":       userType,
	}

	opts := options.Find().SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		r.logger.Error().Err(err).Float64("lat", lat).Float64("lng", lng).Msg("Failed to find users by location")
		return nil, fmt.Errorf("failed to find users by location: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err = cursor.All(ctx, &users); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode users by location")
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

func (r *userRepository) GetActiveDriversInArea(lat, lng, radiusKm float64) ([]*models.User, error) {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query for active drivers with recent location updates
	query := bson.M{
		"role":                  models.RoleDriver,
		"is_active":             true,
		"is_deleted":            false,
		"driver_info.is_online": true,
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lng, lat},
				},
				"$maxDistance": radiusKm * 1000,
			},
		},
		"location.updated_at": bson.M{
			"$gte": time.Now().Add(-5 * time.Minute), // Location updated within last 5 minutes
		},
	}

	cursor, err := collection.Find(ctx, query)
	if err != nil {
		r.logger.Error().Err(err).Float64("lat", lat).Float64("lng", lng).Msg("Failed to find active drivers")
		return nil, fmt.Errorf("failed to find active drivers: %w", err)
	}
	defer cursor.Close(ctx)

	var drivers []*models.User
	if err = cursor.All(ctx, &drivers); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode active drivers")
		return nil, fmt.Errorf("failed to decode drivers: %w", err)
	}

	return drivers, nil
}

func (r *userRepository) UpdateDriverOnlineStatus(userID string, isOnline bool) error {
	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	update := bson.M{
		"$set": bson.M{
			"driver_info.is_online": isOnline,
			"updated_at":            time.Now(),
		},
	}

	if isOnline {
		update["$set"].(bson.M)["driver_info.last_online_at"] = time.Now()
	} else {
		update["$set"].(bson.M)["driver_info.last_offline_at"] = time.Now()
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{
			"_id":  objectID,
			"role": models.RoleDriver,
		},
		update,
	)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Bool("is_online", isOnline).Msg("Failed to update driver online status")
		return fmt.Errorf("failed to update driver online status: %w", err)
	}

	r.logger.Info().Str("user_id", userID).Bool("is_online", isOnline).Msg("Driver online status updated")
	return nil
}

func (r *userRepository) BulkUpdateUserStats(updates []models.UserStatsBulkUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	collection := r.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build bulk write operations
	var operations []mongo.WriteModel
	for _, update := range updates {
		objectID, err := primitive.ObjectIDFromHex(update.UserID)
		if err != nil {
			r.logger.Warn().Err(err).Str("user_id", update.UserID).Msg("Invalid user ID in bulk update")
			continue
		}

		updateModel := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": objectID}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"stats":      update.Stats,
					"updated_at": time.Now(),
				},
			})

		operations = append(operations, updateModel)
	}

	if len(operations) == 0 {
		return fmt.Errorf("no valid operations to perform")
	}

	// Execute bulk write
	_, err := collection.BulkWrite(ctx, operations, options.BulkWrite().SetOrdered(false))
	if err != nil {
		r.logger.Error().Err(err).Int("operations_count", len(operations)).Msg("Failed to bulk update user stats")
		return fmt.Errorf("failed to bulk update user stats: %w", err)
	}

	r.logger.Info().Int("updated_count", len(operations)).Msg("User stats bulk updated successfully")
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
		r.logger.Error().Err(err).Msg("Failed to blacklist token")
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

	var result bson.M
	err := collection.FindOne(ctx, bson.M{"token": token}).Decode(&result)
	return err == nil
}

func (r *tokenRepository) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()

	// Cleanup password reset tokens
	passwordCollection := r.db.Collection("password_reset_tokens")
	_, err := passwordCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": now}})
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup expired password reset tokens")
	}

	// Cleanup email verification tokens
	emailCollection := r.db.Collection("email_verification_tokens")
	_, err = emailCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": now}})
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup expired email verification tokens")
	}

	// Cleanup blacklisted tokens
	blacklistCollection := r.db.Collection("blacklisted_tokens")
	_, err = blacklistCollection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": now}})
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup expired blacklisted tokens")
	}

	r.logger.Info().Msg("Token cleanup completed")
	return nil
}
