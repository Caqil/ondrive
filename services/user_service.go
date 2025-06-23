package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService interface defines user-related business logic
type UserService interface {
	// Avatar & Documents
	UploadAvatar(userID string, file *multipart.FileHeader) (string, error)
	DeleteAvatar(userID string) error
	UploadDocument(userID, documentType string, file *multipart.FileHeader) (*models.VerificationDoc, error)
	GetUserDocuments(userID string) ([]models.VerificationDoc, error)
	DeleteDocument(userID, documentID string) error

	// Location
	UpdateLocation(userID string, location *models.Location) error

	// Emergency Contacts
	GetEmergencyContacts(userID string) ([]models.EmergencyContact, error)
	AddEmergencyContact(userID string, contact *models.EmergencyContact) error
	UpdateEmergencyContact(userID, contactID string, req UpdateEmergencyContactRequest) error
	DeleteEmergencyContact(userID, contactID string) error

	// Favorite Places
	GetFavoritePlaces(userID string) ([]models.FavoritePlace, error)
	AddFavoritePlace(userID string, place *models.FavoritePlace) error
	UpdateFavoritePlace(userID, placeID string, req UpdateFavoritePlaceRequest) error
	DeleteFavoritePlace(userID, placeID string) error

	// Statistics & History
	GetUserStats(userID string) (*models.UserStats, error)
	GetRideHistory(userID string, page, limit int, status string) ([]interface{}, int64, error)
	GetPaymentHistory(userID string, page, limit int, paymentType string) ([]interface{}, int64, error)
}

// userService implements UserService interface
type userService struct {
	userRepo    repositories.UserRepository
	fileStorage FileStorageService
	logger      utils.Logger
}

// UpdateEmergencyContactRequest for service layer
type UpdateEmergencyContactRequest struct {
	Name         *string `json:"name,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	Relationship *string `json:"relationship,omitempty"`
	Email        *string `json:"email,omitempty"`
}

// UpdateFavoritePlaceRequest for service layer
type UpdateFavoritePlaceRequest struct {
	Name       *string  `json:"name,omitempty"`
	Type       *string  `json:"type,omitempty"`
	Icon       *string  `json:"icon,omitempty"`
	Latitude   *float64 `json:"latitude,omitempty"`
	Longitude  *float64 `json:"longitude,omitempty"`
	Address    *string  `json:"address,omitempty"`
	City       *string  `json:"city,omitempty"`
	State      *string  `json:"state,omitempty"`
	Country    *string  `json:"country,omitempty"`
	PostalCode *string  `json:"postal_code,omitempty"`
}

// FileStorageService interface for file operations
type FileStorageService interface {
	UploadFile(file *multipart.FileHeader, path string) (string, error)
	DeleteFile(url string) error
	ValidateImageFile(file *multipart.FileHeader) error
	ValidateDocumentFile(file *multipart.FileHeader) error
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repositories.UserRepository,
	fileStorage FileStorageService,
	logger utils.Logger,
) UserService {
	return &userService{
		userRepo:    userRepo,
		fileStorage: fileStorage,
		logger:      utils.ServiceLogger("user"),
	}
}

// Avatar & Documents Implementation

func (s *userService) UploadAvatar(userID string, file *multipart.FileHeader) (string, error) {
	// Validate file
	if err := s.fileStorage.ValidateImageFile(file); err != nil {
		return "", err
	}

	// Generate file path
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("avatars/%s/avatar_%d%s", userID, time.Now().Unix(), ext)

	// Upload file
	avatarURL, err := s.fileStorage.UploadFile(file, filename)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload avatar file")
		return "", fmt.Errorf("failed to upload avatar")
	}

	// Get current user to get old avatar URL
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user for avatar update")
		return "", fmt.Errorf("failed to get user")
	}

	// Delete old avatar if exists
	if user.Profile.AvatarURL != "" {
		if err := s.fileStorage.DeleteFile(user.Profile.AvatarURL); err != nil {
			s.logger.Warn().Err(err).Str("old_avatar_url", user.Profile.AvatarURL).Msg("Failed to delete old avatar")
		}
	}

	// Update user profile with new avatar URL
	user.Profile.AvatarURL = avatarURL
	user.UpdatedAt = time.Now()

	err = s.userRepo.UpdateProfile(userID, &user.Profile)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user profile with avatar URL")
		// Try to delete the uploaded file
		s.fileStorage.DeleteFile(avatarURL)
		return "", fmt.Errorf("failed to update profile")
	}

	s.logger.Info().Str("user_id", userID).Str("avatar_url", avatarURL).Msg("Avatar uploaded successfully")
	return avatarURL, nil
}

func (s *userService) DeleteAvatar(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.Profile.AvatarURL == "" {
		return fmt.Errorf("no avatar to delete")
	}

	// Delete file from storage
	if err := s.fileStorage.DeleteFile(user.Profile.AvatarURL); err != nil {
		s.logger.Warn().Err(err).Str("avatar_url", user.Profile.AvatarURL).Msg("Failed to delete avatar file")
	}

	// Update user profile
	user.Profile.AvatarURL = ""
	user.UpdatedAt = time.Now()

	err = s.userRepo.UpdateProfile(userID, &user.Profile)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update profile after avatar deletion")
		return fmt.Errorf("failed to update profile")
	}

	return nil
}

func (s *userService) UploadDocument(userID, documentType string, file *multipart.FileHeader) (*models.VerificationDoc, error) {
	// Validate file
	if err := s.fileStorage.ValidateDocumentFile(file); err != nil {
		return nil, err
	}

	// Validate document type
	validTypes := []string{"id_card", "passport", "driver_license", "utility_bill", "bank_statement"}
	if !contains(validTypes, documentType) {
		return nil, fmt.Errorf("invalid document type")
	}

	// Generate file path
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("documents/%s/%s_%d%s", userID, documentType, time.Now().Unix(), ext)

	// Upload file
	documentURL, err := s.fileStorage.UploadFile(file, filename)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload document file")
		return nil, fmt.Errorf("failed to upload document")
	}

	// Create verification document
	document := &models.VerificationDoc{
		ID:           primitive.NewObjectID(),
		Type:         documentType,
		URL:          documentURL,
		Status:       models.VerificationStatus,
		UploadedAt:   time.Now(),
	}

	// Add document to user
	err = s.userRepo.AddVerificationDocument(userID, document)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add verification document")
		// Try to delete the uploaded file
		s.fileStorage.DeleteFile(documentURL)
		return nil, fmt.Errorf("failed to save document")
	}

	s.logger.Info().Str("user_id", userID).Str("document_type", documentType).Msg("Document uploaded successfully")
	return document, nil
}

func (s *userService) GetUserDocuments(userID string) ([]models.VerificationDoc, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user.Verification.Documents, nil
}

func (s *userService) DeleteDocument(userID, documentID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Find document
	var documentURL string
	for _, doc := range user.Verification.Documents {
		if doc.ID.Hex() == documentID {
			documentURL = doc.URL
			break
		}
	}

	if documentURL == "" {
		return fmt.Errorf("document not found")
	}

	// Delete file from storage
	if err := s.fileStorage.DeleteFile(documentURL); err != nil {
		s.logger.Warn().Err(err).Str("document_url", documentURL).Msg("Failed to delete document file")
	}

	// Remove document from user
	err = s.userRepo.RemoveVerificationDocument(userID, documentID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Str("document_id", documentID).Msg("Failed to remove document")
		return fmt.Errorf("failed to delete document")
	}

	return nil
}

// Location Implementation

func (s *userService) UpdateLocation(userID string, location *models.Location) error {
	err := s.userRepo.UpdateLocation(userID, location)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user location")
		return fmt.Errorf("failed to update location")
	}

	return nil
}

// Emergency Contacts Implementation

func (s *userService) GetEmergencyContacts(userID string) ([]models.EmergencyContact, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user.EmergencyContacts, nil
}

func (s *userService) AddEmergencyContact(userID string, contact *models.EmergencyContact) error {
	// Validate maximum contacts
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if len(user.EmergencyContacts) >= 5 {
		return fmt.Errorf("maximum of 5 emergency contacts allowed")
	}

	err = s.userRepo.AddEmergencyContact(userID, contact)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add emergency contact")
		return fmt.Errorf("failed to add emergency contact")
	}

	return nil
}

func (s *userService) UpdateEmergencyContact(userID, contactID string, req UpdateEmergencyContactRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Find and update contact
	var updated bool
	for i, contact := range user.EmergencyContacts {
		if contact.ID.Hex() == contactID {
			if req.Name != nil {
				user.EmergencyContacts[i].Name = *req.Name
			}
			if req.Phone != nil {
				user.EmergencyContacts[i].Phone = *req.Phone
			}
			if req.Relationship != nil {
				user.EmergencyContacts[i].Relationship = *req.Relationship
			}
			if req.Email != nil {
				user.EmergencyContacts[i].Email = *req.Email
			}
			break
		}
	}

	if !updated {
		return fmt.Errorf("emergency contact not found")
	}

	err = s.userRepo.UpdateEmergencyContacts(userID, user.EmergencyContacts)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update emergency contact")
		return fmt.Errorf("failed to update emergency contact")
	}

	return nil
}

func (s *userService) DeleteEmergencyContact(userID, contactID string) error {
	err := s.userRepo.RemoveEmergencyContact(userID, contactID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Str("contact_id", contactID).Msg("Failed to delete emergency contact")
		return fmt.Errorf("failed to delete emergency contact")
	}

	return nil
}

// Favorite Places Implementation

func (s *userService) GetFavoritePlaces(userID string) ([]models.FavoritePlace, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user.FavoritePlaces, nil
}

func (s *userService) AddFavoritePlace(userID string, place *models.FavoritePlace) error {
	// Validate maximum places
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if len(user.FavoritePlaces) >= 10 {
		return fmt.Errorf("maximum of 10 favorite places allowed")
	}

	err = s.userRepo.AddFavoritePlace(userID, place)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add favorite place")
		return fmt.Errorf("failed to add favorite place")
	}

	return nil
}

func (s *userService) UpdateFavoritePlace(userID, placeID string, req UpdateFavoritePlaceRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Find and update place
	var updated bool
	for i, place := range user.FavoritePlaces {
		if place.ID.Hex() == placeID {
			if req.Name != nil {
				user.FavoritePlaces[i].Name = *req.Name
			}
			if req.Type != nil {
				user.FavoritePlaces[i].Type = *req.Type
			}
			if req.Icon != nil {
				user.FavoritePlaces[i].Icon = *req.Icon
			}
			if req.Latitude != nil && req.Longitude != nil {
				user.FavoritePlaces[i].Location.Coordinates = []float64{*req.Longitude, *req.Latitude}
			}
			if req.Address != nil {
				user.FavoritePlaces[i].Location.Address = *req.Address
			}
			if req.City != nil {
				user.FavoritePlaces[i].Location.City = *req.City
			}
			if req.State != nil {
				user.FavoritePlaces[i].Location.State = *req.State
			}
			if req.Country != nil {
				user.FavoritePlaces[i].Location.Country = *req.Country
			}
			if req.PostalCode != nil {
				user.FavoritePlaces[i].Location.PostalCode = *req.PostalCode
			}
			user.FavoritePlaces[i].UpdatedAt = time.Now()
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("favorite place not found")
	}

	err = s.userRepo.UpdateFavoritePlaces(userID, user.FavoritePlaces)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update favorite place")
		return fmt.Errorf("failed to update favorite place")
	}

	return nil
}

func (s *userService) DeleteFavoritePlace(userID, placeID string) error {
	err := s.userRepo.RemoveFavoritePlace(userID, placeID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Str("place_id", placeID).Msg("Failed to delete favorite place")
		return fmt.Errorf("failed to delete favorite place")
	}

	return nil
}

// Statistics & History Implementation

func (s *userService) GetUserStats(userID string) (*models.UserStats, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Calculate additional stats if needed
	stats := user.Stats

	// You can add more complex stat calculations here
	// For example, calculate monthly stats, efficiency metrics, etc.

	return &stats, nil
}

func (s *userService) GetRideHistory(userID string, page, limit int, status string) ([]interface{}, int64, error) {
	// This should interact with a RideRepository
	// For now, return placeholder data
	s.logger.Warn().Str("user_id", userID).Msg("GetRideHistory not fully implemented - requires RideRepository")

	// TODO: Implement with actual ride repository
	// rides, total, err := s.rideRepo.GetUserRides(userID, page, limit, status)

	return []interface{}{}, 0, nil
}

func (s *userService) GetPaymentHistory(userID string, page, limit int, paymentType string) ([]interface{}, int64, error) {
	// This should interact with a PaymentRepository
	// For now, return placeholder data
	s.logger.Warn().Str("user_id", userID).Msg("GetPaymentHistory not fully implemented - requires PaymentRepository")

	// TODO: Implement with actual payment repository
	// payments, total, err := s.paymentRepo.GetUserPayments(userID, page, limit, paymentType)

	return []interface{}{}, 0, nil
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Simple file storage implementation (you should replace this with your actual storage implementation)
type simpleFileStorage struct {
	logger utils.Logger
}

func NewFileStorageService(logger utils.Logger) FileStorageService {
	return &simpleFileStorage{
		logger: logger,
	}
}

func (fs *simpleFileStorage) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	// TODO: Implement actual file upload logic
	// This is a placeholder implementation
	return fmt.Sprintf("https://storage.example.com/%s", path), nil
}

func (fs *simpleFileStorage) DeleteFile(url string) error {
	// TODO: Implement actual file deletion logic
	return nil
}

func (fs *simpleFileStorage) ValidateImageFile(file *multipart.FileHeader) error {
	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return errors.New("file size too large (max 5MB)")
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	if !contains(validExts, ext) {
		return errors.New("invalid file type (allowed: jpg, jpeg, png, gif, webp)")
	}

	return nil
}

func (fs *simpleFileStorage) ValidateDocumentFile(file *multipart.FileHeader) error {
	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return errors.New("file size too large (max 10MB)")
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExts := []string{".jpg", ".jpeg", ".png", ".pdf", ".doc", ".docx"}

	if !contains(validExts, ext) {
		return errors.New("invalid file type (allowed: jpg, jpeg, png, pdf, doc, docx)")
	}

	return nil
}
