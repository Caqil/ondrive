package controllers

import (
	"net/http"
	"strconv"
	"time"

	"ondrive/middleware"
	"ondrive/models"
	"ondrive/repositories"
	"ondrive/services"
	"ondrive/utils"

	"github.com/gin-gonic/gin"
)

type DriverController struct {
	driverService       services.DriverService
	driverRepo          repositories.DriverRepository
	userRepo            repositories.UserRepository
	fileUploadService   services.UploadService
	notificationService services.NotificationService
	logger              utils.Logger
}

// Request/Response structures
type UpdateDriverProfileRequest struct {
	FirstName    string   `json:"first_name" validate:"required"`
	LastName     string   `json:"last_name" validate:"required"`
	Bio          string   `json:"bio"`
	ProfilePhoto string   `json:"profile_photo"`
	DateOfBirth  string   `json:"date_of_birth"`
	Gender       string   `json:"gender"`
	Languages    []string `json:"languages"`
}

type AddVehicleRequest struct {
	Make         string   `json:"make" validate:"required"`
	Model        string   `json:"model" validate:"required"`
	Year         int      `json:"year" validate:"gte=1990,lte=2030"`
	Color        string   `json:"color" validate:"required"`
	LicensePlate string   `json:"license_plate" validate:"required"`
	VIN          string   `json:"vin"`
	FuelType     string   `json:"fuel_type"`
	Transmission string   `json:"transmission"`
	Seats        int      `json:"seats" validate:"gte=1,lte=50"`
	Features     []string `json:"features"`
	Amenities    []string `json:"amenities"`
}

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
	Address   string  `json:"address"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
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
	Subject     string `json:"subject" validate:"required"`
	Description string `json:"description" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Priority    string `json:"priority"`
}

func NewDriverController(
	driverService services.DriverService,
	driverRepo repositories.DriverRepository,
	userRepo repositories.UserRepository,
	fileUploadService services.UploadService,
	notificationService services.NotificationService,
	logger utils.Logger,
) *DriverController {
	return &DriverController{
		driverService:       driverService,
		driverRepo:          driverRepo,
		userRepo:            userRepo,
		fileUploadService:   fileUploadService,
		notificationService: notificationService,
		logger:              logger,
	}
}

// Driver Profile Management

func (dc *DriverController) GetDriverProfile(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	driver, err := dc.driverRepo.GetByUserID(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get driver profile")
		utils.NotFoundResponse(c, "Driver profile")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver profile retrieved successfully", driver)
}

func (dc *DriverController) UpdateDriverProfile(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateDriverProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	updatedDriver, err := dc.driverService.UpdateDriverProfile(userID, &req)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update driver profile")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Driver profile updated successfully", updatedDriver)
}

func (dc *DriverController) UploadDriverDocuments(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequestResponse(c, "Invalid form data")
		return
	}

	documentType := c.PostForm("document_type")
	if documentType == "" {
		utils.BadRequestResponse(c, "Document type is required")
		return
	}

	files := form.File["documents"]
	if len(files) == 0 {
		utils.BadRequestResponse(c, "At least one document file is required")
		return
	}

	uploadedDocs, err := dc.driverService.UploadDriverDocuments(userID, documentType, files)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("document_type", documentType).Msg("Failed to upload driver documents")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Driver documents uploaded successfully", uploadedDocs)
}

func (dc *DriverController) GetDriverDocuments(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	documents, err := dc.driverService.GetDriverDocuments(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get driver documents")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver documents retrieved successfully", documents)
}

// Vehicle Management

func (dc *DriverController) AddVehicle(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req AddVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	vehicle, err := dc.driverService.AddVehicle(userID, &req)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add vehicle")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Vehicle added successfully", vehicle)
}

func (dc *DriverController) GetVehicle(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	vehicle, err := dc.driverService.GetDriverVehicle(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get vehicle")
		utils.NotFoundResponse(c, "Vehicle")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Vehicle retrieved successfully", vehicle)
}

func (dc *DriverController) UpdateVehicle(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req AddVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	vehicle, err := dc.driverService.UpdateVehicle(userID, &req)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update vehicle")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Vehicle updated successfully", vehicle)
}

func (dc *DriverController) UploadVehiclePhotos(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.BadRequestResponse(c, "Invalid form data")
		return
	}

	photoType := c.PostForm("photo_type")
	files := form.File["photos"]

	if len(files) == 0 {
		utils.BadRequestResponse(c, "At least one photo is required")
		return
	}

	photos, err := dc.driverService.UploadVehiclePhotos(userID, photoType, files)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload vehicle photos")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Vehicle photos uploaded successfully", photos)
}

func (dc *DriverController) DeleteVehiclePhoto(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	photoID := c.Param("id")
	if photoID == "" {
		utils.BadRequestResponse(c, "Photo ID is required")
		return
	}

	err := dc.driverService.DeleteVehiclePhoto(userID, photoID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("photo_id", photoID).Msg("Failed to delete vehicle photo")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.DeletedResponse(c, "Vehicle photo deleted successfully")
}

// Driver Status & Availability

func (dc *DriverController) GoOnline(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err := dc.driverService.UpdateOnlineStatus(userID, true)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to go online")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver is now online", nil)
}

func (dc *DriverController) GoOffline(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err := dc.driverService.UpdateOnlineStatus(userID, false)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to go offline")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver is now offline", nil)
}

func (dc *DriverController) GetDriverStatus(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	status, err := dc.driverService.GetDriverStatus(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get driver status")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver status retrieved successfully", status)
}

func (dc *DriverController) UpdateAvailability(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := dc.driverService.UpdateAvailability(userID, &req)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update availability")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Availability updated successfully", nil)
}

// Ride Management

func (dc *DriverController) GetAvailableRides(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	rides, total, err := dc.driverService.GetAvailableRides(userID, params.Page, params.Limit)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get available rides")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Available rides retrieved successfully", rides, meta)
}

func (dc *DriverController) GetCurrentRide(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	ride, err := dc.driverService.GetCurrentRide(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get current ride")
		utils.NotFoundResponse(c, "Current ride")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Current ride retrieved successfully", ride)
}

func (dc *DriverController) GetRideRequests(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	requests, total, err := dc.driverService.GetRideRequests(userID, params.Page, params.Limit)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get ride requests")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Ride requests retrieved successfully", requests, meta)
}

func (dc *DriverController) AcceptRideRequest(c *gin.Context) {
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

	ride, err := dc.driverService.AcceptRideRequest(userID, rideID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("ride_id", rideID).Msg("Failed to accept ride request")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Ride request accepted successfully", ride)
}

func (dc *DriverController) DeclineRideRequest(c *gin.Context) {
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

	err := dc.driverService.DeclineRideRequest(userID, rideID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("ride_id", rideID).Msg("Failed to decline ride request")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Ride request declined successfully", nil)
}

// Location & Navigation

func (dc *DriverController) UpdateDriverLocation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	location := &models.Location{
		Coordinates: []float64{req.Latitude, req.Longitude},
		Address:     req.Address,
		Heading:     req.Heading,
		Speed:       req.Speed,
		UpdatedAt:   time.Now(),
	}

	err := dc.driverService.UpdateDriverLocation(userID, location)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update driver location")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Driver location updated successfully", location)
}

func (dc *DriverController) GetDriverLocation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	location, err := dc.driverService.GetDriverLocation(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get driver location")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver location retrieved successfully", location)
}

func (dc *DriverController) GetNavigation(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rideID := c.Param("ride_id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	navigation, err := dc.driverService.GetNavigation(userID, rideID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("ride_id", rideID).Msg("Failed to get navigation")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Navigation retrieved successfully", navigation)
}

// Earnings & Finance

func (dc *DriverController) GetEarnings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	earnings, total, err := dc.driverService.GetEarnings(userID, params.Page, params.Limit)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get earnings")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Earnings retrieved successfully", earnings, meta)
}

func (dc *DriverController) GetDailyEarnings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	earnings, err := dc.driverService.GetDailyEarnings(userID, date)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("date", date).Msg("Failed to get daily earnings")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daily earnings retrieved successfully", earnings)
}

func (dc *DriverController) GetWeeklyEarnings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	weekStr := c.Query("week")
	if weekStr == "" {
		weekStr = strconv.Itoa(int(time.Now().Weekday()))
	}

	earnings, err := dc.driverService.GetWeeklyEarnings(userID, weekStr)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("week", weekStr).Msg("Failed to get weekly earnings")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Weekly earnings retrieved successfully", earnings)
}

func (dc *DriverController) GetMonthlyEarnings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	earnings, err := dc.driverService.GetMonthlyEarnings(userID, month)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("month", month).Msg("Failed to get monthly earnings")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Monthly earnings retrieved successfully", earnings)
}

func (dc *DriverController) GetPayouts(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	payouts, total, err := dc.driverService.GetPayouts(userID, params.Page, params.Limit)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get payouts")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Payouts retrieved successfully", payouts, meta)
}

func (dc *DriverController) RequestPayout(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req struct {
		Amount float64 `json:"amount" validate:"required,gt=0"`
		Method string  `json:"method" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	payout, err := dc.driverService.RequestPayout(userID, req.Amount, req.Method)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Float64("amount", req.Amount).Msg("Failed to request payout")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Payout request created successfully", payout)
}

// Driver Statistics

func (dc *DriverController) GetDriverStats(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	stats, err := dc.driverService.GetDriverStats(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get driver stats")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver stats retrieved successfully", stats)
}

func (dc *DriverController) GetPerformanceMetrics(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	metrics, err := dc.driverService.GetPerformanceMetrics(userID, period)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("period", period).Msg("Failed to get performance metrics")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Performance metrics retrieved successfully", metrics)
}

func (dc *DriverController) GetRatingSummary(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	ratingSummary, err := dc.driverService.GetRatingSummary(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get rating summary")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating summary retrieved successfully", ratingSummary)
}

// Working Hours & Schedule

func (dc *DriverController) GetWorkingHours(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	workingHours, err := dc.driverService.GetWorkingHours(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get working hours")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Working hours retrieved successfully", workingHours)
}

func (dc *DriverController) UpdateWorkingHours(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req models.WorkingHours
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := dc.driverService.UpdateWorkingHours(userID, &req)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update working hours")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Working hours updated successfully", req)
}

func (dc *DriverController) GetDriverSchedule(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	schedule, err := dc.driverService.GetDriverSchedule(userID, date)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("date", date).Msg("Failed to get driver schedule")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver schedule retrieved successfully", schedule)
}

func (dc *DriverController) TakeBreak(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req struct {
		Duration string `json:"duration" validate:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := dc.driverService.TakeBreak(userID, req.Duration, req.Reason)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("duration", req.Duration).Msg("Failed to take break")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Break started successfully", nil)
}

func (dc *DriverController) ResumeWork(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err := dc.driverService.ResumeWork(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to resume work")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Work resumed successfully", nil)
}

// Driver Preferences

func (dc *DriverController) GetDriverPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	preferences, err := dc.driverService.GetDriverPreferences(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get driver preferences")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver preferences retrieved successfully", preferences)
}

func (dc *DriverController) UpdateDriverPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req UpdateDriverPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	preferences, err := dc.driverService.UpdateDriverPreferences(userID, &req)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update driver preferences")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Driver preferences updated successfully", preferences)
}

func (dc *DriverController) UpdateServiceAreas(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req struct {
		ServiceAreas []models.ServiceArea `json:"service_areas" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := dc.driverService.UpdateServiceAreas(userID, req.ServiceAreas)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update service areas")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Service areas updated successfully", req.ServiceAreas)
}

func (dc *DriverController) UpdateServiceTypes(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req struct {
		ServiceTypes []string `json:"service_types" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	err := dc.driverService.UpdateServiceTypes(userID, req.ServiceTypes)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update service types")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Service types updated successfully", req.ServiceTypes)
}

// Documents & Verification

func (dc *DriverController) UploadDriverLicense(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	file, err := c.FormFile("license")
	if err != nil {
		utils.BadRequestResponse(c, "License file is required")
		return
	}

	expiryDate := c.PostForm("expiry_date")
	licenseNumber := c.PostForm("license_number")

	document, err := dc.driverService.UploadDriverLicense(userID, file, licenseNumber, expiryDate)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload driver license")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Driver license uploaded successfully", document)
}

func (dc *DriverController) UploadInsurance(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	file, err := c.FormFile("insurance")
	if err != nil {
		utils.BadRequestResponse(c, "Insurance file is required")
		return
	}

	expiryDate := c.PostForm("expiry_date")
	policyNumber := c.PostForm("policy_number")

	document, err := dc.driverService.UploadInsurance(userID, file, policyNumber, expiryDate)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload insurance")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Insurance uploaded successfully", document)
}

func (dc *DriverController) UploadVehicleRegistration(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	file, err := c.FormFile("registration")
	if err != nil {
		utils.BadRequestResponse(c, "Registration file is required")
		return
	}

	expiryDate := c.PostForm("expiry_date")
	registrationNumber := c.PostForm("registration_number")

	document, err := dc.driverService.UploadVehicleRegistration(userID, file, registrationNumber, expiryDate)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to upload vehicle registration")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Vehicle registration uploaded successfully", document)
}

func (dc *DriverController) GetVerificationStatus(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	status, err := dc.driverService.GetVerificationStatus(userID)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get verification status")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Verification status retrieved successfully", status)
}

// Driver Support

func (dc *DriverController) CreateSupportTicket(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req CreateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	ticket, err := dc.driverService.CreateSupportTicket(userID, &req)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create support ticket")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Support ticket created successfully", ticket)
}

func (dc *DriverController) GetSupportTickets(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	tickets, total, err := dc.driverService.GetSupportTickets(userID, params.Page, params.Limit)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get support tickets")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Support tickets retrieved successfully", tickets, meta)
}

func (dc *DriverController) UpdateSupportTicket(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	ticketID := c.Param("id")
	if ticketID == "" {
		utils.BadRequestResponse(c, "Ticket ID is required")
		return
	}

	var req struct {
		Message string `json:"message" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	ticket, err := dc.driverService.UpdateSupportTicket(userID, ticketID, req.Message)
	if err != nil {
		dc.logger.Error().Err(err).Str("user_id", userID).Str("ticket_id", ticketID).Msg("Failed to update support ticket")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Support ticket updated successfully", ticket)
}
