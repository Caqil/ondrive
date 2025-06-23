package controllers

import (
	"net/http"
	"strconv"
	"time"

	"ondrive/middleware"
	"ondrive/models"
	"ondrive/services"
	"ondrive/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RideController handles ride-related HTTP requests
type RideController struct {
	rideService services.RideService
	logger      utils.Logger
}

// Request types for controller
type CreateRideControllerRequest struct {
	Type                models.RideType         `json:"type" validate:"required"`
	ServiceType         models.ServiceType      `json:"service_type" validate:"required"`
	PickupLocation      models.RideLocation     `json:"pickup_location" validate:"required"`
	DropoffLocation     models.RideLocation     `json:"dropoff_location" validate:"required"`
	Waypoints           []models.RideLocation   `json:"waypoints"`
	Requirements        models.RideRequirements `json:"requirements"`
	Preferences         models.RidePreferences  `json:"preferences"`
	ScheduledAt         *time.Time              `json:"scheduled_at,omitempty"`
	PaymentMethodID     string                  `json:"payment_method_id" validate:"required"`
	ProposedFare        *float64                `json:"proposed_fare,omitempty"`
	Notes               string                  `json:"notes"`
	SpecialInstructions string                  `json:"special_instructions"`
	Platform            string                  `json:"platform"`
}

type EstimateFareControllerRequest struct {
	ServiceType     models.ServiceType      `json:"service_type" validate:"required"`
	PickupLocation  models.RideLocation     `json:"pickup_location" validate:"required"`
	DropoffLocation models.RideLocation     `json:"dropoff_location" validate:"required"`
	Waypoints       []models.RideLocation   `json:"waypoints"`
	Requirements    models.RideRequirements `json:"requirements"`
	ScheduledAt     *time.Time              `json:"scheduled_at,omitempty"`
}

type NegotiateFareRequest struct {
	Amount  float64 `json:"amount" validate:"required,gt=0"`
	Message string  `json:"message"`
}

type CompleteRideRequest struct {
	ActualDistance float64          `json:"actual_distance" validate:"required,gt=0"`
	ActualDuration int              `json:"actual_duration" validate:"required,gt=0"`
	FinalFare      float64          `json:"final_fare" validate:"required,gt=0"`
	EndLocation    *models.Location `json:"end_location"`
	CompletionCode string           `json:"completion_code" validate:"required"`
	PaymentStatus  string           `json:"payment_status"`
}

type CancelRideRequest struct {
	Reason models.CancellationReason `json:"reason" validate:"required"`
	Notes  string                    `json:"notes"`
}

type RequestSpecificDriverRequest struct {
	DriverID string `json:"driver_id" validate:"required"`
}

type FindNearbyDriversRequest struct {
	Latitude    float64            `json:"latitude" validate:"required"`
	Longitude   float64            `json:"longitude" validate:"required"`
	Radius      float64            `json:"radius" validate:"required,gt=0"`
	ServiceType models.ServiceType `json:"service_type" validate:"required"`
}

// NewRideController creates a new ride controller
func NewRideController(rideService services.RideService) *RideController {
	return &RideController{
		rideService: rideService,
		logger:      utils.ControllerLogger("ride"),
	}
}

// Ride Management

func (rc *RideController) CreateRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req CreateRideControllerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Convert to service request
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	serviceReq := &services.CreateRideRequest{
		Type:                req.Type,
		ServiceType:         req.ServiceType,
		PickupLocation:      req.PickupLocation,
		DropoffLocation:     req.DropoffLocation,
		Waypoints:           req.Waypoints,
		Requirements:        req.Requirements,
		Preferences:         req.Preferences,
		ScheduledAt:         req.ScheduledAt,
		PaymentMethodID:     req.PaymentMethodID,
		ProposedFare:        req.ProposedFare,
		Notes:               req.Notes,
		SpecialInstructions: req.SpecialInstructions,
		Platform:            req.Platform,
	}

	// Set passenger ID
	serviceReq.PassengerID = userObjectID

	ride, err := rc.rideService.CreateRide(serviceReq)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create ride")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Ride created successfully", ride)
}

func (rc *RideController) GetUserRides(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	status := c.Query("status")

	rides, total, err := rc.rideService.GetUserRides(userID, params.Page, params.Limit, status)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user rides")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "User rides retrieved successfully", rides, meta)
}

func (rc *RideController) GetRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		if err.Error() == "ride not found" {
			utils.NotFoundResponse(c, "Ride")
			return
		}
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	ride, err := rc.rideService.GetRide(rideID)
	if err != nil {
		if err.Error() == "ride not found" {
			utils.NotFoundResponse(c, "Ride")
			return
		}
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get ride")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride retrieved successfully", ride)
}

func (rc *RideController) UpdateRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req services.UpdateRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	updatedRide, err := rc.rideService.UpdateRide(rideID, &req)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to update ride")

		if err.Error() == "ride cannot be updated in current status" {
			utils.BadRequestResponse(c, "Ride cannot be updated in current status")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Ride updated successfully", updatedRide)
}

func (rc *RideController) CancelRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req CancelRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	cancelledRide, err := rc.rideService.CancelRide(rideID, userID, req.Reason, req.Notes)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to cancel ride")

		if err.Error() == "you cannot cancel this ride" {
			utils.ForbiddenResponse(c)
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride cancelled successfully", cancelledRide)
}

// Ride Actions

func (rc *RideController) AcceptRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	acceptedRide, err := rc.rideService.AcceptRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("driver_id", userID).Msg("Failed to accept ride")

		if err.Error() == "ride cannot be accepted in current status" {
			utils.BadRequestResponse(c, "Ride cannot be accepted in current status")
			return
		}

		if err.Error() == "driver not found" || err.Error() == "driver is not available" {
			utils.BadRequestResponse(c, err.Error())
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride accepted successfully", acceptedRide)
}

func (rc *RideController) StartRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Create location from request
	location := &models.Location{
		Type:        "Point",
		Coordinates: []float64{req.Longitude, req.Latitude},
	}

	startedRide, err := rc.rideService.StartRide(rideID, userID, location)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("driver_id", userID).Msg("Failed to start ride")

		if err.Error() == "you are not assigned to this ride" {
			utils.ForbiddenResponse(c)
			return
		}

		if err.Error() == "ride cannot be started in current status" {
			utils.BadRequestResponse(c, "Ride cannot be started in current status")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride started successfully", startedRide)
}

func (rc *RideController) CompleteRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req CompleteRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	completionDetails := &services.RideCompletionDetails{
		ActualDistance: req.ActualDistance,
		ActualDuration: req.ActualDuration,
		FinalFare:      req.FinalFare,
		EndLocation:    req.EndLocation,
		CompletionCode: req.CompletionCode,
		PaymentStatus:  req.PaymentStatus,
	}

	completedRide, err := rc.rideService.CompleteRide(rideID, userID, completionDetails)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("driver_id", userID).Msg("Failed to complete ride")

		if err.Error() == "you are not assigned to this ride" {
			utils.ForbiddenResponse(c)
			return
		}

		if err.Error() == "ride cannot be completed in current status" {
			utils.BadRequestResponse(c, "Ride cannot be completed in current status")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride completed successfully", completedRide)
}

// Ride Tracking

func (rc *RideController) TrackRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	trackingInfo, err := rc.rideService.TrackRide(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to track ride")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride tracking info retrieved successfully", trackingInfo)
}

func (rc *RideController) UpdateRideLocation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	location := &models.Location{
		Type:        "Point",
		Coordinates: []float64{req.Longitude, req.Latitude},
	}

	err := rc.rideService.UpdateRideLocation(rideID, userID, location)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("driver_id", userID).Msg("Failed to update ride location")

		if err.Error() == "driver not assigned to this ride" {
			utils.ForbiddenResponse(c)
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride location updated successfully", nil)
}

func (rc *RideController) GetRideRoute(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	route, err := rc.rideService.GetRideRoute(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get ride route")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride route retrieved successfully", route)
}

func (rc *RideController) GetRideETA(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	eta, err := rc.rideService.GetRideETA(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get ride ETA")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride ETA retrieved successfully", gin.H{"eta": eta})
}

// Share Ride

func (rc *RideController) ShareRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	shareResponse, err := rc.rideService.ShareRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to share ride")

		if err.Error() == "you can only share your own rides" {
			utils.ForbiddenResponse(c)
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride shared successfully", shareResponse)
}

func (rc *RideController) GetSharedRide(c *gin.Context) {
	shareCode := c.Param("share_code")
	if shareCode == "" {
		utils.BadRequestResponse(c, "Share code is required")
		return
	}

	ride, err := rc.rideService.GetSharedRide(shareCode)
	if err != nil {
		if err.Error() == "shared ride not found" {
			utils.NotFoundResponse(c, "Shared ride")
			return
		}

		rc.logger.Error().Err(err).Str("share_code", shareCode).Msg("Failed to get shared ride")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Shared ride retrieved successfully", ride)
}

// Ride Scheduling

func (rc *RideController) ScheduleRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req services.ScheduleRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	scheduledRide, err := rc.rideService.ScheduleRide(&req)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to schedule ride")

		if err.Error() == "ride must be scheduled at least 15 minutes in advance" {
			utils.BadRequestResponse(c, err.Error())
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Ride scheduled successfully", scheduledRide)
}

func (rc *RideController) GetScheduledRides(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)

	rides, total, err := rc.rideService.GetScheduledRides(userID, params.Page, params.Limit)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get scheduled rides")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Scheduled rides retrieved successfully", rides, meta)
}

func (rc *RideController) UpdateScheduledRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req services.UpdateScheduledRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	updatedRide, err := rc.rideService.UpdateScheduledRide(rideID, &req)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to update scheduled ride")

		if err.Error() == "only scheduled rides can be updated" ||
			err.Error() == "scheduled ride cannot be updated after driver assignment" ||
			err.Error() == "ride must be scheduled at least 15 minutes in advance" {
			utils.BadRequestResponse(c, err.Error())
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Scheduled ride updated successfully", updatedRide)
}

func (rc *RideController) CancelScheduledRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	err := rc.rideService.CancelScheduledRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to cancel scheduled ride")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Scheduled ride cancelled successfully", nil)
}

// Driver Matching & Nearby

func (rc *RideController) FindNearbyDrivers(c *gin.Context) {
	var req FindNearbyDriversRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	drivers, err := rc.rideService.FindNearbyDrivers(req.Latitude, req.Longitude, req.Radius, req.ServiceType)
	if err != nil {
		rc.logger.Error().Err(err).Msg("Failed to find nearby drivers")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Nearby drivers retrieved successfully", drivers)
}

func (rc *RideController) RequestSpecificDriver(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req RequestSpecificDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := rc.rideService.RequestSpecificDriver(rideID, req.DriverID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("driver_id", req.DriverID).Msg("Failed to request specific driver")

		if err.Error() == "you can only request drivers for your own rides" {
			utils.ForbiddenResponse(c)
			return
		}

		if err.Error() == "driver can only be requested for pending rides" ||
			err.Error() == "requested driver is not available" ||
			err.Error() == "driver not found" {
			utils.BadRequestResponse(c, err.Error())
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver request sent successfully", nil)
}

func (rc *RideController) GetDriverLocation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	location, err := rc.rideService.GetDriverLocation(rideID)
	if err != nil {
		if err.Error() == "no driver assigned to this ride" {
			utils.BadRequestResponse(c, err.Error())
			return
		}

		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get driver location")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver location retrieved successfully", location)
}

// Fare Estimation & Negotiation

func (rc *RideController) EstimateFare(c *gin.Context) {
	var req EstimateFareControllerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	serviceReq := &services.FareEstimationRequest{
		ServiceType:     req.ServiceType,
		PickupLocation:  req.PickupLocation,
		DropoffLocation: req.DropoffLocation,
		Waypoints:       req.Waypoints,
		Requirements:    req.Requirements,
		ScheduledAt:     req.ScheduledAt,
	}

	estimation, err := rc.rideService.EstimateFare(serviceReq)
	if err != nil {
		rc.logger.Error().Err(err).Msg("Failed to estimate fare")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare estimated successfully", estimation)
}

func (rc *RideController) NegotiateFare(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req NegotiateFareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	offer, err := rc.rideService.NegotiateFare(rideID, userID, req.Amount, req.Message)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to negotiate fare")

		if err.Error() == "you cannot negotiate fare for this ride" ||
			err.Error() == "no driver assigned yet" {
			utils.BadRequestResponse(c, err.Error())
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare negotiated successfully", offer)
}

func (rc *RideController) AcceptFare(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req AcceptFareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := rc.rideService.AcceptFare(rideID, userID, req.OfferID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to accept fare")

		if err.Error() == "fare offer not found" ||
			err.Error() == "you cannot accept this offer" ||
			err.Error() == "fare offer has expired" {
			utils.BadRequestResponse(c, err.Error())
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare accepted successfully", nil)
}

func (rc *RideController) CounterOffer(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req struct {
		OfferID string  `json:"offer_id" validate:"required"`
		Amount  float64 `json:"amount" validate:"required,gt=0"`
		Message string  `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	counterOffer, err := rc.rideService.CounterOffer(rideID, userID, req.OfferID, req.Amount, req.Message)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to counter offer")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Counter offer made successfully", counterOffer)
}

func (rc *RideController) GetFareHistory(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	history, err := rc.rideService.GetFareHistory(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get fare history")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fare history retrieved successfully", history)
}

// Ride Preferences

func (rc *RideController) SetRidePreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var preferences models.RidePreferences
	if err := c.ShouldBindJSON(&preferences); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := rc.rideService.SetRidePreferences(rideID, userID, &preferences)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to set ride preferences")

		if err.Error() == "you cannot modify preferences for this ride" {
			utils.ForbiddenResponse(c)
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride preferences set successfully", nil)
}

func (rc *RideController) GetRidePreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	preferences, err := rc.rideService.GetRidePreferences(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get ride preferences")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride preferences retrieved successfully", preferences)
}

// Special Requirements

func (rc *RideController) SetSpecialRequirements(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var requirements models.RideRequirements
	if err := c.ShouldBindJSON(&requirements); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := rc.rideService.SetSpecialRequirements(rideID, userID, &requirements)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to set special requirements")

		if err.Error() == "you cannot modify requirements for this ride" {
			utils.ForbiddenResponse(c)
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Special requirements set successfully", nil)
}

func (rc *RideController) GetSpecialRequirements(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	requirements, err := rc.rideService.GetSpecialRequirements(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get special requirements")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Special requirements retrieved successfully", requirements)
}

// Ride Reports & Issues

func (rc *RideController) ReportRideIssue(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	var req services.ReportRideIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := rc.rideService.ReportRideIssue(rideID, userID, &req)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to report ride issue")

		if err.Error() == "you cannot report issues for this ride" {
			utils.ForbiddenResponse(c)
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride issue reported successfully", nil)
}

func (rc *RideController) GetRideReports(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	// Check if user can access this ride
	canAccess, err := rc.rideService.CanUserAccessRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to check ride access")
		utils.InternalServerErrorResponse(c)
		return
	}

	if !canAccess {
		utils.ForbiddenResponse(c)
		return
	}

	reports, err := rc.rideService.GetRideReports(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get ride reports")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride reports retrieved successfully", reports)
}

// Repeat Rides

func (rc *RideController) RepeatRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	newRide, err := rc.rideService.RepeatRide(rideID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to repeat ride")

		if err.Error() == "you can only repeat your own rides" {
			utils.ForbiddenResponse(c)
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Ride repeated successfully", newRide)
}

func (rc *RideController) GetFrequentRoutes(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	routes, err := rc.rideService.GetFrequentRoutes(userID, limit)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get frequent routes")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Frequent routes retrieved successfully", routes)
}
