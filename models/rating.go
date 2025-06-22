package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RatingType string

const (
	RatingTypeRide      RatingType = "ride"
	RatingTypeDriver    RatingType = "driver"
	RatingTypePassenger RatingType = "passenger"
	RatingTypeService   RatingType = "service"
	RatingTypeApp       RatingType = "app"
)

type Rating struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// Rating Details
	Type   RatingType `json:"type" bson:"type"`
	Score  float64    `json:"score" bson:"score" validate:"gte=1,lte=5"`
	Review string     `json:"review" bson:"review" validate:"max=1000"`

	// Participants
	RaterID     primitive.ObjectID  `json:"rater_id" bson:"rater_id"`           // Who gave the rating
	RatedUserID primitive.ObjectID  `json:"rated_user_id" bson:"rated_user_id"` // Who received the rating
	RideID      *primitive.ObjectID `json:"ride_id,omitempty" bson:"ride_id,omitempty"`

	// Detailed Ratings (breakdown)
	Categories []CategoryRating `json:"categories" bson:"categories"`

	// Feedback
	Feedback RatingFeedback `json:"feedback" bson:"feedback"`

	// Media
	Photos []string `json:"photos" bson:"photos"`

	// Status
	IsPublic    bool       `json:"is_public" bson:"is_public"`
	IsAnonymous bool       `json:"is_anonymous" bson:"is_anonymous"`
	IsEdited    bool       `json:"is_edited" bson:"is_edited"`
	EditedAt    *time.Time `json:"edited_at,omitempty" bson:"edited_at,omitempty"`

	// Response from Rated User
	Response *RatingResponse `json:"response,omitempty" bson:"response,omitempty"`

	// Admin Moderation
	IsReported   bool                `json:"is_reported" bson:"is_reported"`
	ReportedAt   *time.Time          `json:"reported_at,omitempty" bson:"reported_at,omitempty"`
	ReportReason string              `json:"report_reason" bson:"report_reason"`
	IsHidden     bool                `json:"is_hidden" bson:"is_hidden"`
	HiddenAt     *time.Time          `json:"hidden_at,omitempty" bson:"hidden_at,omitempty"`
	HiddenBy     *primitive.ObjectID `json:"hidden_by,omitempty" bson:"hidden_by,omitempty"`

	// Metadata
	Platform   string `json:"platform" bson:"platform"`
	AppVersion string `json:"app_version" bson:"app_version"`
	Language   string `json:"language" bson:"language"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type CategoryRating struct {
	Category string  `json:"category" bson:"category"` // cleanliness, punctuality, communication, safety, vehicle_condition
	Score    float64 `json:"score" bson:"score" validate:"gte=1,lte=5"`
	Weight   float64 `json:"weight" bson:"weight"` // Weight for overall calculation
}

type RatingFeedback struct {
	// Positive Feedback
	PositiveAspects []string `json:"positive_aspects" bson:"positive_aspects"`
	Compliments     []string `json:"compliments" bson:"compliments"`

	// Negative Feedback
	Issues     []string `json:"issues" bson:"issues"`
	Complaints []string `json:"complaints" bson:"complaints"`

	// Suggestions
	Suggestions []string `json:"suggestions" bson:"suggestions"`

	// Quick Feedback (predefined options)
	QuickFeedback []string `json:"quick_feedback" bson:"quick_feedback"`

	// Additional Notes
	Notes string `json:"notes" bson:"notes"`
}

type RatingResponse struct {
	Content     string    `json:"content" bson:"content" validate:"max=500"`
	RespondedAt time.Time `json:"responded_at" bson:"responded_at"`
	IsPublic    bool      `json:"is_public" bson:"is_public"`
}

type RatingSummary struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	UserType string             `json:"user_type" bson:"user_type"` // driver, passenger

	// Overall Rating
	OverallRating float64 `json:"overall_rating" bson:"overall_rating"`
	TotalRatings  int     `json:"total_ratings" bson:"total_ratings"`
	TotalReviews  int     `json:"total_reviews" bson:"total_reviews"`

	// Rating Distribution
	FiveStars  int `json:"five_stars" bson:"five_stars"`
	FourStars  int `json:"four_stars" bson:"four_stars"`
	ThreeStars int `json:"three_stars" bson:"three_stars"`
	TwoStars   int `json:"two_stars" bson:"two_stars"`
	OneStar    int `json:"one_star" bson:"one_star"`

	// Category Ratings
	CategoryScores []CategoryScore `json:"category_scores" bson:"category_scores"`

	// Recent Performance
	RecentRating float64 `json:"recent_rating" bson:"recent_rating"` // Last 30 days
	RatingTrend  string  `json:"rating_trend" bson:"rating_trend"`   // improving, declining, stable

	// Achievement Badges
	Badges []RatingBadge `json:"badges" bson:"badges"`

	// Top Reviews (most helpful)
	TopPositiveReview *Rating `json:"top_positive_review,omitempty" bson:"top_positive_review,omitempty"`
	TopNegativeReview *Rating `json:"top_negative_review,omitempty" bson:"top_negative_review,omitempty"`

	// Statistics
	Stats RatingStats `json:"stats" bson:"stats"`

	LastUpdatedAt time.Time `json:"last_updated_at" bson:"last_updated_at"`
}

type CategoryScore struct {
	Category     string  `json:"category" bson:"category"`
	AverageScore float64 `json:"average_score" bson:"average_score"`
	TotalRatings int     `json:"total_ratings" bson:"total_ratings"`
}

type RatingBadge struct {
	ID          string    `json:"id" bson:"id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	Icon        string    `json:"icon" bson:"icon"`
	Color       string    `json:"color" bson:"color"`
	EarnedAt    time.Time `json:"earned_at" bson:"earned_at"`
	Level       int       `json:"level" bson:"level"` // Bronze=1, Silver=2, Gold=3
	IsActive    bool      `json:"is_active" bson:"is_active"`
}

type RatingStats struct {
	// Response Statistics
	ResponseRate        float64 `json:"response_rate" bson:"response_rate"`                 // % of ratings with responses
	AverageResponseTime float64 `json:"average_response_time" bson:"average_response_time"` // hours

	// Review Statistics
	ReviewRate          float64 `json:"review_rate" bson:"review_rate"` // % of ratings with reviews
	AverageReviewLength int     `json:"average_review_length" bson:"average_review_length"`

	// Quality Metrics
	HelpfulnessScore float64 `json:"helpfulness_score" bson:"helpfulness_score"`
	VerifiedRatings  int     `json:"verified_ratings" bson:"verified_ratings"`

	// Time-based Statistics
	RatingsThisMonth int    `json:"ratings_this_month" bson:"ratings_this_month"`
	RatingsLastMonth int    `json:"ratings_last_month" bson:"ratings_last_month"`
	BestMonth        string `json:"best_month" bson:"best_month"`
	WorstMonth       string `json:"worst_month" bson:"worst_month"`

	// Comparative Statistics
	BetterThanPercent float64 `json:"better_than_percent" bson:"better_than_percent"` // Better than X% of users
	RankInCity        int     `json:"rank_in_city" bson:"rank_in_city"`

	// Improvement Metrics
	RatingImprovement float64 `json:"rating_improvement" bson:"rating_improvement"` // Change over time
	ConsistencyScore  float64 `json:"consistency_score" bson:"consistency_score"`   // How consistent ratings are
}

type RatingCriteria struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Category    string             `json:"category" bson:"category"`
	Type        RatingType         `json:"type" bson:"type"`
	Weight      float64            `json:"weight" bson:"weight"`
	IsActive    bool               `json:"is_active" bson:"is_active"`
	SortOrder   int                `json:"sort_order" bson:"sort_order"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

type RatingIncentive struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name"`
	Description    string             `json:"description" bson:"description"`
	Type           string             `json:"type" bson:"type"` // points, discount, badge, cash
	Value          float64            `json:"value" bson:"value"`
	RequiredRating float64            `json:"required_rating" bson:"required_rating"`
	RequiredCount  int                `json:"required_count" bson:"required_count"`
	ValidityDays   int                `json:"validity_days" bson:"validity_days"`
	MaxRedemptions int                `json:"max_redemptions" bson:"max_redemptions"`
	IsActive       bool               `json:"is_active" bson:"is_active"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
}
