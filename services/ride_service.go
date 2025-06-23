package services

import (
	"errors"
	"fmt"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RideService interface defines ride-related business logic
type RideService interface {
	// Basic ride operations
	CreateRide(req *CreateRideRequest) (*models.Ride, error)
	GetRide(rideID string) (*models.Ride, error)
	UpdateRide(rideID string, req *UpdateRideRequest) (*models.Ride, error)
	CancelRide(rideID, userID string, reason models.CancellationReason, notes string) (*models.Ride, error)

	// User rides
	GetUserRides(userID string, page, limit int, status string) ([]*models.Ride, int64, error)

	// Ride actions
	AcceptRide(rideID, driverID string) (*models.Ride, error)
	StartRide(rideID, driverID string, location *models.Location) (*models.Ride, error)
	CompleteRide(rideID, driverID string, completionDetails *RideCompletionDetails) (*models.Ride, error)

	// Ride tracking
	TrackRide(rideID string) (*RideTrackingInfo, error)
	UpdateRideLocation(rideID, driverID string, location *models.Location) error
	GetRideRoute(rideID string) ([]models.Location, error)
	GetRideETA(rideID string) (*time.Time, error)

	// Share ride
	ShareRide(rideID, userID string) (*ShareRideResponse, error)
	GetSharedRide(shareCode string) (*models.Ride, error)

	// Ride scheduling
	ScheduleRide(req *ScheduleRideRequest) (*models.Ride, error)
	GetScheduledRides(userID string, page, limit int) ([]*models.Ride, int64, error)
	UpdateScheduledRide(rideID string, req *UpdateScheduledRideRequest) (*models.Ride, error)
	CancelScheduledRide(rideID, userID string) error

	// Driver matching & nearby
	FindNearbyDrivers(lat, lng, radius float64, serviceType models.ServiceType) ([]*DriverInfo, error)
	RequestSpecificDriver(rideID, driverID, userID string) error
	GetDriverLocation(rideID string) (*models.Location, error)

	// Fare estimation & negotiation
	EstimateFare(req *FareEstimationRequest) (*FareEstimation, error)
	NegotiateFare(rideID, userID string, amount float64, message string) (*models.FareOffer, error)
	AcceptFare(rideID, userID string, offerID string) error
	CounterOffer(rideID, userID string, offerID string, amount float64, message string) (*models.FareOffer, error)
	GetFareHistory(rideID string) ([]*models.FareOffer, error)

	// Ride preferences
	SetRidePreferences(rideID, userID string, preferences *models.RidePreferences) error
	GetRidePreferences(rideID string) (*models.RidePreferences, error)

	// Special requirements
	SetSpecialRequirements(rideID, userID string, requirements *models.RideRequirements) error
	GetSpecialRequirements(rideID string) (*models.RideRequirements, error)

	// Ride reports & issues
	ReportRideIssue(rideID, userID string, report *ReportRideIssueRequest) error
	GetRideReports(rideID string) ([]*repositories.RideReport, error)

	// Repeat rides
	RepeatRide(rideID, userID string) (*models.Ride, error)
	GetFrequentRoutes(userID string, limit int) ([]*repositories.FrequentRoute, error)

	// Business validation and helpers
	ValidateRideRequest(req *CreateRideRequest) error
	CanUserAccessRide(rideID, userID string) (bool, error)
	CalculateRideDistance(pickup, dropoff *models.RideLocation) (float64, error)
	ProcessScheduledRides() error
}

// Request and response types

type CreateRideRequest struct {
	Type             models.RideType         `json:"type" validate:"required"`
	ServiceType      models.ServiceType      `json:"service_type" validate:"required"`
	PickupLocation   models.RideLocation     `json:"pickup_location" validate:"required"`
	DropoffLocation  models.RideLocation     `json:"dropoff_location" validate:"required"`
	Waypoints        []models.RideLocation   `json:"waypoints"`
	Requirements     models.RideRequirements `json:"requirements"`
	Preferences      models.RidePreferences  `json:"preferences"`
	ScheduledAt      *time.Time              `json:"scheduled_at,omitempty"`
	PaymentMethodID  string                  `json:"payment_method_id" validate:"required"`
	ProposedFare     *float64                `json:"proposed_fare,omitempty"`
	Notes            string                  `json:"notes"`
	SpecialInstructions string               `json:"special_instructions"`
	Platform         string                  `json:"platform"`
}

type UpdateRideRequest struct {
	PickupLocation      *models.RideLocation     `json:"pickup_location,omitempty"`
	DropoffLocation     *models.RideLocation     `json:"dropoff_location,omitempty"`
	Waypoints           []models.RideLocation    `json:"waypoints,omitempty"`
	Requirements        *models.RideRequirements `json:"requirements,omitempty"`
	Preferences         *models.RidePreferences  `json:"preferences,omitempty"`
	Notes               *string                  `json:"notes,omitempty"`
	SpecialInstructions *string                  `json:"special_instructions,omitempty"`
}

type RideCompletionDetails struct {
	ActualDistance   float64                 `json:"actual_distance"`
	ActualDuration   int                     `json:"actual_duration"`
	FinalFare        float64                 `json:"final_fare"`
	EndLocation      *models.Location        `json:"end_location"`
	CompletionCode   string                  `json:"completion_code"`
	PaymentStatus    string                  `json:"payment_status"`
}

type RideTrackingInfo struct {
	RideID           string            `json:"ride_id"`
	Status           models.RideStatus `json:"status"`
	DriverLocation   *models.Location  `json:"driver_location,omitempty"`
	EstimatedArrival *time.Time        `json:"estimated_arrival,omitempty"`
	Route            []models.Location `json:"route,omitempty"`
	Distance         float64           `json:"distance"`
	Duration         int               `json:"duration"`
	Progress         float64           `json:"progress"` // Percentage completed
}

type ShareRideResponse struct {
	ShareCode string `json:"share_code"`
	ShareURL  string `json:"share_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ScheduleRideRequest struct {
	Type             models.RideType         `json:"type" validate:"required"`
	ServiceType      models.ServiceType      `json:"service_type" validate:"required"`
	PickupLocation   models.RideLocation     `json:"pickup_location" validate:"required"`
	DropoffLocation  models.RideLocation     `json:"dropoff_location" validate:"required"`
	ScheduledAt      time.Time               `json:"scheduled_at" validate:"required"`
	Requirements     models.RideRequirements `json:"requirements"`
	Preferences      models.RidePreferences  `json:"preferences"`
	PaymentMethodID  string                  `json:"payment_method_id" validate:"required"`
	Notes            string                  `json:"notes"`
	RecurringPattern *RecurringPattern       `json:"recurring_pattern,omitempty"`
}

type UpdateScheduledRideRequest struct {
	ScheduledAt     *time.Time               `json:"scheduled_at,omitempty"`
	PickupLocation  *models.RideLocation     `json:"pickup_location,omitempty"`
	DropoffLocation *models.RideLocation     `json:"dropoff_location,omitempty"`
	Requirements    *models.RideRequirements `json:"requirements,omitempty"`
	Preferences     *models.RidePreferences  `json:"preferences,omitempty"`
	Notes           *string                  `json:"notes,omitempty"`
}

type RecurringPattern struct {
	Type        string    `json:"type"` // daily, weekly, monthly
	DaysOfWeek  []int     `json:"days_of_week,omitempty"` // 0=Sunday, 1=Monday, etc.
	Frequency   int       `json:"frequency"` // Every N days/weeks/months
	EndDate     *time.Time `json:"end_date,omitempty"`
	Occurrences *int      `json:"occurrences,omitempty"` // Number of occurrences
}

type DriverInfo struct {
	DriverID     string            `json:"driver_id"`
	Name         string            `json:"name"`
	Rating       float64           `json:"rating"`
	TotalRides   int               `json:"total_rides"`
	VehicleInfo  VehicleInfo       `json:"vehicle_info"`
	Location     *models.Location  `json:"location"`
	Distance     float64           `json:"distance"`
	ETA          int               `json:"eta"`
	ServiceTypes []models.ServiceType `json:"service_types"`
	IsAvailable  bool              `json:"is_available"`
}

type VehicleInfo struct {
	Type         models.VehicleType `json:"type"`
	Make         string             `json:"make"`
	Model        string             `json:"model"`
	Year         int                `json:"year"`
	Color        string             `json:"color"`
	LicensePlate string             `json:"license_plate"`
	PhotoURL     string             `json:"photo_url"`
}

type FareEstimationRequest struct {
	ServiceType     models.ServiceType      `json:"service_type" validate:"required"`
	PickupLocation  models.RideLocation     `json:"pickup_location" validate:"required"`
	DropoffLocation models.RideLocation     `json:"dropoff_location" validate:"required"`
	Waypoints       []models.RideLocation   `json:"waypoints"`
	Requirements    models.RideRequirements `json:"requirements"`
	ScheduledAt     *time.Time              `json:"scheduled_at,omitempty"`
}

type FareEstimation struct {
	EstimatedFare   float64                 `json:"estimated_fare"`
	MinFare         float64                 `json:"min_fare"`
	MaxFare         float64                 `json:"max_fare"`
	Currency        string                  `json:"currency"`
	Distance        float64                 `json:"distance"`
	Duration        int                     `json:"duration"`
	FareBreakdown   models.FareDetails      `json:"fare_breakdown"`
	SuggestedFare   float64                 `json:"suggested_fare"`
	MarketRate      float64                 `json:"market_rate"`
}

type ReportRideIssueRequest struct {
	Type        string   `json:"type" validate:"required"`
	Category    string   `json:"category" validate:"required"`
	Subject     string   `json:"subject" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Priority    string   `json:"priority"`
	Photos      []string `json:"photos"`
}

// rideService implements RideService interface
type rideService struct {
	rideRepo        repositories.RideRepository
	driverRepo      repositories.DriverRepository
	userRepo        repositories.UserRepository
	fareService     FareService
	locationService LocationService
	notificationSvc NotificationService
	paymentService  PaymentService
	logger          utils.Logger
}

// NewRideService creates a new ride service
func NewRideService(
	rideRepo repositories.RideRepository,
	driverRepo repositories.DriverRepository,
	userRepo repositories.UserRepository,
	fareService FareService,
	locationService LocationService,
	notificationSvc NotificationService,
	paymentService PaymentService,
	logger utils.Logger,
) RideService {
	return &rideService{
		rideRepo:        rideRepo,
		driverRepo:      driverRepo,
		userRepo:        userRepo,
		fareService:     fareService,
		locationService: locationService,
		notificationSvc: notificationSvc,
		paymentService:  paymentService,
		logger:          utils.ServiceLogger("ride"),
	}
}

// Basic ride operations

func (s *rideService) CreateRide(req *CreateRideRequest) (*models.Ride, error) {
	// Validate request
	if err := s.ValidateRideRequest(req); err != nil {
		return nil, err
	}

	// Get estimated fare if not provided
	if req.ProposedFare == nil {
		fareReq := &FareEstimationRequest{
			ServiceType:     req.ServiceType,
			PickupLocation:  req.PickupLocation,
			DropoffLocation: req.DropoffLocation,
			Waypoints:       req.Waypoints,
			Requirements:    req.Requirements,
			ScheduledAt:     req.ScheduledAt,
		}
		
		estimation, err := s.EstimateFare(fareReq)
		if err != nil {
			s.logger.Warn().Err(err).Msg("Failed to estimate fare, using default")
			proposedFare := 10.0 // Default minimum fare
			req.ProposedFare = &proposedFare
		} else {
			req.ProposedFare = &estimation.EstimatedFare
		}
	}

	// Create ride model
	ride := &models.Ride{
		Type:                req.Type,
		ServiceType:         req.ServiceType,
		PickupLocation:      req.PickupLocation,
		DropoffLocation:     req.DropoffLocation,
		Waypoints:           req.Waypoints,
		Requirements:        req.Requirements,
		Preferences:         req.Preferences,
		PaymentMethodID:     req.PaymentMethodID,
		Notes:               req.Notes,
		SpecialInstructions: req.SpecialInstructions,
		Platform:            req.Platform,
		RequestedAt:         time.Now(),
		TrackingEnabled:     true,
		ChatEnabled:         true,
	}

	// Set fare details
	ride.FareDetails = models.FareDetails{
		ProposedFare: *req.ProposedFare,
		Currency:     "USD", // Should be configurable
		Status:       models.FareStatusPending,
	}

	// Set scheduled time if provided
	if req.ScheduledAt != nil {
		ride.ScheduledAt = req.ScheduledAt
		ride.Type = models.RideTypeScheduled
	}

	// Calculate estimated distance and duration
	if distance, err := s.CalculateRideDistance(&req.PickupLocation, &req.DropoffLocation); err == nil {
		ride.EstimatedDistance = distance
		ride.EstimatedDuration = int(distance * 2) // Rough estimate: 2 minutes per km
	}

	// Create ride in database
	createdRide, err := s.rideRepo.Create(ride)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create ride")
		return nil, fmt.Errorf("failed to create ride")
	}

	// For instant rides, start driver matching process
	if ride.Type == models.RideTypeInstant {
		go s.startDriverMatching(createdRide)
	}

	// Send notifications
	go s.notifyRideCreated(createdRide)

	s.logger.Info().
		Str("ride_id", createdRide.ID.Hex()).
		Str("passenger_id", createdRide.PassengerID.Hex()).
		Str("service_type", string(createdRide.ServiceType)).
		Float64("proposed_fare", *req.ProposedFare).
		Msg("Ride created successfully")

	return createdRide, nil
}

func (s *rideService) GetRide(rideID string) (*models.Ride, error) {
	return s.rideRepo.GetByID(rideID)
}

func (s *rideService) UpdateRide(rideID string, req *UpdateRideRequest) (*models.Ride, error) {
	// Get existing ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Check if ride can be updated
	if !s.canUpdateRide(ride) {
		return nil, errors.New("ride cannot be updated in current status")
	}

	// Update fields
	if req.PickupLocation != nil {
		ride.PickupLocation = *req.PickupLocation
	}
	if req.DropoffLocation != nil {
		ride.DropoffLocation = *req.DropoffLocation
	}
	if req.Waypoints != nil {
		ride.Waypoints = req.Waypoints
	}
	if req.Requirements != nil {
		ride.Requirements = *req.Requirements
	}
	if req.Preferences != nil {
		ride.Preferences = *req.Preferences
	}
	if req.Notes != nil {
		ride.Notes = *req.Notes
	}
	if req.SpecialInstructions != nil {
		ride.SpecialInstructions = *req.SpecialInstructions
	}

	// Recalculate distance if locations changed
	if req.PickupLocation != nil || req.DropoffLocation != nil {
		if distance, err := s.CalculateRideDistance(&ride.PickupLocation, &ride.DropoffLocation); err == nil {
			ride.EstimatedDistance = distance
			ride.EstimatedDuration = int(distance * 2)
		}
	}

	// Update in database
	updatedRide, err := s.rideRepo.Update(ride)
	if err != nil {
		s.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update ride")
		return nil, fmt.Errorf("failed to update ride")
	}

	// Notify driver if assigned
	if ride.DriverID != nil {
		go s.notifyRideUpdated(updatedRide)
	}

	return updatedRide, nil
}

func (s *rideService) CancelRide(rideID, userID string, reason models.CancellationReason, notes string) (*models.Ride, error) {
	// Get ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Check if user can cancel
	canCancel, err := s.canUserCancelRide(ride, userID)
	if err != nil {
		return nil, err
	}
	if !canCancel {
		return nil, errors.New("you cannot cancel this ride")
	}

	// Update ride status
	ride.Status = models.RideStatusCancelled
	ride.CancellationReason = reason
	ride.CancellationNotes = notes
	ride.CancelledAt = &[]time.Time{time.Now()}[0]

	// Calculate cancellation fee if applicable
	cancellationFee := s.calculateCancellationFee(ride, userID)
	if cancellationFee > 0 {
		ride.FareDetails.CancellationFee = cancellationFee
	}

	// Update in database
	updatedRide, err := s.rideRepo.Update(ride)
	if err != nil {
		s.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to cancel ride")
		return nil, fmt.Errorf("failed to cancel ride")
	}

	// Process cancellation fee payment if applicable
	if cancellationFee > 0 {
		go s.processCancellationFeePayment(updatedRide, userID, cancellationFee)
	}

	// Notify participants
	go s.notifyRideCancelled(updatedRide, userID, reason)

	// Make driver available again if assigned
	if ride.DriverID != nil {
		go s.makeDriverAvailable(ride.DriverID.Hex())
	}

	s.logger.Info().
		Str("ride_id", rideID).
		Str("user_id", userID).
		Str("reason", string(reason)).
		Msg("Ride cancelled successfully")

	return updatedRide, nil
}

// User rides

func (s *rideService) GetUserRides(userID string, page, limit int, status string) ([]*models.Ride, int64, error) {
	return s.rideRepo.GetUserRides(userID, page, limit, status)
}

// Ride actions

func (s *rideService) AcceptRide(rideID, driverID string) (*models.Ride, error) {
	// Get ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Validate ride can be accepted
	if ride.Status != models.RideStatusPending && ride.Status != models.RideStatusFareNegotiation {
		return nil, errors.New("ride cannot be accepted in current status")
	}

	// Check if driver is available
	driver, err := s.driverRepo.GetByUserID(driverID)
	if err != nil {
		return nil, fmt.Errorf("driver not found")
	}
	if !driver.IsAvailable {
		return nil, errors.New("driver is not available")
	}

	// Assign driver to ride
	err = s.rideRepo.AssignDriver(rideID, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to assign driver")
	}

	// Update ride status
	err = s.rideRepo.UpdateRideStatus(rideID, models.RideStatusAccepted)
	if err != nil {
		return nil, fmt.Errorf("failed to update ride status")
	}

	// Make driver unavailable
	err = s.driverRepo.UpdateAvailability(driverID, false, 0, nil)
	if err != nil {
		s.logger.Warn().Err(err).Str("driver_id", driverID).Msg("Failed to update driver availability")
	}

	// Get updated ride
	updatedRide, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Notify passenger
	go s.notifyRideAccepted(updatedRide, driverID)

	// Calculate estimated arrival
	go s.calculateAndUpdateETA(updatedRide)

	s.logger.Info().
		Str("ride_id", rideID).
		Str("driver_id", driverID).
		Msg("Ride accepted successfully")

	return updatedRide, nil
}

func (s *rideService) StartRide(rideID, driverID string, location *models.Location) (*models.Ride, error) {
	// Get ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Validate driver can start ride
	if ride.DriverID == nil || ride.DriverID.Hex() != driverID {
		return nil, errors.New("you are not assigned to this ride")
	}

	if ride.Status != models.RideStatusDriverArrived && ride.Status != models.RideStatusAccepted {
		return nil, errors.New("ride cannot be started in current status")
	}

	// Update ride status and location
	err = s.rideRepo.UpdateRideStatus(rideID, models.RideStatusStarted)
	if err != nil {
		return nil, fmt.Errorf("failed to update ride status")
	}

	if location != nil {
		err = s.rideRepo.UpdateRideLocation(rideID, location)
		if err != nil {
			s.logger.Warn().Err(err).Msg("Failed to update ride location")
		}
	}

	// Generate completion code
	completionCode := utils.GenerateShortCode(4)
	ride.CompletionCode = completionCode

	// Update ride
	updatedRide, err := s.rideRepo.Update(ride)
	if err != nil {
		return nil, fmt.Errorf("failed to update ride")
	}

	// Notify passenger
	go s.notifyRideStarted(updatedRide)

	s.logger.Info().
		Str("ride_id", rideID).
		Str("driver_id", driverID).
		Msg("Ride started successfully")

	return updatedRide, nil
}

func (s *rideService) CompleteRide(rideID, driverID string, details *RideCompletionDetails) (*models.Ride, error) {
	// Get ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Validate driver can complete ride
	if ride.DriverID == nil || ride.DriverID.Hex() != driverID {
		return nil, errors.New("you are not assigned to this ride")
	}

	if ride.Status != models.RideStatusStarted && ride.Status != models.RideStatusInProgress {
		return nil, errors.New("ride cannot be completed in current status")
	}

	// Update ride with completion details
	ride.Status = models.RideStatusCompleted
	ride.ActualDistance = details.ActualDistance
	ride.ActualDuration = details.ActualDuration
	ride.FareDetails.FinalFare = details.FinalFare
	ride.PaymentStatus = details.PaymentStatus
	
	now := time.Now()
	ride.CompletedAt = &now

	// Update ride
	updatedRide, err := s.rideRepo.Update(ride)
	if err != nil {
		return nil, fmt.Errorf("failed to complete ride")
	}

	// Process payment
	go s.processRidePayment(updatedRide)

	// Make driver available again
	go s.makeDriverAvailable(driverID)

	// Notify passenger
	go s.notifyRideCompleted(updatedRide)

	s.logger.Info().
		Str("ride_id", rideID).
		Str("driver_id", driverID).
		Float64("final_fare", details.FinalFare).
		Msg("Ride completed successfully")

	return updatedRide, nil
}

// Ride tracking

func (s *rideService) TrackRide(rideID string) (*RideTrackingInfo, error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	tracking := &RideTrackingInfo{
		RideID:           rideID,
		Status:           ride.Status,
		DriverLocation:   ride.DriverLocation,
		EstimatedArrival: ride.EstimatedArrival,
		Route:            ride.Route,
		Distance:         ride.EstimatedDistance,
		Duration:         ride.EstimatedDuration,
		Progress:         s.calculateRideProgress(ride),
	}

	return tracking, nil
}

func (s *rideService) UpdateRideLocation(rideID, driverID string, location *models.Location) error {
	// Verify driver is assigned to ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return err
	}

	if ride.DriverID == nil || ride.DriverID.Hex() != driverID {
		return errors.New("driver not assigned to this ride")
	}

	// Update location
	err = s.rideRepo.UpdateRideLocation(rideID, location)
	if err != nil {
		return err
	}

	// Update ETA if ride is in progress
	if ride.Status == models.RideStatusDriverEnRoute || ride.Status == models.RideStatusStarted {
		go s.updateETABasedOnLocation(rideID, location)
	}

	// Broadcast location update via websocket
	go s.broadcastLocationUpdate(rideID, location)

	return nil
}

func (s *rideService) GetRideRoute(rideID string) ([]models.Location, error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	return ride.Route, nil
}

func (s *rideService) GetRideETA(rideID string) (*time.Time, error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	return ride.EstimatedArrival, nil
}

// Share ride

func (s *rideService) ShareRide(rideID, userID string) (*ShareRideResponse, error) {
	// Get ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Check if user can share this ride
	if ride.PassengerID.Hex() != userID {
		return nil, errors.New("you can only share your own rides")
	}

	// Enable sharing
	ride.IsShared = true
	if ride.ShareCode == "" {
		ride.ShareCode = utils.GenerateShortCode(8)
	}

	// Update ride
	_, err = s.rideRepo.Update(ride)
	if err != nil {
		return nil, fmt.Errorf("failed to enable ride sharing")
	}

	response := &ShareRideResponse{
		ShareCode: ride.ShareCode,
		ShareURL:  fmt.Sprintf("https://app.ondrive.com/shared-ride/%s", ride.ShareCode),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Share link expires in 24 hours
	}

	return response, nil
}

func (s *rideService) GetSharedRide(shareCode string) (*models.Ride, error) {
	return s.rideRepo.GetRidesByShareCode(shareCode)
}

// Ride scheduling

func (s *rideService) ScheduleRide(req *ScheduleRideRequest) (*models.Ride, error) {
	// Validate scheduled time
	if req.ScheduledAt.Before(time.Now().Add(15 * time.Minute)) {
		return nil, errors.New("ride must be scheduled at least 15 minutes in advance")
	}

	// Create ride request
	createReq := &CreateRideRequest{
		Type:             models.RideTypeScheduled,
		ServiceType:      req.ServiceType,
		PickupLocation:   req.PickupLocation,
		DropoffLocation:  req.DropoffLocation,
		Requirements:     req.Requirements,
		Preferences:      req.Preferences,
		PaymentMethodID:  req.PaymentMethodID,
		Notes:            req.Notes,
		ScheduledAt:      &req.ScheduledAt,
	}

	ride, err := s.CreateRide(createReq)
	if err != nil {
		return nil, err
	}

	// Schedule recurring rides if pattern provided
	if req.RecurringPattern != nil {
		go s.createRecurringRides(ride, req.RecurringPattern)
	}

	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Time("scheduled_at", req.ScheduledAt).
		Msg("Ride scheduled successfully")

	return ride, nil
}

func (s *rideService) GetScheduledRides(userID string, page, limit int) ([]*models.Ride, int64, error) {
	return s.rideRepo.GetScheduledRides(userID, page, limit)
}

func (s *rideService) UpdateScheduledRide(rideID string, req *UpdateScheduledRideRequest) (*models.Ride, error) {
	// Get ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Check if ride is scheduled and can be updated
	if ride.Type != models.RideTypeScheduled {
		return nil, errors.New("only scheduled rides can be updated")
	}

	if ride.Status != models.RideStatusPending {
		return nil, errors.New("scheduled ride cannot be updated after driver assignment")
	}

	// Update scheduled time if provided
	if req.ScheduledAt != nil {
		if req.ScheduledAt.Before(time.Now().Add(15 * time.Minute)) {
			return nil, errors.New("ride must be scheduled at least 15 minutes in advance")
		}
		err = s.rideRepo.UpdateScheduledRide(rideID, *req.ScheduledAt)
		if err != nil {
			return nil, fmt.Errorf("failed to update scheduled time")
		}
	}

	// Update other fields
	updateReq := &UpdateRideRequest{
		PickupLocation:  req.PickupLocation,
		DropoffLocation: req.DropoffLocation,
		Requirements:    req.Requirements,
		Preferences:     req.Preferences,
		Notes:           req.Notes,
	}

	return s.UpdateRide(rideID, updateReq)
}

func (s *rideService) CancelScheduledRide(rideID, userID string) error {
	_, err := s.CancelRide(rideID, userID, models.CancelByPassenger, "Scheduled ride cancelled")
	return err
}

// Driver matching & nearby

func (s *rideService) FindNearbyDrivers(lat, lng, radius float64, serviceType models.ServiceType) ([]*DriverInfo, error) {
	drivers, err := s.driverRepo.GetAvailableDrivers(string(serviceType), lat, lng, radius, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby drivers: %w", err)
	}

	var driverInfos []*DriverInfo
	for _, driver := range drivers {
		// Calculate distance and ETA
		distance := s.locationService.CalculateDistance(
			&models.Location{Coordinates: []float64{lng, lat}},
			driver.CurrentLocation,
		)
		eta := s.locationService.CalculateETA(
			&models.Location{Coordinates: []float64{lng, lat}},
			driver.CurrentLocation,
		)

		driverInfo := &DriverInfo{
			DriverID:     driver.UserID.Hex(),
			Name:         fmt.Sprintf("%s %s", driver.Profile.FirstName, driver.Profile.LastName),
			Rating:       driver.Rating.Average,
			TotalRides:   driver.Rating.TotalRides,
			Location:     driver.CurrentLocation,
			Distance:     distance,
			ETA:          eta,
			ServiceTypes: driver.ServiceTypes,
			IsAvailable:  driver.IsAvailable,
			VehicleInfo: VehicleInfo{
				Type:         driver.VehicleDetails.VehicleType,
				Make:         driver.VehicleDetails.Make,
				Model:        driver.VehicleDetails.Model,
				Year:         driver.VehicleDetails.Year,
				Color:        driver.VehicleDetails.Color,
				LicensePlate: driver.VehicleDetails.LicensePlate,
				PhotoURL:     driver.VehicleDetails.PhotoURL,
			},
		}

		driverInfos = append(driverInfos, driverInfo)
	}

	return driverInfos, nil
}

func (s *rideService) RequestSpecificDriver(rideID, driverID, userID string) error {
	// Get ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return err
	}

	// Check if user owns the ride
	if ride.PassengerID.Hex() != userID {
		return errors.New("you can only request drivers for your own rides")
	}

	// Check if ride is in pending status
	if ride.Status != models.RideStatusPending {
		return errors.New("driver can only be requested for pending rides")
	}

	// Check if driver is available
	driver, err := s.driverRepo.GetByUserID(driverID)
	if err != nil {
		return fmt.Errorf("driver not found")
	}

	if !driver.IsAvailable {
		return errors.New("requested driver is not available")
	}

	// Send notification to specific driver
	go s.notifySpecificDriverRequest(ride, driverID)

	s.logger.Info().
		Str("ride_id", rideID).
		Str("driver_id", driverID).
		Str("passenger_id", userID).
		Msg("Specific driver requested")

	return nil
}

func (s *rideService) GetDriverLocation(rideID string) (*models.Location, error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	if ride.DriverID == nil {
		return nil, errors.New("no driver assigned to this ride")
	}

	return ride.DriverLocation, nil
}

// Fare estimation & negotiation

func (s *rideService) EstimateFare(req *FareEstimationRequest) (*FareEstimation, error) {
	// Calculate distance
	distance, err := s.CalculateRideDistance(&req.PickupLocation, &req.DropoffLocation)
	if err != nil {
		distance = 5.0 // Default estimate
	}

	// Calculate duration (rough estimate)
	duration := int(distance * 2) // 2 minutes per km

	// Get fare estimation from fare service
	fareDetails, err := s.fareService.CalculateBaseFare(string(req.ServiceType), distance, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fare: %w", err)
	}

	estimation := &FareEstimation{
		EstimatedFare: fareDetails.ProposedFare,
		MinFare:       fareDetails.ProposedFare * 0.8,
		MaxFare:       fareDetails.ProposedFare * 1.5,
		Currency:      "USD",
		Distance:      distance,
		Duration:      duration,
		FareBreakdown: *fareDetails,
		SuggestedFare: fareDetails.ProposedFare,
		MarketRate:    fareDetails.ProposedFare,
	}

	return estimation, nil
}

func (s *rideService) NegotiateFare(rideID, userID string, amount float64, message string) (*models.FareOffer, error) {
	// Get ride
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Check if user can negotiate
	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	if ride.PassengerID != userObjectID && (ride.DriverID == nil || *ride.DriverID != userObjectID) {
		return nil, errors.New("you cannot negotiate fare for this ride")
	}

	// Create fare offer
	var offerTo primitive.ObjectID
	if ride.PassengerID == userObjectID {
		if ride.DriverID == nil {
			return nil, errors.New("no driver assigned yet")
		}
		offerTo = *ride.DriverID
	} else {
		offerTo = ride.PassengerID
	}

	offer := &models.FareOffer{
		OfferBy:   userObjectID,
		OfferTo:   offerTo,
		Amount:    amount,
		Message:   message,
		ExpiresAt: time.Now().Add(5 * time.Minute), // Offer expires in 5 minutes
	}

	// Save offer
	err = s.rideRepo.SaveFareOffer(rideID, offer)
	if err != nil {
		return nil, fmt.Errorf("failed to save fare offer")
	}

	// Update ride status to fare negotiation
	err = s.rideRepo.UpdateRideStatus(rideID, models.RideStatusFareNegotiation)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to update ride status to fare negotiation")
	}

	// Notify other party
	go s.notifyFareProposed(ride, offer)

	s.logger.Info().
		Str("ride_id", rideID).
		Str("user_id", userID).
		Float64("amount", amount).
		Msg("Fare negotiated")

	return offer, nil
}

func (s *rideService) AcceptFare(rideID, userID string, offerID string) error {
	// Get ride and fare history
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return err
	}

	offers, err := s.rideRepo.GetFareHistory(rideID)
	if err != nil {
		return err
	}

	// Find the specific offer
	var targetOffer *models.FareOffer
	for _, offer := range offers {
		if offer.ID.Hex() == offerID {
			targetOffer = offer
			break
		}
	}

	if targetOffer == nil {
		return errors.New("fare offer not found")
	}

	// Check if user can accept this offer
	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	if targetOffer.OfferTo != userObjectID {
		return errors.New("you cannot accept this offer")
	}

	// Check if offer is still valid
	if time.Now().After(targetOffer.ExpiresAt) {
		return errors.New("fare offer has expired")
	}

	// Update fare details
	ride.FareDetails.FinalFare = targetOffer.Amount
	ride.FareDetails.Status = models.FareStatusAccepted

	// Mark offer as accepted
	targetOffer.IsAccepted = true
	now := time.Now()
	targetOffer.ResponseAt = &now

	// Update ride
	err = s.rideRepo.Update(ride)
	if err != nil {
		return fmt.Errorf("failed to accept fare")
	}

	// Notify other party
	go s.notifyFareAccepted(ride, targetOffer)

	s.logger.Info().
		Str("ride_id", rideID).
		Str("user_id", userID).
		Str("offer_id", offerID).
		Float64("amount", targetOffer.Amount).
		Msg("Fare accepted")

	return nil
}

func (s *rideService) CounterOffer(rideID, userID string, offerID string, amount float64, message string) (*models.FareOffer, error) {
	// First reject the original offer
	// Then create a new counter offer
	return s.NegotiateFare(rideID, userID, amount, message)
}

func (s *rideService) GetFareHistory(rideID string) ([]*models.FareOffer, error) {
	return s.rideRepo.GetFareHistory(rideID)
}

// Ride preferences and requirements

func (s *rideService) SetRidePreferences(rideID, userID string, preferences *models.RidePreferences) error {
	// Verify user can modify preferences
	canAccess, err := s.CanUserAccessRide(rideID, userID)
	if err != nil {
		return err
	}
	if !canAccess {
		return errors.New("you cannot modify preferences for this ride")
	}

	return s.rideRepo.UpdateRidePreferences(rideID, preferences)
}

func (s *rideService) GetRidePreferences(rideID string) (*models.RidePreferences, error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	return &ride.Preferences, nil
}

func (s *rideService) SetSpecialRequirements(rideID, userID string, requirements *models.RideRequirements) error {
	// Verify user can modify requirements
	canAccess, err := s.CanUserAccessRide(rideID, userID)
	if err != nil {
		return err
	}
	if !canAccess {
		return errors.New("you cannot modify requirements for this ride")
	}

	return s.rideRepo.UpdateRideRequirements(rideID, requirements)
}

func (s *rideService) GetSpecialRequirements(rideID string) (*models.RideRequirements, error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	return &ride.Requirements, nil
}

// Ride reports & issues

func (s *rideService) ReportRideIssue(rideID, userID string, req *ReportRideIssueRequest) error {
	// Verify user can report this ride
	canAccess, err := s.CanUserAccessRide(rideID, userID)
	if err != nil {
		return err
	}
	if !canAccess {
		return errors.New("you cannot report issues for this ride")
	}

	userObjectID, _ := primitive.ObjectIDFromHex(userID)
	report := &repositories.RideReport{
		ReporterID:  userObjectID,
		Type:        req.Type,
		Category:    req.Category,
		Subject:     req.Subject,
		Description: req.Description,
		Priority:    req.Priority,
		Photos:      req.Photos,
	}

	err = s.rideRepo.SaveRideReport(rideID, report)
	if err != nil {
		return fmt.Errorf("failed to save ride report")
	}

	// Notify support team
	go s.notifyRideIssueReported(rideID, userID, report)

	s.logger.Info().
		Str("ride_id", rideID).
		Str("user_id", userID).
		Str("type", req.Type).
		Str("category", req.Category).
		Msg("Ride issue reported")

	return nil
}

func (s *rideService) GetRideReports(rideID string) ([]*repositories.RideReport, error) {
	return s.rideRepo.GetRideReports(rideID)
}

// Repeat rides

func (s *rideService) RepeatRide(rideID, userID string) (*models.Ride, error) {
	// Get original ride
	originalRide, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}

	// Check if user owns the ride
	if originalRide.PassengerID.Hex() != userID {
		return nil, errors.New("you can only repeat your own rides")
	}

	// Create new ride request based on original
	req := &CreateRideRequest{
		Type:                models.RideTypeInstant,
		ServiceType:         originalRide.ServiceType,
		PickupLocation:      originalRide.PickupLocation,
		DropoffLocation:     originalRide.DropoffLocation,
		Waypoints:           originalRide.Waypoints,
		Requirements:        originalRide.Requirements,
		Preferences:         originalRide.Preferences,
		PaymentMethodID:     originalRide.PaymentMethodID,
		Notes:               originalRide.Notes,
		SpecialInstructions: originalRide.SpecialInstructions,
		Platform:            originalRide.Platform,
	}

	return s.CreateRide(req)
}

func (s *rideService) GetFrequentRoutes(userID string, limit int) ([]*repositories.FrequentRoute, error) {
	return s.rideRepo.GetFrequentRoutes(userID, limit)
}

// Business validation and helpers

func (s *rideService) ValidateRideRequest(req *CreateRideRequest) error {
	// Validate pickup and dropoff locations are different
	if req.PickupLocation.Address == req.DropoffLocation.Address {
		return errors.New("pickup and dropoff locations must be different")
	}

	// Validate service type is supported
	validServiceTypes := []models.ServiceType{
		models.ServiceTypeRide,
		models.ServiceTypeCourier,
		models.ServiceTypeFreight,
	}
	
	isValid := false
	for _, validType := range validServiceTypes {
		if req.ServiceType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("invalid service type")
	}

	// Validate passenger count
	if req.Requirements.PassengerCount < 1 || req.Requirements.PassengerCount > 8 {
		return errors.New("passenger count must be between 1 and 8")
	}

	return nil
}

func (s *rideService) CanUserAccessRide(rideID, userID string) (bool, error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return false, err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	// Check if user is passenger or driver
	if ride.PassengerID == userObjectID {
		return true, nil
	}

	if ride.DriverID != nil && *ride.DriverID == userObjectID {
		return true, nil
	}

	return false, nil
}

func (s *rideService) CalculateRideDistance(pickup, dropoff *models.RideLocation) (float64, error) {
	if s.locationService == nil {
		// Fallback calculation using coordinates
		return s.calculateDistanceFromCoordinates(pickup.Coordinates, dropoff.Coordinates), nil
	}

	pickupLocation := &models.Location{
		Type:        "Point",
		Coordinates: pickup.Coordinates,
		Address:     pickup.Address,
	}

	dropoffLocation := &models.Location{
		Type:        "Point",
		Coordinates: dropoff.Coordinates,
		Address:     dropoff.Address,
	}

	return s.locationService.CalculateDistance(pickupLocation, dropoffLocation), nil
}

func (s *rideService) ProcessScheduledRides() error {
	// Get rides scheduled for the next 15 minutes
	targetTime := time.Now().Add(15 * time.Minute)
	rides, err := s.rideRepo.GetScheduledRidesForTime(targetTime)
	if err != nil {
		return fmt.Errorf("failed to get scheduled rides: %w", err)
	}

	for _, ride := range rides {
		// Start driver matching for scheduled rides
		go s.startDriverMatching(ride)
		
		s.logger.Info().
			Str("ride_id", ride.ID.Hex()).
			Time("scheduled_at", *ride.ScheduledAt).
			Msg("Processing scheduled ride")
	}

	return nil
}

// Private helper methods

func (s *rideService) canUpdateRide(ride *models.Ride) bool {
	// Can only update pending rides or rides in fare negotiation
	return ride.Status == models.RideStatusPending || 
		   ride.Status == models.RideStatusFareNegotiation
}

func (s *rideService) canUserCancelRide(ride *models.Ride, userID string) (bool, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	// Check if user is participant
	if ride.PassengerID != userObjectID && (ride.DriverID == nil || *ride.DriverID != userObjectID) {
		return false, nil
	}

	// Check if ride can be cancelled
	cancelableStatuses := []models.RideStatus{
		models.RideStatusPending,
		models.RideStatusFareNegotiation,
		models.RideStatusAccepted,
		models.RideStatusDriverEnRoute,
		models.RideStatusDriverArrived,
	}

	for _, status := range cancelableStatuses {
		if ride.Status == status {
			return true, nil
		}
	}

	return false, nil
}

func (s *rideService) calculateCancellationFee(ride *models.Ride, userID string) float64 {
	// Calculate cancellation fee based on ride status and timing
	if ride.Status == models.RideStatusPending || ride.Status == models.RideStatusFareNegotiation {
		return 0 // No fee for early cancellation
	}

	// If driver is en route or arrived, charge cancellation fee
	if ride.Status == models.RideStatusDriverEnRoute || ride.Status == models.RideStatusDriverArrived {
		return 5.0 // Base cancellation fee
	}

	return 0
}

func (s *rideService) calculateRideProgress(ride *models.Ride) float64 {
	switch ride.Status {
	case models.RideStatusPending, models.RideStatusFareNegotiation:
		return 0
	case models.RideStatusAccepted:
		return 10
	case models.RideStatusDriverEnRoute:
		return 25
	case models.RideStatusDriverArrived:
		return 40
	case models.RideStatusStarted:
		return 50
	case models.RideStatusInProgress:
		return 75
	case models.RideStatusCompleted:
		return 100
	default:
		return 0
	}
}

func (s *rideService) calculateDistanceFromCoordinates(pickup, dropoff []float64) float64 {
	if len(pickup) < 2 || len(dropoff) < 2 {
		return 5.0 // Default estimate
	}

	// Simple haversine distance calculation
	return utils.CalculateHaversineDistance(pickup[1], pickup[0], dropoff[1], dropoff[0])
}

// Notification methods (async operations)

func (s *rideService) startDriverMatching(ride *models.Ride) {
	// Find available drivers near pickup location
	if len(ride.PickupLocation.Coordinates) < 2 {
		s.logger.Warn().Str("ride_id", ride.ID.Hex()).Msg("Invalid pickup coordinates for driver matching")
		return
	}

	lat := ride.PickupLocation.Coordinates[1]
	lng := ride.PickupLocation.Coordinates[0]

	drivers, err := s.FindNearbyDrivers(lat, lng, 10.0, ride.ServiceType)
	if err != nil {
		s.logger.Error().Err(err).Str("ride_id", ride.ID.Hex()).Msg("Failed to find nearby drivers")
		return
	}

	if len(drivers) == 0 {
		s.logger.Warn().Str("ride_id", ride.ID.Hex()).Msg("No available drivers found")
		return
	}

	// Notify drivers about the ride request
	for _, driver := range drivers {
		s.notifyDriverAboutRide(ride, driver.DriverID)
	}
}

func (s *rideService) notifyRideCreated(ride *models.Ride) {
	if s.notificationSvc == nil {
		return
	}

	notification := &NotificationRequest{
		Type:    "ride_created",
		Title:   "Ride Requested",
		Message: "Your ride has been requested successfully",
		Data: map[string]interface{}{
			"ride_id": ride.ID.Hex(),
		},
	}

	err := s.notificationSvc.SendNotificationToUser(ride.PassengerID.Hex(), notification)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to send ride created notification")
	}
}

func (s *rideService) notifyDriverAboutRide(ride *models.Ride, driverID string) {
	// Implementation would send push notification to driver
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Str("driver_id", driverID).
		Msg("Notifying driver about ride request")
}

func (s *rideService) notifyRideAccepted(ride *models.Ride, driverID string) {
	// Implementation would notify passenger about driver acceptance
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Str("driver_id", driverID).
		Msg("Notifying passenger about ride acceptance")
}

func (s *rideService) notifyRideStarted(ride *models.Ride) {
	// Implementation would notify passenger that ride has started
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Msg("Notifying passenger that ride has started")
}

func (s *rideService) notifyRideCompleted(ride *models.Ride) {
	// Implementation would notify passenger about ride completion
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Msg("Notifying passenger about ride completion")
}

func (s *rideService) notifyRideCancelled(ride *models.Ride, cancelledBy string, reason models.CancellationReason) {
	// Implementation would notify other party about cancellation
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Str("cancelled_by", cancelledBy).
		Str("reason", string(reason)).
		Msg("Notifying about ride cancellation")
}

func (s *rideService) notifyRideUpdated(ride *models.Ride) {
	// Implementation would notify driver about ride updates
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Msg("Notifying driver about ride updates")
}

func (s *rideService) notifySpecificDriverRequest(ride *models.Ride, driverID string) {
	// Implementation would send special notification to requested driver
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Str("driver_id", driverID).
		Msg("Notifying specific driver about ride request")
}

func (s *rideService) notifyFareProposed(ride *models.Ride, offer *models.FareOffer) {
	// Implementation would notify about fare proposal
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Str("offer_id", offer.ID.Hex()).
		Float64("amount", offer.Amount).
		Msg("Notifying about fare proposal")
}

func (s *rideService) notifyFareAccepted(ride *models.Ride, offer *models.FareOffer) {
	// Implementation would notify about fare acceptance
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Str("offer_id", offer.ID.Hex()).
		Msg("Notifying about fare acceptance")
}

func (s *rideService) notifyRideIssueReported(rideID, userID string, report *repositories.RideReport) {
	// Implementation would notify support team
	s.logger.Info().
		Str("ride_id", rideID).
		Str("user_id", userID).
		Str("report_id", report.ID.Hex()).
		Msg("Notifying support team about ride issue")
}

// Payment and driver management methods

func (s *rideService) processRidePayment(ride *models.Ride) {
	// Implementation would process payment through payment service
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Float64("amount", ride.FareDetails.FinalFare).
		Msg("Processing ride payment")
}

func (s *rideService) processCancellationFeePayment(ride *models.Ride, userID string, amount float64) {
	// Implementation would charge cancellation fee
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Str("user_id", userID).
		Float64("amount", amount).
		Msg("Processing cancellation fee payment")
}

func (s *rideService) makeDriverAvailable(driverID string) {
	err := s.driverRepo.UpdateAvailability(driverID, true, 0, nil)
	if err != nil {
		s.logger.Warn().Err(err).Str("driver_id", driverID).Msg("Failed to make driver available")
	}
}

func (s *rideService) calculateAndUpdateETA(ride *models.Ride) {
	// Implementation would calculate ETA and update ride
	s.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Msg("Calculating and updating ETA")
}

func (s *rideService) updateETABasedOnLocation(rideID string, location *models.Location) {
	// Implementation would update ETA based on current location
	s.logger.Info().
		Str("ride_id", rideID).
		Msg("Updating ETA based on location")
}

func (s *rideService) broadcastLocationUpdate(rideID string, location *models.Location) {
	// Implementation would broadcast location via websocket
	s.logger.Info().
		Str("ride_id", rideID).
		Msg("Broadcasting location update")
}

func (s *rideService) createRecurringRides(baseRide *models.Ride, pattern *RecurringPattern) {
	// Implementation would create recurring rides based on pattern
	s.logger.Info().
		Str("ride_id", baseRide.ID.Hex()).
		Str("pattern_type", pattern.Type).
		Msg("Creating recurring rides")
}