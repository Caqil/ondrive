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

// RatingController handles rating-related HTTP requests
type RatingController struct {
	ratingService services.RatingService
	logger        utils.Logger
}

// Request types for controller
type CreateRatingRequest struct {
	Type        models.RatingType       `json:"type" validate:"required"`
	Score       float64                 `json:"score" validate:"required,gte=1,lte=5"`
	Review      string                  `json:"review" validate:"max=1000"`
	RatedUserID string                  `json:"rated_user_id" validate:"required"`
	RideID      string                  `json:"ride_id,omitempty"`
	Categories  []models.CategoryRating `json:"categories"`
	Feedback    models.RatingFeedback   `json:"feedback"`
	Photos      []string                `json:"photos"`
	IsPublic    bool                    `json:"is_public"`
	IsAnonymous bool                    `json:"is_anonymous"`
	Platform    string                  `json:"platform"`
}

// NewRatingController creates a new rating controller
func NewRatingController(ratingService services.RatingService) *RatingController {
	return &RatingController{
		ratingService: ratingService,
		logger:        utils.ControllerLogger("rating"),
	}
}

// Rating Management

func (rc *RatingController) CreateRating(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Convert user IDs to ObjectIDs
	raterObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	ratedUserObjectID, err := primitive.ObjectIDFromHex(req.RatedUserID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid rated user ID")
		return
	}

	// Create rating model
	rating := &models.Rating{
		Type:        req.Type,
		Score:       req.Score,
		Review:      req.Review,
		RaterID:     raterObjectID,
		RatedUserID: ratedUserObjectID,
		Categories:  req.Categories,
		Feedback:    req.Feedback,
		Photos:      req.Photos,
		IsPublic:    req.IsPublic,
		IsAnonymous: req.IsAnonymous,
		Platform:    req.Platform,
	}

	// Set ride ID if provided
	if req.RideID != "" {
		rideObjectID, err := primitive.ObjectIDFromHex(req.RideID)
		if err != nil {
			utils.BadRequestResponse(c, "Invalid ride ID")
			return
		}
		rating.RideID = &rideObjectID
	}

	// Create rating
	createdRating, err := rc.ratingService.CreateRating(rating)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to create rating")

		if err.Error() == "user has already rated this ride" {
			utils.ConflictResponse(c, "You have already rated this ride")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Rating created successfully", createdRating)
}

func (rc *RatingController) GetRating(c *gin.Context) {
	ratingID := c.Param("id")
	if ratingID == "" {
		utils.BadRequestResponse(c, "Rating ID is required")
		return
	}

	rating, err := rc.ratingService.GetRating(ratingID)
	if err != nil {
		if err.Error() == "rating not found" {
			utils.NotFoundResponse(c, "Rating")
			return
		}

		rc.logger.Error().Err(err).Str("rating_id", ratingID).Msg("Failed to get rating")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating retrieved successfully", rating)
}

func (rc *RatingController) UpdateRating(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	ratingID := c.Param("id")
	if ratingID == "" {
		utils.BadRequestResponse(c, "Rating ID is required")
		return
	}

	var req services.UpdateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	updatedRating, err := rc.ratingService.UpdateRating(ratingID, &req)
	if err != nil {
		rc.logger.Error().Err(err).Str("rating_id", ratingID).Str("user_id", userID).Msg("Failed to update rating")

		if err.Error() == "rating not found" {
			utils.NotFoundResponse(c, "Rating")
			return
		}

		if err.Error() == "rating can only be edited within 24 hours of creation" {
			utils.BadRequestResponse(c, "Rating can only be edited within 24 hours of creation")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.UpdatedResponse(c, "Rating updated successfully", updatedRating)
}

func (rc *RatingController) DeleteRating(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	ratingID := c.Param("id")
	if ratingID == "" {
		utils.BadRequestResponse(c, "Rating ID is required")
		return
	}

	err := rc.ratingService.DeleteRating(ratingID, userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("rating_id", ratingID).Str("user_id", userID).Msg("Failed to delete rating")

		if err.Error() == "rating not found" {
			utils.NotFoundResponse(c, "Rating")
			return
		}

		if err.Error() == "you can only delete your own ratings" {
			utils.ForbiddenResponse(c)
			return
		}

		if err.Error() == "rating can only be deleted within 1 hour of creation" {
			utils.BadRequestResponse(c, "Rating can only be deleted within 1 hour of creation")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating deleted successfully", nil)
}

// Ride Ratings

func (rc *RatingController) RateRide(c *gin.Context) {
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

	var req services.RideRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	rating, err := rc.ratingService.RateRide(rideID, userID, &req)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Str("user_id", userID).Msg("Failed to rate ride")

		if err.Error() == "you cannot rate this ride" {
			utils.BadRequestResponse(c, "You cannot rate this ride")
			return
		}

		if err.Error() == "ride not found" {
			utils.NotFoundResponse(c, "Ride")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.CreatedResponse(c, "Ride rated successfully", rating)
}

func (rc *RatingController) GetRideRating(c *gin.Context) {
	rideID := c.Param("ride_id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	ratings, err := rc.ratingService.GetRideRating(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get ride ratings")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride ratings retrieved successfully", ratings)
}

func (rc *RatingController) GetMutualRating(c *gin.Context) {
	rideID := c.Param("ride_id")
	if rideID == "" {
		utils.BadRequestResponse(c, "Ride ID is required")
		return
	}

	mutualRating, err := rc.ratingService.GetMutualRating(rideID)
	if err != nil {
		rc.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get mutual rating")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Mutual rating retrieved successfully", mutualRating)
}

// User Ratings

func (rc *RatingController) GetUserRatings(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		utils.BadRequestResponse(c, "User ID is required")
		return
	}

	params := utils.GetPaginationParams(c)
	ratings, total, err := rc.ratingService.GetUserRatings(userID, params.Page, params.Limit)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user ratings")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "User ratings retrieved successfully", ratings, meta)
}

func (rc *RatingController) GetRatingSummary(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		utils.BadRequestResponse(c, "User ID is required")
		return
	}

	summary, err := rc.ratingService.GetRatingSummary(userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get rating summary")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating summary retrieved successfully", summary)
}

func (rc *RatingController) GetAverageRating(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		utils.BadRequestResponse(c, "User ID is required")
		return
	}

	averageRating, err := rc.ratingService.GetAverageRating(userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get average rating")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Average rating retrieved successfully", averageRating)
}

// Driver Ratings

func (rc *RatingController) GetDriverRatings(c *gin.Context) {
	driverID := c.Param("driver_id")
	if driverID == "" {
		utils.BadRequestResponse(c, "Driver ID is required")
		return
	}

	params := utils.GetPaginationParams(c)
	ratings, total, err := rc.ratingService.GetDriverRatings(driverID, params.Page, params.Limit)
	if err != nil {
		rc.logger.Error().Err(err).Str("driver_id", driverID).Msg("Failed to get driver ratings")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Driver ratings retrieved successfully", ratings, meta)
}

func (rc *RatingController) GetDriverRatingStats(c *gin.Context) {
	driverID := c.Param("driver_id")
	if driverID == "" {
		utils.BadRequestResponse(c, "Driver ID is required")
		return
	}

	stats, err := rc.ratingService.GetDriverRatingStats(driverID)
	if err != nil {
		rc.logger.Error().Err(err).Str("driver_id", driverID).Msg("Failed to get driver rating stats")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver rating stats retrieved successfully", stats)
}

func (rc *RatingController) GetRatingBreakdown(c *gin.Context) {
	driverID := c.Param("driver_id")
	if driverID == "" {
		utils.BadRequestResponse(c, "Driver ID is required")
		return
	}

	breakdown, err := rc.ratingService.GetRatingBreakdown(driverID)
	if err != nil {
		rc.logger.Error().Err(err).Str("driver_id", driverID).Msg("Failed to get rating breakdown")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating breakdown retrieved successfully", breakdown)
}

// Rating History

func (rc *RatingController) GetRatingHistory(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	ratings, total, err := rc.ratingService.GetRatingHistory(userID, params.Page, params.Limit)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get rating history")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Rating history retrieved successfully", ratings, meta)
}

func (rc *RatingController) GetGivenRatings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	ratings, total, err := rc.ratingService.GetGivenRatings(userID, params.Page, params.Limit)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get given ratings")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Given ratings retrieved successfully", ratings, meta)
}

func (rc *RatingController) GetReceivedRatings(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	params := utils.GetPaginationParams(c)
	ratings, total, err := rc.ratingService.GetReceivedRatings(userID, params.Page, params.Limit)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get received ratings")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(params.Page, params.Limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Received ratings retrieved successfully", ratings, meta)
}

// Rating Analytics

func (rc *RatingController) GetRatingAnalytics(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	analytics, err := rc.ratingService.GetRatingAnalytics(userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get rating analytics")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating analytics retrieved successfully", analytics)
}

func (rc *RatingController) GetRatingTrends(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	// Get days parameter (default to 30)
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	trends, err := rc.ratingService.GetRatingTrends(userID, days)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get rating trends")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating trends retrieved successfully", trends)
}

func (rc *RatingController) GetRatingComparison(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	compareWithID := c.Query("compare_with")
	if compareWithID == "" {
		utils.BadRequestResponse(c, "compare_with parameter is required")
		return
	}

	comparison, err := rc.ratingService.GetRatingComparison(userID, compareWithID)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Str("compare_with", compareWithID).Msg("Failed to get rating comparison")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating comparison retrieved successfully", comparison)
}

// Rating Categories

func (rc *RatingController) GetRatingCategories(c *gin.Context) {
	categories, err := rc.ratingService.GetRatingCategories()
	if err != nil {
		rc.logger.Error().Err(err).Msg("Failed to get rating categories")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating categories retrieved successfully", categories)
}

func (rc *RatingController) GetRatingCriteria(c *gin.Context) {
	ratingTypeStr := c.Query("type")
	if ratingTypeStr == "" {
		utils.BadRequestResponse(c, "Rating type is required")
		return
	}

	ratingType := models.RatingType(ratingTypeStr)
	criteria, err := rc.ratingService.GetRatingCriteria(ratingType)
	if err != nil {
		rc.logger.Error().Err(err).Str("rating_type", ratingTypeStr).Msg("Failed to get rating criteria")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating criteria retrieved successfully", criteria)
}

// Feedback & Reviews

func (rc *RatingController) AddFeedback(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	ratingID := c.Param("id")
	if ratingID == "" {
		utils.BadRequestResponse(c, "Rating ID is required")
		return
	}

	var req services.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := rc.ratingService.AddFeedback(ratingID, userID, &req)
	if err != nil {
		rc.logger.Error().Err(err).Str("rating_id", ratingID).Str("user_id", userID).Msg("Failed to add feedback")

		if err.Error() == "rating not found" {
			utils.NotFoundResponse(c, "Rating")
			return
		}

		if err.Error() == "you can only respond to ratings about you" {
			utils.ForbiddenResponse(c)
			return
		}

		if err.Error() == "you have already responded to this rating" {
			utils.ConflictResponse(c, "You have already responded to this rating")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Feedback added successfully", nil)
}

func (rc *RatingController) GetFeedback(c *gin.Context) {
	ratingID := c.Param("id")
	if ratingID == "" {
		utils.BadRequestResponse(c, "Rating ID is required")
		return
	}

	feedback, err := rc.ratingService.GetFeedback(ratingID)
	if err != nil {
		rc.logger.Error().Err(err).Str("rating_id", ratingID).Msg("Failed to get feedback")

		if err.Error() == "rating not found" {
			utils.NotFoundResponse(c, "Rating")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Feedback retrieved successfully", feedback)
}

func (rc *RatingController) ReportRating(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	ratingID := c.Param("id")
	if ratingID == "" {
		utils.BadRequestResponse(c, "Rating ID is required")
		return
	}

	var req services.ReportRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format")
		return
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	err := rc.ratingService.ReportRating(ratingID, userID, &req)
	if err != nil {
		rc.logger.Error().Err(err).Str("rating_id", ratingID).Str("user_id", userID).Msg("Failed to report rating")

		if err.Error() == "rating not found" {
			utils.NotFoundResponse(c, "Rating")
			return
		}

		if err.Error() == "rating has already been reported" {
			utils.ConflictResponse(c, "Rating has already been reported")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating reported successfully", nil)
}

// Rating Incentives

func (rc *RatingController) GetRatingRewards(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	rewards, err := rc.ratingService.GetRatingRewards(userID)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get rating rewards")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating rewards retrieved successfully", rewards)
}

func (rc *RatingController) ClaimRatingReward(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	var req struct {
		RewardID string `json:"reward_id" validate:"required"`
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

	reward, err := rc.ratingService.ClaimRatingReward(userID, req.RewardID)
	if err != nil {
		rc.logger.Error().Err(err).Str("user_id", userID).Str("reward_id", req.RewardID).Msg("Failed to claim rating reward")

		if err.Error() == "reward not found" {
			utils.NotFoundResponse(c, "Reward")
			return
		}

		if err.Error() == "not eligible for this reward" {
			utils.BadRequestResponse(c, "Not eligible for this reward")
			return
		}

		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rating reward claimed successfully", reward)
}
