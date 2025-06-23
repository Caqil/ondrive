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

// RatingService interface defines rating-related business logic
type RatingService interface {
	// Basic rating operations
	CreateRating(rating *models.Rating) (*models.Rating, error)
	GetRating(ratingID string) (*models.Rating, error)
	UpdateRating(ratingID string, req *UpdateRatingRequest) (*models.Rating, error)
	DeleteRating(ratingID, userID string) error

	// Ride ratings
	RateRide(rideID, raterID string, req *RideRatingRequest) (*models.Rating, error)
	GetRideRating(rideID string) ([]*models.Rating, error)
	GetMutualRating(rideID string) (*MutualRatingResponse, error)

	// User ratings
	GetUserRatings(userID string, page, limit int) ([]*models.Rating, int64, error)
	GetRatingSummary(userID string) (*models.RatingSummary, error)
	GetAverageRating(userID string) (*AverageRatingResponse, error)

	// Driver ratings
	GetDriverRatings(driverID string, page, limit int) ([]*models.Rating, int64, error)
	GetDriverRatingStats(driverID string) (*DriverRatingStats, error)
	GetRatingBreakdown(userID string) (*repositories.RatingBreakdown, error)

	// Rating history
	GetRatingHistory(userID string, page, limit int) ([]*models.Rating, int64, error)
	GetGivenRatings(userID string, page, limit int) ([]*models.Rating, int64, error)
	GetReceivedRatings(userID string, page, limit int) ([]*models.Rating, int64, error)

	// Rating analytics
	GetRatingAnalytics(userID string) (*RatingAnalytics, error)
	GetRatingTrends(userID string, days int) (*repositories.RatingTrends, error)
	GetRatingComparison(userID, compareWithID string) (*repositories.RatingComparison, error)

	// Rating categories and criteria
	GetRatingCategories() ([]RatingCategory, error)
	GetRatingCriteria(ratingType models.RatingType) ([]models.RatingCriteria, error)

	// Feedback and reporting
	AddFeedback(ratingID, userID string, req *FeedbackRequest) error
	GetFeedback(ratingID string) (*FeedbackResponse, error)
	ReportRating(ratingID, reporterID string, req *ReportRatingRequest) error

	// Rating incentives
	GetRatingRewards(userID string) ([]RatingReward, error)
	ClaimRatingReward(userID, rewardID string) (*RatingReward, error)

	// Validation and business rules
	ValidateRating(rating *models.Rating) error
	CanUserRateRide(userID, rideID string) (bool, error)
	CalculateRatingImpact(userID string, newRating float64) (*RatingImpact, error)
}

// Request and response types

type UpdateRatingRequest struct {
	Score      *float64                `json:"score,omitempty" validate:"omitempty,gte=1,lte=5"`
	Review     *string                 `json:"review,omitempty" validate:"omitempty,max=1000"`
	Categories []models.CategoryRating `json:"categories,omitempty"`
	Photos     []string                `json:"photos,omitempty"`
	IsPublic   *bool                   `json:"is_public,omitempty"`
}

type RideRatingRequest struct {
	Score       float64                 `json:"score" validate:"required,gte=1,lte=5"`
	Review      string                  `json:"review" validate:"max=1000"`
	Categories  []models.CategoryRating `json:"categories"`
	Feedback    models.RatingFeedback   `json:"feedback"`
	Photos      []string                `json:"photos"`
	IsPublic    bool                    `json:"is_public"`
	IsAnonymous bool                    `json:"is_anonymous"`
	Platform    string                  `json:"platform"`
}

type MutualRatingResponse struct {
	RideID          string         `json:"ride_id"`
	PassengerRating *models.Rating `json:"passenger_rating,omitempty"`
	DriverRating    *models.Rating `json:"driver_rating,omitempty"`
	IsComplete      bool           `json:"is_complete"`
	AverageScore    float64        `json:"average_score"`
}

type AverageRatingResponse struct {
	UserID        string      `json:"user_id"`
	AverageRating float64     `json:"average_rating"`
	TotalRatings  int         `json:"total_ratings"`
	Distribution  map[int]int `json:"distribution"`
	RecentTrend   string      `json:"recent_trend"`
}

type DriverRatingStats struct {
	Summary      *models.RatingSummary  `json:"summary"`
	MonthlyStats []MonthlyRatingStats   `json:"monthly_stats"`
	CategoryAvg  []models.CategoryScore `json:"category_averages"`
	TopReviews   []*models.Rating       `json:"top_reviews"`
}

type MonthlyRatingStats struct {
	Month        string  `json:"month"`
	AverageScore float64 `json:"average_score"`
	TotalRatings int     `json:"total_ratings"`
	Trend        string  `json:"trend"`
}

type RatingAnalytics struct {
	UserID              string                     `json:"user_id"`
	OverallStats        *models.RatingSummary      `json:"overall_stats"`
	TrendAnalysis       *repositories.RatingTrends `json:"trend_analysis"`
	CategoryPerformance []CategoryPerformance      `json:"category_performance"`
	ComparisonStats     *ComparisonStats           `json:"comparison_stats"`
	Recommendations     []string                   `json:"recommendations"`
}

type CategoryPerformance struct {
	Category    string   `json:"category"`
	Score       float64  `json:"score"`
	Improvement float64  `json:"improvement"`
	Rank        int      `json:"rank"`
	Feedback    []string `json:"feedback"`
}

type ComparisonStats struct {
	BetterThanPercent float64 `json:"better_than_percent"`
	CityRank          int     `json:"city_rank"`
	ServiceTypeRank   int     `json:"service_type_rank"`
}

type RatingCategory struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Weight      float64 `json:"weight"`
}

type FeedbackRequest struct {
	Content  string `json:"content" validate:"required,max=500"`
	IsPublic bool   `json:"is_public"`
}

type FeedbackResponse struct {
	RatingID string                  `json:"rating_id"`
	Feedback []models.RatingFeedback `json:"feedback"`
	CanReply bool                    `json:"can_reply"`
}

type ReportRatingRequest struct {
	Reason      string `json:"reason" validate:"required"`
	Description string `json:"description" validate:"max=500"`
	Category    string `json:"category"`
}

type RatingReward struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Type           string     `json:"type"`
	Value          float64    `json:"value"`
	RequiredRating float64    `json:"required_rating"`
	RequiredCount  int        `json:"required_count"`
	IsEligible     bool       `json:"is_eligible"`
	ClaimedAt      *time.Time `json:"claimed_at,omitempty"`
}

type RatingImpact struct {
	CurrentRating float64 `json:"current_rating"`
	NewRating     float64 `json:"new_rating"`
	Change        float64 `json:"change"`
	Impact        string  `json:"impact"` // positive, negative, neutral
	NewRank       int     `json:"new_rank"`
}

// ratingService implements RatingService interface
type ratingService struct {
	ratingRepo      repositories.RatingRepository
	userRepo        repositories.UserRepository
	rideRepo        repositories.RideRepository
	notificationSvc NotificationService
	logger          utils.Logger
}

// NewRatingService creates a new rating service
func NewRatingService(
	ratingRepo repositories.RatingRepository,
	userRepo repositories.UserRepository,
	rideRepo repositories.RideRepository,
	notificationSvc NotificationService,
	logger utils.Logger,
) RatingService {
	return &ratingService{
		ratingRepo:      ratingRepo,
		userRepo:        userRepo,
		rideRepo:        rideRepo,
		notificationSvc: notificationSvc,
		logger:          utils.ServiceLogger("rating"),
	}
}

// Basic rating operations

func (s *ratingService) CreateRating(rating *models.Rating) (*models.Rating, error) {
	// Validate rating
	if err := s.ValidateRating(rating); err != nil {
		return nil, err
	}

	// Check if user already rated this ride
	if rating.RideID != nil {
		hasRated, err := s.ratingRepo.HasUserRatedRide(rating.RaterID.Hex(), rating.RideID.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to check if user already rated: %w", err)
		}
		if hasRated {
			return nil, errors.New("user has already rated this ride")
		}
	}

	// Set additional fields
	rating.CreatedAt = time.Now()
	rating.UpdatedAt = time.Now()

	// Create rating
	createdRating, err := s.ratingRepo.Create(rating)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create rating")
		return nil, fmt.Errorf("failed to create rating")
	}

	// Send notification to rated user
	go s.notifyUserRated(createdRating)

	// Update rating summary
	go s.updateUserRatingSummary(rating.RatedUserID.Hex())

	s.logger.Info().
		Str("rating_id", createdRating.ID.Hex()).
		Str("rater_id", rating.RaterID.Hex()).
		Str("rated_user_id", rating.RatedUserID.Hex()).
		Float64("score", rating.Score).
		Msg("Rating created successfully")

	return createdRating, nil
}

func (s *ratingService) GetRating(ratingID string) (*models.Rating, error) {
	rating, err := s.ratingRepo.GetByID(ratingID)
	if err != nil {
		return nil, err
	}

	// Check if rating is hidden
	if rating.IsHidden {
		return nil, errors.New("rating not found")
	}

	return rating, nil
}

func (s *ratingService) UpdateRating(ratingID string, req *UpdateRatingRequest) (*models.Rating, error) {
	// Get existing rating
	rating, err := s.ratingRepo.GetByID(ratingID)
	if err != nil {
		return nil, err
	}

	// Check if rating can be edited (within time limit)
	timeSinceCreation := time.Since(rating.CreatedAt)
	if timeSinceCreation > 24*time.Hour {
		return nil, errors.New("rating can only be edited within 24 hours of creation")
	}

	// Update fields
	if req.Score != nil {
		rating.Score = *req.Score
	}
	if req.Review != nil {
		rating.Review = *req.Review
	}
	if req.Categories != nil {
		rating.Categories = req.Categories
	}
	if req.Photos != nil {
		rating.Photos = req.Photos
	}
	if req.IsPublic != nil {
		rating.IsPublic = *req.IsPublic
	}

	// Validate updated rating
	if err := s.ValidateRating(rating); err != nil {
		return nil, err
	}

	// Update in database
	updatedRating, err := s.ratingRepo.Update(rating)
	if err != nil {
		s.logger.Error().Err(err).Str("rating_id", ratingID).Msg("Failed to update rating")
		return nil, fmt.Errorf("failed to update rating")
	}

	// Update rating summary
	go s.updateUserRatingSummary(rating.RatedUserID.Hex())

	return updatedRating, nil
}

func (s *ratingService) DeleteRating(ratingID, userID string) error {
	// Get rating to verify ownership
	rating, err := s.ratingRepo.GetByID(ratingID)
	if err != nil {
		return err
	}

	// Check if user owns the rating
	if rating.RaterID.Hex() != userID {
		return errors.New("you can only delete your own ratings")
	}

	// Check if rating can be deleted (within time limit)
	timeSinceCreation := time.Since(rating.CreatedAt)
	if timeSinceCreation > 1*time.Hour {
		return errors.New("rating can only be deleted within 1 hour of creation")
	}

	err = s.ratingRepo.Delete(ratingID)
	if err != nil {
		s.logger.Error().Err(err).Str("rating_id", ratingID).Msg("Failed to delete rating")
		return fmt.Errorf("failed to delete rating")
	}

	// Update rating summary
	go s.updateUserRatingSummary(rating.RatedUserID.Hex())

	return nil
}

// Ride ratings

func (s *ratingService) RateRide(rideID, raterID string, req *RideRatingRequest) (*models.Rating, error) {
	// Validate ride exists and user participated
	canRate, err := s.CanUserRateRide(raterID, rideID)
	if err != nil {
		return nil, err
	}
	if !canRate {
		return nil, errors.New("you cannot rate this ride")
	}

	// Get ride details
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, fmt.Errorf("ride not found")
	}

	// Determine rated user
	raterObjectID, _ := primitive.ObjectIDFromHex(raterID)
	var ratedUserID primitive.ObjectID
	if ride.PassengerID == raterObjectID {
		ratedUserID = ride.DriverID
	} else {
		ratedUserID = ride.PassengerID
	}

	// Create rating
	rating := &models.Rating{
		Type:        models.RatingTypeRide,
		Score:       req.Score,
		Review:      req.Review,
		RaterID:     raterObjectID,
		RatedUserID: ratedUserID,
		RideID:      &ride.ID,
		Categories:  req.Categories,
		Feedback:    req.Feedback,
		Photos:      req.Photos,
		IsPublic:    req.IsPublic,
		IsAnonymous: req.IsAnonymous,
		Platform:    req.Platform,
	}

	return s.CreateRating(rating)
}

func (s *ratingService) GetRideRating(rideID string) ([]*models.Rating, error) {
	return s.ratingRepo.GetByRideID(rideID)
}

func (s *ratingService) GetMutualRating(rideID string) (*MutualRatingResponse, error) {
	ratings, err := s.ratingRepo.GetMutualRating(rideID)
	if err != nil {
		return nil, err
	}

	response := &MutualRatingResponse{
		RideID:     rideID,
		IsComplete: len(ratings) == 2,
	}

	var totalScore float64
	for _, rating := range ratings {
		// Determine if this is passenger or driver rating based on ride
		ride, err := s.rideRepo.GetByID(rideID)
		if err != nil {
			continue
		}

		if rating.RaterID == ride.PassengerID {
			response.DriverRating = rating
		} else {
			response.PassengerRating = rating
		}

		totalScore += rating.Score
	}

	if len(ratings) > 0 {
		response.AverageScore = totalScore / float64(len(ratings))
	}

	return response, nil
}

// User ratings

func (s *ratingService) GetUserRatings(userID string, page, limit int) ([]*models.Rating, int64, error) {
	return s.ratingRepo.GetUserRatings(userID, page, limit)
}

func (s *ratingService) GetRatingSummary(userID string) (*models.RatingSummary, error) {
	return s.ratingRepo.GetDriverRatingStats(userID)
}

func (s *ratingService) GetAverageRating(userID string) (*AverageRatingResponse, error) {
	average, err := s.ratingRepo.GetAverageRating(userID)
	if err != nil {
		return nil, err
	}

	// Get rating breakdown for distribution
	breakdown, err := s.ratingRepo.GetRatingBreakdown(userID)
	if err != nil {
		breakdown = &repositories.RatingBreakdown{Distribution: make(map[int]int)}
	}

	// Get trend
	trends, err := s.ratingRepo.GetRatingTrends(userID, 30)
	if err != nil {
		trends = &repositories.RatingTrends{Trend: "stable"}
	}

	return &AverageRatingResponse{
		UserID:        userID,
		AverageRating: average,
		TotalRatings:  breakdown.TotalRatings,
		Distribution:  breakdown.Distribution,
		RecentTrend:   trends.Trend,
	}, nil
}

// Driver ratings

func (s *ratingService) GetDriverRatings(driverID string, page, limit int) ([]*models.Rating, int64, error) {
	return s.ratingRepo.GetDriverRatings(driverID, page, limit)
}

func (s *ratingService) GetDriverRatingStats(driverID string) (*DriverRatingStats, error) {
	summary, err := s.ratingRepo.GetDriverRatingStats(driverID)
	if err != nil {
		return nil, err
	}

	// Get recent top reviews
	ratings, _, err := s.ratingRepo.GetDriverRatings(driverID, 1, 5)
	if err != nil {
		ratings = []*models.Rating{}
	}

	// Filter for high-rated reviews with content
	var topReviews []*models.Rating
	for _, rating := range ratings {
		if rating.Score >= 4.0 && rating.Review != "" {
			topReviews = append(topReviews, rating)
		}
	}

	return &DriverRatingStats{
		Summary:      summary,
		MonthlyStats: []MonthlyRatingStats{}, // Would need additional aggregation
		CategoryAvg:  summary.CategoryScores,
		TopReviews:   topReviews,
	}, nil
}

func (s *ratingService) GetRatingBreakdown(userID string) (*repositories.RatingBreakdown, error) {
	return s.ratingRepo.GetRatingBreakdown(userID)
}

// Rating history

func (s *ratingService) GetRatingHistory(userID string, page, limit int) ([]*models.Rating, int64, error) {
	return s.ratingRepo.GetUserRatings(userID, page, limit)
}

func (s *ratingService) GetGivenRatings(userID string, page, limit int) ([]*models.Rating, int64, error) {
	return s.ratingRepo.GetUserRatingsAsRater(userID, page, limit)
}

func (s *ratingService) GetReceivedRatings(userID string, page, limit int) ([]*models.Rating, int64, error) {
	return s.ratingRepo.GetUserRatingsAsRated(userID, page, limit)
}

// Rating analytics

func (s *ratingService) GetRatingAnalytics(userID string) (*RatingAnalytics, error) {
	// Get overall stats
	summary, err := s.ratingRepo.GetDriverRatingStats(userID)
	if err != nil {
		return nil, err
	}

	// Get trend analysis
	trends, err := s.ratingRepo.GetRatingTrends(userID, 90)
	if err != nil {
		trends = &repositories.RatingTrends{}
	}

	// Generate recommendations based on ratings
	recommendations := s.generateRecommendations(summary, trends)

	return &RatingAnalytics{
		UserID:              userID,
		OverallStats:        summary,
		TrendAnalysis:       trends,
		CategoryPerformance: []CategoryPerformance{}, // Would need category analysis
		ComparisonStats:     &ComparisonStats{},      // Would need city/service comparison
		Recommendations:     recommendations,
	}, nil
}

func (s *ratingService) GetRatingTrends(userID string, days int) (*repositories.RatingTrends, error) {
	return s.ratingRepo.GetRatingTrends(userID, days)
}

func (s *ratingService) GetRatingComparison(userID, compareWithID string) (*repositories.RatingComparison, error) {
	return s.ratingRepo.GetRatingComparison(userID, compareWithID)
}

// Rating categories and criteria

func (s *ratingService) GetRatingCategories() ([]RatingCategory, error) {
	// Return predefined categories
	categories := []RatingCategory{
		{
			ID:          "cleanliness",
			Name:        "Cleanliness",
			Description: "Vehicle cleanliness and hygiene",
			Icon:        "🧽",
			Weight:      0.2,
		},
		{
			ID:          "punctuality",
			Name:        "Punctuality",
			Description: "On-time arrival and departure",
			Icon:        "⏰",
			Weight:      0.25,
		},
		{
			ID:          "communication",
			Name:        "Communication",
			Description: "Friendliness and communication skills",
			Icon:        "💬",
			Weight:      0.2,
		},
		{
			ID:          "safety",
			Name:        "Safety",
			Description: "Safe driving and behavior",
			Icon:        "🛡️",
			Weight:      0.35,
		},
	}

	return categories, nil
}

func (s *ratingService) GetRatingCriteria(ratingType models.RatingType) ([]models.RatingCriteria, error) {
	// This would typically come from database
	// Return empty slice for now - implement based on your requirements
	return []models.RatingCriteria{}, nil
}

// Feedback and reporting

func (s *ratingService) AddFeedback(ratingID, userID string, req *FeedbackRequest) error {
	// Get rating to verify rated user
	rating, err := s.ratingRepo.GetByID(ratingID)
	if err != nil {
		return err
	}

	// Check if user is the rated user
	if rating.RatedUserID.Hex() != userID {
		return errors.New("you can only respond to ratings about you")
	}

	// Check if already responded
	if rating.Response != nil {
		return errors.New("you have already responded to this rating")
	}

	// Create response
	response := &models.RatingResponse{
		Content:     req.Content,
		RespondedAt: time.Now(),
		IsPublic:    req.IsPublic,
	}

	rating.Response = response
	_, err = s.ratingRepo.Update(rating)
	if err != nil {
		s.logger.Error().Err(err).Str("rating_id", ratingID).Msg("Failed to add feedback")
		return fmt.Errorf("failed to add feedback")
	}

	return nil
}

func (s *ratingService) GetFeedback(ratingID string) (*FeedbackResponse, error) {
	rating, err := s.ratingRepo.GetByID(ratingID)
	if err != nil {
		return nil, err
	}

	response := &FeedbackResponse{
		RatingID: ratingID,
		Feedback: []models.RatingFeedback{rating.Feedback},
		CanReply: rating.Response == nil,
	}

	return response, nil
}

func (s *ratingService) ReportRating(ratingID, reporterID string, req *ReportRatingRequest) error {
	// Get rating
	rating, err := s.ratingRepo.GetByID(ratingID)
	if err != nil {
		return err
	}

	// Check if already reported
	if rating.IsReported {
		return errors.New("rating has already been reported")
	}

	// Update rating as reported
	now := time.Now()
	rating.IsReported = true
	rating.ReportedAt = &now
	rating.ReportReason = req.Reason

	_, err = s.ratingRepo.Update(rating)
	if err != nil {
		s.logger.Error().Err(err).Str("rating_id", ratingID).Msg("Failed to report rating")
		return fmt.Errorf("failed to report rating")
	}

	// Log report for admin review
	s.logger.Warn().
		Str("rating_id", ratingID).
		Str("reporter_id", reporterID).
		Str("reason", req.Reason).
		Msg("Rating reported")

	return nil
}

// Rating incentives

func (s *ratingService) GetRatingRewards(userID string) ([]RatingReward, error) {
	// Get user's rating stats
	summary, err := s.ratingRepo.GetDriverRatingStats(userID)
	if err != nil {
		return []RatingReward{}, nil
	}

	// Define available rewards
	rewards := []RatingReward{
		{
			ID:             "excellent_rating",
			Name:           "Excellent Rating Bonus",
			Description:    "Bonus for maintaining 4.8+ rating",
			Type:           "bonus",
			Value:          10.00,
			RequiredRating: 4.8,
			RequiredCount:  20,
			IsEligible:     summary.OverallRating >= 4.8 && summary.TotalRatings >= 20,
		},
		{
			ID:             "consistency_reward",
			Name:           "Consistency Reward",
			Description:    "Reward for 100+ consistent ratings",
			Type:           "badge",
			Value:          0,
			RequiredRating: 4.0,
			RequiredCount:  100,
			IsEligible:     summary.OverallRating >= 4.0 && summary.TotalRatings >= 100,
		},
	}

	return rewards, nil
}

func (s *ratingService) ClaimRatingReward(userID, rewardID string) (*RatingReward, error) {
	// Get available rewards
	rewards, err := s.GetRatingRewards(userID)
	if err != nil {
		return nil, err
	}

	// Find specific reward
	for _, reward := range rewards {
		if reward.ID == rewardID {
			if !reward.IsEligible {
				return nil, errors.New("not eligible for this reward")
			}

			// Mark as claimed
			now := time.Now()
			reward.ClaimedAt = &now

			// Log reward claim
			s.logger.Info().
				Str("user_id", userID).
				Str("reward_id", rewardID).
				Msg("Rating reward claimed")

			return &reward, nil
		}
	}

	return nil, errors.New("reward not found")
}

// Validation and business rules

func (s *ratingService) ValidateRating(rating *models.Rating) error {
	if rating.Score < 1 || rating.Score > 5 {
		return errors.New("rating score must be between 1 and 5")
	}

	if len(rating.Review) > 1000 {
		return errors.New("review text too long (max 1000 characters)")
	}

	if rating.RaterID == rating.RatedUserID {
		return errors.New("users cannot rate themselves")
	}

	return nil
}

func (s *ratingService) CanUserRateRide(userID, rideID string) (bool, error) {
	// Check if user already rated this ride
	hasRated, err := s.ratingRepo.HasUserRatedRide(userID, rideID)
	if err != nil {
		return false, err
	}
	if hasRated {
		return false, nil
	}

	// Get ride to verify user participated
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return false, err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	// Check if user was passenger or driver
	isParticipant := ride.PassengerID == userObjectID || ride.DriverID == userObjectID

	// Check if ride is completed
	isCompleted := ride.Status == "completed"

	return isParticipant && isCompleted, nil
}

func (s *ratingService) CalculateRatingImpact(userID string, newRating float64) (*RatingImpact, error) {
	currentAvg, err := s.ratingRepo.GetAverageRating(userID)
	if err != nil {
		currentAvg = 0
	}

	// Simple calculation - in practice you'd want more sophisticated logic
	summary, err := s.ratingRepo.GetDriverRatingStats(userID)
	if err != nil {
		return nil, err
	}

	totalRatings := float64(summary.TotalRatings)
	newAverage := ((currentAvg * totalRatings) + newRating) / (totalRatings + 1)
	change := newAverage - currentAvg

	impact := "neutral"
	if change > 0.1 {
		impact = "positive"
	} else if change < -0.1 {
		impact = "negative"
	}

	return &RatingImpact{
		CurrentRating: currentAvg,
		NewRating:     newAverage,
		Change:        change,
		Impact:        impact,
		NewRank:       0, // Would need ranking calculation
	}, nil
}

// Helper methods

func (s *ratingService) notifyUserRated(rating *models.Rating) {
	if s.notificationSvc == nil {
		return
	}

	notification := &NotificationRequest{
		Type:    "rating_received",
		Title:   "New Rating Received",
		Message: fmt.Sprintf("You received a %.1f star rating!", rating.Score),
		Data: map[string]interface{}{
			"rating_id": rating.ID.Hex(),
			"score":     rating.Score,
		},
	}

	err := s.notificationSvc.SendNotificationToUser(rating.RatedUserID.Hex(), notification)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to send rating notification")
	}
}

func (s *ratingService) updateUserRatingSummary(userID string) {
	err := s.ratingRepo.UpdateRatingSummary(userID)
	if err != nil {
		s.logger.Warn().Err(err).Str("user_id", userID).Msg("Failed to update rating summary")
	}
}

func (s *ratingService) generateRecommendations(summary *models.RatingSummary, trends *repositories.RatingTrends) []string {
	var recommendations []string

	if summary.OverallRating < 4.0 {
		recommendations = append(recommendations, "Focus on improving customer service to increase your rating")
	}

	if trends.Trend == "declining" {
		recommendations = append(recommendations, "Your rating trend is declining - consider reviewing recent feedback")
	}

	if summary.TotalRatings < 10 {
		recommendations = append(recommendations, "Complete more rides to establish a solid rating profile")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Keep up the excellent work!")
	}

	return recommendations
}
