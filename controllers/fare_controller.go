package controllers

import (
	"net/http"
	"strconv"

	"ondrive/middleware"
	"ondrive/models"
	"ondrive/services"
	"ondrive/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FareController handles fare-related HTTP requests
type FareController struct {
	fareService services.FareService
	logger      utils.Logger
}

// NewFareController creates a new fare controller
func NewFareController(fareService services.FareService) *FareController {
	return &FareController{
		fareService: fareService,
		logger:      utils.ControllerLogger("fare"),
	}
}

// Request structures
type EstimateFareRequest struct {
	PickupLocation    models.Location    `json:"pickup_location" binding:"required"`
	DropoffLocation   models.Location    `json:"dropoff_location" binding:"required"`
	ServiceType       models.ServiceType `json:"service_type" binding:"required"`
	VehicleType       models.VehicleType `json:"vehicle_type" binding:"required"`
	RequestedDateTime *string            `json:"requested_date_time,omitempty"`
	PromoCode         string             `json:"promo_code,omitempty"`
}

type CalculateFareRequest struct {
	RideID      primitive.ObjectID `json:"ride_id" binding:"required"`
	Distance    float64            `json:"distance" binding:"required"`
	Duration    int                `json:"duration" binding:"required"`
	WaitingTime int                `json:"waiting_time,omitempty"`
	TollsFee    float64            `json:"tolls_fee,omitempty"`
	PromoCode   string             `json:"promo_code,omitempty"`
	TipAmount   float64            `json:"tip_amount,omitempty"`
}

type ProposeFareRequest struct {
	RideID       primitive.ObjectID `json:"ride_id" binding:"required"`
	ProposedFare float64            `json:"proposed_fare" binding:"required"`
	Message      string             `json:"message,omitempty"`
}

type CounterOfferRequest struct {
	NegotiationID primitive.ObjectID `json:"negotiation_id" binding:"required"`
	CounterFare   float64            `json:"counter_fare" binding:"required"`
	Message       string             `json:"message,omitempty"`
}

type AcceptFareRequest struct {
	NegotiationID primitive.ObjectID `json:"negotiation_id" binding:"required"`
	OfferID       primitive.ObjectID `json:"offer_id" binding:"required"`
}

type RejectFareRequest struct {
	NegotiationID primitive.ObjectID `json:"negotiation_id" binding:"required"`
	OfferID       primitive.ObjectID `json:"offer_id" binding:"required"`
	Reason        string             `json:"reason,omitempty"`
}

type ApplyDiscountRequest struct {
	RideID    primitive.ObjectID `json:"ride_id" binding:"required"`
	PromoCode string             `json:"promo_code" binding:"required"`
}

// Fare Calculation

func (fc *FareController) EstimateFare(c *gin.Context) {
	var req EstimateFareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	estimate, err := fc.fareService.EstimateFare(userID, req.PickupLocation, req.DropoffLocation,
		req.ServiceType, req.VehicleType, req.RequestedDateTime, req.PromoCode)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to estimate fare")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare estimated successfully", estimate)
}

func (fc *FareController) CalculateFare(c *gin.Context) {
	var req CalculateFareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	fare, err := fc.fareService.CalculateFare(userID, req.RideID, req.Distance, req.Duration,
		req.WaitingTime, req.TollsFee, req.PromoCode, req.TipAmount)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to calculate fare")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare calculated successfully", fare)
}

func (fc *FareController) GetBaseRates(c *gin.Context) {
	serviceType := c.Query("service_type")
	vehicleType := c.Query("vehicle_type")
	city := c.Query("city")
	country := c.Query("country")

	rates, err := fc.fareService.GetBaseRates(serviceType, vehicleType, city, country)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get base rates")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Base rates retrieved successfully", rates)
}

func (fc *FareController) GetSurgeInfo(c *gin.Context) {
	lng, _ := strconv.ParseFloat(c.Query("longitude"), 64)
	lat, _ := strconv.ParseFloat(c.Query("latitude"), 64)
	serviceType := c.Query("service_type")

	if lat == 0 || lng == 0 {
		utils.BadRequestResponse(c, "Valid coordinates are required")
		return
	}

	location := models.Location{
		Type:        "Point",
		Coordinates: []float64{lng, lat}, // GeoJSON format: [longitude, latitude]
	}

	surgeInfo, err := fc.fareService.GetSurgeInfo(location, serviceType)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get surge info")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Surge information retrieved successfully", surgeInfo)
}

// Fare Negotiation

func (fc *FareController) ProposeFare(c *gin.Context) {
	var req ProposeFareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	negotiation, err := fc.fareService.ProposeFare(userID, req.RideID, req.ProposedFare, req.Message)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to propose fare")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Fare proposed successfully", negotiation)
}

func (fc *FareController) CounterOffer(c *gin.Context) {
	var req CounterOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	negotiation, err := fc.fareService.CounterOffer(userID, req.NegotiationID, req.CounterFare, req.Message)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to make counter offer")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Counter offer made successfully", negotiation)
}

func (fc *FareController) AcceptFare(c *gin.Context) {
	var req AcceptFareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	negotiation, err := fc.fareService.AcceptFare(userID, req.NegotiationID, req.OfferID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to accept fare")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare accepted successfully", negotiation)
}

func (fc *FareController) RejectFare(c *gin.Context) {
	var req RejectFareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	negotiation, err := fc.fareService.RejectFare(userID, req.NegotiationID, req.OfferID, req.Reason)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to reject fare")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare rejected successfully", negotiation)
}

func (fc *FareController) GetNegotiationHistory(c *gin.Context) {
	rideID := c.Param("ride_id")
	objID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid ride ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	history, err := fc.fareService.GetNegotiationHistory(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get negotiation history")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Negotiation history retrieved successfully", history)
}

// Fare Comparison

func (fc *FareController) CompareFares(c *gin.Context) {
	lng, _ := strconv.ParseFloat(c.Query("pickup_lng"), 64)
	lat, _ := strconv.ParseFloat(c.Query("pickup_lat"), 64)
	destLng, _ := strconv.ParseFloat(c.Query("dropoff_lng"), 64)
	destLat, _ := strconv.ParseFloat(c.Query("dropoff_lat"), 64)

	if lat == 0 || lng == 0 || destLat == 0 || destLng == 0 {
		utils.BadRequestResponse(c, "Valid pickup and dropoff coordinates are required")
		return
	}

	pickup := models.Location{
		Type:        "Point",
		Coordinates: []float64{lng, lat}, // GeoJSON format: [longitude, latitude]
	}
	dropoff := models.Location{
		Type:        "Point",
		Coordinates: []float64{destLng, destLat},
	}

	comparison, err := fc.fareService.CompareFares(pickup, dropoff)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to compare fares")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare comparison retrieved successfully", comparison)
}

func (fc *FareController) GetMarketRates(c *gin.Context) {
	serviceType := c.Query("service_type")
	city := c.Query("city")
	country := c.Query("country")

	rates, err := fc.fareService.GetMarketRates(serviceType, city, country)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get market rates")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Market rates retrieved successfully", rates)
}

func (fc *FareController) GetSuggestedFare(c *gin.Context) {
	lng, _ := strconv.ParseFloat(c.Query("pickup_lng"), 64)
	lat, _ := strconv.ParseFloat(c.Query("pickup_lat"), 64)
	destLng, _ := strconv.ParseFloat(c.Query("dropoff_lng"), 64)
	destLat, _ := strconv.ParseFloat(c.Query("dropoff_lat"), 64)
	serviceType := c.Query("service_type")

	if lat == 0 || lng == 0 || destLat == 0 || destLng == 0 {
		utils.BadRequestResponse(c, "Valid pickup and dropoff coordinates are required")
		return
	}

	pickup := models.Location{
		Type:        "Point",
		Coordinates: []float64{lng, lat}, // GeoJSON format: [longitude, latitude]
	}
	dropoff := models.Location{
		Type:        "Point",
		Coordinates: []float64{destLng, destLat},
	}

	suggested, err := fc.fareService.GetSuggestedFare(pickup, dropoff, serviceType)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get suggested fare")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Suggested fare retrieved successfully", suggested)
}

// Fare Rules & Settings

func (fc *FareController) GetFareRules(c *gin.Context) {
	serviceType := c.Query("service_type")
	city := c.Query("city")
	country := c.Query("country")

	rules, err := fc.fareService.GetFareRules(serviceType, city, country)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get fare rules")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare rules retrieved successfully", rules)
}

func (fc *FareController) GetMinimumFare(c *gin.Context) {
	serviceType := c.Query("service_type")
	city := c.Query("city")

	minFare, err := fc.fareService.GetMinimumFare(serviceType, city)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get minimum fare")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Minimum fare retrieved successfully", gin.H{
		"minimum_fare": minFare,
		"service_type": serviceType,
		"city":         city,
	})
}

func (fc *FareController) GetMaximumFare(c *gin.Context) {
	serviceType := c.Query("service_type")
	city := c.Query("city")

	maxFare, err := fc.fareService.GetMaximumFare(serviceType, city)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get maximum fare")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Maximum fare retrieved successfully", gin.H{
		"maximum_fare": maxFare,
		"service_type": serviceType,
		"city":         city,
	})
}

// Fare History & Analytics

func (fc *FareController) GetFareHistory(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	history, total, err := fc.fareService.GetFareHistory(userID, page, limit, startDate, endDate)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get fare history")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(page, limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Fare history retrieved successfully", history, meta)
}

func (fc *FareController) GetFareStatistics(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	period := c.DefaultQuery("period", "month") // day, week, month, year

	stats, err := fc.fareService.GetFareStatistics(userID, period)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get fare statistics")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare statistics retrieved successfully", stats)
}

func (fc *FareController) GetFareTrends(c *gin.Context) {
	serviceType := c.Query("service_type")
	city := c.Query("city")
	period := c.DefaultQuery("period", "month")

	trends, err := fc.fareService.GetFareTrends(serviceType, city, period)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get fare trends")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare trends retrieved successfully", trends)
}

// Special Pricing

func (fc *FareController) GetPromotionalPricing(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	serviceType := c.Query("service_type")
	city := c.Query("city")

	promotions, err := fc.fareService.GetPromotionalPricing(userID, serviceType, city)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get promotional pricing")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Promotional pricing retrieved successfully", promotions)
}

func (fc *FareController) GetDiscountCodes(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	discounts, err := fc.fareService.GetDiscountCodes(userID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get discount codes")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Discount codes retrieved successfully", discounts)
}

func (fc *FareController) ApplyDiscountCode(c *gin.Context) {
	var req ApplyDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	result, err := fc.fareService.ApplyDiscountCode(userID, req.RideID, req.PromoCode)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to apply discount code")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Discount code applied successfully", result)
}

// Commission & Fees

func (fc *FareController) GetCommissionRates(c *gin.Context) {
	serviceType := c.Query("service_type")
	city := c.Query("city")

	rates, err := fc.fareService.GetCommissionRates(serviceType, city)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get commission rates")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Commission rates retrieved successfully", rates)
}

func (fc *FareController) GetServiceFees(c *gin.Context) {
	serviceType := c.Query("service_type")
	city := c.Query("city")

	fees, err := fc.fareService.GetServiceFees(serviceType, city)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get service fees")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Service fees retrieved successfully", fees)
}

func (fc *FareController) GetFareBreakdown(c *gin.Context) {
	rideID := c.Param("ride_id")
	objID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid ride ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	breakdown, err := fc.fareService.GetFareBreakdown(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get fare breakdown")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare breakdown retrieved successfully", breakdown)
}
