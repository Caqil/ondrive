package repositories

import (
	"context"
	"fmt"
	"time"

	"ondrive/models"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RatingRepository interface for rating operations
type RatingRepository interface {
	// Basic CRUD operations
	Create(rating *models.Rating) (*models.Rating, error)
	GetByID(id string) (*models.Rating, error)
	Update(rating *models.Rating) (*models.Rating, error)
	Delete(id string) error

	// Ride-specific ratings
	GetByRideID(rideID string) ([]*models.Rating, error)
	GetMutualRating(rideID string) ([]*models.Rating, error)
	HasUserRatedRide(userID, rideID string) (bool, error)

	// User ratings
	GetUserRatings(userID string, page, limit int) ([]*models.Rating, int64, error)
	GetUserRatingsAsRater(userID string, page, limit int) ([]*models.Rating, int64, error)
	GetUserRatingsAsRated(userID string, page, limit int) ([]*models.Rating, int64, error)

	// Driver ratings
	GetDriverRatings(driverID string, page, limit int) ([]*models.Rating, int64, error)
	GetDriverRatingStats(driverID string) (*models.RatingSummary, error)
	GetRatingBreakdown(userID string) (*RatingBreakdown, error)

	// Rating analytics
	GetAverageRating(userID string) (float64, error)
	GetRatingHistory(userID string, days int) ([]*RatingHistoryItem, error)
	GetRatingTrends(userID string, days int) (*RatingTrends, error)
	GetRatingComparison(userID string, compareWithID string) (*RatingComparison, error)

	// Rating management
	GetReportedRatings(page, limit int) ([]*models.Rating, int64, error)
	HideRating(ratingID, adminID string) error
	UnhideRating(ratingID string) error

	// Bulk operations
	GetRatingsByUserIDs(userIDs []string) ([]*models.Rating, error)
	UpdateRatingSummary(userID string) error
	DeleteUserRatings(userID string) error
}

// Additional types for analytics
type RatingBreakdown struct {
	Overall        float64                 `json:"overall"`
	Categories     []models.CategoryScore  `json:"categories"`
	Distribution   map[int]int            `json:"distribution"`
	TotalRatings   int                    `json:"total_ratings"`
	RecentRatings  []*models.Rating       `json:"recent_ratings"`
}

type RatingHistoryItem struct {
	Date         time.Time `json:"date"`
	AverageScore float64   `json:"average_score"`
	Count        int       `json:"count"`
}

type RatingTrends struct {
	UserID       string                `json:"user_id"`
	Period       string                `json:"period"`
	Trend        string                `json:"trend"` // improving, declining, stable
	CurrentAvg   float64               `json:"current_avg"`
	PreviousAvg  float64               `json:"previous_avg"`
	Change       float64               `json:"change"`
	HistoryItems []*RatingHistoryItem  `json:"history_items"`
}

type RatingComparison struct {
	UserA        string  `json:"user_a"`
	UserB        string  `json:"user_b"`
	UserARating  float64 `json:"user_a_rating"`
	UserBRating  float64 `json:"user_b_rating"`
	Difference   float64 `json:"difference"`
	BetterUser   string  `json:"better_user"`
}

// Implementation
type ratingRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

// NewRatingRepository creates a new rating repository
func NewRatingRepository(db *mongo.Database, logger utils.Logger) RatingRepository {
	return &ratingRepository{
		db:     db,
		logger: logger,
	}
}

// Basic CRUD operations

func (r *ratingRepository) Create(rating *models.Rating) (*models.Rating, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rating.ID = primitive.NewObjectID()
	rating.CreatedAt = time.Now()
	rating.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, rating)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create rating")
		return nil, fmt.Errorf("failed to create rating: %w", err)
	}

	r.logger.Info().
		Str("rating_id", rating.ID.Hex()).
		Str("rater_id", rating.RaterID.Hex()).
		Str("rated_user_id", rating.RatedUserID.Hex()).
		Float64("score", rating.Score).
		Msg("Rating created successfully")

	return rating, nil
}

func (r *ratingRepository) GetByID(id string) (*models.Rating, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var rating models.Rating
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&rating)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("rating not found")
		}
		r.logger.Error().Err(err).Str("rating_id", id).Msg("Failed to get rating")
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	return &rating, nil
}

func (r *ratingRepository) Update(rating *models.Rating) (*models.Rating, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rating.UpdatedAt = time.Now()
	rating.IsEdited = true
	rating.EditedAt = &rating.UpdatedAt

	_, err := collection.ReplaceOne(ctx, bson.M{"_id": rating.ID}, rating)
	if err != nil {
		r.logger.Error().Err(err).Str("rating_id", rating.ID.Hex()).Msg("Failed to update rating")
		return nil, fmt.Errorf("failed to update rating: %w", err)
	}

	return rating, nil
}

func (r *ratingRepository) Delete(id string) error {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		r.logger.Error().Err(err).Str("rating_id", id).Msg("Failed to delete rating")
		return fmt.Errorf("failed to delete rating: %w", err)
	}

	return nil
}

// Ride-specific ratings

func (r *ratingRepository) GetByRideID(rideID string) ([]*models.Rating, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rideObjectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(ctx, bson.M{"ride_id": rideObjectID})
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to get ratings by ride ID")
		return nil, fmt.Errorf("failed to get ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*models.Rating
	if err = cursor.All(ctx, &ratings); err != nil {
		return nil, fmt.Errorf("failed to decode ratings: %w", err)
	}

	return ratings, nil
}

func (r *ratingRepository) GetMutualRating(rideID string) ([]*models.Rating, error) {
	return r.GetByRideID(rideID)
}

func (r *ratingRepository) HasUserRatedRide(userID, rideID string) (bool, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	rideObjectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return false, err
	}

	count, err := collection.CountDocuments(ctx, bson.M{
		"rater_id": userObjectID,
		"ride_id":  rideObjectID,
	})

	return count > 0, err
}

// User ratings

func (r *ratingRepository) GetUserRatings(userID string, page, limit int) ([]*models.Rating, int64, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"$or": []bson.M{
			{"rater_id": userObjectID},
			{"rated_user_id": userObjectID},
		},
	}

	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ratings: %w", err)
	}

	// Calculate pagination
	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*models.Rating
	if err = cursor.All(ctx, &ratings); err != nil {
		return nil, 0, fmt.Errorf("failed to decode ratings: %w", err)
	}

	return ratings, total, nil
}

func (r *ratingRepository) GetUserRatingsAsRater(userID string, page, limit int) ([]*models.Rating, int64, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"rater_id": userObjectID}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ratings: %w", err)
	}

	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ratings as rater: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*models.Rating
	if err = cursor.All(ctx, &ratings); err != nil {
		return nil, 0, fmt.Errorf("failed to decode ratings: %w", err)
	}

	return ratings, total, nil
}

func (r *ratingRepository) GetUserRatingsAsRated(userID string, page, limit int) ([]*models.Rating, int64, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"rated_user_id": userObjectID}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ratings: %w", err)
	}

	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ratings as rated: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*models.Rating
	if err = cursor.All(ctx, &ratings); err != nil {
		return nil, 0, fmt.Errorf("failed to decode ratings: %w", err)
	}

	return ratings, total, nil
}

// Driver ratings

func (r *ratingRepository) GetDriverRatings(driverID string, page, limit int) ([]*models.Rating, int64, error) {
	return r.GetUserRatingsAsRated(driverID, page, limit)
}

func (r *ratingRepository) GetDriverRatingStats(driverID string) (*models.RatingSummary, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	driverObjectID, err := primitive.ObjectIDFromHex(driverID)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{"$match": bson.M{"rated_user_id": driverObjectID}},
		{"$group": bson.M{
			"_id":           nil,
			"total_ratings": bson.M{"$sum": 1},
			"average_score": bson.M{"$avg": "$score"},
			"five_stars":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$score", 5}}, 1, 0}}},
			"four_stars":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$score", 4}}, 1, 0}}},
			"three_stars":   bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$score", 3}}, 1, 0}}},
			"two_stars":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$score", 2}}, 1, 0}}},
			"one_star":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$score", 1}}, 1, 0}}},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error().Err(err).Str("driver_id", driverID).Msg("Failed to get driver rating stats")
		return nil, fmt.Errorf("failed to get driver rating stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalRatings int     `bson:"total_ratings"`
		AverageScore float64 `bson:"average_score"`
		FiveStars    int     `bson:"five_stars"`
		FourStars    int     `bson:"four_stars"`
		ThreeStars   int     `bson:"three_stars"`
		TwoStars     int     `bson:"two_stars"`
		OneStar      int     `bson:"one_star"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode stats: %w", err)
		}
	}

	summary := &models.RatingSummary{
		UserID:        driverObjectID,
		UserType:      "driver",
		OverallRating: result.AverageScore,
		TotalRatings:  result.TotalRatings,
		FiveStars:     result.FiveStars,
		FourStars:     result.FourStars,
		ThreeStars:    result.ThreeStars,
		TwoStars:      result.TwoStars,
		OneStar:       result.OneStar,
		LastUpdatedAt: time.Now(),
	}

	return summary, nil
}

func (r *ratingRepository) GetRatingBreakdown(userID string) (*RatingBreakdown, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Get recent ratings
	cursor, err := collection.Find(ctx, 
		bson.M{"rated_user_id": userObjectID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(10),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var recentRatings []*models.Rating
	if err = cursor.All(ctx, &recentRatings); err != nil {
		return nil, fmt.Errorf("failed to decode ratings: %w", err)
	}

	// Calculate overall average
	average, err := r.GetAverageRating(userID)
	if err != nil {
		average = 0
	}

	// Create distribution map
	distribution := make(map[int]int)
	for _, rating := range recentRatings {
		distribution[int(rating.Score)]++
	}

	breakdown := &RatingBreakdown{
		Overall:       average,
		Categories:    []models.CategoryScore{}, // Would need to aggregate category data
		Distribution:  distribution,
		TotalRatings:  len(recentRatings),
		RecentRatings: recentRatings,
	}

	return breakdown, nil
}

// Rating analytics

func (r *ratingRepository) GetAverageRating(userID string) (float64, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	pipeline := []bson.M{
		{"$match": bson.M{"rated_user_id": userObjectID}},
		{"$group": bson.M{
			"_id":           nil,
			"average_score": bson.M{"$avg": "$score"},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate average rating: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		AverageScore float64 `bson:"average_score"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, fmt.Errorf("failed to decode average: %w", err)
		}
		return result.AverageScore, nil
	}

	return 0, nil
}

func (r *ratingRepository) GetRatingHistory(userID string, days int) ([]*RatingHistoryItem, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	startDate := time.Now().AddDate(0, 0, -days)

	pipeline := []bson.M{
		{"$match": bson.M{
			"rated_user_id": userObjectID,
			"created_at":    bson.M{"$gte": startDate},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"date": bson.M{"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$created_at",
				}},
			},
			"average_score": bson.M{"$avg": "$score"},
			"count":         bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id.date": 1}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating history: %w", err)
	}
	defer cursor.Close(ctx)

	var history []*RatingHistoryItem
	for cursor.Next(ctx) {
		var result struct {
			ID struct {
				Date string `bson:"date"`
			} `bson:"_id"`
			AverageScore float64 `bson:"average_score"`
			Count        int     `bson:"count"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		date, _ := time.Parse("2006-01-02", result.ID.Date)
		history = append(history, &RatingHistoryItem{
			Date:         date,
			AverageScore: result.AverageScore,
			Count:        result.Count,
		})
	}

	return history, nil
}

func (r *ratingRepository) GetRatingTrends(userID string, days int) (*RatingTrends, error) {
	history, err := r.GetRatingHistory(userID, days)
	if err != nil {
		return nil, err
	}

	if len(history) < 2 {
		return &RatingTrends{
			UserID:       userID,
			Period:       fmt.Sprintf("%d days", days),
			Trend:        "stable",
			HistoryItems: history,
		}, nil
	}

	currentAvg := history[len(history)-1].AverageScore
	previousAvg := history[0].AverageScore
	change := currentAvg - previousAvg

	trend := "stable"
	if change > 0.1 {
		trend = "improving"
	} else if change < -0.1 {
		trend = "declining"
	}

	return &RatingTrends{
		UserID:       userID,
		Period:       fmt.Sprintf("%d days", days),
		Trend:        trend,
		CurrentAvg:   currentAvg,
		PreviousAvg:  previousAvg,
		Change:       change,
		HistoryItems: history,
	}, nil
}

func (r *ratingRepository) GetRatingComparison(userID string, compareWithID string) (*RatingComparison, error) {
	userARating, err := r.GetAverageRating(userID)
	if err != nil {
		return nil, err
	}

	userBRating, err := r.GetAverageRating(compareWithID)
	if err != nil {
		return nil, err
	}

	difference := userARating - userBRating
	betterUser := userID
	if userBRating > userARating {
		betterUser = compareWithID
	}

	return &RatingComparison{
		UserA:       userID,
		UserB:       compareWithID,
		UserARating: userARating,
		UserBRating: userBRating,
		Difference:  difference,
		BetterUser:  betterUser,
	}, nil
}

// Rating management

func (r *ratingRepository) GetReportedRatings(page, limit int) ([]*models.Rating, int64, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"is_reported": true}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reported ratings: %w", err)
	}

	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "reported_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get reported ratings: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*models.Rating
	if err = cursor.All(ctx, &ratings); err != nil {
		return nil, 0, fmt.Errorf("failed to decode reported ratings: %w", err)
	}

	return ratings, total, nil
}

func (r *ratingRepository) HideRating(ratingID, adminID string) error {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ratingObjectID, err := primitive.ObjectIDFromHex(ratingID)
	if err != nil {
		return err
	}

	adminObjectID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return err
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_hidden":  true,
			"hidden_at":  now,
			"hidden_by":  adminObjectID,
			"updated_at": now,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": ratingObjectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("rating_id", ratingID).Msg("Failed to hide rating")
		return fmt.Errorf("failed to hide rating: %w", err)
	}

	return nil
}

func (r *ratingRepository) UnhideRating(ratingID string) error {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ratingObjectID, err := primitive.ObjectIDFromHex(ratingID)
	if err != nil {
		return err
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_hidden":  false,
			"updated_at": now,
		},
		"$unset": bson.M{
			"hidden_at": "",
			"hidden_by": "",
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": ratingObjectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("rating_id", ratingID).Msg("Failed to unhide rating")
		return fmt.Errorf("failed to unhide rating: %w", err)
	}

	return nil
}

// Bulk operations

func (r *ratingRepository) GetRatingsByUserIDs(userIDs []string) ([]*models.Rating, error) {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var objectIDs []primitive.ObjectID
	for _, id := range userIDs {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objectID)
	}

	filter := bson.M{
		"$or": []bson.M{
			{"rater_id": bson.M{"$in": objectIDs}},
			{"rated_user_id": bson.M{"$in": objectIDs}},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings by user IDs: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []*models.Rating
	if err = cursor.All(ctx, &ratings); err != nil {
		return nil, fmt.Errorf("failed to decode ratings: %w", err)
	}

	return ratings, nil
}

func (r *ratingRepository) UpdateRatingSummary(userID string) error {
	// This would update the rating summary for a user
	// Implementation depends on whether you store summaries in a separate collection
	r.logger.Info().Str("user_id", userID).Msg("UpdateRatingSummary called - implementation needed")
	return nil
}

func (r *ratingRepository) DeleteUserRatings(userID string) error {
	collection := r.db.Collection("ratings")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"$or": []bson.M{
			{"rater_id": userObjectID},
			{"rated_user_id": userObjectID},
		},
	}

	_, err = collection.DeleteMany(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user ratings")
		return fmt.Errorf("failed to delete user ratings: %w", err)
	}

	return nil
}