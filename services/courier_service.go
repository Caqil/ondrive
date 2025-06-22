package services

import (
	"fmt"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CourierService interface for courier business logic
type CourierService interface {
	// Core Operations
	CreateRequest(request *models.CourierRequest) (*models.CourierRequest, error)
	AcceptRequest(requestID, courierID primitive.ObjectID) error
	CancelRequest(requestID, userID primitive.ObjectID, reason string) error

	// Package Management
	UpdatePackageDetails(requestID primitive.ObjectID, packageDetails models.CourierPackage) error
	AddPackagePhotos(requestID primitive.ObjectID, photoURLs []string) error

	// Tracking & Status
	UpdateStatus(requestID primitive.ObjectID, status models.CourierStatus, courierID primitive.ObjectID) error
	ConfirmPickup(requestID, courierID primitive.ObjectID, photos []string, notes string, location *models.Location) error
	ConfirmDelivery(requestID, courierID primitive.ObjectID, deliveryProof *models.DeliveryProof) error
	GetTrackingInfo(requestID primitive.ObjectID) (*TrackingInfo, error)
	AddTrackingEvent(requestID primitive.ObjectID, event models.TrackingEvent) error
	UpdateLocation(requestID, courierID primitive.ObjectID, location models.Location) error

	// Recipient Management
	UpdateRecipient(requestID primitive.ObjectID, recipient *models.CourierRecipient) error
	NotifyRecipient(requestID primitive.ObjectID, message string, sendSMS, sendEmail, sendPush bool) error

	// Delivery Management
	ScheduleDelivery(requestID primitive.ObjectID, scheduledTime time.Time, timeWindow *models.TimeWindow) error
	GetAvailableTimeSlots(date, location string) ([]TimeSlot, error)

	// Special Services
	AddInsurance(requestID primitive.ObjectID, insurance *models.CourierInsurance) error
	SetSignatureRequirement(requestID primitive.ObjectID, required bool, signatureType string) error

	// Pricing & Fare
	CalculateFare(request *models.CourierRequest) (*models.CourierFare, error)
	EstimatePrice(req EstimatePriceRequest) (*PriceEstimate, error)
	ApplyPromoCode(requestID primitive.ObjectID, promoCode string) error

	// Delivery Proof
	UpdateDeliveryProofPhotos(requestID primitive.ObjectID, photoURLs []string) error

	// Address Management
	GetAddressBook(userID primitive.ObjectID) ([]*models.SavedAddress, error)
	AddAddress(userID primitive.ObjectID, address *models.SavedAddress) (*models.SavedAddress, error)
	UpdateAddress(userID, addressID primitive.ObjectID, address *models.SavedAddress) (*models.SavedAddress, error)
	DeleteAddress(userID, addressID primitive.ObjectID) error

	// Support & Issues
	ReportIssue(requestID, reporterID primitive.ObjectID, issueType, description, priority string, photos []string) (primitive.ObjectID, error)
	FileClaim(requestID, claimantID primitive.ObjectID, claimType string, amount float64, description string, evidence []string) (primitive.ObjectID, error)

	// Analytics & Stats
	GetCourierStats(courierID primitive.ObjectID) (*repositories.CourierStats, error)
	GetSenderStats(senderID primitive.ObjectID) (*repositories.SenderStats, error)

	// Courier Assignment
	FindNearbyRequests(courierID primitive.ObjectID, location models.Location, radius float64) ([]*models.CourierRequest, error)
	AssignCourier(requestID, courierID primitive.ObjectID) error
	UnassignCourier(requestID primitive.ObjectID) error

	// Validation
	ValidateRequest(request *models.CourierRequest) error
	CanCancelRequest(requestID, userID primitive.ObjectID) (bool, error)
	CanUpdateRequest(requestID, userID primitive.ObjectID) (bool, error)
}

// Supporting types
type TrackingInfo struct {
	Request          *models.CourierRequest `json:"request"`
	TrackingEvents   []models.TrackingEvent `json:"tracking_events"`
	CurrentStatus    models.CourierStatus   `json:"current_status"`
	CurrentLocation  *models.Location       `json:"current_location,omitempty"`
	EstimatedArrival *time.Time             `json:"estimated_arrival,omitempty"`
	CourierInfo      *CourierInfo           `json:"courier_info,omitempty"`
	LastUpdate       time.Time              `json:"last_update"`
}

type CourierInfo struct {
	ID     primitive.ObjectID `json:"id"`
	Name   string             `json:"name"`
	Phone  string             `json:"phone"`
	Rating float64            `json:"rating"`
	Photo  string             `json:"photo,omitempty"`
}

type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
	Price     float64   `json:"price,omitempty"`
}

type PriceEstimate struct {
	BaseFare          float64            `json:"base_fare"`
	DistanceFare      float64            `json:"distance_fare"`
	WeightFare        float64            `json:"weight_fare"`
	SizeFare          float64            `json:"size_fare"`
	PriorityFare      float64            `json:"priority_fare"`
	Surcharges        map[string]float64 `json:"surcharges"`
	InsuranceFee      float64            `json:"insurance_fee"`
	TaxAmount         float64            `json:"tax_amount"`
	SubTotal          float64            `json:"sub_total"`
	TotalAmount       float64            `json:"total_amount"`
	Currency          string             `json:"currency"`
	BreakdownDetails  []PriceBreakdown   `json:"breakdown_details"`
	EstimatedDuration string             `json:"estimated_duration"`
}

type PriceBreakdown struct {
	Item        string  `json:"item"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type EstimatePriceRequest struct {
	Priority              models.DeliverySpeed   `json:"priority"`
	Package               CourierPackageDTO      `json:"package"`
	PickupLocation        models.CourierLocation `json:"pickup_location"`
	DeliveryLocation      models.CourierLocation `json:"delivery_location"`
	ScheduledPickupTime   *time.Time             `json:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *time.Time             `json:"scheduled_delivery_time,omitempty"`
	RequiresInsurance     bool                   `json:"requires_insurance"`
	InsuranceValue        float64                `json:"insurance_value"`
}

type CourierPackageDTO struct {
	Description       string                   `json:"description"`
	Category          models.PackageCategory   `json:"category"`
	Quantity          int                      `json:"quantity"`
	Weight            float64                  `json:"weight"`
	Dimensions        models.PackageDimensions `json:"dimensions"`
	Contents          []models.PackageItem     `json:"contents"`
	Value             float64                  `json:"value"`
	Currency          string                   `json:"currency"`
	IsFragile         bool                     `json:"is_fragile"`
	IsPerishable      bool                     `json:"is_perishable"`
	RequiresColdChain bool                     `json:"requires_cold_chain"`
	RequiresUpright   bool                     `json:"requires_upright"`
	SpecialHandling   []string                 `json:"special_handling"`
	RequiresSignature bool                     `json:"requires_signature"`
	RequiresID        bool                     `json:"requires_id"`
	AgeRestricted     bool                     `json:"age_restricted"`
	PackagingType     string                   `json:"packaging_type"`
}

// Implementation
type courierService struct {
	courierRepo     repositories.CourierRepository
	addressRepo     repositories.SavedAddressRepository
	issueRepo       repositories.CourierIssueRepository
	userRepo        repositories.UserRepository
	notificationSvc NotificationService
	pricingSvc      PricingService
	logger          utils.Logger
}

// Constructor
func NewCourierService(
	courierRepo repositories.CourierRepository,
	addressRepo repositories.SavedAddressRepository,
	issueRepo repositories.CourierIssueRepository,
	userRepo repositories.UserRepository,
	notificationSvc NotificationService,
	pricingSvc PricingService,
	logger utils.Logger,
) CourierService {
	return &courierService{
		courierRepo:     courierRepo,
		addressRepo:     addressRepo,
		issueRepo:       issueRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
		pricingSvc:      pricingSvc,
		logger:          logger,
	}
}

// Core Operations

func (s *courierService) CreateRequest(request *models.CourierRequest) (*models.CourierRequest, error) {
	// Validate request
	if err := s.ValidateRequest(request); err != nil {
		return nil, err
	}

	// Calculate volume
	if request.Package.Dimensions.Length > 0 && request.Package.Dimensions.Width > 0 && request.Package.Dimensions.Height > 0 {
		request.Package.Volume = request.Package.Dimensions.Length * request.Package.Dimensions.Width * request.Package.Dimensions.Height
	}

	// Set defaults
	request.Status = models.CourierStatusPending
	request.TrackingCode = s.generateTrackingCode()
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()

	// Create initial tracking event
	initialEvent := models.TrackingEvent{
		ID:          primitive.NewObjectID(),
		Status:      models.CourierStatusPending,
		Description: "Courier request created and waiting for courier assignment",
		Timestamp:   time.Now(),
		IsPublic:    true,
	}
	request.TrackingHistory = []models.TrackingEvent{initialEvent}

	// Save to repository
	createdRequest, err := s.courierRepo.Create(request)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create courier request")
		return nil, err
	}

	// Send notification to nearby couriers (async)
	go s.notifyNearbyCouriers(createdRequest)

	s.logger.Info().Str("request_id", createdRequest.ID.Hex()).Msg("Courier request created successfully")
	return createdRequest, nil
}

func (s *courierService) AcceptRequest(requestID, courierID primitive.ObjectID) error {
	// Get request
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	// Check if request is still available
	if request.Status != models.CourierStatusPending {
		return fmt.Errorf("request is no longer available")
	}

	// Assign courier
	err = s.courierRepo.AssignCourier(requestID, courierID)
	if err != nil {
		return err
	}

	// Add tracking event
	event := models.TrackingEvent{
		ID:          primitive.NewObjectID(),
		Status:      models.CourierStatusAccepted,
		Description: "Request accepted by courier",
		Timestamp:   time.Now(),
		CreatedBy:   &courierID,
		IsPublic:    true,
	}

	err = s.courierRepo.AddTrackingEvent(requestID, event)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to add tracking event")
	}

	// Notify sender
	go s.notificationSvc.NotifyRequestAccepted(request.SenderID, requestID, courierID)

	s.logger.Info().
		Str("request_id", requestID.Hex()).
		Str("courier_id", courierID.Hex()).
		Msg("Courier request accepted")

	return nil
}

func (s *courierService) CancelRequest(requestID, userID primitive.ObjectID, reason string) error {
	// Get request
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return fmt.Errorf("request not found")
	}

	// Check permissions
	if request.SenderID != userID && (request.CourierID == nil || *request.CourierID != userID) {
		return fmt.Errorf("unauthorized")
	}

	// Check if cancellation is allowed
	if request.Status == models.CourierStatusDelivered {
		return fmt.Errorf("cannot cancel delivered request")
	}

	// Cancel the request
	err = s.courierRepo.Cancel(requestID, reason, userID)
	if err != nil {
		return err
	}

	// Add tracking event
	event := models.TrackingEvent{
		ID:          primitive.NewObjectID(),
		Status:      models.CourierStatusCancelled,
		Description: fmt.Sprintf("Request cancelled: %s", reason),
		Timestamp:   time.Now(),
		CreatedBy:   &userID,
		IsPublic:    true,
	}

	err = s.courierRepo.AddTrackingEvent(requestID, event)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to add cancellation tracking event")
	}

	// Process refund if payment was made
	go s.processRefund(request)

	// Notify relevant parties
	go s.notifyRequestCancellation(request, userID, reason)

	s.logger.Info().
		Str("request_id", requestID.Hex()).
		Str("cancelled_by", userID.Hex()).
		Str("reason", reason).
		Msg("Courier request cancelled")

	return nil
}

// Package Management

func (s *courierService) UpdatePackageDetails(requestID primitive.ObjectID, packageDetails models.CourierPackage) error {
	// Calculate volume
	packageDetails.Volume = packageDetails.Dimensions.Length * packageDetails.Dimensions.Width * packageDetails.Dimensions.Height

	return s.courierRepo.UpdatePackageDetails(requestID, packageDetails)
}

func (s *courierService) AddPackagePhotos(requestID primitive.ObjectID, photoURLs []string) error {
	return s.courierRepo.AddPackagePhotos(requestID, photoURLs)
}

// Tracking & Status

func (s *courierService) UpdateStatus(requestID primitive.ObjectID, status models.CourierStatus, courierID primitive.ObjectID) error {
	// Update status
	err := s.courierRepo.UpdateStatus(requestID, status)
	if err != nil {
		return err
	}

	// Add tracking event
	event := models.TrackingEvent{
		ID:          primitive.NewObjectID(),
		Status:      status,
		Description: s.getStatusDescription(status),
		Timestamp:   time.Now(),
		CreatedBy:   &courierID,
		IsPublic:    true,
	}

	return s.courierRepo.AddTrackingEvent(requestID, event)
}

func (s *courierService) ConfirmPickup(requestID, courierID primitive.ObjectID, photos []string, notes string, location *models.Location) error {
	// Update status to picked up
	err := s.courierRepo.UpdateStatus(requestID, models.CourierStatusPickedUp)
	if err != nil {
		return err
	}

	// Add tracking event
	event := models.TrackingEvent{
		ID:          primitive.NewObjectID(),
		Status:      models.CourierStatusPickedUp,
		Description: "Package picked up by courier",
		Location:    location,
		Timestamp:   time.Now(),
		CreatedBy:   &courierID,
		Notes:       notes,
		Photos:      photos,
		IsPublic:    true,
	}

	err = s.courierRepo.AddTrackingEvent(requestID, event)
	if err != nil {
		return err
	}

	// Get request for notifications
	request, err := s.courierRepo.GetByID(requestID)
	if err == nil {
		// Notify sender and recipient
		go s.notificationSvc.NotifyPickupConfirmed(request.SenderID, requestID)
		go s.notifyRecipientOfPickup(request)
	}

	s.logger.Info().
		Str("request_id", requestID.Hex()).
		Str("courier_id", courierID.Hex()).
		Msg("Package pickup confirmed")

	return nil
}

func (s *courierService) ConfirmDelivery(requestID, courierID primitive.ObjectID, deliveryProof *models.DeliveryProof) error {
	// Update status to delivered
	err := s.courierRepo.UpdateStatus(requestID, models.CourierStatusDelivered)
	if err != nil {
		return err
	}

	// Set delivered time
	now := time.Now()
	err = s.courierRepo.MarkAsCompleted(requestID, now)
	if err != nil {
		return err
	}

	// Update delivery proof
	err = s.courierRepo.UpdateDeliveryProof(requestID, deliveryProof)
	if err != nil {
		return err
	}

	// Add tracking event
	event := models.TrackingEvent{
		ID:          primitive.NewObjectID(),
		Status:      models.CourierStatusDelivered,
		Description: fmt.Sprintf("Package delivered to %s", deliveryProof.DeliveredTo),
		Location:    &deliveryProof.DeliveryLocation,
		Timestamp:   now,
		CreatedBy:   &courierID,
		Photos:      deliveryProof.DeliveryPhotos,
		IsPublic:    true,
	}

	err = s.courierRepo.AddTrackingEvent(requestID, event)
	if err != nil {
		return err
	}

	// Get request for notifications
	request, err := s.courierRepo.GetByID(requestID)
	if err == nil {
		// Notify sender and recipient
		go s.notificationSvc.NotifyDeliveryCompleted(request.SenderID, requestID)
		go s.notifySuccessfulDelivery(request, deliveryProof)
	}

	s.logger.Info().
		Str("request_id", requestID.Hex()).
		Str("courier_id", courierID.Hex()).
		Str("delivered_to", deliveryProof.DeliveredTo).
		Msg("Package delivery confirmed")

	return nil
}

func (s *courierService) GetTrackingInfo(requestID primitive.ObjectID) (*TrackingInfo, error) {
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return nil, err
	}

	trackingInfo := &TrackingInfo{
		Request:        request,
		TrackingEvents: request.TrackingHistory,
		CurrentStatus:  request.Status,
		LastUpdate:     request.UpdatedAt,
	}

	// Add courier info if assigned
	if request.CourierID != nil {
		courier, err := s.userRepo.GetByID(request.CourierID.Hex())
		if err == nil {
			trackingInfo.CourierInfo = &CourierInfo{
				ID:     courier.ID,
				Name:   courier.Profile.FullName,
				Phone:  courier.Phone,
				Rating: courier.Rating,
				Photo:  courier.Profile.AvatarURL,
			}
		}
	}

	// Add current location if available
	if request.CurrentLocation != nil {
		trackingInfo.CurrentLocation = request.CurrentLocation
	}

	// Add estimated arrival
	if request.EstimatedArrival != nil {
		trackingInfo.EstimatedArrival = request.EstimatedArrival
	}

	return trackingInfo, nil
}

func (s *courierService) AddTrackingEvent(requestID primitive.ObjectID, event models.TrackingEvent) error {
	event.ID = primitive.NewObjectID()
	event.Timestamp = time.Now()
	return s.courierRepo.AddTrackingEvent(requestID, event)
}

func (s *courierService) UpdateLocation(requestID, courierID primitive.ObjectID, location models.Location) error {
	// Get request to verify courier
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	if request.CourierID == nil || *request.CourierID != courierID {
		return fmt.Errorf("unauthorized to update location")
	}

	// Update location and timestamp
	location.UpdatedAt = time.Now()
	request.CurrentLocation = &location

	// Calculate estimated arrival if in transit
	if request.Status == models.CourierStatusInTransit {
		estimatedArrival := s.calculateEstimatedArrival(location, request.DeliveryLocation.Location)
		request.EstimatedArrival = &estimatedArrival
	}

	_, err = s.courierRepo.Update(requestID, request)
	return err
}

// Recipient Management

func (s *courierService) UpdateRecipient(requestID primitive.ObjectID, recipient *models.CourierRecipient) error {
	return s.courierRepo.UpdateRecipient(requestID, *recipient)
}

func (s *courierService) NotifyRecipient(requestID primitive.ObjectID, message string, sendSMS, sendEmail, sendPush bool) error {
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	// Send notifications based on preferences
	if sendSMS && request.Recipient.Phone != "" {
		go s.notificationSvc.SendSMS(request.Recipient.Phone, message)
	}

	if sendEmail && request.Recipient.Email != "" {
		go s.notificationSvc.SendEmail(request.Recipient.Email, "Package Update", message)
	}

	if sendPush {
		// Send push notification if recipient has the app
		go s.notificationSvc.SendPushNotificationByContact(request.Recipient.Phone, message)
	}

	return nil
}

// Delivery Management

func (s *courierService) ScheduleDelivery(requestID primitive.ObjectID, scheduledTime time.Time, timeWindow *models.TimeWindow) error {
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	request.ScheduledDeliveryTime = &scheduledTime
	if timeWindow != nil {
		request.TimeWindow = timeWindow
	}
	request.UpdatedAt = time.Now()

	_, err = s.courierRepo.Update(requestID, request)
	return err
}

func (s *courierService) GetAvailableTimeSlots(date, location string) ([]TimeSlot, error) {
	// Parse date
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format")
	}

	// Generate time slots (example implementation)
	var slots []TimeSlot
	startHour := 9 // 9 AM
	endHour := 21  // 9 PM

	for hour := startHour; hour < endHour; hour += 2 {
		startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), hour, 0, 0, 0, targetDate.Location())
		endTime := startTime.Add(2 * time.Hour)

		// Check availability (simplified - would check against existing bookings)
		available := s.isTimeSlotAvailable(startTime, location)

		slot := TimeSlot{
			StartTime: startTime,
			EndTime:   endTime,
			Available: available,
		}

		// Add price premium for peak hours
		if hour >= 18 || hour <= 10 {
			slot.Price = 5.0 // Peak hour surcharge
		}

		slots = append(slots, slot)
	}

	return slots, nil
}

// Special Services

func (s *courierService) AddInsurance(requestID primitive.ObjectID, insurance *models.CourierInsurance) error {
	// Set insurance details
	insurance.ValidFrom = time.Now()
	insurance.ValidTo = time.Now().AddDate(0, 0, 30) // 30 days coverage
	insurance.PolicyNumber = s.generatePolicyNumber()

	return s.courierRepo.AddInsurance(requestID, insurance)
}

func (s *courierService) SetSignatureRequirement(requestID primitive.ObjectID, required bool, signatureType string) error {
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	request.Package.RequiresSignature = required
	request.UpdatedAt = time.Now()

	_, err = s.courierRepo.Update(requestID, request)
	return err
}

// Pricing & Fare

func (s *courierService) CalculateFare(request *models.CourierRequest) (*models.CourierFare, error) {
	// Calculate distance between pickup and delivery
	distance := s.calculateDistance(request.PickupLocation.Location, request.DeliveryLocation.Location)

	// Base fare by priority
	baseFares := map[models.DeliverySpeed]float64{
		models.DeliverySpeedStandard: 15.0,
		models.DeliverySpeedExpress:  25.0,
		models.DeliverySpeedPriority: 35.0,
		models.DeliverySpeedUrgent:   50.0,
	}

	baseFare := baseFares[request.Priority]

	// Distance fare
	distanceRates := map[models.DeliverySpeed]float64{
		models.DeliverySpeedStandard: 1.5,
		models.DeliverySpeedExpress:  2.0,
		models.DeliverySpeedPriority: 2.5,
		models.DeliverySpeedUrgent:   3.0,
	}

	distanceFare := distance * distanceRates[request.Priority]

	// Weight fare
	weightFare := request.Package.Weight * 2.0

	// Size surcharge
	sizeFare := 0.0
	volume := request.Package.Volume
	if volume > 27000 { // > 30cm x 30cm x 30cm
		sizeFare = 10.0
	}

	// Special handling surcharges
	var surcharges float64
	if request.Package.IsFragile {
		surcharges += 5.0
	}
	if request.Package.IsPerishable {
		surcharges += 8.0
	}
	if request.Package.RequiresColdChain {
		surcharges += 15.0
	}

	// Time-based surcharges
	now := time.Now()
	if now.Hour() >= 22 || now.Hour() <= 6 {
		surcharges += 20.0 // Night surcharge
	}
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		surcharges += 15.0 // Weekend surcharge
	}

	// Insurance fee
	insuranceFee := 0.0
	if request.Insurance != nil {
		insuranceFee = request.Insurance.Premium
	}

	// Calculate totals
	subTotal := baseFare + distanceFare + weightFare + sizeFare + surcharges + insuranceFee
	taxRate := 0.10 // 10% tax
	taxAmount := subTotal * taxRate
	totalAmount := subTotal + taxAmount

	// Commission calculation
	commissionRate := 0.20 // 20% commission
	commission := totalAmount * commissionRate
	courierEarnings := totalAmount - commission

	fare := &models.CourierFare{
		BaseFare:        baseFare,
		DistanceFare:    distanceFare,
		WeightFare:      weightFare,
		SizeFare:        sizeFare,
		SurchargeAmount: surcharges,
		InsuranceFee:    insuranceFee,
		SubTotal:        subTotal,
		TaxAmount:       taxAmount,
		TotalAmount:     totalAmount,
		CourierEarnings: courierEarnings,
		Commission:      commission,
		CommissionRate:  commissionRate,
		Currency:        "USD",
	}

	return fare, nil
}

func (s *courierService) EstimatePrice(req EstimatePriceRequest) (*PriceEstimate, error) {
	// Convert DTO to model for calculation
	courierRequest := &models.CourierRequest{
		Priority:         req.Priority,
		PickupLocation:   req.PickupLocation,
		DeliveryLocation: req.DeliveryLocation,
		Package: models.CourierPackage{
			Weight:            req.Package.Weight,
			Dimensions:        req.Package.Dimensions,
			IsFragile:         req.Package.IsFragile,
			IsPerishable:      req.Package.IsPerishable,
			RequiresColdChain: req.Package.RequiresColdChain,
		},
	}

	// Add insurance if required
	if req.RequiresInsurance {
		courierRequest.Insurance = &models.CourierInsurance{
			CoverageAmount: req.InsuranceValue,
			Premium:        req.InsuranceValue * 0.02, // 2% of value
		}
	}

	// Calculate fare
	fare, err := s.CalculateFare(courierRequest)
	if err != nil {
		return nil, err
	}

	// Build detailed breakdown
	surcharges := make(map[string]float64)
	if req.Package.IsFragile {
		surcharges["fragile"] = 5.0
	}
	if req.Package.IsPerishable {
		surcharges["perishable"] = 8.0
	}
	if req.Package.RequiresColdChain {
		surcharges["cold_chain"] = 15.0
	}

	breakdown := []PriceBreakdown{
		{Item: "Base Fare", Description: "Basic delivery fee", Amount: fare.BaseFare},
		{Item: "Distance", Description: "Per kilometer charge", Amount: fare.DistanceFare},
		{Item: "Weight", Description: "Per kilogram charge", Amount: fare.WeightFare},
		{Item: "Size", Description: "Package size surcharge", Amount: fare.SizeFare},
		{Item: "Insurance", Description: "Package insurance", Amount: fare.InsuranceFee},
		{Item: "Tax", Description: "Service tax", Amount: fare.TaxAmount},
	}

	estimate := &PriceEstimate{
		BaseFare:          fare.BaseFare,
		DistanceFare:      fare.DistanceFare,
		WeightFare:        fare.WeightFare,
		SizeFare:          fare.SizeFare,
		PriorityFare:      0, // Included in base fare
		Surcharges:        surcharges,
		InsuranceFee:      fare.InsuranceFee,
		TaxAmount:         fare.TaxAmount,
		SubTotal:          fare.SubTotal,
		TotalAmount:       fare.TotalAmount,
		Currency:          fare.Currency,
		BreakdownDetails:  breakdown,
		EstimatedDuration: s.getEstimatedDuration(req.Priority),
	}

	return estimate, nil
}

func (s *courierService) ApplyPromoCode(requestID primitive.ObjectID, promoCode string) error {
	// Get request
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	// Validate promo code (simplified)
	discount, err := s.validatePromoCode(promoCode, request.Fare.TotalAmount)
	if err != nil {
		return err
	}

	// Apply discount
	request.Fare.PromoCode = promoCode
	request.Fare.DiscountAmount = discount
	request.Fare.TotalAmount -= discount
	request.UpdatedAt = time.Now()

	_, err = s.courierRepo.Update(requestID, request)
	return err
}

// Delivery Proof

func (s *courierService) UpdateDeliveryProofPhotos(requestID primitive.ObjectID, photoURLs []string) error {
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	if request.DeliveryProof == nil {
		request.DeliveryProof = &models.DeliveryProof{}
	}

	request.DeliveryProof.DeliveryPhotos = append(request.DeliveryProof.DeliveryPhotos, photoURLs...)
	request.UpdatedAt = time.Now()

	_, err = s.courierRepo.Update(requestID, request)
	return err
}

// Address Management

func (s *courierService) GetAddressBook(userID primitive.ObjectID) ([]*models.SavedAddress, error) {
	return s.addressRepo.GetByUserID(userID)
}

func (s *courierService) AddAddress(userID primitive.ObjectID, address *models.SavedAddress) (*models.SavedAddress, error) {
	address.UserID = userID
	address.ID = primitive.NewObjectID()
	return s.addressRepo.Create(address)
}

func (s *courierService) UpdateAddress(userID, addressID primitive.ObjectID, address *models.SavedAddress) (*models.SavedAddress, error) {
	// Verify ownership
	existingAddress, err := s.addressRepo.GetByID(addressID)
	if err != nil {
		return nil, err
	}

	if existingAddress.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	address.UserID = userID
	return s.addressRepo.Update(addressID, address)
}

func (s *courierService) DeleteAddress(userID, addressID primitive.ObjectID) error {
	// Verify ownership
	existingAddress, err := s.addressRepo.GetByID(addressID)
	if err != nil {
		return err
	}

	if existingAddress.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	return s.addressRepo.Delete(addressID)
}

// Support & Issues

func (s *courierService) ReportIssue(requestID, reporterID primitive.ObjectID, issueType, description, priority string, photos []string) (primitive.ObjectID, error) {
	// For now, return a mock ID since CourierIssue model isn't fully defined
	issueID := primitive.NewObjectID()

	s.logger.Info().
		Str("request_id", requestID.Hex()).
		Str("reporter_id", reporterID.Hex()).
		Str("issue_type", issueType).
		Str("issue_id", issueID.Hex()).
		Msg("Issue reported")

	// In a real implementation, this would create an issue record
	// and trigger support workflows

	return issueID, nil
}

func (s *courierService) FileClaim(requestID, claimantID primitive.ObjectID, claimType string, amount float64, description string, evidence []string) (primitive.ObjectID, error) {
	// For now, return a mock ID since CourierClaim model isn't fully defined
	claimID := primitive.NewObjectID()

	s.logger.Info().
		Str("request_id", requestID.Hex()).
		Str("claimant_id", claimantID.Hex()).
		Str("claim_type", claimType).
		Float64("amount", amount).
		Str("claim_id", claimID.Hex()).
		Msg("Claim filed")

	// In a real implementation, this would create a claim record
	// and trigger claims processing workflows

	return claimID, nil
}

// Analytics & Stats

func (s *courierService) GetCourierStats(courierID primitive.ObjectID) (*repositories.CourierStats, error) {
	return s.courierRepo.GetCourierStats(courierID)
}

func (s *courierService) GetSenderStats(senderID primitive.ObjectID) (*repositories.SenderStats, error) {
	return s.courierRepo.GetSenderStats(senderID)
}

// Courier Assignment

func (s *courierService) FindNearbyRequests(courierID primitive.ObjectID, location models.Location, radius float64) ([]*models.CourierRequest, error) {
	return s.courierRepo.GetPendingRequests(location, radius)
}

func (s *courierService) AssignCourier(requestID, courierID primitive.ObjectID) error {
	return s.courierRepo.AssignCourier(requestID, courierID)
}

func (s *courierService) UnassignCourier(requestID primitive.ObjectID) error {
	return s.courierRepo.UnassignCourier(requestID)
}

// Validation

func (s *courierService) ValidateRequest(request *models.CourierRequest) error {
	if request.Package.Weight <= 0 {
		return fmt.Errorf("package weight must be greater than 0")
	}

	if request.Package.Weight > 20 {
		return fmt.Errorf("package weight cannot exceed 20kg for courier service")
	}

	if request.Package.Description == "" {
		return fmt.Errorf("package description is required")
	}

	if request.Recipient.Name == "" {
		return fmt.Errorf("recipient name is required")
	}

	if request.Recipient.Phone == "" && request.Recipient.Email == "" {
		return fmt.Errorf("recipient phone or email is required")
	}

	// Validate coordinates
	if len(request.PickupLocation.Coordinates) != 2 {
		return fmt.Errorf("invalid pickup location coordinates")
	}

	if len(request.DeliveryLocation.Coordinates) != 2 {
		return fmt.Errorf("invalid delivery location coordinates")
	}

	return nil
}

func (s *courierService) CanCancelRequest(requestID, userID primitive.ObjectID) (bool, error) {
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return false, err
	}

	// Sender can cancel before pickup
	if request.SenderID == userID && request.Status != models.CourierStatusPickedUp && request.Status != models.CourierStatusInTransit && request.Status != models.CourierStatusDelivered {
		return true, nil
	}

	// Courier can cancel before pickup
	if request.CourierID != nil && *request.CourierID == userID && request.Status == models.CourierStatusAccepted {
		return true, nil
	}

	return false, nil
}

func (s *courierService) CanUpdateRequest(requestID, userID primitive.ObjectID) (bool, error) {
	request, err := s.courierRepo.GetByID(requestID)
	if err != nil {
		return false, err
	}

	// Only sender can update and only if pending
	return request.SenderID == userID && request.Status == models.CourierStatusPending, nil
}

// Helper methods

func (s *courierService) generateTrackingCode() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("OD%d%04d", timestamp, time.Now().Nanosecond()%10000)
}

func (s *courierService) generatePolicyNumber() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("INS%d", timestamp)
}

func (s *courierService) getStatusDescription(status models.CourierStatus) string {
	descriptions := map[models.CourierStatus]string{
		models.CourierStatusPending:   "Request is pending courier assignment",
		models.CourierStatusAccepted:  "Request accepted by courier",
		models.CourierStatusPickedUp:  "Package picked up by courier",
		models.CourierStatusInTransit: "Package is in transit",
		models.CourierStatusDelivered: "Package delivered successfully",
		models.CourierStatusCancelled: "Request cancelled",
		models.CourierStatusReturned:  "Package returned to sender",
		models.CourierStatusFailed:    "Delivery failed",
	}
	return descriptions[status]
}

func (s *courierService) getEstimatedDuration(priority models.DeliverySpeed) string {
	durations := map[models.DeliverySpeed]string{
		models.DeliverySpeedStandard: "6-8 hours",
		models.DeliverySpeedExpress:  "2-4 hours",
		models.DeliverySpeedPriority: "1-2 hours",
		models.DeliverySpeedUrgent:   "Within 1 hour",
	}
	return durations[priority]
}

func (s *courierService) calculateDistance(pickup, delivery models.Location) float64 {
	// Simplified distance calculation (Haversine formula would be more accurate)
	// This is a placeholder - in production, use a proper geospatial library
	latDiff := pickup.Coordinates[1] - delivery.Coordinates[1]
	lonDiff := pickup.Coordinates[0] - delivery.Coordinates[0]
	return (latDiff*latDiff + lonDiff*lonDiff) * 111 // Rough conversion to km
}

func (s *courierService) calculateEstimatedArrival(currentLocation models.Location, destination models.Location) time.Time {
	distance := s.calculateDistance(currentLocation, destination)
	averageSpeed := 30.0 // 30 km/h average speed
	estimatedHours := distance / averageSpeed
	return time.Now().Add(time.Duration(estimatedHours * float64(time.Hour)))
}

func (s *courierService) isTimeSlotAvailable(startTime time.Time, location string) bool {
	// Simplified availability check - in production, check against existing bookings
	hour := startTime.Hour()
	// Peak hours (6-9 AM, 5-8 PM) are less available
	if (hour >= 6 && hour <= 9) || (hour >= 17 && hour <= 20) {
		return time.Now().Unix()%3 != 0 // Randomly 33% available
	}
	return time.Now().Unix()%4 != 0 // Randomly 75% available
}

func (s *courierService) validatePromoCode(promoCode string, totalAmount float64) (float64, error) {
	// Simplified promo code validation
	promoCodes := map[string]float64{
		"SAVE10": 10.0,
		"SAVE20": 20.0,
		"FIRST":  totalAmount * 0.15, // 15% discount
	}

	discount, exists := promoCodes[promoCode]
	if !exists {
		return 0, fmt.Errorf("invalid promo code")
	}

	// Cap discount at 50% of total
	maxDiscount := totalAmount * 0.5
	if discount > maxDiscount {
		discount = maxDiscount
	}

	return discount, nil
}

func (s *courierService) notifyNearbyCouriers(request *models.CourierRequest) {
	// Find nearby couriers and send notifications
	// This would typically query courier locations and send push notifications
	s.logger.Info().
		Str("request_id", request.ID.Hex()).
		Msg("Notifying nearby couriers of new request")
}

func (s *courierService) notifyRecipientOfPickup(request *models.CourierRequest) {
	message := fmt.Sprintf("Your package from %s has been picked up and is on the way! Tracking: %s",
		request.PickupLocation.Address, request.TrackingCode)

	if request.Recipient.Phone != "" {
		s.notificationSvc.SendSMS(request.Recipient.Phone, message)
	}
}

func (s *courierService) notifySuccessfulDelivery(request *models.CourierRequest, proof *models.DeliveryProof) {
	message := fmt.Sprintf("Your package has been delivered successfully to %s", proof.DeliveredTo)

	if request.Recipient.Phone != "" {
		s.notificationSvc.SendSMS(request.Recipient.Phone, message)
	}
}

func (s *courierService) notifyRequestCancellation(request *models.CourierRequest, cancelledBy primitive.ObjectID, reason string) {
	// Notify relevant parties about cancellation
	if request.CourierID != nil && *request.CourierID != cancelledBy {
		s.notificationSvc.NotifyRequestCancelled(*request.CourierID, request.ID, reason)
	}

	if request.SenderID != cancelledBy {
		s.notificationSvc.NotifyRequestCancelled(request.SenderID, request.ID, reason)
	}
}

func (s *courierService) processRefund(request *models.CourierRequest) {
	// Process refund logic based on cancellation timing and policy
	s.logger.Info().
		Str("request_id", request.ID.Hex()).
		Msg("Processing refund for cancelled request")
}
