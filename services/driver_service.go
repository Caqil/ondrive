package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DriverService interface {
	// Profile Management
	UpdateDriverProfile(userID string, req interface{}) (*models.DriverInfo, error)
	GetDriverDocuments(userID string) ([]models.VerificationDoc, error)
	UploadDriverDocuments(userID, documentType string, files []*multipart.FileHeader) ([]*models.VerificationDoc, error)

	// Vehicle Management
	AddVehicle(userID string, req interface{}) (*models.VehicleDetails, error)
	UpdateVehicle(userID string, req interface{}) (*models.VehicleDetails, error)
	GetDriverVehicle(userID string) (*models.VehicleDetails, error)
	UploadVehiclePhotos(userID, photoType string, files []*multipart.FileHeader) ([]*models.VehiclePhoto, error)
	DeleteVehiclePhoto(userID, photoID string) error

	// Status & Availability
	UpdateOnlineStatus(userID string, isOnline bool) error
	GetDriverStatus(userID string) (*DriverStatusResponse, error)
	UpdateAvailability(userID string, req interface{}) error

	// Ride Management
	GetAvailableRides(userID string, page, limit int) ([]interface{}, int64, error)
	GetCurrentRide(userID string) (interface{}, error)
	GetRideRequests(userID string, page, limit int) ([]interface{}, int64, error)
	AcceptRideRequest(userID, rideID string) (interface{}, error)
	DeclineRideRequest(userID, rideID string) error

	// Location & Navigation
	UpdateDriverLocation(userID string, location *models.Location) error
	GetDriverLocation(userID string) (*models.Location, error)
	GetNavigation(userID, rideID string) (*NavigationResponse, error)

	// Earnings & Finance
	GetEarnings(userID string, page, limit int) ([]*models.Earning, int64, error)
	GetDailyEarnings(userID, date string) (*EarningsResponse, error)
	GetWeeklyEarnings(userID, week string) (*EarningsResponse, error)
	GetMonthlyEarnings(userID, month string) (*EarningsResponse, error)
	GetPayouts(userID string, page, limit int) ([]*models.Payout, int64, error)
	RequestPayout(userID string, amount float64, method string) (*models.Payout, error)

	// Statistics
	GetDriverStats(userID string) (*models.DriverStats, error)
	GetPerformanceMetrics(userID, period string) (*models.PerformanceMetrics, error)
	GetRatingSummary(userID string) (*models.RatingSummary, error)

	// Working Hours & Schedule
	GetWorkingHours(userID string) (*models.WorkingHours, error)
	UpdateWorkingHours(userID string, workingHours *models.WorkingHours) error
	GetDriverSchedule(userID, date string) (*models.DriverSchedule, error)
	TakeBreak(userID, duration, reason string) error
	ResumeWork(userID string) error

	// Preferences
	GetDriverPreferences(userID string) (*models.DriverPreferences, error)
	UpdateDriverPreferences(userID string, req interface{}) (*models.DriverPreferences, error)
	UpdateServiceAreas(userID string, serviceAreas []models.ServiceArea) error
	UpdateServiceTypes(userID string, serviceTypes []string) error

	// Verification & Documents
	UploadDriverLicense(userID string, file *multipart.FileHeader, licenseNumber, expiryDate string) (*models.VerificationDoc, error)
	UploadInsurance(userID string, file *multipart.FileHeader, policyNumber, expiryDate string) (*models.VerificationDoc, error)
	UploadVehicleRegistration(userID string, file *multipart.FileHeader, registrationNumber, expiryDate string) (*models.VerificationDoc, error)
	GetVerificationStatus(userID string) (*models.VerificationStatus, error)

	// Support
	CreateSupportTicket(userID string, req interface{}) (*models.SupportTicket, error)
	GetSupportTickets(userID string, page, limit int) ([]*models.SupportTicket, int64, error)
	UpdateSupportTicket(userID, ticketID, message string) (*models.SupportTicket, error)

	// Utility Methods
	ValidateDriverProfile(profile interface{}) error
	ValidateVehicleDetails(vehicle interface{}) error
	CalculateEarnings(userID string, period string) (*EarningsCalculation, error)
	GetDriverMetrics(userID string) (*DriverMetrics, error)
}

// Response structures
type DriverStatusResponse struct {
	IsOnline      bool             `json:"is_online"`
	IsAvailable   bool             `json:"is_available"`
	Status        string           `json:"status"`
	OnlineSince   *time.Time       `json:"online_since,omitempty"`
	OfflineSince  *time.Time       `json:"offline_since,omitempty"`
	CurrentRideID *string          `json:"current_ride_id,omitempty"`
	Location      *models.Location `json:"location,omitempty"`
}

type NavigationResponse struct {
	DestinationAddress string           `json:"destination_address"`
	Distance           float64          `json:"distance"`
	Duration           int              `json:"duration"`
	Steps              []string         `json:"steps"`
	RoutePolyline      string           `json:"route_polyline"`
	CurrentLocation    *models.Location `json:"current_location"`
}

type EarningsResponse struct {
	Period         string  `json:"period"`
	TotalAmount    float64 `json:"total_amount"`
	TotalRides     int     `json:"total_rides"`
	AveragePerRide float64 `json:"average_per_ride"`
	Tips           float64 `json:"tips"`
	Bonuses        float64 `json:"bonuses"`
	Deductions     float64 `json:"deductions"`
	NetEarnings    float64 `json:"net_earnings"`
}

type EarningsCalculation struct {
	GrossEarnings float64 `json:"gross_earnings"`
	Commission    float64 `json:"commission"`
	NetEarnings   float64 `json:"net_earnings"`
	Tips          float64 `json:"tips"`
	Bonuses       float64 `json:"bonuses"`
	Deductions    float64 `json:"deductions"`
}

type DriverMetrics struct {
	TotalRides       int     `json:"total_rides"`
	AcceptanceRate   float64 `json:"acceptance_rate"`
	CancellationRate float64 `json:"cancellation_rate"`
	AverageRating    float64 `json:"average_rating"`
	OnlineHoursToday float64 `json:"online_hours_today"`
	OnlineHoursWeek  float64 `json:"online_hours_week"`
	EarningsToday    float64 `json:"earnings_today"`
	EarningsWeek     float64 `json:"earnings_week"`
}

// Request structures (interfaces for flexibility)
type UpdateDriverProfileRequest struct {
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	Bio          string   `json:"bio"`
	ProfilePhoto string   `json:"profile_photo"`
	DateOfBirth  string   `json:"date_of_birth"`
	Gender       string   `json:"gender"`
	Languages    []string `json:"languages"`
}

type AddVehicleRequest struct {
	Make         string   `json:"make"`
	Model        string   `json:"model"`
	Year         int      `json:"year"`
	Color        string   `json:"color"`
	LicensePlate string   `json:"license_plate"`
	VIN          string   `json:"vin"`
	FuelType     string   `json:"fuel_type"`
	Transmission string   `json:"transmission"`
	Seats        int      `json:"seats"`
	Features     []string `json:"features"`
	Amenities    []string `json:"amenities"`
}

type UpdateAvailabilityRequest struct {
	IsAvailable  bool                `json:"is_available"`
	MaxDistance  int                 `json:"max_distance"`
	ServiceTypes []string            `json:"service_types"`
	WorkingHours models.WorkingHours `json:"working_hours"`
}

type UpdateDriverPreferencesRequest struct {
	MaxDistance          int                         `json:"max_distance"`
	ServiceTypes         []string                    `json:"service_types"`
	AutoAcceptRides      bool                        `json:"auto_accept_rides"`
	AcceptCashPayments   bool                        `json:"accept_cash_payments"`
	AcceptCardPayments   bool                        `json:"accept_card_payments"`
	NotificationSettings models.NotificationSettings `json:"notification_settings"`
}

type CreateSupportTicketRequest struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

// Service implementation
type driverService struct {
	driverRepo          repositories.DriverRepository
	userRepo            repositories.UserRepository
	fileUploadService   FileStorageService
	notificationService NotificationService
	locationService     LocationService
	rideService         RideService
	logger              utils.Logger
}

func NewDriverService(
	driverRepo repositories.DriverRepository,
	userRepo repositories.UserRepository,
	fileUploadService FileStorageService,
	notificationService NotificationService,
	locationService LocationService,
	rideService RideService,
	logger utils.Logger,
) DriverService {
	return &driverService{
		driverRepo:          driverRepo,
		userRepo:            userRepo,
		fileUploadService:   fileUploadService,
		notificationService: notificationService,
		locationService:     locationService,
		rideService:         rideService,
		logger:              logger,
	}
}

// Profile Management

func (s *driverService) UpdateDriverProfile(userID string, reqInterface interface{}) (*models.DriverInfo, error) {
	req, ok := reqInterface.(*UpdateDriverProfileRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	if err := s.ValidateDriverProfile(req); err != nil {
		return nil, err
	}

	driver, err := s.driverRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}

	// Update driver fields
	if req.FirstName != "" {
		driver.FirstName = req.FirstName
	}
	if req.LastName != "" {
		driver.LastName = req.LastName
	}
	if req.Bio != "" {
		driver.Bio = req.Bio
	}
	if req.ProfilePhoto != "" {
		driver.ProfilePhoto = req.ProfilePhoto
	}
	if req.Gender != "" {
		driver.Gender = req.Gender
	}
	if len(req.Languages) > 0 {
		driver.Languages = req.Languages
	}

	// Parse and update date of birth
	if req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err == nil {
			driver.DateOfBirth = &dob
		}
	}

	driver.LastActive = time.Now()

	return s.driverRepo.Update(driver)
}

func (s *driverService) GetDriverDocuments(userID string) ([]models.VerificationDoc, error) {
	return s.driverRepo.GetDocuments(userID)
}

func (s *driverService) UploadDriverDocuments(userID, documentType string, files []*multipart.FileHeader) ([]*models.VerificationDoc, error) {
	var uploadedDocs []*models.VerificationDoc

	for _, file := range files {
		// Validate file
		if err := s.validateDocumentFile(file); err != nil {
			return nil, err
		}

		// Upload file
		fileURL, err := s.fileUploadService.UploadFile(file, "documents/drivers/"+userID)
		if err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload document file")
			return nil, fmt.Errorf("failed to upload document: %w", err)
		}

		// Create document record
		doc := &models.VerificationDoc{
			Type:       documentType,
			URL:        fileURL,
			FileName:   file.Filename,
			FileSize:   file.Size,
			Status:     "pending",
			UploadedAt: time.Now(),
		}

		// Add to driver
		if err := s.driverRepo.AddDocument(userID, doc); err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add document to driver")
			return nil, fmt.Errorf("failed to save document: %w", err)
		}

		uploadedDocs = append(uploadedDocs, doc)
	}

	// Trigger document verification process
	go s.triggerDocumentVerification(userID, uploadedDocs)

	return uploadedDocs, nil
}

// Vehicle Management

func (s *driverService) AddVehicle(userID string, reqInterface interface{}) (*models.VehicleDetails, error) {
	req, ok := reqInterface.(*AddVehicleRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	if err := s.ValidateVehicleDetails(req); err != nil {
		return nil, err
	}

	vehicle := &models.VehicleDetails{
		Make:            req.Make,
		Model:           req.Model,
		Year:            req.Year,
		Color:           req.Color,
		LicensePlate:    req.LicensePlate,
		VIN:             req.VIN,
		FuelType:        req.FuelType,
		Transmission:    req.Transmission,
		Seats:           req.Seats,
		AirConditioning: true, // Default
		Features:        req.Features,
		Amenities:       req.Amenities,
		IsVerified:      false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.driverRepo.AddVehicle(userID, vehicle); err != nil {
		return nil, fmt.Errorf("failed to add vehicle: %w", err)
	}

	// Trigger vehicle verification process
	go s.triggerVehicleVerification(userID, vehicle)

	return vehicle, nil
}

func (s *driverService) UpdateVehicle(userID string, reqInterface interface{}) (*models.VehicleDetails, error) {
	req, ok := reqInterface.(*AddVehicleRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	if err := s.ValidateVehicleDetails(req); err != nil {
		return nil, err
	}

	// Get existing vehicle
	existingVehicle, err := s.driverRepo.GetDriverVehicle(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing vehicle: %w", err)
	}

	// Update fields
	vehicle := &models.VehicleDetails{
		Make:            req.Make,
		Model:           req.Model,
		Year:            req.Year,
		Color:           req.Color,
		LicensePlate:    req.LicensePlate,
		VIN:             req.VIN,
		FuelType:        req.FuelType,
		Transmission:    req.Transmission,
		Seats:           req.Seats,
		AirConditioning: existingVehicle.AirConditioning,
		Features:        req.Features,
		Amenities:       req.Amenities,
		Photos:          existingVehicle.Photos, // Keep existing photos
		IsVerified:      existingVehicle.IsVerified,
		VerifiedAt:      existingVehicle.VerifiedAt,
		CreatedAt:       existingVehicle.CreatedAt,
		UpdatedAt:       time.Now(),
	}

	if err := s.driverRepo.UpdateVehicle(userID, vehicle); err != nil {
		return nil, fmt.Errorf("failed to update vehicle: %w", err)
	}

	return vehicle, nil
}

func (s *driverService) GetDriverVehicle(userID string) (*models.VehicleDetails, error) {
	return s.driverRepo.GetDriverVehicle(userID)
}

func (s *driverService) UploadVehiclePhotos(userID, photoType string, files []*multipart.FileHeader) ([]*models.VehiclePhoto, error) {
	var uploadedPhotos []*models.VehiclePhoto

	for _, file := range files {
		// Validate file
		if err := s.validateImageFile(file); err != nil {
			return nil, err
		}

		// Upload file
		fileURL, err := s.fileUploadService.UploadFile(file, "vehicles/"+userID)
		if err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload vehicle photo")
			return nil, fmt.Errorf("failed to upload photo: %w", err)
		}

		// Create photo record
		photo := &models.VehiclePhoto{
			URL:         fileURL,
			Type:        photoType,
			Description: fmt.Sprintf("%s photo", photoType),
			UploadedAt:  time.Now(),
		}

		// Add to vehicle
		if err := s.driverRepo.AddVehiclePhoto(userID, photo); err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add vehicle photo")
			return nil, fmt.Errorf("failed to save photo: %w", err)
		}

		uploadedPhotos = append(uploadedPhotos, photo)
	}

	return uploadedPhotos, nil
}

func (s *driverService) DeleteVehiclePhoto(userID, photoID string) error {
	return s.driverRepo.DeleteVehiclePhoto(userID, photoID)
}

// Status & Availability

func (s *driverService) UpdateOnlineStatus(userID string, isOnline bool) error {
	err := s.driverRepo.UpdateOnlineStatus(userID, isOnline)
	if err != nil {
		return err
	}

	// Update user repo as well for consistency
	if err := s.userRepo.UpdateDriverOnlineStatus(userID, isOnline); err != nil {
		s.logger.Warn().Err(err).Str("user_id", userID).Msg("Failed to update user online status")
	}

	// Send notification to relevant parties
	go s.notifyStatusChange(userID, isOnline)

	return nil
}

func (s *driverService) GetDriverStatus(userID string) (*DriverStatusResponse, error) {
	driver, err := s.driverRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}

	location, _ := s.driverRepo.GetDriverLocation(userID)

	status := &DriverStatusResponse{
		IsOnline:    driver.IsOnline,
		IsAvailable: driver.IsAvailable,
		Status:      driver.Status,
		Location:    location,
	}

	if driver.IsOnline && driver.OnlineSince != nil {
		status.OnlineSince = driver.OnlineSince
	}

	// Get current ride if any
	if currentRide, err := s.rideService.GetCurrentRideForDriver(userID); err == nil && currentRide != nil {
		rideID := currentRide.ID.Hex()
		status.CurrentRideID = &rideID
	}

	return status, nil
}

func (s *driverService) UpdateAvailability(userID string, reqInterface interface{}) error {
	req, ok := reqInterface.(*UpdateAvailabilityRequest)
	if !ok {
		return errors.New("invalid request type")
	}

	return s.driverRepo.UpdateAvailability(userID, req.IsAvailable, req.MaxDistance, req.ServiceTypes)
}

// Ride Management

func (s *driverService) GetAvailableRides(userID string, page, limit int) ([]interface{}, int64, error) {
	// Get driver's location and preferences
	driver, err := s.driverRepo.GetByUserID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get driver: %w", err)
	}

	location, err := s.driverRepo.GetDriverLocation(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get driver location: %w", err)
	}

	// Get available rides based on driver's criteria
	rides, total, err := s.rideService.GetAvailableRidesForDriver(userID, driver.ServiceTypes, location, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get available rides: %w", err)
	}

	// Convert to interface{} slice
	result := make([]interface{}, len(rides))
	for i, ride := range rides {
		result[i] = ride
	}

	return result, total, nil
}

func (s *driverService) GetCurrentRide(userID string) (interface{}, error) {
	return s.rideService.GetCurrentRideForDriver(userID)
}

func (s *driverService) GetRideRequests(userID string, page, limit int) ([]interface{}, int64, error) {
	requests, total, err := s.rideService.GetRideRequestsForDriver(userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ride requests: %w", err)
	}

	// Convert to interface{} slice
	result := make([]interface{}, len(requests))
	for i, request := range requests {
		result[i] = request
	}

	return result, total, nil
}

func (s *driverService) AcceptRideRequest(userID, rideID string) (interface{}, error) {
	// Check if driver is available
	driver, err := s.driverRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}

	if !driver.IsOnline || !driver.IsAvailable {
		return nil, errors.New("driver is not available")
	}

	// Accept the ride
	ride, err := s.rideService.AcceptRide(rideID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to accept ride: %w", err)
	}

	// Update driver availability
	s.driverRepo.UpdateAvailability(userID, false, 0, nil)

	return ride, nil
}

func (s *driverService) DeclineRideRequest(userID, rideID string) error {
	return s.rideService.DeclineRide(rideID, userID)
}

// Location & Navigation

func (s *driverService) UpdateDriverLocation(userID string, location *models.Location) error {
	err := s.driverRepo.UpdateLocation(userID, location)
	if err != nil {
		return err
	}

	// Update user location as well
	if err := s.userRepo.UpdateLocation(userID, location); err != nil {
		s.logger.Warn().Err(err).Str("user_id", userID).Msg("Failed to update user location")
	}

	// Broadcast location to relevant parties (current passengers, dispatch, etc.)
	go s.broadcastLocationUpdate(userID, location)

	return nil
}

func (s *driverService) GetDriverLocation(userID string) (*models.Location, error) {
	return s.driverRepo.GetDriverLocation(userID)
}

func (s *driverService) GetNavigation(userID, rideID string) (*NavigationResponse, error) {
	// Get current ride details
	ride, err := s.rideService.GetRideByID(rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride: %w", err)
	}

	// Get driver's current location
	location, err := s.driverRepo.GetDriverLocation(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver location: %w", err)
	}

	// Calculate navigation
	navigation, err := s.locationService.CalculateRoute(
		location.Latitude, location.Longitude,
		ride.Destination.Latitude, ride.Destination.Longitude,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate navigation: %w", err)
	}

	return &NavigationResponse{
		DestinationAddress: ride.Destination.Address,
		Distance:           navigation.Distance,
		Duration:           navigation.Duration,
		Steps:              navigation.Steps,
		RoutePolyline:      navigation.Polyline,
		CurrentLocation:    location,
	}, nil
}

// Earnings & Finance

func (s *driverService) GetEarnings(userID string, page, limit int) ([]*models.Earning, int64, error) {
	return s.driverRepo.GetEarnings(userID, page, limit)
}

func (s *driverService) GetDailyEarnings(userID, date string) (*EarningsResponse, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	earnings, err := s.driverRepo.GetEarningsByPeriod(userID, "daily", startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily earnings: %w", err)
	}

	return s.calculateEarningsResponse(earnings, "daily"), nil
}

func (s *driverService) GetWeeklyEarnings(userID, week string) (*EarningsResponse, error) {
	// Parse week number or use current week
	var startOfWeek time.Time
	if week == "" {
		now := time.Now()
		startOfWeek = now.AddDate(0, 0, -int(now.Weekday()))
	} else {
		// Parse week format (could be week number or date)
		weekNum, err := strconv.Atoi(week)
		if err != nil {
			return nil, fmt.Errorf("invalid week format: %w", err)
		}
		// Calculate start of week based on week number
		year := time.Now().Year()
		startOfWeek = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, (weekNum-1)*7)
	}

	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	earnings, err := s.driverRepo.GetEarningsByPeriod(userID, "weekly", startOfWeek, endOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly earnings: %w", err)
	}

	return s.calculateEarningsResponse(earnings, "weekly"), nil
}

func (s *driverService) GetMonthlyEarnings(userID, month string) (*EarningsResponse, error) {
	var startOfMonth time.Time
	if month == "" {
		now := time.Now()
		startOfMonth = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return nil, fmt.Errorf("invalid month format: %w", err)
		}
		startOfMonth = time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, parsedMonth.Location())
	}

	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	earnings, err := s.driverRepo.GetEarningsByPeriod(userID, "monthly", startOfMonth, endOfMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly earnings: %w", err)
	}

	return s.calculateEarningsResponse(earnings, "monthly"), nil
}

func (s *driverService) GetPayouts(userID string, page, limit int) ([]*models.Payout, int64, error) {
	return s.driverRepo.GetPayouts(userID, page, limit)
}

func (s *driverService) RequestPayout(userID string, amount float64, method string) (*models.Payout, error) {
	// Validate payout request
	if amount <= 0 {
		return nil, errors.New("payout amount must be positive")
	}

	// Check available balance
	stats, err := s.driverRepo.GetDriverStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver stats: %w", err)
	}

	availableBalance := stats.TotalEarnings - stats.TotalWithdrawn
	if amount > availableBalance {
		return nil, errors.New("insufficient balance for payout")
	}

	// Create payout request
	payout := &models.Payout{
		ID:            primitive.NewObjectID(),
		UserID:        primitive.ObjectID{}, // Would be set from userID
		Amount:        amount,
		Method:        method,
		Status:        "pending",
		RequestedAt:   time.Now(),
		ProcessingFee: amount * 0.02, // 2% processing fee
		NetAmount:     amount * 0.98,
	}

	if err := s.driverRepo.AddPayout(userID, payout); err != nil {
		return nil, fmt.Errorf("failed to create payout request: %w", err)
	}

	// Trigger payout processing
	go s.processPayout(payout)

	return payout, nil
}

// Statistics

func (s *driverService) GetDriverStats(userID string) (*models.DriverStats, error) {
	return s.driverRepo.GetDriverStats(userID)
}

func (s *driverService) GetPerformanceMetrics(userID, period string) (*models.PerformanceMetrics, error) {
	return s.driverRepo.GetPerformanceMetrics(userID, period)
}

func (s *driverService) GetRatingSummary(userID string) (*models.RatingSummary, error) {
	return s.driverRepo.GetRatingSummary(userID)
}

// Working Hours & Schedule

func (s *driverService) GetWorkingHours(userID string) (*models.WorkingHours, error) {
	return s.driverRepo.GetWorkingHours(userID)
}

func (s *driverService) UpdateWorkingHours(userID string, workingHours *models.WorkingHours) error {
	return s.driverRepo.UpdateWorkingHours(userID, workingHours)
}

func (s *driverService) GetDriverSchedule(userID, date string) (*models.DriverSchedule, error) {
	return s.driverRepo.GetDriverSchedule(userID, date)
}

func (s *driverService) TakeBreak(userID, duration, reason string) error {
	// Parse duration
	breakDuration, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	breakInfo := &models.BreakInfo{
		StartTime: time.Now(),
		Duration:  breakDuration,
		Reason:    reason,
		IsActive:  true,
	}

	return s.driverRepo.AddBreak(userID, breakInfo)
}

func (s *driverService) ResumeWork(userID string) error {
	return s.driverRepo.EndBreak(userID)
}

// Preferences

func (s *driverService) GetDriverPreferences(userID string) (*models.DriverPreferences, error) {
	return s.driverRepo.GetDriverPreferences(userID)
}

func (s *driverService) UpdateDriverPreferences(userID string, reqInterface interface{}) (*models.DriverPreferences, error) {
	req, ok := reqInterface.(*UpdateDriverPreferencesRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	preferences := &models.DriverPreferences{
		MaxDistance:          req.MaxDistance,
		ServiceTypes:         req.ServiceTypes,
		AutoAcceptRides:      req.AutoAcceptRides,
		AcceptCashPayments:   req.AcceptCashPayments,
		AcceptCardPayments:   req.AcceptCardPayments,
		NotificationSettings: req.NotificationSettings,
		UpdatedAt:            time.Now(),
	}

	if err := s.driverRepo.UpdateDriverPreferences(userID, preferences); err != nil {
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}

	return preferences, nil
}

func (s *driverService) UpdateServiceAreas(userID string, serviceAreas []models.ServiceArea) error {
	return s.driverRepo.UpdateServiceAreas(userID, serviceAreas)
}

func (s *driverService) UpdateServiceTypes(userID string, serviceTypes []string) error {
	return s.driverRepo.UpdateServiceTypes(userID, serviceTypes)
}

// Verification & Documents

func (s *driverService) UploadDriverLicense(userID string, file *multipart.FileHeader, licenseNumber, expiryDate string) (*models.VerificationDoc, error) {
	if err := s.validateDocumentFile(file); err != nil {
		return nil, err
	}

	// Upload file
	fileURL, err := s.fileUploadService.UploadFile(file, "documents/licenses/"+userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload license: %w", err)
	}

	// Parse expiry date
	var expiryTime *time.Time
	if expiryDate != "" {
		parsed, err := time.Parse("2006-01-02", expiryDate)
		if err == nil {
			expiryTime = &parsed
		}
	}

	doc := &models.VerificationDoc{
		Type:       "driver_license",
		URL:        fileURL,
		FileName:   file.Filename,
		FileSize:   file.Size,
		Number:     licenseNumber,
		ExpiryDate: expiryTime,
		Status:     "pending",
		UploadedAt: time.Now(),
	}

	if err := s.driverRepo.AddDocument(userID, doc); err != nil {
		return nil, fmt.Errorf("failed to save license document: %w", err)
	}

	return doc, nil
}

func (s *driverService) UploadInsurance(userID string, file *multipart.FileHeader, policyNumber, expiryDate string) (*models.VerificationDoc, error) {
	if err := s.validateDocumentFile(file); err != nil {
		return nil, err
	}

	fileURL, err := s.fileUploadService.UploadFile(file, "documents/insurance/"+userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload insurance: %w", err)
	}

	var expiryTime *time.Time
	if expiryDate != "" {
		parsed, err := time.Parse("2006-01-02", expiryDate)
		if err == nil {
			expiryTime = &parsed
		}
	}

	doc := &models.VerificationDoc{
		Type:       "insurance",
		URL:        fileURL,
		FileName:   file.Filename,
		FileSize:   file.Size,
		Number:     policyNumber,
		ExpiryDate: expiryTime,
		Status:     "pending",
		UploadedAt: time.Now(),
	}

	if err := s.driverRepo.AddDocument(userID, doc); err != nil {
		return nil, fmt.Errorf("failed to save insurance document: %w", err)
	}

	return doc, nil
}

func (s *driverService) UploadVehicleRegistration(userID string, file *multipart.FileHeader, registrationNumber, expiryDate string) (*models.VerificationDoc, error) {
	if err := s.validateDocumentFile(file); err != nil {
		return nil, err
	}

	fileURL, err := s.fileUploadService.UploadFile(file, "documents/registration/"+userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload registration: %w", err)
	}

	var expiryTime *time.Time
	if expiryDate != "" {
		parsed, err := time.Parse("2006-01-02", expiryDate)
		if err == nil {
			expiryTime = &parsed
		}
	}

	doc := &models.VerificationDoc{
		Type:       "vehicle_registration",
		URL:        fileURL,
		FileName:   file.Filename,
		FileSize:   file.Size,
		Number:     registrationNumber,
		ExpiryDate: expiryTime,
		Status:     "pending",
		UploadedAt: time.Now(),
	}

	if err := s.driverRepo.AddDocument(userID, doc); err != nil {
		return nil, fmt.Errorf("failed to save registration document: %w", err)
	}

	return doc, nil
}

func (s *driverService) GetVerificationStatus(userID string) (*models.VerificationStatus, error) {
	return s.driverRepo.GetVerificationStatus(userID)
}

// Support

func (s *driverService) CreateSupportTicket(userID string, reqInterface interface{}) (*models.SupportTicket, error) {
	req, ok := reqInterface.(*CreateSupportTicketRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	ticket := &models.SupportTicket{
		Subject:     req.Subject,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		Status:      "open",
	}

	return s.driverRepo.CreateSupportTicket(userID, ticket)
}

func (s *driverService) GetSupportTickets(userID string, page, limit int) ([]*models.SupportTicket, int64, error) {
	return s.driverRepo.GetSupportTickets(userID, page, limit)
}

func (s *driverService) UpdateSupportTicket(userID, ticketID, message string) (*models.SupportTicket, error) {
	return s.driverRepo.UpdateSupportTicket(ticketID, message)
}

// Utility Methods

func (s *driverService) ValidateDriverProfile(profileInterface interface{}) error {
	profile, ok := profileInterface.(*UpdateDriverProfileRequest)
	if !ok {
		return errors.New("invalid profile type")
	}

	if profile.FirstName == "" || profile.LastName == "" {
		return errors.New("first name and last name are required")
	}

	if len(profile.FirstName) < 2 || len(profile.LastName) < 2 {
		return errors.New("names must be at least 2 characters long")
	}

	// Validate date of birth if provided
	if profile.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", profile.DateOfBirth)
		if err != nil {
			return errors.New("invalid date of birth format (YYYY-MM-DD)")
		}

		// Check minimum age (18 years)
		if time.Since(dob).Hours()/24/365 < 18 {
			return errors.New("driver must be at least 18 years old")
		}
	}

	return nil
}

func (s *driverService) ValidateVehicleDetails(vehicleInterface interface{}) error {
	vehicle, ok := vehicleInterface.(*AddVehicleRequest)
	if !ok {
		return errors.New("invalid vehicle type")
	}

	if vehicle.Make == "" || vehicle.Model == "" {
		return errors.New("vehicle make and model are required")
	}

	if vehicle.Year < 1990 || vehicle.Year > time.Now().Year()+1 {
		return errors.New("invalid vehicle year")
	}

	if vehicle.Color == "" {
		return errors.New("vehicle color is required")
	}

	if vehicle.LicensePlate == "" {
		return errors.New("license plate is required")
	}

	if vehicle.Seats < 1 || vehicle.Seats > 50 {
		return errors.New("invalid number of seats")
	}

	return nil
}

func (s *driverService) CalculateEarnings(userID string, period string) (*EarningsCalculation, error) {
	// This would implement complex earnings calculation logic
	return &EarningsCalculation{
		GrossEarnings: 0,
		Commission:    0,
		NetEarnings:   0,
		Tips:          0,
		Bonuses:       0,
		Deductions:    0,
	}, nil
}

func (s *driverService) GetDriverMetrics(userID string) (*DriverMetrics, error) {
	stats, err := s.driverRepo.GetDriverStats(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver stats: %w", err)
	}

	// Calculate metrics from stats
	return &DriverMetrics{
		TotalRides:       stats.TotalRides,
		AcceptanceRate:   stats.AcceptanceRate,
		CancellationRate: stats.CancellationRate,
		AverageRating:    stats.AverageRating,
		OnlineHoursToday: 0, // Would calculate from actual data
		OnlineHoursWeek:  0, // Would calculate from actual data
		EarningsToday:    0, // Would calculate from actual data
		EarningsWeek:     0, // Would calculate from actual data
	}, nil
}

// Helper methods

func (s *driverService) validateDocumentFile(file *multipart.FileHeader) error {
	// Check file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return errors.New("file size too large (max 10MB)")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".pdf"}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return errors.New("invalid file type (allowed: jpg, jpeg, png, pdf)")
}

func (s *driverService) validateImageFile(file *multipart.FileHeader) error {
	// Check file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return errors.New("image size too large (max 5MB)")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png"}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return errors.New("invalid image type (allowed: jpg, jpeg, png)")
}

func (s *driverService) calculateEarningsResponse(earnings []*models.Earning, period string) *EarningsResponse {
	var totalAmount, tips, bonuses, deductions float64
	totalRides := len(earnings)

	for _, earning := range earnings {
		totalAmount += earning.Amount
		tips += earning.Tips
		bonuses += earning.Bonuses
		deductions += earning.Deductions
	}

	averagePerRide := float64(0)
	if totalRides > 0 {
		averagePerRide = totalAmount / float64(totalRides)
	}

	netEarnings := totalAmount + tips + bonuses - deductions

	return &EarningsResponse{
		Period:         period,
		TotalAmount:    totalAmount,
		TotalRides:     totalRides,
		AveragePerRide: averagePerRide,
		Tips:           tips,
		Bonuses:        bonuses,
		Deductions:     deductions,
		NetEarnings:    netEarnings,
	}
}

func (s *driverService) triggerDocumentVerification(userID string, documents []*models.VerificationDoc) {
	// This would trigger an automated document verification process
	s.logger.Info().Str("user_id", userID).Int("doc_count", len(documents)).Msg("Document verification triggered")
}

func (s *driverService) triggerVehicleVerification(userID string, vehicle *models.VehicleDetails) {
	// This would trigger an automated vehicle verification process
	s.logger.Info().Str("user_id", userID).Str("vehicle", vehicle.Make+" "+vehicle.Model).Msg("Vehicle verification triggered")
}

func (s *driverService) notifyStatusChange(userID string, isOnline bool) {
	// This would send notifications about driver status changes
	status := "offline"
	if isOnline {
		status = "online"
	}
	s.logger.Info().Str("user_id", userID).Str("status", status).Msg("Driver status change notified")
}

func (s *driverService) broadcastLocationUpdate(userID string, location *models.Location) {
	// This would broadcast location updates to relevant parties
	s.logger.Debug().Str("user_id", userID).Float64("lat", location.Latitude).Float64("lng", location.Longitude).Msg("Location update broadcasted")
}

func (s *driverService) processPayout(payout *models.Payout) {
	// This would process the actual payout through payment providers
	s.logger.Info().Str("payout_id", payout.ID.Hex()).Float64("amount", payout.Amount).Msg("Payout processing initiated")
}
