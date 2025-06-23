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

// RideRepository interface for ride operations
type RideRepository interface {
	// Basic CRUD operations
	Create(ride *models.Ride) (*models.Ride, error)
	GetByID(id string) (*models.Ride, error)
	Update(ride *models.Ride) (*models.Ride, error)
	Delete(id string) error

	// User rides
	GetUserRides(userID string, page, limit int, status string) ([]*models.Ride, int64, error)
	GetRidesByPassenger(passengerID string, page, limit int) ([]*models.Ride, int64, error)
	GetRidesByDriver(driverID string, page, limit int) ([]*models.Ride, int64, error)

	// Ride status management
	UpdateRideStatus(rideID string, status models.RideStatus) error
	UpdateRideLocation(rideID string, driverLocation *models.Location) error
	AssignDriver(rideID, driverID string) error

	// Scheduled rides
	GetScheduledRides(userID string, page, limit int) ([]*models.Ride, int64, error)
	GetScheduledRidesForTime(targetTime time.Time) ([]*models.Ride, error)
	UpdateScheduledRide(rideID string, scheduledAt time.Time) error

	// Driver matching
	GetPendingRides(serviceType models.ServiceType, lat, lng, radius float64) ([]*models.Ride, error)
	GetRidesByStatus(status models.RideStatus, page, limit int) ([]*models.Ride, int64, error)

	// Fare management
	UpdateFareDetails(rideID string, fareDetails *models.FareDetails) error
	SaveFareOffer(rideID string, offer *models.FareOffer) error
	GetFareHistory(rideID string) ([]*models.FareOffer, error)

	// Ride tracking
	GetActiveRideForUser(userID string) (*models.Ride, error)
	GetActiveRideForDriver(driverID string) (*models.Ride, error)
	GetRidesByShareCode(shareCode string) (*models.Ride, error)

	// Route and tracking
	UpdateRoute(rideID string, route []models.Location) error
	UpdateEstimatedArrival(rideID string, eta time.Time) error
	UpdateActualDistance(rideID string, distance float64) error

	// Requirements and preferences
	UpdateRideRequirements(rideID string, requirements *models.RideRequirements) error
	UpdateRidePreferences(rideID string, preferences *models.RidePreferences) error

	// Reports and issues
	SaveRideReport(rideID string, report *RideReport) error
	GetRideReports(rideID string) ([]*RideReport, error)

	// Analytics and history
	GetRideHistory(userID string, page, limit int, filters RideHistoryFilter) ([]*models.Ride, int64, error)
	GetFrequentRoutes(userID string, limit int) ([]*FrequentRoute, error)
	GetRideStats(userID string, period string) (*RideStats, error)

	// Bulk operations
	GetRidesByIDs(rideIDs []string) ([]*models.Ride, error)
	UpdateMultipleRideStatuses(rideIDs []string, status models.RideStatus) error
	DeleteUserRides(userID string) error
}

// Additional types for repository operations
type RideHistoryFilter struct {
	Status      *models.RideStatus  `json:"status,omitempty"`
	ServiceType *models.ServiceType `json:"service_type,omitempty"`
	RideType    *models.RideType    `json:"ride_type,omitempty"`
	FromDate    *time.Time          `json:"from_date,omitempty"`
	ToDate      *time.Time          `json:"to_date,omitempty"`
	MinFare     *float64            `json:"min_fare,omitempty"`
	MaxFare     *float64            `json:"max_fare,omitempty"`
	City        string              `json:"city,omitempty"`
}

type FrequentRoute struct {
	PickupLocation  models.RideLocation `json:"pickup_location"`
	DropoffLocation models.RideLocation `json:"dropoff_location"`
	Count           int                 `json:"count"`
	LastUsed        time.Time           `json:"last_used"`
	AverageFare     float64             `json:"average_fare"`
	AverageDuration int                 `json:"average_duration"`
}

type RideStats struct {
	UserID         string          `json:"user_id"`
	Period         string          `json:"period"`
	TotalRides     int             `json:"total_rides"`
	CompletedRides int             `json:"completed_rides"`
	CancelledRides int             `json:"cancelled_rides"`
	TotalSpent     float64         `json:"total_spent"`
	TotalDistance  float64         `json:"total_distance"`
	TotalDuration  int             `json:"total_duration"`
	AverageRating  float64         `json:"average_rating"`
	TopRoutes      []FrequentRoute `json:"top_routes"`
}

type RideReport struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RideID      primitive.ObjectID `json:"ride_id" bson:"ride_id"`
	ReporterID  primitive.ObjectID `json:"reporter_id" bson:"reporter_id"`
	Type        string             `json:"type" bson:"type"`
	Category    string             `json:"category" bson:"category"`
	Subject     string             `json:"subject" bson:"subject"`
	Description string             `json:"description" bson:"description"`
	Status      string             `json:"status" bson:"status"`
	Priority    string             `json:"priority" bson:"priority"`
	Photos      []string           `json:"photos" bson:"photos"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Implementation
type rideRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

// NewRideRepository creates a new ride repository
func NewRideRepository(db *mongo.Database, logger utils.Logger) RideRepository {
	return &rideRepository{
		db:     db,
		logger: logger,
	}
}

// Basic CRUD operations

func (r *rideRepository) Create(ride *models.Ride) (*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ride.ID = primitive.NewObjectID()
	ride.CreatedAt = time.Now()
	ride.UpdatedAt = time.Now()
	ride.Status = models.RideStatusPending

	// Generate share code
	if ride.ShareCode == "" {
		ride.ShareCode = utils.GenerateShortCode(8)
	}

	_, err := collection.InsertOne(ctx, ride)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create ride")
		return nil, fmt.Errorf("failed to create ride: %w", err)
	}

	r.logger.Info().
		Str("ride_id", ride.ID.Hex()).
		Str("passenger_id", ride.PassengerID.Hex()).
		Str("service_type", string(ride.ServiceType)).
		Msg("Ride created successfully")

	return ride, nil
}

func (r *rideRepository) GetByID(id string) (*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var ride models.Ride
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&ride)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("ride not found")
		}
		r.logger.Error().Err(err).Str("ride_id", id).Msg("Failed to get ride")
		return nil, fmt.Errorf("failed to get ride: %w", err)
	}

	return &ride, nil
}

func (r *rideRepository) Update(ride *models.Ride) (*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ride.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(ctx, bson.M{"_id": ride.ID}, ride)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", ride.ID.Hex()).Msg("Failed to update ride")
		return nil, fmt.Errorf("failed to update ride: %w", err)
	}

	return ride, nil
}

func (r *rideRepository) Delete(id string) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", id).Msg("Failed to delete ride")
		return fmt.Errorf("failed to delete ride: %w", err)
	}

	return nil
}

// User rides

func (r *rideRepository) GetUserRides(userID string, page, limit int, status string) ([]*models.Ride, int64, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"$or": []bson.M{
			{"passenger_id": userObjectID},
			{"driver_id": userObjectID},
		},
	}

	if status != "" {
		filter["status"] = status
	}

	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count rides: %w", err)
	}

	// Calculate pagination
	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user rides: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, 0, fmt.Errorf("failed to decode rides: %w", err)
	}

	return rides, total, nil
}

func (r *rideRepository) GetRidesByPassenger(passengerID string, page, limit int) ([]*models.Ride, int64, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	passengerObjectID, err := primitive.ObjectIDFromHex(passengerID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"passenger_id": passengerObjectID}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count passenger rides: %w", err)
	}

	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get passenger rides: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, 0, fmt.Errorf("failed to decode rides: %w", err)
	}

	return rides, total, nil
}

func (r *rideRepository) GetRidesByDriver(driverID string, page, limit int) ([]*models.Ride, int64, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	driverObjectID, err := primitive.ObjectIDFromHex(driverID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"driver_id": driverObjectID}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count driver rides: %w", err)
	}

	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get driver rides: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, 0, fmt.Errorf("failed to decode rides: %w", err)
	}

	return rides, total, nil
}

// Ride status management

func (r *rideRepository) UpdateRideStatus(rideID string, status models.RideStatus) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// Set timestamp based on status
	now := time.Now()
	switch status {
	case models.RideStatusAccepted:
		update["$set"].(bson.M)["accepted_at"] = now
	case models.RideStatusStarted:
		update["$set"].(bson.M)["started_at"] = now
	case models.RideStatusCompleted:
		update["$set"].(bson.M)["completed_at"] = now
	case models.RideStatusCancelled:
		update["$set"].(bson.M)["cancelled_at"] = now
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Str("status", string(status)).Msg("Failed to update ride status")
		return fmt.Errorf("failed to update ride status: %w", err)
	}

	return nil
}

func (r *rideRepository) UpdateRideLocation(rideID string, driverLocation *models.Location) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"driver_location":      driverLocation,
			"last_location_update": time.Now(),
			"updated_at":           time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update ride location")
		return fmt.Errorf("failed to update ride location: %w", err)
	}

	return nil
}

func (r *rideRepository) AssignDriver(rideID, driverID string) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rideObjectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	driverObjectID, err := primitive.ObjectIDFromHex(driverID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"driver_id":  driverObjectID,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": rideObjectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Str("driver_id", driverID).Msg("Failed to assign driver")
		return fmt.Errorf("failed to assign driver: %w", err)
	}

	return nil
}

// Scheduled rides

func (r *rideRepository) GetScheduledRides(userID string, page, limit int) ([]*models.Ride, int64, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"passenger_id": userObjectID,
		"type":         models.RideTypeScheduled,
		"scheduled_at": bson.M{"$gt": time.Now()},
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count scheduled rides: %w", err)
	}

	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "scheduled_at", Value: 1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get scheduled rides: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, 0, fmt.Errorf("failed to decode rides: %w", err)
	}

	return rides, total, nil
}

func (r *rideRepository) GetScheduledRidesForTime(targetTime time.Time) ([]*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get rides scheduled within 15 minutes of target time
	startTime := targetTime.Add(-15 * time.Minute)
	endTime := targetTime.Add(15 * time.Minute)

	filter := bson.M{
		"type":         models.RideTypeScheduled,
		"status":       models.RideStatusPending,
		"scheduled_at": bson.M{"$gte": startTime, "$lte": endTime},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled rides for time: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, fmt.Errorf("failed to decode rides: %w", err)
	}

	return rides, nil
}

func (r *rideRepository) UpdateScheduledRide(rideID string, scheduledAt time.Time) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"scheduled_at": scheduledAt,
			"updated_at":   time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update scheduled ride")
		return fmt.Errorf("failed to update scheduled ride: %w", err)
	}

	return nil
}

// Driver matching

func (r *rideRepository) GetPendingRides(serviceType models.ServiceType, lat, lng, radius float64) ([]*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"status":       models.RideStatusPending,
		"service_type": serviceType,
		"driver_id":    bson.M{"$exists": false},
	}

	// Add geospatial filter if coordinates provided
	if lat != 0 && lng != 0 && radius > 0 {
		filter["pickup_location.coordinates"] = bson.M{
			"$geoWithin": bson.M{
				"$centerSphere": []interface{}{
					[]float64{lng, lat},
					radius / 6378.1, // Convert km to radians
				},
			},
		}
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending rides: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, fmt.Errorf("failed to decode rides: %w", err)
	}

	return rides, nil
}

func (r *rideRepository) GetRidesByStatus(status models.RideStatus, page, limit int) ([]*models.Ride, int64, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"status": status}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count rides by status: %w", err)
	}

	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get rides by status: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, 0, fmt.Errorf("failed to decode rides: %w", err)
	}

	return rides, total, nil
}

// Fare management

func (r *rideRepository) UpdateFareDetails(rideID string, fareDetails *models.FareDetails) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"fare_details": fareDetails,
			"updated_at":   time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update fare details")
		return fmt.Errorf("failed to update fare details: %w", err)
	}

	return nil
}

func (r *rideRepository) SaveFareOffer(rideID string, offer *models.FareOffer) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	offer.ID = primitive.NewObjectID()
	offer.OfferedAt = time.Now()

	update := bson.M{
		"$push": bson.M{
			"fare_details.negotiations": offer,
		},
		"$inc": bson.M{
			"fare_details.negotiation_count": 1,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to save fare offer")
		return fmt.Errorf("failed to save fare offer: %w", err)
	}

	return nil
}

func (r *rideRepository) GetFareHistory(rideID string) ([]*models.FareOffer, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return nil, err
	}

	var ride models.Ride
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&ride)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride for fare history: %w", err)
	}

	return ride.FareDetails.Negotiations, nil
}

// Ride tracking

func (r *rideRepository) GetActiveRideForUser(userID string) (*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	activeStatuses := []models.RideStatus{
		models.RideStatusPending,
		models.RideStatusFareNegotiation,
		models.RideStatusAccepted,
		models.RideStatusDriverEnRoute,
		models.RideStatusDriverArrived,
		models.RideStatusStarted,
		models.RideStatusInProgress,
	}

	filter := bson.M{
		"$or": []bson.M{
			{"passenger_id": userObjectID},
			{"driver_id": userObjectID},
		},
		"status": bson.M{"$in": activeStatuses},
	}

	var ride models.Ride
	err = collection.FindOne(ctx, filter, options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})).Decode(&ride)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No active ride found
		}
		return nil, fmt.Errorf("failed to get active ride: %w", err)
	}

	return &ride, nil
}

func (r *rideRepository) GetActiveRideForDriver(driverID string) (*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	driverObjectID, err := primitive.ObjectIDFromHex(driverID)
	if err != nil {
		return nil, err
	}

	activeStatuses := []models.RideStatus{
		models.RideStatusAccepted,
		models.RideStatusDriverEnRoute,
		models.RideStatusDriverArrived,
		models.RideStatusStarted,
		models.RideStatusInProgress,
	}

	filter := bson.M{
		"driver_id": driverObjectID,
		"status":    bson.M{"$in": activeStatuses},
	}

	var ride models.Ride
	err = collection.FindOne(ctx, filter).Decode(&ride)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No active ride found
		}
		return nil, fmt.Errorf("failed to get active ride for driver: %w", err)
	}

	return &ride, nil
}

func (r *rideRepository) GetRidesByShareCode(shareCode string) (*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ride models.Ride
	err := collection.FindOne(ctx, bson.M{"share_code": shareCode, "is_shared": true}).Decode(&ride)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("shared ride not found")
		}
		return nil, fmt.Errorf("failed to get shared ride: %w", err)
	}

	return &ride, nil
}

// Route and tracking

func (r *rideRepository) UpdateRoute(rideID string, route []models.Location) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"route":      route,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update route")
		return fmt.Errorf("failed to update route: %w", err)
	}

	return nil
}

func (r *rideRepository) UpdateEstimatedArrival(rideID string, eta time.Time) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"estimated_arrival": eta,
			"updated_at":        time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update estimated arrival")
		return fmt.Errorf("failed to update estimated arrival: %w", err)
	}

	return nil
}

func (r *rideRepository) UpdateActualDistance(rideID string, distance float64) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"actual_distance": distance,
			"updated_at":      time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update actual distance")
		return fmt.Errorf("failed to update actual distance: %w", err)
	}

	return nil
}

// Requirements and preferences

func (r *rideRepository) UpdateRideRequirements(rideID string, requirements *models.RideRequirements) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"requirements": requirements,
			"updated_at":   time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update ride requirements")
		return fmt.Errorf("failed to update ride requirements: %w", err)
	}

	return nil
}

func (r *rideRepository) UpdateRidePreferences(rideID string, preferences *models.RidePreferences) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"preferences": preferences,
			"updated_at":  time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to update ride preferences")
		return fmt.Errorf("failed to update ride preferences: %w", err)
	}

	return nil
}

// Reports and issues

func (r *rideRepository) SaveRideReport(rideID string, report *RideReport) error {
	collection := r.db.Collection("ride_reports")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rideObjectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return err
	}

	report.ID = primitive.NewObjectID()
	report.RideID = rideObjectID
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()
	report.Status = "pending"

	_, err = collection.InsertOne(ctx, report)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID).Msg("Failed to save ride report")
		return fmt.Errorf("failed to save ride report: %w", err)
	}

	return nil
}

func (r *rideRepository) GetRideReports(rideID string) ([]*RideReport, error) {
	collection := r.db.Collection("ride_reports")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rideObjectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(ctx, bson.M{"ride_id": rideObjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to get ride reports: %w", err)
	}
	defer cursor.Close(ctx)

	var reports []*RideReport
	if err = cursor.All(ctx, &reports); err != nil {
		return nil, fmt.Errorf("failed to decode ride reports: %w", err)
	}

	return reports, nil
}

// Analytics and history

func (r *rideRepository) GetRideHistory(userID string, page, limit int, filters RideHistoryFilter) ([]*models.Ride, int64, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"$or": []bson.M{
			{"passenger_id": userObjectID},
			{"driver_id": userObjectID},
		},
	}

	// Apply filters
	if filters.Status != nil {
		filter["status"] = *filters.Status
	}
	if filters.ServiceType != nil {
		filter["service_type"] = *filters.ServiceType
	}
	if filters.RideType != nil {
		filter["type"] = *filters.RideType
	}
	if filters.FromDate != nil || filters.ToDate != nil {
		dateFilter := bson.M{}
		if filters.FromDate != nil {
			dateFilter["$gte"] = *filters.FromDate
		}
		if filters.ToDate != nil {
			dateFilter["$lte"] = *filters.ToDate
		}
		filter["created_at"] = dateFilter
	}
	if filters.MinFare != nil || filters.MaxFare != nil {
		fareFilter := bson.M{}
		if filters.MinFare != nil {
			fareFilter["$gte"] = *filters.MinFare
		}
		if filters.MaxFare != nil {
			fareFilter["$lte"] = *filters.MaxFare
		}
		filter["fare_details.final_fare"] = fareFilter
	}
	if filters.City != "" {
		filter["pickup_location.city"] = filters.City
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ride history: %w", err)
	}

	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ride history: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, 0, fmt.Errorf("failed to decode ride history: %w", err)
	}

	return rides, total, nil
}

func (r *rideRepository) GetFrequentRoutes(userID string, limit int) ([]*FrequentRoute, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"passenger_id": userObjectID,
			"status":       models.RideStatusCompleted,
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"pickup_address":  "$pickup_location.address",
				"dropoff_address": "$dropoff_location.address",
			},
			"count":            bson.M{"$sum": 1},
			"last_used":        bson.M{"$max": "$created_at"},
			"average_fare":     bson.M{"$avg": "$fare_details.final_fare"},
			"average_duration": bson.M{"$avg": "$actual_duration"},
			"pickup_location":  bson.M{"$first": "$pickup_location"},
			"dropoff_location": bson.M{"$first": "$dropoff_location"},
		}},
		{"$match": bson.M{"count": bson.M{"$gte": 2}}},
		{"$sort": bson.M{"count": -1, "last_used": -1}},
		{"$limit": limit},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get frequent routes: %w", err)
	}
	defer cursor.Close(ctx)

	var routes []*FrequentRoute
	for cursor.Next(ctx) {
		var result struct {
			ID struct {
				PickupAddress  string `bson:"pickup_address"`
				DropoffAddress string `bson:"dropoff_address"`
			} `bson:"_id"`
			Count           int                 `bson:"count"`
			LastUsed        time.Time           `bson:"last_used"`
			AverageFare     float64             `bson:"average_fare"`
			AverageDuration int                 `bson:"average_duration"`
			PickupLocation  models.RideLocation `bson:"pickup_location"`
			DropoffLocation models.RideLocation `bson:"dropoff_location"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		route := &FrequentRoute{
			PickupLocation:  result.PickupLocation,
			DropoffLocation: result.DropoffLocation,
			Count:           result.Count,
			LastUsed:        result.LastUsed,
			AverageFare:     result.AverageFare,
			AverageDuration: result.AverageDuration,
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func (r *rideRepository) GetRideStats(userID string, period string) (*RideStats, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Calculate start date based on period
	var startDate time.Time
	now := time.Now()
	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0) // Default to month
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"passenger_id": userObjectID,
			"created_at":   bson.M{"$gte": startDate},
		}},
		{"$group": bson.M{
			"_id":             nil,
			"total_rides":     bson.M{"$sum": 1},
			"completed_rides": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$status", models.RideStatusCompleted}}, 1, 0}}},
			"cancelled_rides": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$status", models.RideStatusCancelled}}, 1, 0}}},
			"total_spent":     bson.M{"$sum": "$fare_details.final_fare"},
			"total_distance":  bson.M{"$sum": "$actual_distance"},
			"total_duration":  bson.M{"$sum": "$actual_duration"},
			"average_rating":  bson.M{"$avg": "$passenger_rating.score"},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalRides     int     `bson:"total_rides"`
		CompletedRides int     `bson:"completed_rides"`
		CancelledRides int     `bson:"cancelled_rides"`
		TotalSpent     float64 `bson:"total_spent"`
		TotalDistance  float64 `bson:"total_distance"`
		TotalDuration  int     `bson:"total_duration"`
		AverageRating  float64 `bson:"average_rating"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode ride stats: %w", err)
		}
	}

	// Get top routes
	topRoutes, err := r.GetFrequentRoutes(userID, 5)
	if err != nil {
		topRoutes = []*FrequentRoute{}
	}

	stats := &RideStats{
		UserID:         userID,
		Period:         period,
		TotalRides:     result.TotalRides,
		CompletedRides: result.CompletedRides,
		CancelledRides: result.CancelledRides,
		TotalSpent:     result.TotalSpent,
		TotalDistance:  result.TotalDistance,
		TotalDuration:  result.TotalDuration,
		AverageRating:  result.AverageRating,
		TopRoutes:      topRoutes,
	}

	return stats, nil
}

// Bulk operations

func (r *rideRepository) GetRidesByIDs(rideIDs []string) ([]*models.Ride, error) {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var objectIDs []primitive.ObjectID
	for _, id := range rideIDs {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objectID)
	}

	cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
	if err != nil {
		return nil, fmt.Errorf("failed to get rides by IDs: %w", err)
	}
	defer cursor.Close(ctx)

	var rides []*models.Ride
	if err = cursor.All(ctx, &rides); err != nil {
		return nil, fmt.Errorf("failed to decode rides: %w", err)
	}

	return rides, nil
}

func (r *rideRepository) UpdateMultipleRideStatuses(rideIDs []string, status models.RideStatus) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var objectIDs []primitive.ObjectID
	for _, id := range rideIDs {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objectID)
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": objectIDs}}, update)
	if err != nil {
		r.logger.Error().Err(err).Strs("ride_ids", rideIDs).Str("status", string(status)).Msg("Failed to update multiple ride statuses")
		return fmt.Errorf("failed to update multiple ride statuses: %w", err)
	}

	return nil
}

func (r *rideRepository) DeleteUserRides(userID string) error {
	collection := r.db.Collection("rides")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"$or": []bson.M{
			{"passenger_id": userObjectID},
			{"driver_id": userObjectID},
		},
	}

	_, err = collection.DeleteMany(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user rides")
		return fmt.Errorf("failed to delete user rides: %w", err)
	}

	return nil
}
