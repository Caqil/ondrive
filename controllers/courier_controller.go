package controllers

import (
	"net/http"
	"strconv"
	"time"

	"ondrive/models"
	"ondrive/services"
	"ondrive/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CourierController struct {
	courierService services.CourierService
	uploadService  services.UploadService
	logger         utils.Logger
}

func NewCourierController(
	courierService services.CourierService,
	uploadService services.UploadService,
	logger utils.Logger,
) *CourierController {
	return &CourierController{
		courierService: courierService,
		uploadService:  uploadService,
		logger:         logger,
	}
}

// Request/Response DTOs
type CreateCourierRequest struct {
	Type                  string                  `json:"type" binding:"required"`
	Priority              models.DeliverySpeed    `json:"priority" binding:"required"`
	Package               CourierPackageDTO       `json:"package" binding:"required"`
	PickupLocation        models.CourierLocation  `json:"pickup_location" binding:"required"`
	DeliveryLocation      models.CourierLocation  `json:"delivery_location" binding:"required"`
	Recipient             models.CourierRecipient `json:"recipient" binding:"required"`
	ScheduledPickupTime   *time.Time              `json:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *time.Time              `json:"scheduled_delivery_time,omitempty"`
	FlexibleTiming        bool                    `json:"flexible_timing"`
	TimeWindow            *models.TimeWindow      `json:"time_window,omitempty"`
	SpecialInstructions   string                  `json:"special_instructions"`
	PaymentMethodID       string                  `json:"payment_method_id" binding:"required"`
}

type CourierPackageDTO struct {
	Description       string                   `json:"description" binding:"required"`
	Category          models.PackageCategory   `json:"category" binding:"required"`
	Quantity          int                      `json:"quantity" binding:"required,gte=1"`
	Weight            float64                  `json:"weight" binding:"required,gte=0.1,lte=20"`
	Dimensions        models.PackageDimensions `json:"dimensions" binding:"required"`
	Contents          []models.PackageItem     `json:"contents"`
	Value             float64                  `json:"value" binding:"gte=0"`
	Currency          string                   `json:"currency" binding:"required"`
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

type UpdateCourierRequest struct {
	Priority              *models.DeliverySpeed `json:"priority,omitempty"`
	ScheduledPickupTime   *time.Time            `json:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *time.Time            `json:"scheduled_delivery_time,omitempty"`
	FlexibleTiming        *bool                 `json:"flexible_timing,omitempty"`
	TimeWindow            *models.TimeWindow    `json:"time_window,omitempty"`
	SpecialInstructions   *string               `json:"special_instructions,omitempty"`
}

type EstimatePriceRequest struct {
	Priority              models.DeliverySpeed   `json:"priority" binding:"required"`
	Package               CourierPackageDTO      `json:"package" binding:"required"`
	PickupLocation        models.CourierLocation `json:"pickup_location" binding:"required"`
	DeliveryLocation      models.CourierLocation `json:"delivery_location" binding:"required"`
	ScheduledPickupTime   *time.Time             `json:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *time.Time             `json:"scheduled_delivery_time,omitempty"`
	RequiresInsurance     bool                   `json:"requires_insurance"`
	InsuranceValue        float64                `json:"insurance_value"`
}

// Courier Service Management

func (cc *CourierController) CreateCourierRequest(c *gin.Context) {
	var req CreateCourierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		cc.logger.Warn().Err(err).Msg("Invalid courier request data")
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", "")
		return
	}

	senderID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", "")
		return
	}

	// Create courier request
	courierRequest := &models.CourierRequest{
		ID:                    primitive.NewObjectID(),
		Type:                  req.Type,
		Priority:              req.Priority,
		SenderID:              senderID,
		Package:               cc.convertToPackageModel(req.Package),
		PickupLocation:        req.PickupLocation,
		DeliveryLocation:      req.DeliveryLocation,
		Recipient:             req.Recipient,
		ScheduledPickupTime:   req.ScheduledPickupTime,
		ScheduledDeliveryTime: req.ScheduledDeliveryTime,
		FlexibleTiming:        req.FlexibleTiming,
		TimeWindow:            req.TimeWindow,
		SpecialInstructions:   req.SpecialInstructions,
		PaymentMethodID:       req.PaymentMethodID,
		PaymentStatus:         "pending",
		Platform:              "mobile",
	}

	// Save to repository
	savedRequest, err := cc.courierService.CreateRequest(courierRequest)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to create courier request")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Courier request created successfully", savedRequest)
}

func (cc *CourierController) GetCourierRequests(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", "")
		return
	}

	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", "")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	filter := map[string]interface{}{
		"sender_id": objID,
	}

	if status != "" {
		filter["status"] = status
	}

	// Use service to get requests (the service will call repository)
	requests, total, err := cc.courierService.GetRequestsByFilter(filter, page, limit)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to get courier requests")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := &utils.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
		HasNext:    int64(page*limit) < total,
	}

	if page > 1 {
		meta.HasPrevious = true
	}

	utils.PaginatedResponse(c, http.StatusOK, "Courier requests retrieved successfully", requests, meta)
}

func (cc *CourierController) GetCourierRequest(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	request, err := cc.courierService.GetRequest(objID)
	if err != nil {
		cc.logger.Error().Err(err).Str("id", id).Msg("Failed to get courier request")
		utils.NotFoundResponse(c, "Courier request not found")
		return
	}

	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))

	// Check if user has access to this request
	if request.SenderID != userObjID && (request.CourierID == nil || *request.CourierID != userObjID) {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Courier request retrieved successfully", request)
}

func (cc *CourierController) UpdateCourierRequest(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var req UpdateCourierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	// Get existing request
	existingRequest, err := cc.courierService.GetRequest(objID)
	if err != nil {
		utils.NotFoundResponse(c, "Courier request not found")
		return
	}

	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))

	// Check if user is the sender
	if existingRequest.SenderID != userObjID {
		utils.ErrorResponse(c, http.StatusForbidden, "Only sender can update request", "")
		return
	}

	// Check if request can be updated
	if existingRequest.Status != models.CourierStatusPending {
		utils.ErrorResponse(c, http.StatusBadRequest, "Cannot update request after it has been accepted", "")
		return
	}

	// Build update data
	updateData := make(map[string]interface{})
	if req.Priority != nil {
		updateData["priority"] = *req.Priority
	}
	if req.ScheduledPickupTime != nil {
		updateData["scheduled_pickup_time"] = req.ScheduledPickupTime
	}
	if req.ScheduledDeliveryTime != nil {
		updateData["scheduled_delivery_time"] = req.ScheduledDeliveryTime
	}
	if req.FlexibleTiming != nil {
		updateData["flexible_timing"] = *req.FlexibleTiming
	}
	if req.TimeWindow != nil {
		updateData["time_window"] = req.TimeWindow
	}
	if req.SpecialInstructions != nil {
		updateData["special_instructions"] = *req.SpecialInstructions
	}

	updatedRequest, err := cc.courierService.UpdateRequest(objID, updateData)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to update courier request")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Courier request updated successfully", updatedRequest)
}

func (cc *CourierController) CancelCourierRequest(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	reason := c.Query("reason")
	if reason == "" {
		reason = "Cancelled by user"
	}

	userID, _ := c.Get("user_id")
	userObjID, _ := primitive.ObjectIDFromHex(userID.(string))

	err = cc.courierService.CancelRequest(objID, userObjID, reason)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to cancel courier request")
		if err.Error() == "request not found" {
			utils.NotFoundResponse(c, "Courier request not found")
			return
		}
		if err.Error() == "unauthorized" {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "")
			return
		}
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Courier request cancelled successfully", nil)
}

// Package Management

func (cc *CourierController) SetPackageDetails(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var packageDetails CourierPackageDTO
	if err := c.ShouldBindJSON(&packageDetails); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid package data"})
		return
	}

	err = cc.courierService.UpdatePackageDetails(objID, cc.convertToPackageModel(packageDetails))
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to update package details")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Package details updated successfully", nil)
}

func (cc *CourierController) GetPackageDetails(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	request, err := cc.courierService.GetRequest(objID)
	if err != nil {
		utils.NotFoundResponse(c, "Courier request not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Package details retrieved successfully", request.Package)
}

func (cc *CourierController) UpdatePackageDetails(c *gin.Context) {
	// Same as SetPackageDetails for this implementation
	cc.SetPackageDetails(c)
}

func (cc *CourierController) UploadPackagePhotos(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	// Handle file uploads
	form, err := c.MultipartForm()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file upload", "")
		return
	}

	files := form.File["photos"]
	var photoURLs []string

	for _, file := range files {
		url, err := cc.uploadService.UploadPackagePhoto(file, objID.Hex())
		if err != nil {
			cc.logger.Error().Err(err).Msg("Failed to upload package photo")
			continue
		}
		photoURLs = append(photoURLs, url)
	}

	if len(photoURLs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No photos uploaded successfully", "")
		return
	}

	err = cc.courierService.AddPackagePhotos(objID, photoURLs)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to save package photos")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Package photos uploaded successfully", map[string]interface{}{
		"uploaded_count": len(photoURLs),
		"photo_urls":     photoURLs,
	})
}

func (cc *CourierController) GetPackagePhotos(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	request, err := cc.courierService.GetRequest(objID)
	if err != nil {
		utils.NotFoundResponse(c, "Courier request not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Package photos retrieved successfully", request.Package.Photos)
}

// Courier Tracking

func (cc *CourierController) TrackCourierDelivery(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	trackingInfo, err := cc.courierService.GetTrackingInfo(objID)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to get tracking info")
		if err.Error() == "request not found" {
			utils.NotFoundResponse(c, "Courier request not found")
			return
		}
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tracking information retrieved successfully", trackingInfo)
}

func (cc *CourierController) ConfirmPickup(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var req struct {
		Photos   []string         `json:"photos"`
		Notes    string           `json:"notes"`
		Location *models.Location `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid confirmation data"})
		return
	}

	userID, _ := c.Get("user_id")
	courierID, _ := primitive.ObjectIDFromHex(userID.(string))

	err = cc.courierService.ConfirmPickup(objID, courierID, req.Photos, req.Notes, req.Location)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to confirm pickup")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pickup confirmed successfully", nil)
}

func (cc *CourierController) ConfirmDelivery(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var req struct {
		DeliveredTo    string           `json:"delivered_to" binding:"required"`
		Photos         []string         `json:"photos"`
		Notes          string           `json:"notes"`
		Location       *models.Location `json:"location"`
		SignatureURL   string           `json:"signature_url"`
		RecipientPhoto string           `json:"recipient_photo"`
		OTPCode        string           `json:"otp_code"`
		Condition      string           `json:"condition"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid delivery confirmation data"})
		return
	}

	userID, _ := c.Get("user_id")
	courierID, _ := primitive.ObjectIDFromHex(userID.(string))

	deliveryProof := &models.DeliveryProof{
		Type:           "photo_signature",
		DeliveryPhotos: req.Photos,
		SignatureURL:   req.SignatureURL,
		SignedBy:       req.DeliveredTo,
		RecipientPhoto: req.RecipientPhoto,
		OTPCode:        req.OTPCode,
		DeliveredTo:    req.DeliveredTo,
		DeliveredBy:    courierID,
		DeliveryTime:   time.Now(),
		Condition:      req.Condition,
		DeliveryNotes:  req.Notes,
		IsVerified:     true,
		VerifiedAt:     &[]time.Time{time.Now()}[0],
		VerifiedBy:     &courierID,
	}

	if req.Location != nil {
		deliveryProof.DeliveryLocation = *req.Location
	}

	err = cc.courierService.ConfirmDelivery(objID, courierID, deliveryProof)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to confirm delivery")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery confirmed successfully", nil)
}

func (cc *CourierController) GetDeliveryProof(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	request, err := cc.courierService.GetRequest(objID)
	if err != nil {
		utils.NotFoundResponse(c, "Courier request not found")
		return
	}

	if request.DeliveryProof == nil {
		utils.NotFoundResponse(c, "No delivery proof available")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery proof retrieved successfully", request.DeliveryProof)
}

func (cc *CourierController) UploadDeliveryProof(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file upload", "")
		return
	}

	files := form.File["proof_photos"]
	var photoURLs []string

	for _, file := range files {
		url, err := cc.uploadService.UploadDeliveryProof(file, objID.Hex())
		if err != nil {
			cc.logger.Error().Err(err).Msg("Failed to upload delivery proof")
			continue
		}
		photoURLs = append(photoURLs, url)
	}

	err = cc.courierService.UpdateDeliveryProofPhotos(objID, photoURLs)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to save delivery proof photos")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery proof uploaded successfully", map[string]interface{}{
		"uploaded_count": len(photoURLs),
		"photo_urls":     photoURLs,
	})
}

// Recipient Management

func (cc *CourierController) SetRecipientDetails(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var recipient models.CourierRecipient
	if err := c.ShouldBindJSON(&recipient); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid recipient data"})
		return
	}

	err = cc.courierService.UpdateRecipient(objID, &recipient)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to update recipient details")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Recipient details updated successfully", nil)
}

func (cc *CourierController) GetRecipientDetails(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	request, err := cc.courierService.GetRequest(objID)
	if err != nil {
		utils.NotFoundResponse(c, "Courier request not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Recipient details retrieved successfully", request.Recipient)
}

func (cc *CourierController) UpdateRecipientDetails(c *gin.Context) {
	cc.SetRecipientDetails(c)
}

func (cc *CourierController) NotifyRecipient(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var req struct {
		Message   string `json:"message"`
		SendSMS   bool   `json:"send_sms"`
		SendEmail bool   `json:"send_email"`
		SendPush  bool   `json:"send_push"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid notification data"})
		return
	}

	err = cc.courierService.NotifyRecipient(objID, req.Message, req.SendSMS, req.SendEmail, req.SendPush)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to notify recipient")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Recipient notified successfully", nil)
}

// Delivery Options

func (cc *CourierController) GetDeliveryOptions(c *gin.Context) {
	options := map[string]interface{}{
		"speeds": []map[string]interface{}{
			{"code": "standard", "name": "Standard Delivery", "description": "Same day delivery", "duration": "6-8 hours"},
			{"code": "express", "name": "Express Delivery", "description": "Fast delivery", "duration": "2-4 hours"},
			{"code": "priority", "name": "Priority Delivery", "description": "High priority", "duration": "1-2 hours"},
			{"code": "urgent", "name": "Urgent Delivery", "description": "Emergency delivery", "duration": "Within 1 hour"},
		},
		"special_services": []string{
			"fragile_handling",
			"signature_required",
			"age_verification",
			"cold_chain",
			"insurance",
		},
		"package_categories": []string{
			"documents",
			"electronics",
			"clothing",
			"food",
			"medicine",
			"fragile",
			"personal",
			"other",
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery options retrieved successfully", options)
}

func (cc *CourierController) ScheduleDelivery(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var req struct {
		ScheduledTime time.Time          `json:"scheduled_time" binding:"required"`
		TimeWindow    *models.TimeWindow `json:"time_window"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid schedule data"})
		return
	}

	err = cc.courierService.ScheduleDelivery(objID, req.ScheduledTime, req.TimeWindow)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to schedule delivery")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Delivery scheduled successfully", nil)
}

func (cc *CourierController) RescheduleDelivery(c *gin.Context) {
	cc.ScheduleDelivery(c) // Same logic for rescheduling
}

func (cc *CourierController) GetDeliveryTimeSlots(c *gin.Context) {
	date := c.Query("date")
	location := c.Query("location")

	slots, err := cc.courierService.GetAvailableTimeSlots(date, location)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to get time slots")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Time slots retrieved successfully", slots)
}

// Special Services

func (cc *CourierController) GetFragileHandlingOptions(c *gin.Context) {
	options := map[string]interface{}{
		"handling_types": []map[string]interface{}{
			{"code": "standard", "name": "Standard Fragile", "description": "Basic fragile handling", "fee": 5.0},
			{"code": "extra_care", "name": "Extra Care", "description": "Enhanced protection", "fee": 10.0},
			{"code": "white_glove", "name": "White Glove", "description": "Premium handling service", "fee": 25.0},
		},
		"packaging_options": []string{
			"bubble_wrap",
			"foam_padding",
			"custom_box",
			"rigid_container",
		},
		"insurance_recommended": true,
		"max_weight":            15.0,
	}

	utils.SuccessResponse(c, http.StatusOK, "Fragile handling options retrieved successfully", options)
}

func (cc *CourierController) AddInsurance(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var insurance models.CourierInsurance
	if err := c.ShouldBindJSON(&insurance); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid insurance data"})
		return
	}

	err = cc.courierService.AddInsurance(objID, &insurance)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to add insurance")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Insurance added successfully", nil)
}

func (cc *CourierController) GetInsuranceDetails(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	request, err := cc.courierService.GetRequest(objID)
	if err != nil {
		utils.NotFoundResponse(c, "Courier request not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Insurance details retrieved successfully", request.Insurance)
}

func (cc *CourierController) RequireSignature(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var req struct {
		Required bool   `json:"required"`
		Type     string `json:"type"` // digital, physical
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid signature requirement data"})
		return
	}

	err = cc.courierService.SetSignatureRequirement(objID, req.Required, req.Type)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to set signature requirement")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Signature requirement updated successfully", nil)
}

// Courier Pricing

func (cc *CourierController) EstimateCourierPrice(c *gin.Context) {
	var req EstimatePriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid price estimation data"})
		return
	}

	estimate, err := cc.courierService.EstimatePrice(req)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to estimate price")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Price estimated successfully", estimate)
}

func (cc *CourierController) GetPricingCalculator(c *gin.Context) {
	calculator := map[string]interface{}{
		"base_rates": map[string]float64{
			"standard": 15.0,
			"express":  25.0,
			"priority": 35.0,
			"urgent":   50.0,
		},
		"distance_rates": map[string]float64{
			"per_km_standard": 1.5,
			"per_km_express":  2.0,
			"per_km_priority": 2.5,
			"per_km_urgent":   3.0,
		},
		"weight_rates": map[string]float64{
			"per_kg_standard": 2.0,
			"per_kg_express":  3.0,
			"per_kg_priority": 4.0,
			"per_kg_urgent":   5.0,
		},
		"surcharges": map[string]float64{
			"fragile":    5.0,
			"perishable": 10.0,
			"cold_chain": 15.0,
			"oversized":  20.0,
			"peak_hour":  10.0,
			"weekend":    15.0,
			"night_time": 20.0,
		},
		"limits": map[string]interface{}{
			"max_weight":     20.0,
			"max_dimensions": map[string]float64{"length": 100, "width": 80, "height": 80},
			"max_value":      10000.0,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Pricing calculator retrieved successfully", calculator)
}

func (cc *CourierController) GetWeightLimits(c *gin.Context) {
	limits := map[string]interface{}{
		"standard_max": 20.0,
		"express_max":  15.0,
		"priority_max": 10.0,
		"urgent_max":   5.0,
		"unit":         "kg",
		"overweight_handling": map[string]interface{}{
			"available":         true,
			"max_weight":        50.0,
			"extra_charge":      "per_kg",
			"rate_per_kg":       5.0,
			"requires_approval": true,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Weight limits retrieved successfully", limits)
}

func (cc *CourierController) GetSizeLimits(c *gin.Context) {
	limits := map[string]interface{}{
		"max_dimensions": map[string]float64{
			"length": 100.0,
			"width":  80.0,
			"height": 80.0,
		},
		"unit":         "cm",
		"volume_limit": 640000.0, // cm³
		"oversized_handling": map[string]interface{}{
			"available":         true,
			"max_length":        200.0,
			"max_width":         150.0,
			"max_height":        150.0,
			"extra_charge":      50.0,
			"requires_approval": true,
			"special_vehicle":   true,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Size limits retrieved successfully", limits)
}

// Courier History

func (cc *CourierController) GetCourierHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	objID, _ := primitive.ObjectIDFromHex(userID.(string))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	filter := map[string]interface{}{
		"sender_id": objID,
	}

	if status != "" {
		filter["status"] = status
	}

	if dateFrom != "" || dateTo != "" {
		dateFilter := make(map[string]interface{})
		if dateFrom != "" {
			if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
				dateFilter["$gte"] = parsedDate
			}
		}
		if dateTo != "" {
			if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
				dateFilter["$lte"] = parsedDate.AddDate(0, 0, 1)
			}
		}
		if len(dateFilter) > 0 {
			filter["created_at"] = dateFilter
		}
	}

	requests, total, err := cc.courierService.GetRequestsByFilter(filter, page, limit)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to get courier history")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := &utils.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
		HasNext:    int64(page*limit) < total,
	}

	if page > 1 {
		meta.HasPrevious = true
	}

	utils.PaginatedResponse(c, http.StatusOK, "Courier history retrieved successfully", requests, meta)
}

func (cc *CourierController) GetSentPackages(c *gin.Context) {
	// Same as GetCourierHistory but with specific filtering
	cc.GetCourierHistory(c)
}

func (cc *CourierController) GetReceivedPackages(c *gin.Context) {
	phone := c.GetString("user_phone") // Assuming phone is available in context

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Find packages where the user is the recipient
	filter := map[string]interface{}{
		"$or": []interface{}{
			map[string]interface{}{"recipient.phone": phone},
			map[string]interface{}{"recipient.email": c.GetString("user_email")},
		},
	}

	requests, total, err := cc.courierService.GetRequestsByFilter(filter, page, limit)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to get received packages")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := &utils.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
		HasNext:    int64(page*limit) < total,
	}

	if page > 1 {
		meta.HasPrevious = true
	}

	utils.PaginatedResponse(c, http.StatusOK, "Received packages retrieved successfully", requests, meta)
}

// Courier Support

func (cc *CourierController) ReportCourierIssue(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var req struct {
		IssueType   string   `json:"issue_type" binding:"required"`
		Description string   `json:"description" binding:"required"`
		Priority    string   `json:"priority"`
		Photos      []string `json:"photos"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid issue report data"})
		return
	}

	userID, _ := c.Get("user_id")
	reporterID, _ := primitive.ObjectIDFromHex(userID.(string))

	issueID, err := cc.courierService.ReportIssue(objID, reporterID, req.IssueType, req.Description, req.Priority, req.Photos)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to report issue")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Issue reported successfully", map[string]interface{}{
		"issue_id": issueID,
	})
}

func (cc *CourierController) GetCourierSupport(c *gin.Context) {
	support := map[string]interface{}{
		"contact_methods": []map[string]interface{}{
			{"type": "phone", "value": "+1-800-COURIER", "hours": "24/7"},
			{"type": "email", "value": "support@ondrive.com", "response_time": "2-4 hours"},
			{"type": "chat", "value": "in_app", "hours": "6AM-10PM"},
		},
		"faq_categories": []string{
			"delivery_issues",
			"pricing_questions",
			"package_damage",
			"scheduling_problems",
			"payment_issues",
		},
		"emergency_contact": "+1-800-EMERGENCY",
		"claim_process": map[string]interface{}{
			"deadline_days":   7,
			"required_docs":   []string{"photos", "receipt", "description"},
			"processing_time": "3-5 business days",
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Support information retrieved successfully", support)
}

func (cc *CourierController) FileClaim(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid courier request ID", "")
		return
	}

	var req struct {
		ClaimType   string   `json:"claim_type" binding:"required"` // damage, loss, delay
		Amount      float64  `json:"amount" binding:"required"`
		Description string   `json:"description" binding:"required"`
		Evidence    []string `json:"evidence"` // photo URLs
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid claim data"})
		return
	}

	userID, _ := c.Get("user_id")
	claimantID, _ := primitive.ObjectIDFromHex(userID.(string))

	claimID, err := cc.courierService.FileClaim(objID, claimantID, req.ClaimType, req.Amount, req.Description, req.Evidence)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to file claim")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Claim filed successfully", map[string]interface{}{
		"claim_id": claimID,
	})
}

// Address Book

func (cc *CourierController) GetAddressBook(c *gin.Context) {
	userID, _ := c.Get("user_id")
	objID, _ := primitive.ObjectIDFromHex(userID.(string))

	addresses, err := cc.courierService.GetAddressBook(objID)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to get address book")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address book retrieved successfully", addresses)
}

func (cc *CourierController) AddAddress(c *gin.Context) {
	var address models.SavedAddress
	if err := c.ShouldBindJSON(&address); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid address data"})
		return
	}

	userID, _ := c.Get("user_id")
	objID, _ := primitive.ObjectIDFromHex(userID.(string))

	address.ID = primitive.NewObjectID()
	address.UserID = objID
	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()

	savedAddress, err := cc.courierService.AddAddress(objID, &address)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to add address")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Address added successfully", savedAddress)
}

func (cc *CourierController) UpdateAddress(c *gin.Context) {
	id := c.Param("id")
	addressID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID", "")
		return
	}

	var address models.SavedAddress
	if err := c.ShouldBindJSON(&address); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid address data"})
		return
	}

	userID, _ := c.Get("user_id")
	objID, _ := primitive.ObjectIDFromHex(userID.(string))

	address.ID = addressID
	address.UserID = objID
	address.UpdatedAt = time.Now()

	updatedAddress, err := cc.courierService.UpdateAddress(objID, addressID, &address)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to update address")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address updated successfully", updatedAddress)
}

func (cc *CourierController) DeleteAddress(c *gin.Context) {
	id := c.Param("id")
	addressID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID", "")
		return
	}

	userID, _ := c.Get("user_id")
	objID, _ := primitive.ObjectIDFromHex(userID.(string))

	err = cc.courierService.DeleteAddress(objID, addressID)
	if err != nil {
		cc.logger.Error().Err(err).Msg("Failed to delete address")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address deleted successfully", nil)
}

// Helper functions

func (cc *CourierController) convertToPackageModel(dto CourierPackageDTO) models.CourierPackage {
	return models.CourierPackage{
		Description:       dto.Description,
		Category:          dto.Category,
		Quantity:          dto.Quantity,
		Weight:            dto.Weight,
		Dimensions:        dto.Dimensions,
		Contents:          dto.Contents,
		Value:             dto.Value,
		Currency:          dto.Currency,
		IsFragile:         dto.IsFragile,
		IsPerishable:      dto.IsPerishable,
		RequiresColdChain: dto.RequiresColdChain,
		RequiresUpright:   dto.RequiresUpright,
		SpecialHandling:   dto.SpecialHandling,
		RequiresSignature: dto.RequiresSignature,
		RequiresID:        dto.RequiresID,
		AgeRestricted:     dto.AgeRestricted,
		PackagingType:     dto.PackagingType,
		Photos:            []models.PackagePhoto{}, // Initialize empty
		Volume:            dto.Dimensions.Length * dto.Dimensions.Width * dto.Dimensions.Height,
	}
}
