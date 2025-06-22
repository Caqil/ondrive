package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FareService interface defines fare-related business logic
type FareService interface {
	// Fare Calculation
	EstimateFare(userID string, pickup, dropoff models.Location, serviceType models.ServiceType,
		vehicleType models.VehicleType, requestedDateTime *string, promoCode string) (*models.FareEstimate, error)
	CalculateFare(userID string, rideID primitive.ObjectID, distance float64, duration int,
		waitingTime int, tollsFee float64, promoCode string, tipAmount float64) (*models.FareDetails, error)
	GetBaseRates(serviceType, vehicleType, city, country string) ([]*models.RateCard, error)
	GetSurgeInfo(location models.Location, serviceType string) (*models.DynamicPricing, error)

	// Fare Negotiation
	ProposeFare(userID string, rideID primitive.ObjectID, proposedFare float64, message string) (*models.FareNegotiation, error)
	CounterOffer(userID string, negotiationID primitive.ObjectID, counterFare float64, message string) (*models.FareNegotiation, error)
	AcceptFare(userID string, negotiationID primitive.ObjectID, offerID primitive.ObjectID) (*models.FareNegotiation, error)
	RejectFare(userID string, negotiationID primitive.ObjectID, offerID primitive.ObjectID, reason string) (*models.FareNegotiation, error)
	GetNegotiationHistory(userID string, rideID primitive.ObjectID) ([]*models.FareNegotiation, error)

	// Fare Comparison
	CompareFares(pickup, dropoff models.Location) (*models.FareComparison, error)
	GetMarketRates(serviceType, city, country string) (*models.MarketRate, error)
	GetSuggestedFare(pickup, dropoff models.Location, serviceType string) (*models.SuggestedFare, error)

	// Fare Rules & Settings
	GetFareRules(serviceType, city, country string) ([]*models.FareRule, error)
	GetMinimumFare(serviceType, city string) (float64, error)
	GetMaximumFare(serviceType, city string) (float64, error)

	// Fare History & Analytics
	GetFareHistory(userID string, page, limit int, startDate, endDate string) ([]*models.FareDetails, int64, error)
	GetFareStatistics(userID string, period string) (*models.FareStatistics, error)
	GetFareTrends(serviceType, city, period string) ([]*models.FareTrend, error)

	// Special Pricing
	GetPromotionalPricing(userID, serviceType, city string) ([]*models.PromotionalRate, error)
	GetDiscountCodes(userID string) ([]*models.DiscountCode, error)
	ApplyDiscountCode(userID string, rideID primitive.ObjectID, promoCode string) (*models.DiscountResult, error)

	// Commission & Fees
	GetCommissionRates(serviceType, city string) (*models.CommissionRate, error)
	GetServiceFees(serviceType, city string) (*models.ServiceFee, error)
	GetFareBreakdown(userID string, rideID primitive.ObjectID) (*models.FareBreakdown, error)
}

// fareService implements FareService interface
type fareService struct {
	fareRepo        repositories.FareRepository
	rideRepo        repositories.RideRepository
	userRepo        repositories.UserRepository
	locationService LocationService
	logger          utils.Logger
}

// NewFareService creates a new fare service
func NewFareService(
	fareRepo repositories.FareRepository,
	rideRepo repositories.RideRepository,
	userRepo repositories.UserRepository,
	locationService LocationService,
) FareService {
	return &fareService{
		fareRepo:        fareRepo,
		rideRepo:        rideRepo,
		userRepo:        userRepo,
		locationService: locationService,
		logger:          utils.ServiceLogger("fare"),
	}
}

// Fare Calculation

func (s *fareService) EstimateFare(userID string, pickup, dropoff models.Location,
	serviceType models.ServiceType, vehicleType models.VehicleType,
	requestedDateTime *string, promoCode string) (*models.FareEstimate, error) {

	// Get distance and duration
	distance, duration, err := s.locationService.CalculateRoute(pickup, dropoff)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate route: %w", err)
	}

	// Get rate card for the service
	rateCard, err := s.fareRepo.GetRateCard(string(serviceType), string(vehicleType), pickup)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate card: %w", err)
	}

	// Calculate base fare
	baseFare := rateCard.BaseFare
	distanceFare := distance * rateCard.PerKmRate
	timeFare := float64(duration) * rateCard.PerMinuteRate

	// Calculate surge pricing
	surge, err := s.calculateSurgePricing(pickup, serviceType)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to calculate surge pricing")
		surge = &models.DynamicPricing{IsActive: false, Multiplier: 1.0}
	}

	// Apply surge if active
	totalBeforeSurge := baseFare + distanceFare + timeFare
	surgeFare := 0.0
	if surge.IsActive {
		surgeFare = totalBeforeSurge * (surge.Multiplier - 1.0)
	}

	// Calculate service fee
	serviceFee := totalBeforeSurge * rateCard.ServiceFeeRate

	// Calculate subtotal
	subtotal := totalBeforeSurge + surgeFare + serviceFee

	// Apply promo code discount if provided
	discountAmount := 0.0
	if promoCode != "" {
		discount, err := s.calculateDiscount(userID, promoCode, subtotal)
		if err == nil {
			discountAmount = discount
		}
	}

	// Calculate final total
	total := math.Max(subtotal-discountAmount, rateCard.MinimumFare)

	estimate := &models.FareEstimate{
		ServiceType:    serviceType,
		VehicleType:    vehicleType,
		Distance:       distance,
		Duration:       duration,
		BaseFare:       baseFare,
		DistanceFare:   distanceFare,
		TimeFare:       timeFare,
		SurgeFare:      surgeFare,
		ServiceFee:     serviceFee,
		DiscountAmount: discountAmount,
		PromoCode:      promoCode,
		Subtotal:       subtotal,
		Total:          total,
		Currency:       "USD",
		EstimatedAt:    time.Now(),
		ValidUntil:     time.Now().Add(10 * time.Minute),
	}

	return estimate, nil
}

func (s *fareService) CalculateFare(userID string, rideID primitive.ObjectID,
	distance float64, duration int, waitingTime int, tollsFee float64,
	promoCode string, tipAmount float64) (*models.FareDetails, error) {

	// Get ride details
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, fmt.Errorf("ride not found: %w", err)
	}

	// Verify user has access to this ride
	userObjID, _ := primitive.ObjectIDFromHex(userID)
	if ride.PassengerID != userObjID && ride.DriverID != userObjID {
		return nil, errors.New("unauthorized access to ride")
	}

	// Get rate card
	rateCard, err := s.fareRepo.GetRateCard(string(ride.ServiceType), string(ride.VehicleType), ride.PickupLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate card: %w", err)
	}

	// Calculate fare components
	baseFare := rateCard.BaseFare
	distanceFare := distance * rateCard.PerKmRate
	timeFare := float64(duration) * rateCard.PerMinuteRate

	// Calculate waiting fee
	waitingFee := 0.0
	if waitingTime > rateCard.FreeWaitingTime {
		chargeableWaitingTime := waitingTime - rateCard.FreeWaitingTime
		waitingFee = float64(chargeableWaitingTime) * rateCard.WaitingTimeRate
	}

	// Apply peak hour surcharge if applicable
	peakHourFare := s.calculatePeakHourSurcharge(rateCard, time.Now())

	// Calculate subtotal before fees
	subtotal := baseFare + distanceFare + timeFare + waitingFee + tollsFee + peakHourFare

	// Calculate service fee and commission
	serviceFee := subtotal * rateCard.ServiceFeeRate
	commission := subtotal * rateCard.CommissionRate
	driverEarnings := subtotal - commission + tipAmount

	// Apply discount if promo code provided
	discountAmount := 0.0
	if promoCode != "" {
		discount, err := s.calculateDiscount(userID, promoCode, subtotal+serviceFee)
		if err == nil {
			discountAmount = discount
		}
	}

	// Calculate final total
	totalFare := math.Max(subtotal+serviceFee-discountAmount+tipAmount, rateCard.MinimumFare)

	fareDetails := &models.FareDetails{
		ProposedFare:    totalFare,
		FinalFare:       totalFare,
		Currency:        "USD",
		Status:          models.FareStatusAccepted,
		BaseFare:        baseFare,
		DistanceFare:    distanceFare,
		TimeFare:        timeFare,
		WaitingFee:      waitingFee,
		TollsFee:        tollsFee,
		PeakHourFare:    peakHourFare,
		ServiceFee:      serviceFee,
		DiscountAmount:  discountAmount,
		PromoCode:       promoCode,
		TipAmount:       tipAmount,
		Commission:      commission,
		CommissionRate:  rateCard.CommissionRate,
		DriverEarnings:  driverEarnings,
		PlatformFee:     serviceFee,
		CalculationType: models.FareTypeDistance,
		CalculatedAt:    time.Now(),
	}

	// Save fare details
	err = s.fareRepo.SaveFareDetails(rideID, fareDetails)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to save fare details")
	}

	return fareDetails, nil
}

func (s *fareService) GetBaseRates(serviceType, vehicleType, city, country string) ([]*models.RateCard, error) {
	return s.fareRepo.GetRateCards(serviceType, vehicleType, city, country)
}

func (s *fareService) GetSurgeInfo(location models.Location, serviceType string) (*models.DynamicPricing, error) {
	return s.calculateSurgePricing(location, models.ServiceType(serviceType))
}

// Fare Negotiation

func (s *fareService) ProposeFare(userID string, rideID primitive.ObjectID,
	proposedFare float64, message string) (*models.FareNegotiation, error) {

	// Get ride details
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, fmt.Errorf("ride not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)

	// Check if user is passenger for this ride
	if ride.PassengerID != userObjID {
		return nil, errors.New("only passengers can propose fares")
	}

	// Check if ride is in correct status for negotiation
	if ride.Status != models.RideStatusPending {
		return nil, errors.New("ride is not available for fare negotiation")
	}

	// Check for existing active negotiations
	existing, err := s.fareRepo.GetActiveNegotiation(rideID)
	if err == nil && existing != nil {
		return nil, errors.New("active negotiation already exists for this ride")
	}

	// Validate proposed fare
	minFare, err := s.GetMinimumFare(string(ride.ServiceType), "")
	if err == nil && proposedFare < minFare {
		return nil, fmt.Errorf("proposed fare below minimum of %.2f", minFare)
	}

	// Create negotiation
	negotiation := &models.FareNegotiation{
		ID:          primitive.NewObjectID(),
		RideID:      rideID,
		PassengerID: userObjID,
		DriverID:    ride.DriverID,
		Status:      models.FareStatusProposed,
		InitialFare: proposedFare,
		FinalFare:   proposedFare,
		Currency:    "USD",
		StartedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(15 * time.Minute), // 15 minute expiry
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add initial offer
	offer := models.FareOffer{
		ID:        primitive.NewObjectID(),
		OfferBy:   userObjID,
		OfferTo:   ride.DriverID,
		Amount:    proposedFare,
		Message:   message,
		OfferedAt: time.Now(),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	negotiation.Offers = append(negotiation.Offers, offer)
	negotiation.TotalOffers = 1

	// Save negotiation
	return s.fareRepo.CreateNegotiation(negotiation)
}

func (s *fareService) CounterOffer(userID string, negotiationID primitive.ObjectID,
	counterFare float64, message string) (*models.FareNegotiation, error) {

	// Get negotiation
	negotiation, err := s.fareRepo.GetNegotiation(negotiationID)
	if err != nil {
		return nil, fmt.Errorf("negotiation not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)

	// Check if user is part of this negotiation
	if negotiation.PassengerID != userObjID && negotiation.DriverID != userObjID {
		return nil, errors.New("unauthorized access to negotiation")
	}

	// Check if negotiation is still active
	if negotiation.Status != models.FareStatusProposed && negotiation.Status != models.FareStatusCountered {
		return nil, errors.New("negotiation is not active")
	}

	// Check if negotiation has expired
	if time.Now().After(negotiation.ExpiresAt) {
		return nil, errors.New("negotiation has expired")
	}

	// Determine offer recipient
	var offerTo primitive.ObjectID
	if userObjID == negotiation.PassengerID {
		offerTo = negotiation.DriverID
	} else {
		offerTo = negotiation.PassengerID
	}

	// Create counter offer
	offer := models.FareOffer{
		ID:          primitive.NewObjectID(),
		OfferBy:     userObjID,
		OfferTo:     offerTo,
		Amount:      counterFare,
		Message:     message,
		OfferedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(15 * time.Minute),
		IsCountered: true,
	}

	// Update negotiation
	negotiation.Offers = append(negotiation.Offers, offer)
	negotiation.TotalOffers++
	negotiation.Status = models.FareStatusCountered
	negotiation.FinalFare = counterFare
	negotiation.UpdatedAt = time.Now()

	return s.fareRepo.UpdateNegotiation(negotiation)
}

func (s *fareService) AcceptFare(userID string, negotiationID primitive.ObjectID,
	offerID primitive.ObjectID) (*models.FareNegotiation, error) {

	// Get negotiation
	negotiation, err := s.fareRepo.GetNegotiation(negotiationID)
	if err != nil {
		return nil, fmt.Errorf("negotiation not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)

	// Find the offer
	var targetOffer *models.FareOffer
	for i := range negotiation.Offers {
		if negotiation.Offers[i].ID == offerID {
			targetOffer = &negotiation.Offers[i]
			break
		}
	}

	if targetOffer == nil {
		return nil, errors.New("offer not found")
	}

	// Check if user is the intended recipient
	if targetOffer.OfferTo != userObjID {
		return nil, errors.New("unauthorized to accept this offer")
	}

	// Accept the offer
	targetOffer.IsAccepted = true
	targetOffer.ResponseAt = &[]time.Time{time.Now()}[0]

	// Update negotiation
	negotiation.Status = models.FareStatusAccepted
	negotiation.FinalFare = targetOffer.Amount
	negotiation.AcceptedBy = &userObjID
	negotiation.CompletedAt = &[]time.Time{time.Now()}[0]
	negotiation.UpdatedAt = time.Now()

	// Update ride with accepted fare
	err = s.rideRepo.UpdateFare(negotiation.RideID, targetOffer.Amount)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to update ride fare")
	}

	return s.fareRepo.UpdateNegotiation(negotiation)
}

func (s *fareService) RejectFare(userID string, negotiationID primitive.ObjectID,
	offerID primitive.ObjectID, reason string) (*models.FareNegotiation, error) {

	// Get negotiation
	negotiation, err := s.fareRepo.GetNegotiation(negotiationID)
	if err != nil {
		return nil, fmt.Errorf("negotiation not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)

	// Find the offer
	var targetOffer *models.FareOffer
	for i := range negotiation.Offers {
		if negotiation.Offers[i].ID == offerID {
			targetOffer = &negotiation.Offers[i]
			break
		}
	}

	if targetOffer == nil {
		return nil, errors.New("offer not found")
	}

	// Check if user is the intended recipient
	if targetOffer.OfferTo != userObjID {
		return nil, errors.New("unauthorized to reject this offer")
	}

	// Reject the offer
	targetOffer.IsRejected = true
	targetOffer.Response = reason
	targetOffer.ResponseAt = &[]time.Time{time.Now()}[0]

	// Update negotiation
	negotiation.Status = models.FareStatusRejected
	negotiation.RejectedBy = &userObjID
	negotiation.RejectionReason = reason
	negotiation.CompletedAt = &[]time.Time{time.Now()}[0]
	negotiation.UpdatedAt = time.Now()

	return s.fareRepo.UpdateNegotiation(negotiation)
}

func (s *fareService) GetNegotiationHistory(userID string, rideID primitive.ObjectID) ([]*models.FareNegotiation, error) {
	userObjID, _ := primitive.ObjectIDFromHex(userID)
	return s.fareRepo.GetNegotiationHistory(rideID, userObjID)
}

// Helper methods

func (s *fareService) calculateSurgePricing(location models.Location, serviceType models.ServiceType) (*models.DynamicPricing, error) {
	// This is a simplified surge calculation
	// In reality, this would consider demand/supply, weather, events, etc.

	surge := &models.DynamicPricing{
		IsActive:     false,
		Multiplier:   1.0,
		Reason:       "Normal pricing",
		DemandLevel:  "medium",
		SupplyLevel:  "medium",
		CalculatedAt: time.Now(),
		ExpiresAt:    time.Now().Add(15 * time.Minute),
	}

	// Simple rules for demonstration
	hour := time.Now().Hour()
	if hour >= 7 && hour <= 9 || hour >= 17 && hour <= 19 {
		// Peak hours
		surge.IsActive = true
		surge.Multiplier = 1.5
		surge.Reason = "Peak hour pricing"
		surge.DemandLevel = "high"
	}

	return surge, nil
}

func (s *fareService) calculatePeakHourSurcharge(rateCard *models.RateCard, rideTime time.Time) float64 {
	// Check if current time falls within peak hours
	for _, peakHour := range rateCard.PeakHours {
		if int(rideTime.Weekday()) == peakHour.DayOfWeek {
			// Parse time and check if within range
			// Simplified implementation
			return rateCard.BaseFare * (peakHour.Multiplier - 1.0)
		}
	}
	return 0.0
}

func (s *fareService) calculateDiscount(userID, promoCode string, amount float64) (float64, error) {
	// Get promo code details
	promo, err := s.fareRepo.GetPromoCode(promoCode)
	if err != nil {
		return 0, err
	}

	// Check if user has already used this promo
	used, err := s.fareRepo.HasUserUsedPromo(userID, promoCode)
	if err != nil {
		return 0, err
	}

	if used && promo.FirstRideOnly {
		return 0, errors.New("promo code already used")
	}

	// Check minimum fare requirement
	if amount < promo.MinFareAmount {
		return 0, fmt.Errorf("minimum fare of %.2f required", promo.MinFareAmount)
	}

	// Calculate discount
	var discount float64
	switch promo.DiscountType {
	case "percentage":
		discount = amount * (promo.DiscountValue / 100.0)
		if promo.MaxDiscount > 0 {
			discount = math.Min(discount, promo.MaxDiscount)
		}
	case "fixed":
		discount = promo.DiscountValue
	default:
		return 0, errors.New("invalid discount type")
	}

	return discount, nil
}

// Additional implementations for the remaining interface methods would follow similar patterns...

func (s *fareService) CompareFares(pickup, dropoff models.Location) (*models.FareComparison, error) {
	// Implementation for fare comparison across different service types
	return nil, errors.New("not implemented")
}

func (s *fareService) GetMarketRates(serviceType, city, country string) (*models.MarketRate, error) {
	// Implementation for getting market rates
	return nil, errors.New("not implemented")
}

func (s *fareService) GetSuggestedFare(pickup, dropoff models.Location, serviceType string) (*models.SuggestedFare, error) {
	// Implementation for suggested fare calculation
	return nil, errors.New("not implemented")
}

func (s *fareService) GetFareRules(serviceType, city, country string) ([]*models.FareRule, error) {
	// Implementation for getting fare rules
	return nil, errors.New("not implemented")
}

func (s *fareService) GetMinimumFare(serviceType, city string) (float64, error) {
	rateCard, err := s.fareRepo.GetRateCardByType(serviceType, city)
	if err != nil {
		return 0, err
	}
	return rateCard.MinimumFare, nil
}

func (s *fareService) GetMaximumFare(serviceType, city string) (float64, error) {
	rateCard, err := s.fareRepo.GetRateCardByType(serviceType, city)
	if err != nil {
		return 0, err
	}
	return rateCard.MaximumFare, nil
}

func (s *fareService) GetFareHistory(userID string, page, limit int, startDate, endDate string) ([]*models.FareDetails, int64, error) {
	// Implementation for fare history
	return nil, 0, errors.New("not implemented")
}

func (s *fareService) GetFareStatistics(userID string, period string) (*models.FareStatistics, error) {
	// Implementation for fare statistics
	return nil, errors.New("not implemented")
}

func (s *fareService) GetFareTrends(serviceType, city, period string) ([]*models.FareTrend, error) {
	// Implementation for fare trends
	return nil, errors.New("not implemented")
}

func (s *fareService) GetPromotionalPricing(userID, serviceType, city string) ([]*models.PromotionalRate, error) {
	// Implementation for promotional pricing
	return nil, errors.New("not implemented")
}

func (s *fareService) GetDiscountCodes(userID string) ([]*models.DiscountCode, error) {
	// Implementation for getting discount codes
	return nil, errors.New("not implemented")
}

func (s *fareService) ApplyDiscountCode(userID string, rideID primitive.ObjectID, promoCode string) (*models.DiscountResult, error) {
	// Implementation for applying discount codes
	return nil, errors.New("not implemented")
}

func (s *fareService) GetCommissionRates(serviceType, city string) (*models.CommissionRate, error) {
	// Implementation for getting commission rates
	return nil, errors.New("not implemented")
}

func (s *fareService) GetServiceFees(serviceType, city string) (*models.ServiceFee, error) {
	// Implementation for getting service fees
	return nil, errors.New("not implemented")
}

func (s *fareService) GetFareBreakdown(userID string, rideID primitive.ObjectID) (*models.FareBreakdown, error) {
	// Implementation for fare breakdown
	return nil, errors.New("not implemented")
}
