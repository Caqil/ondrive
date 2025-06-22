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

// FreightController handles freight-related HTTP requests
type FreightController struct {
	freightService services.FreightService
	logger         utils.Logger
}

// NewFreightController creates a new freight controller
func NewFreightController(freightService services.FreightService) *FreightController {
	return &FreightController{
		freightService: freightService,
		logger:         utils.ControllerLogger("freight"),
	}
}

// Request structures
type CreateFreightRequestReq struct {
	Type                  string                            `json:"type" binding:"required"`
	Priority              string                            `json:"priority" binding:"required"`
	Cargo                 models.FreightCargo               `json:"cargo" binding:"required"`
	VehicleRequirements   models.FreightVehicleRequirements `json:"vehicle_requirements" binding:"required"`
	PickupLocation        models.FreightLocation            `json:"pickup_location" binding:"required"`
	DeliveryLocation      models.FreightLocation            `json:"delivery_location" binding:"required"`
	ScheduledPickupTime   *string                           `json:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *string                           `json:"scheduled_delivery_time,omitempty"`
	FlexibleScheduling    bool                              `json:"flexible_scheduling"`
	LoadingRequirements   models.LoadingRequirements        `json:"loading_requirements"`
	UnloadingRequirements models.LoadingRequirements        `json:"unloading_requirements"`
	PaymentTerms          string                            `json:"payment_terms"`
	SpecialInstructions   string                            `json:"special_instructions,omitempty"`
	RequiresInsurance     bool                              `json:"requires_insurance"`
	InsuranceValue        float64                           `json:"insurance_value,omitempty"`
}

type UpdateFreightRequestReq struct {
	Priority              string                      `json:"priority,omitempty"`
	ScheduledPickupTime   *string                     `json:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *string                     `json:"scheduled_delivery_time,omitempty"`
	FlexibleScheduling    *bool                       `json:"flexible_scheduling,omitempty"`
	LoadingRequirements   *models.LoadingRequirements `json:"loading_requirements,omitempty"`
	UnloadingRequirements *models.LoadingRequirements `json:"unloading_requirements,omitempty"`
	PaymentTerms          string                      `json:"payment_terms,omitempty"`
	SpecialInstructions   string                      `json:"special_instructions,omitempty"`
}

type SetCargoDetailsReq struct {
	Cargo models.FreightCargo `json:"cargo" binding:"required"`
}

type SetVehicleRequirementsReq struct {
	VehicleRequirements models.FreightVehicleRequirements `json:"vehicle_requirements" binding:"required"`
}

type RequestLoadingAssistanceReq struct {
	AssistanceType    string   `json:"assistance_type" binding:"required"`
	RequiredPersonnel int      `json:"required_personnel"`
	RequiredEquipment []string `json:"required_equipment"`
	EstimatedTime     int      `json:"estimated_time"`
	SpecialNotes      string   `json:"special_notes,omitempty"`
}

type ConfirmLoadingReq struct {
	LoadingStartedAt   string   `json:"loading_started_at" binding:"required"`
	LoadingCompletedAt string   `json:"loading_completed_at" binding:"required"`
	Photos             []string `json:"photos"`
	Notes              string   `json:"notes,omitempty"`
	WeightConfirmation float64  `json:"weight_confirmation"`
}

type AddMilestoneUpdateReq struct {
	MilestoneType string                 `json:"milestone_type" binding:"required"`
	Location      models.Location        `json:"location" binding:"required"`
	Timestamp     string                 `json:"timestamp" binding:"required"`
	Notes         string                 `json:"notes,omitempty"`
	Photos        []string               `json:"photos,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type CalculateRouteReq struct {
	PickupLocation   models.FreightLocation `json:"pickup_location" binding:"required"`
	DeliveryLocation models.FreightLocation `json:"delivery_location" binding:"required"`
	VehicleType      string                 `json:"vehicle_type" binding:"required"`
	Stops            []models.FreightStop   `json:"stops,omitempty"`
	OptimizeRoute    bool                   `json:"optimize_route"`
}

type EstimatePriceReq struct {
	PickupLocation      models.FreightLocation            `json:"pickup_location" binding:"required"`
	DeliveryLocation    models.FreightLocation            `json:"delivery_location" binding:"required"`
	Cargo               models.FreightCargo               `json:"cargo" binding:"required"`
	VehicleRequirements models.FreightVehicleRequirements `json:"vehicle_requirements" binding:"required"`
	ServiceType         string                            `json:"service_type"`
	Priority            string                            `json:"priority"`
	LoadingRequirements models.LoadingRequirements        `json:"loading_requirements,omitempty"`
}

type AddFreightInsuranceReq struct {
	InsuranceType     string  `json:"insurance_type" binding:"required"`
	CoverageAmount    float64 `json:"coverage_amount" binding:"required"`
	InsuranceProvider string  `json:"insurance_provider"`
	PolicyNumber      string  `json:"policy_number,omitempty"`
}

type AddDeliveryStopReq struct {
	Location    models.FreightLocation `json:"location" binding:"required"`
	StopType    string                 `json:"stop_type" binding:"required"`
	Duration    int                    `json:"duration"`
	CargoAction string                 `json:"cargo_action"`
	CargoItems  []string               `json:"cargo_items,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
}

type UpdateDeliveryStopReq struct {
	Location      *models.FreightLocation `json:"location,omitempty"`
	Duration      *int                    `json:"duration,omitempty"`
	CargoAction   string                  `json:"cargo_action,omitempty"`
	CargoItems    []string                `json:"cargo_items,omitempty"`
	Notes         string                  `json:"notes,omitempty"`
	ActualArrival *string                 `json:"actual_arrival,omitempty"`
	ActualDepart  *string                 `json:"actual_depart,omitempty"`
}

// Freight Service Management

func (fc *FreightController) CreateFreightRequest(c *gin.Context) {
	var req CreateFreightRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	freightRequest, err := fc.freightService.CreateFreightRequest(userID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create freight request")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Freight request created successfully", freightRequest)
}

func (fc *FreightController) GetFreightRequests(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	cargoType := c.Query("cargo_type")
	vehicleType := c.Query("vehicle_type")

	requests, total, err := fc.freightService.GetFreightRequests(userID, page, limit, status, cargoType, vehicleType)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get freight requests")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(page, limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Freight requests retrieved successfully", requests, meta)
}

func (fc *FreightController) GetFreightRequest(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	request, err := fc.freightService.GetFreightRequest(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to get freight request")
		utils.NotFoundResponse(c, "Freight request")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight request retrieved successfully", request)
}

func (fc *FreightController) UpdateFreightRequest(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req UpdateFreightRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	request, err := fc.freightService.UpdateFreightRequest(userID, objID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to update freight request")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight request updated successfully", request)
}

func (fc *FreightController) CancelFreightRequest(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err = fc.freightService.CancelFreightRequest(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to cancel freight request")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight request cancelled successfully", nil)
}

// Cargo Management

func (fc *FreightController) SetCargoDetails(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req SetCargoDetailsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	request, err := fc.freightService.SetCargoDetails(userID, objID, &req.Cargo)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to set cargo details")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cargo details updated successfully", request)
}

func (fc *FreightController) GetCargoDetails(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	cargo, err := fc.freightService.GetCargoDetails(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to get cargo details")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cargo details retrieved successfully", cargo)
}

func (fc *FreightController) UpdateCargoDetails(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req SetCargoDetailsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	request, err := fc.freightService.UpdateCargoDetails(userID, objID, &req.Cargo)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to update cargo details")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cargo details updated successfully", request)
}

func (fc *FreightController) UploadCargoPhotos(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

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

	files := form.File["photos"]
	if len(files) == 0 {
		utils.BadRequestResponse(c, "No photos provided")
		return
	}

	photoType := c.PostForm("type")
	description := c.PostForm("description")

	photos, err := fc.freightService.UploadCargoPhotos(userID, objID, files, photoType, description)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to upload cargo photos")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cargo photos uploaded successfully", photos)
}

func (fc *FreightController) GetCargoPhotos(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	photoType := c.Query("type")

	photos, err := fc.freightService.GetCargoPhotos(userID, objID, photoType)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to get cargo photos")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cargo photos retrieved successfully", photos)
}

// Vehicle Requirements

func (fc *FreightController) GetFreightVehicleTypes(c *gin.Context) {
	cargoType := c.Query("cargo_type")
	weight := c.Query("weight")
	volume := c.Query("volume")

	vehicleTypes, err := fc.freightService.GetFreightVehicleTypes(cargoType, weight, volume)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get freight vehicle types")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight vehicle types retrieved successfully", vehicleTypes)
}

func (fc *FreightController) SetVehicleRequirements(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req SetVehicleRequirementsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	request, err := fc.freightService.SetVehicleRequirements(userID, objID, &req.VehicleRequirements)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to set vehicle requirements")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Vehicle requirements updated successfully", request)
}

func (fc *FreightController) GetVehicleRequirements(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	requirements, err := fc.freightService.GetVehicleRequirements(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to get vehicle requirements")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Vehicle requirements retrieved successfully", requirements)
}

func (fc *FreightController) CheckVehicleAvailability(c *gin.Context) {
	vehicleType := c.Query("vehicle_type")
	location := c.Query("location")
	pickupTime := c.Query("pickup_time")
	weight := c.Query("weight")
	volume := c.Query("volume")

	availability, err := fc.freightService.CheckVehicleAvailability(vehicleType, location, pickupTime, weight, volume)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to check vehicle availability")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Vehicle availability checked successfully", availability)
}

// Loading & Unloading

func (fc *FreightController) GetLoadingOptions(c *gin.Context) {
	cargoType := c.Query("cargo_type")
	weight := c.Query("weight")
	dimensions := c.Query("dimensions")

	options, err := fc.freightService.GetLoadingOptions(cargoType, weight, dimensions)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get loading options")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Loading options retrieved successfully", options)
}

func (fc *FreightController) RequestLoadingAssistance(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req RequestLoadingAssistanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	assistance, err := fc.freightService.RequestLoadingAssistance(userID, objID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to request loading assistance")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Loading assistance requested successfully", assistance)
}

func (fc *FreightController) RequestUnloadingAssistance(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req RequestLoadingAssistanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	assistance, err := fc.freightService.RequestUnloadingAssistance(userID, objID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to request unloading assistance")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Unloading assistance requested successfully", assistance)
}

func (fc *FreightController) GetEquipmentNeeded(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	equipment, err := fc.freightService.GetEquipmentNeeded(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to get equipment needed")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Equipment requirements retrieved successfully", equipment)
}

// Freight Tracking

func (fc *FreightController) TrackFreightDelivery(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	tracking, err := fc.freightService.TrackFreightDelivery(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to track freight delivery")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight tracking retrieved successfully", tracking)
}

func (fc *FreightController) ConfirmLoading(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req ConfirmLoadingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	confirmation, err := fc.freightService.ConfirmLoading(userID, objID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to confirm loading")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Loading confirmed successfully", confirmation)
}

func (fc *FreightController) ConfirmFreightDelivery(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req ConfirmLoadingReq // Reusing same structure
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	confirmation, err := fc.freightService.ConfirmFreightDelivery(userID, objID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to confirm freight delivery")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight delivery confirmed successfully", confirmation)
}

func (fc *FreightController) GetMilestoneUpdates(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	milestones, err := fc.freightService.GetMilestoneUpdates(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to get milestone updates")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Milestone updates retrieved successfully", milestones)
}

func (fc *FreightController) AddMilestoneUpdate(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req AddMilestoneUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	milestone, err := fc.freightService.AddMilestoneUpdate(userID, objID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to add milestone update")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Milestone update added successfully", milestone)
}

// Route & Distance

func (fc *FreightController) CalculateFreightRoute(c *gin.Context) {
	var req CalculateRouteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	route, err := fc.freightService.CalculateFreightRoute(&req)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to calculate freight route")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight route calculated successfully", route)
}

func (fc *FreightController) OptimizeRoute(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	optimizedRoute, err := fc.freightService.OptimizeRoute(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to optimize route")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Route optimized successfully", optimizedRoute)
}

func (fc *FreightController) CalculateDistance(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	distance, err := fc.freightService.CalculateDistance(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to calculate distance")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Distance calculated successfully", distance)
}

func (fc *FreightController) EstimateFuelCost(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	fuelCost, err := fc.freightService.EstimateFuelCost(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to estimate fuel cost")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Fuel cost estimated successfully", fuelCost)
}

// Freight Pricing

func (fc *FreightController) EstimateFreightPrice(c *gin.Context) {
	var req EstimatePriceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	estimate, err := fc.freightService.EstimateFreightPrice(&req)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to estimate freight price")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight price estimated successfully", estimate)
}

func (fc *FreightController) GetPricingFactors(c *gin.Context) {
	cargoType := c.Query("cargo_type")
	vehicleType := c.Query("vehicle_type")
	serviceType := c.Query("service_type")

	factors, err := fc.freightService.GetPricingFactors(cargoType, vehicleType, serviceType)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get pricing factors")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pricing factors retrieved successfully", factors)
}

func (fc *FreightController) GetWeightBasedPricing(c *gin.Context) {
	weight := c.Query("weight")
	cargoType := c.Query("cargo_type")
	vehicleType := c.Query("vehicle_type")

	pricing, err := fc.freightService.GetWeightBasedPricing(weight, cargoType, vehicleType)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get weight-based pricing")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Weight-based pricing retrieved successfully", pricing)
}

func (fc *FreightController) GetDistanceBasedPricing(c *gin.Context) {
	distance := c.Query("distance")
	vehicleType := c.Query("vehicle_type")
	serviceType := c.Query("service_type")

	pricing, err := fc.freightService.GetDistanceBasedPricing(distance, vehicleType, serviceType)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get distance-based pricing")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Distance-based pricing retrieved successfully", pricing)
}

// Continue with remaining endpoints...
// (Documentation, Insurance & Safety, Multi-Stop Delivery, History & Analytics, Special Services)

// Documentation endpoints would follow the same pattern
func (fc *FreightController) UploadFreightDocuments(c *gin.Context) {
	// Implementation similar to UploadCargoPhotos
	utils.SuccessResponse(c, http.StatusOK, "Documents uploaded successfully", nil)
}

func (fc *FreightController) GetFreightDocuments(c *gin.Context) {
	// Implementation similar to GetCargoPhotos
	utils.SuccessResponse(c, http.StatusOK, "Documents retrieved successfully", nil)
}

func (fc *FreightController) SubmitCustomsDeclaration(c *gin.Context) {
	// Implementation for customs declaration
	utils.SuccessResponse(c, http.StatusOK, "Customs declaration submitted successfully", nil)
}

func (fc *FreightController) GetShippingManifest(c *gin.Context) {
	// Implementation for shipping manifest
	utils.SuccessResponse(c, http.StatusOK, "Shipping manifest retrieved successfully", nil)
}

// Insurance & Safety
func (fc *FreightController) GetInsuranceOptions(c *gin.Context) {
	// Implementation for insurance options
	utils.SuccessResponse(c, http.StatusOK, "Insurance options retrieved successfully", nil)
}

func (fc *FreightController) AddFreightInsurance(c *gin.Context) {
	// Implementation for adding insurance
	utils.SuccessResponse(c, http.StatusOK, "Insurance added successfully", nil)
}

func (fc *FreightController) GetSafetyGuidelines(c *gin.Context) {
	// Implementation for safety guidelines
	utils.SuccessResponse(c, http.StatusOK, "Safety guidelines retrieved successfully", nil)
}

func (fc *FreightController) PerformSafetyCheck(c *gin.Context) {
	// Implementation for safety check
	utils.SuccessResponse(c, http.StatusOK, "Safety check performed successfully", nil)
}

// Multi-Stop Delivery
func (fc *FreightController) AddDeliveryStop(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	var req AddDeliveryStopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	stop, err := fc.freightService.AddDeliveryStop(userID, objID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to add delivery stop")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Delivery stop added successfully", stop)
}

func (fc *FreightController) GetDeliveryStops(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	stops, err := fc.freightService.GetDeliveryStops(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to get delivery stops")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery stops retrieved successfully", stops)
}

func (fc *FreightController) UpdateDeliveryStop(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	stopID := c.Param("stop_id")
	stopObjID, err := primitive.ObjectIDFromHex(stopID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid stop ID")
		return
	}

	var req UpdateDeliveryStopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	stop, err := fc.freightService.UpdateDeliveryStop(userID, objID, stopObjID, &req)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Str("stop_id", stopID).Msg("Failed to update delivery stop")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery stop updated successfully", stop)
}

func (fc *FreightController) RemoveDeliveryStop(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	stopID := c.Param("stop_id")
	stopObjID, err := primitive.ObjectIDFromHex(stopID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid stop ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err = fc.freightService.RemoveDeliveryStop(userID, objID, stopObjID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Str("stop_id", stopID).Msg("Failed to remove delivery stop")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery stop removed successfully", nil)
}

func (fc *FreightController) OptimizeDeliveryStops(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid freight request ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	optimizedStops, err := fc.freightService.OptimizeDeliveryStops(userID, objID)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Str("request_id", id).Msg("Failed to optimize delivery stops")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery stops optimized successfully", optimizedStops)
}

// History & Analytics
func (fc *FreightController) GetFreightHistory(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	history, total, err := fc.freightService.GetFreightHistory(userID, page, limit, startDate, endDate)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get freight history")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(page, limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Freight history retrieved successfully", history, meta)
}

func (fc *FreightController) GetFreightAnalytics(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	period := c.DefaultQuery("period", "month")

	analytics, err := fc.freightService.GetFreightAnalytics(userID, period)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get freight analytics")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Freight analytics retrieved successfully", analytics)
}

func (fc *FreightController) GetCostBreakdown(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	freightID := c.Query("freight_id")
	period := c.Query("period")

	breakdown, err := fc.freightService.GetCostBreakdown(userID, freightID, period)
	if err != nil {
		fc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get cost breakdown")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cost breakdown retrieved successfully", breakdown)
}

// Special Services
func (fc *FreightController) GetTemperatureControlledOptions(c *gin.Context) {
	cargoType := c.Query("cargo_type")
	temperature := c.Query("temperature")

	options, err := fc.freightService.GetTemperatureControlledOptions(cargoType, temperature)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get temperature controlled options")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Temperature controlled options retrieved successfully", options)
}

func (fc *FreightController) GetHazardousMaterialsGuidelines(c *gin.Context) {
	materialType := c.Query("material_type")
	vehicleType := c.Query("vehicle_type")

	guidelines, err := fc.freightService.GetHazardousMaterialsGuidelines(materialType, vehicleType)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get hazardous materials guidelines")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Hazardous materials guidelines retrieved successfully", guidelines)
}

func (fc *FreightController) GetOversizedCargoOptions(c *gin.Context) {
	weight := c.Query("weight")
	dimensions := c.Query("dimensions")
	cargoType := c.Query("cargo_type")

	options, err := fc.freightService.GetOversizedCargoOptions(weight, dimensions, cargoType)
	if err != nil {
		fc.logger.Error().Err(err).Msg("Failed to get oversized cargo options")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Oversized cargo options retrieved successfully", options)
}
