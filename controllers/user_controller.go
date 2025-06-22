package controllers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"ondrive/middleware"
	"ondrive/models"
	"ondrive/repositories"
	"ondrive/services"
	"ondrive/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct {
	userRepo    repositories.UserRepository
	userService services.UserService
	logger      utils.Logger
}

func NewUserController(
	userRepo repositories.UserRepository,
	userService services.UserService,
	logger utils.Logger,
) *UserController {
	return &UserController{
		userRepo:    userRepo,
		userService: userService,
		logger:      utils.ControllerLogger("user"),
	}
}

// Profile Management

func (uc *UserController) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user profile")
		utils.NotFoundResponse(c, "User")
		return
	}

	profile := uc.sanitizeUserProfile(user)
	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", profile)
}

func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid profile update request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Validate request
	if err := uc.validateProfileRequest(req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	// Get current user
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user for profile update")
		utils.NotFoundResponse(c, "User")
		return
	}

	// Update profile fields
	updatedProfile := user.Profile
	if req.FirstName != nil {
		updatedProfile.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		updatedProfile.LastName = *req.LastName
	}
	if req.FirstName != nil || req.LastName != nil {
		updatedProfile.FullName = updatedProfile.FirstName + " " + updatedProfile.LastName
	}
	if req.DateOfBirth != nil {
		updatedProfile.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		updatedProfile.Gender = models.Gender(*req.Gender)
	}
	if req.Bio != nil {
		updatedProfile.Bio = *req.Bio
	}
	if req.Language != nil {
		updatedProfile.Language = *req.Language
	}
	if req.City != nil {
		updatedProfile.City = *req.City
	}
	if req.State != nil {
		updatedProfile.State = *req.State
	}
	if req.Country != nil {
		updatedProfile.Country = *req.Country
	}
	if req.PostalCode != nil {
		updatedProfile.PostalCode = *req.PostalCode
	}

	// Update in database
	err = uc.userRepo.UpdateProfile(userID, &updatedProfile)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user profile")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Msg("User profile updated successfully")
	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", uc.sanitizeUserProfile(&models.User{Profile: updatedProfile}))
}

func (uc *UserController) DeleteProfile(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.DeleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Verify password
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User")
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		utils.BadRequestResponse(c, "Invalid password")
		return
	}

	// Soft delete user
	err = uc.userRepo.SoftDelete(userID, req.Reason)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user profile")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("reason", req.Reason).Msg("User profile deleted")
	utils.SuccessResponse(c, http.StatusOK, "Profile deleted successfully", nil)
}

// Avatar & Documents

func (uc *UserController) UploadAvatar(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		utils.BadRequestResponse(c, "No avatar file provided")
		return
	}

	// Validate file
	if err := uc.validateImageFile(file); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Upload file to storage (implement your file storage logic)
	avatarURL, err := uc.userService.UploadAvatar(userID, file)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload avatar")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("avatar_url", avatarURL).Msg("Avatar uploaded successfully")
	utils.SuccessResponse(c, http.StatusOK, "Avatar uploaded successfully", gin.H{"avatar_url": avatarURL})
}

func (uc *UserController) DeleteAvatar(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err := uc.userService.DeleteAvatar(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete avatar")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Msg("Avatar deleted successfully")
	utils.SuccessResponse(c, http.StatusOK, "Avatar deleted successfully", nil)
}

func (uc *UserController) UploadDocument(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	documentType := c.PostForm("type")
	if documentType == "" {
		utils.BadRequestResponse(c, "Document type is required")
		return
	}

	file, err := c.FormFile("document")
	if err != nil {
		utils.BadRequestResponse(c, "No document file provided")
		return
	}

	// Validate document
	if err := uc.validateDocumentFile(file); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// Upload document
	document, err := uc.userService.UploadDocument(userID, documentType, file)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload document")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("document_type", documentType).Msg("Document uploaded successfully")
	utils.SuccessResponse(c, http.StatusOK, "Document uploaded successfully", document)
}

func (uc *UserController) GetDocuments(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	documents, err := uc.userService.GetUserDocuments(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user documents")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Documents retrieved successfully", documents)
}

func (uc *UserController) DeleteDocument(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		utils.BadRequestResponse(c, "Document ID is required")
		return
	}

	err := uc.userService.DeleteDocument(userID, documentID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Str("document_id", documentID).Msg("Failed to delete document")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("document_id", documentID).Msg("Document deleted successfully")
	utils.SuccessResponse(c, http.StatusOK, "Document deleted successfully", nil)
}

// Location Management

func (uc *UserController) UpdateLocation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid location update request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Validate coordinates
	if !utils.ValidateCoordinates(req.Latitude, req.Longitude) {
		utils.BadRequestResponse(c, "Invalid coordinates")
		return
	}

	location := &models.Location{
		Type:        "Point",
		Coordinates: []float64{req.Longitude, req.Latitude},
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		PostalCode:  req.PostalCode,
		UpdatedAt:   time.Now(),
	}

	err := uc.userService.UpdateLocation(userID, location)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user location")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Msg("User location updated successfully")
	utils.SuccessResponse(c, http.StatusOK, "Location updated successfully", location)
}

func (uc *UserController) GetLocation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user for location")
		utils.NotFoundResponse(c, "User")
		return
	}

	if user.Location == nil {
		utils.NotFoundResponse(c, "Location")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location retrieved successfully", user.Location)
}

// Settings

func (uc *UserController) GetSettings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user settings")
		utils.NotFoundResponse(c, "User")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Settings retrieved successfully", user.Settings)
}

func (uc *UserController) UpdateSettings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid settings update request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Get current settings
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User")
		return
	}

	settings := user.Settings

	// Update settings fields
	if req.Language != nil {
		settings.Language = *req.Language
	}
	if req.Currency != nil {
		settings.Currency = *req.Currency
	}
	if req.Theme != nil {
		settings.Theme = *req.Theme
	}
	if req.DistanceUnit != nil {
		settings.DistanceUnit = *req.DistanceUnit
	}

	err = uc.userRepo.UpdateSettings(userID, &settings)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user settings")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Msg("User settings updated successfully")
	utils.SuccessResponse(c, http.StatusOK, "Settings updated successfully", settings)
}

func (uc *UserController) UpdatePrivacySettings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.UpdatePrivacySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid privacy settings update request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Get current settings
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User")
		return
	}

	settings := user.Settings

	// Update privacy settings
	if req.ShowLastSeen != nil {
		settings.PrivacySettings.ShowLastSeen = *req.ShowLastSeen
	}
	if req.ShowOnlineStatus != nil {
		settings.ShowOnlineStatus = *req.ShowOnlineStatus
	}
	if req.ShowPhoneNumber != nil {
		settings.PrivacySettings.ShowLastSeen = *req.ShowPhoneNumber
	}
	if req.AllowLocationSharing != nil {
		settings.PrivacySettings.ShareLocation = *req.AllowLocationSharing
	}

	err = uc.userRepo.UpdateSettings(userID, &settings)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update privacy settings")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Msg("Privacy settings updated successfully")
	utils.SuccessResponse(c, http.StatusOK, "Privacy settings updated successfully", settings)
}

func (uc *UserController) UpdateNotificationSettings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid notification settings update request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Get current settings
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User")
		return
	}

	settings := user.Settings

	// Update notification settings
	if req.NotificationsEnabled != nil {
		settings.NotificationsEnabled = *req.NotificationsEnabled
	}
	if req.PushNotifications != nil {
		settings.PushNotifications = *req.PushNotifications
	}
	if req.EmailNotifications != nil {
		settings.EmailNotifications = *req.EmailNotifications
	}
	if req.SMSNotifications != nil {
		settings.SMSNotifications = *req.SMSNotifications
	}
	if req.RideNotifications != nil {
		settings.RideNotifications = *req.RideNotifications
	}
	if req.PromotionalNotifications != nil {
		settings.PromotionalNotifications = *req.PromotionalNotifications
	}

	err = uc.userRepo.UpdateSettings(userID, &settings)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update notification settings")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Msg("Notification settings updated successfully")
	utils.SuccessResponse(c, http.StatusOK, "Notification settings updated successfully", settings)
}

// Emergency Contacts

func (uc *UserController) GetEmergencyContacts(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	contacts, err := uc.userService.GetEmergencyContacts(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get emergency contacts")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Emergency contacts retrieved successfully", contacts)
}

func (uc *UserController) AddEmergencyContact(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.AddEmergencyContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid emergency contact request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Validate request
	if err := uc.validateEmergencyContactRequest(req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	contact := &models.EmergencyContact{
		ID:           primitive.NewObjectID(),
		Name:         req.Name,
		Phone:        req.Phone,
		Relationship: req.Relationship,
		CreatedAt:    time.Now(),
	}

	err := uc.userService.AddEmergencyContact(userID, contact)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add emergency contact")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("contact_name", req.Name).Msg("Emergency contact added successfully")
	utils.SuccessResponse(c, http.StatusCreated, "Emergency contact added successfully", contact)
}

func (uc *UserController) UpdateEmergencyContact(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	contactID := c.Param("id")
	if contactID == "" {
		utils.BadRequestResponse(c, "Contact ID is required")
		return
	}

	var req models.UpdateEmergencyContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid emergency contact update request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	updateReq := services.UpdateEmergencyContactRequest{
		Name:         req.Name,
		Phone:        req.Phone,
		Relationship: req.Relationship,
	}
	err := uc.userService.UpdateEmergencyContact(userID, contactID, updateReq)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Str("contact_id", contactID).Msg("Failed to update emergency contact")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("contact_id", contactID).Msg("Emergency contact updated successfully")
	utils.SuccessResponse(c, http.StatusOK, "Emergency contact updated successfully", nil)
}

func (uc *UserController) DeleteEmergencyContact(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	contactID := c.Param("id")
	if contactID == "" {
		utils.BadRequestResponse(c, "Contact ID is required")
		return
	}

	err := uc.userService.DeleteEmergencyContact(userID, contactID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Str("contact_id", contactID).Msg("Failed to delete emergency contact")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("contact_id", contactID).Msg("Emergency contact deleted successfully")
	utils.SuccessResponse(c, http.StatusOK, "Emergency contact deleted successfully", nil)
}

// Favorite Places

func (uc *UserController) GetFavoritePlaces(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	places, err := uc.userService.GetFavoritePlaces(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get favorite places")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Favorite places retrieved successfully", places)
}

func (uc *UserController) AddFavoritePlace(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.AddFavoritePlaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid favorite place request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Validate request
	if err := uc.validateFavoritePlaceRequest(req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": err.Error()})
		return
	}

	place := &models.FavoritePlace{
		ID:   primitive.NewObjectID(),
		Name: req.Name,
		Type: req.Type,
		Icon: req.Icon,
		Location: models.Location{
			Type:        "Point",
			Coordinates: []float64{req.Longitude, req.Latitude},
			Address:     req.Address,
			City:        req.City,
			State:       req.State,
			Country:     req.Country,
			PostalCode:  req.PostalCode,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := uc.userService.AddFavoritePlace(userID, place)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add favorite place")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("place_name", req.Name).Msg("Favorite place added successfully")
	utils.SuccessResponse(c, http.StatusCreated, "Favorite place added successfully", place)
}

func (uc *UserController) UpdateFavoritePlace(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	placeID := c.Param("id")
	if placeID == "" {
		utils.BadRequestResponse(c, "Place ID is required")
		return
	}

	var req models.UpdateFavoritePlaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn().Err(err).Str("user_id", userID).Msg("Invalid favorite place update request")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	updateReq := services.UpdateFavoritePlaceRequest{
		Name:       req.Name,
		Type:       req.Type,
		Icon:       req.Icon,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Address:    req.Address,
		City:       req.City,
		State:      req.State,
		Country:    req.Country,
		PostalCode: req.PostalCode,
	}
	err := uc.userService.UpdateFavoritePlace(userID, placeID, updateReq)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Str("place_id", placeID).Msg("Failed to update favorite place")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("place_id", placeID).Msg("Favorite place updated successfully")
	utils.SuccessResponse(c, http.StatusOK, "Favorite place updated successfully", nil)
}

func (uc *UserController) DeleteFavoritePlace(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	placeID := c.Param("id")
	if placeID == "" {
		utils.BadRequestResponse(c, "Place ID is required")
		return
	}

	err := uc.userService.DeleteFavoritePlace(userID, placeID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Str("place_id", placeID).Msg("Failed to delete favorite place")
		utils.InternalServerErrorResponse(c)
		return
	}

	uc.logger.Info().Str("user_id", userID).Str("place_id", placeID).Msg("Favorite place deleted successfully")
	utils.SuccessResponse(c, http.StatusOK, "Favorite place deleted successfully", nil)
}

// User Statistics

func (uc *UserController) GetUserStats(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	stats, err := uc.userService.GetUserStats(userID)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user stats")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User statistics retrieved successfully", stats)
}

func (uc *UserController) GetRideHistory(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	rides, total, err := uc.userService.GetRideHistory(userID, page, limit, status)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get ride history")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := &utils.Meta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  (total + int64(limit) - 1) / int64(limit),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	utils.PaginatedResponse(c, http.StatusOK, "Ride history retrieved successfully", rides, meta)
}

func (uc *UserController) GetPaymentHistory(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	paymentType := c.Query("type")

	payments, total, err := uc.userService.GetPaymentHistory(userID, page, limit, paymentType)
	if err != nil {
		uc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get payment history")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := &utils.Meta{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  (total + int64(limit) - 1) / int64(limit),
		HasNext:     int64(page*limit) < total,
		HasPrevious: page > 1,
	}

	utils.PaginatedResponse(c, http.StatusOK, "Payment history retrieved successfully", payments, meta)
}

// Helper Methods

func (uc *UserController) sanitizeUserProfile(user *models.User) gin.H {
	return gin.H{
		"id":       user.ID,
		"phone":    user.Phone,
		"email":    user.Email,
		"role":     user.Role,
		"profile":  user.Profile,
		"location": user.Location,
		"settings": user.Settings,
		"verification": gin.H{
			"phone_verified":    user.Verification.PhoneVerified,
			"email_verified":    user.Verification.EmailVerified,
			"identity_verified": user.Verification.IdentityVerified,
		},
		"stats":      user.Stats,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}

func (uc *UserController) validateProfileRequest(req models.UpdateProfileRequest) error {
	if req.FirstName != nil && (len(*req.FirstName) < 2 || len(*req.FirstName) > 50) {
		return fmt.Errorf("first name must be between 2 and 50 characters")
	}
	if req.LastName != nil && (len(*req.LastName) < 2 || len(*req.LastName) > 50) {
		return fmt.Errorf("last name must be between 2 and 50 characters")
	}
	if req.Bio != nil && len(*req.Bio) > 500 {
		return fmt.Errorf("bio cannot exceed 500 characters")
	}
	if req.Gender != nil {
		validGenders := []string{"male", "female", "other", "prefer_not_to_say"}
		if !contains(validGenders, *req.Gender) {
			return fmt.Errorf("invalid gender value")
		}
	}
	return nil
}

func (uc *UserController) validateImageFile(file *multipart.FileHeader) error {
	// Add image validation logic here
	// Check file size, format, etc.
	return nil
}

func (uc *UserController) validateDocumentFile(file *multipart.FileHeader) error {
	// Add document validation logic here
	// Check file size, format, etc.
	return nil
}

func (uc *UserController) validateEmergencyContactRequest(req models.AddEmergencyContactRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Phone == "" {
		return fmt.Errorf("phone number is required")
	}
	if req.Relationship == "" {
		return fmt.Errorf("relationship is required")
	}
	return nil
}

func (uc *UserController) validateFavoritePlaceRequest(req models.AddFavoritePlaceRequest) error {
	if req.Name == "" {
		return fmt.Errorf("place name is required")
	}
	if req.Type == "" {
		return fmt.Errorf("place type is required")
	}
	if !utils.ValidateCoordinates(req.Latitude, req.Longitude) {
		return fmt.Errorf("invalid coordinates")
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Request/Response Types
