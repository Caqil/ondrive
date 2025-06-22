package repositories

import (
	"context"
	"fmt"
	"time"

	"ondrive/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DriverRepository interface {
	// Basic CRUD operations
	Create(driver *models.DriverInfo) (*models.DriverInfo, error)
	GetByID(id string) (*models.DriverInfo, error)
	GetByUserID(userID string) (*models.DriverInfo, error)
	Update(driver *models.DriverInfo) (*models.DriverInfo, error)
	Delete(id string) error
	List(filter DriverFilter) ([]*models.DriverInfo, int64, error)

	// Status & Availability
	UpdateOnlineStatus(userID string, isOnline bool) error
	GetOnlineDriversInArea(lat, lng, radiusKm float64, serviceType string) ([]*models.DriverInfo, error)
	UpdateAvailability(userID string, isAvailable bool, maxDistance int, serviceTypes []string) error
	GetAvailableDrivers(serviceType string, lat, lng float64, radiusKm float64, limit int) ([]*models.DriverInfo, error)

	// Location Management
	UpdateLocation(userID string, location *models.Location) error
	GetDriverLocation(userID string) (*models.Location, error)
	GetNearbyDrivers(lat, lng, radiusKm float64, limit int) ([]*models.DriverInfo, error)

	// Vehicle Management
	AddVehicle(userID string, vehicle *models.VehicleDetails) error
	UpdateVehicle(userID string, vehicle *models.VehicleDetails) error
	GetDriverVehicle(userID string) (*models.VehicleDetails, error)
	AddVehiclePhoto(userID string, photo *models.VehiclePhoto) error
	DeleteVehiclePhoto(userID, photoID string) error

	// Documents Management
	AddDocument(userID string, document *models.VerificationDoc) error
	GetDocuments(userID string) ([]models.VerificationDoc, error)
	UpdateDocumentStatus(userID, documentID, status string) error
	DeleteDocument(userID, documentID string) error

	// Earnings & Finance
	GetEarnings(userID string, page, limit int) ([]*models.Earning, int64, error)
	GetEarningsByPeriod(userID, period string, startDate, endDate time.Time) ([]*models.Earning, error)
	AddEarning(userID string, earning *models.Earning) error
	GetPayouts(userID string, page, limit int) ([]*models.Payout, int64, error)
	AddPayout(userID string, payout *models.Payout) error

	// Statistics
	GetDriverStats(userID string) (*models.DriverStats, error)
	UpdateDriverStats(userID string, stats *models.DriverStats) error
	GetPerformanceMetrics(userID, period string) (*models.PerformanceMetrics, error)
	GetRatingSummary(userID string) (*models.RatingSummary, error)

	// Working Hours & Schedule
	GetWorkingHours(userID string) (*models.WorkingHours, error)
	UpdateWorkingHours(userID string, workingHours *models.WorkingHours) error
	GetDriverSchedule(userID, date string) (*models.DriverSchedule, error)
	AddBreak(userID string, breakInfo *models.BreakInfo) error
	EndBreak(userID string) error

	// Preferences
	GetDriverPreferences(userID string) (*models.DriverPreferences, error)
	UpdateDriverPreferences(userID string, preferences *models.DriverPreferences) error
	UpdateServiceAreas(userID string, serviceAreas []models.ServiceArea) error
	UpdateServiceTypes(userID string, serviceTypes []string) error

	// Verification
	GetVerificationStatus(userID string) (*models.VerificationStatus, error)
	UpdateVerificationStatus(userID string, status *models.VerificationStatus) error

	// Support
	CreateSupportTicket(userID string, ticket *models.SupportTicket) (*models.SupportTicket, error)
	GetSupportTickets(userID string, page, limit int) ([]*models.SupportTicket, int64, error)
	UpdateSupportTicket(ticketID, message string) (*models.SupportTicket, error)

	// Bulk Operations
	BulkUpdateDriverStats(updates []models.DriverStatsBulkUpdate) error
	GetDriversByServiceType(serviceType string, limit int) ([]*models.DriverInfo, error)

	// Advanced Queries
	SearchDrivers(query string, filter DriverFilter) ([]*models.DriverInfo, int64, error)
	GetTopRatedDrivers(limit int) ([]*models.DriverInfo, error)
	GetActiveDriversCount() (int64, error)
}

type DriverFilter struct {
	Status       string
	ServiceType  string
	IsOnline     *bool
	IsAvailable  *bool
	IsVerified   *bool
	LocationLat  float64
	LocationLng  float64
	RadiusKm     float64
	MinRating    float64
	JoinedAfter  *time.Time
	JoinedBefore *time.Time
	Page         int
	Limit        int
}

type driverRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewDriverRepository(db *mongo.Database) DriverRepository {
	return &driverRepository{
		collection: db.Collection("drivers"),
		db:         db,
	}
}

// Basic CRUD operations
func (r *driverRepository) Create(driverInfo *models.DriverInfo) (*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	driver := &models.Driver{
		ID:         primitive.NewObjectID(),
		UserID:     *driverInfo.CurrentRideID,
		DriverInfo: driverInfo,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	driverInfo.JoinedAt = time.Now()
	driverInfo.LastActive = time.Now()

	_, err := r.collection.InsertOne(ctx, driver)
	if err != nil {
		return nil, err
	}

	return driverInfo, nil
}

func (r *driverRepository) GetByID(id string) (*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var driver models.DriverInfo
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&driver)

	if err != nil {
		return nil, err
	}

	return &driver, nil
}

func (r *driverRepository) GetByUserID(userID string) (*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var driver models.DriverInfo
	err = r.collection.FindOne(ctx, bson.M{
		"user_id":    userObjectID,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&driver)

	if err != nil {
		return nil, err
	}

	return &driver, nil
}

func (r *driverRepository) Update(driver *models.DriverInfo) (*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	driver.LastActive = time.Now()

	filter := bson.M{"_id": driver.CurrentRideID}
	update := bson.M{"$set": driver}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return driver, nil
}

func (r *driverRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"deleted_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *driverRepository) List(filter DriverFilter) ([]*models.DriverInfo, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build MongoDB filter
	mongoFilter := bson.M{"is_deleted": bson.M{"$ne": true}}

	if filter.Status != "" {
		mongoFilter["status"] = filter.Status
	}
	if filter.ServiceType != "" {
		mongoFilter["service_types"] = bson.M{"$in": []string{filter.ServiceType}}
	}
	if filter.IsOnline != nil {
		mongoFilter["is_online"] = *filter.IsOnline
	}
	if filter.IsAvailable != nil {
		mongoFilter["is_available"] = *filter.IsAvailable
	}
	if filter.IsVerified != nil {
		mongoFilter["is_verified"] = *filter.IsVerified
	}
	if filter.MinRating > 0 {
		mongoFilter["rating"] = bson.M{"$gte": filter.MinRating}
	}
	if filter.JoinedAfter != nil {
		mongoFilter["joined_at"] = bson.M{"$gte": *filter.JoinedAfter}
	}
	if filter.JoinedBefore != nil {
		if existing, ok := mongoFilter["joined_at"]; ok {
			mongoFilter["joined_at"] = bson.M{"$gte": existing.(bson.M)["$gte"], "$lte": *filter.JoinedBefore}
		} else {
			mongoFilter["joined_at"] = bson.M{"$lte": *filter.JoinedBefore}
		}
	}

	// Location-based filtering
	if filter.LocationLat != 0 && filter.LocationLng != 0 && filter.RadiusKm > 0 {
		mongoFilter["current_location"] = bson.M{
			"$geoWithin": bson.M{
				"$centerSphere": []interface{}{
					[]float64{filter.LocationLng, filter.LocationLat},
					filter.RadiusKm / 6378.1, // Earth radius in km
				},
			},
		}
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	skip := (filter.Page - 1) * filter.Limit
	opts := options.Find().
		SetSort(bson.D{{Key: "last_active", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(filter.Limit))

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var drivers []*models.DriverInfo
	if err = cursor.All(ctx, &drivers); err != nil {
		return nil, 0, err
	}

	return drivers, total, nil
}

// Status & Availability

func (r *driverRepository) UpdateOnlineStatus(userID string, isOnline bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_online":   isOnline,
			"last_active": time.Now(),
		},
	}

	if isOnline {
		update["$set"].(bson.M)["online_since"] = time.Now()
		update["$unset"] = bson.M{"offline_since": ""}
	} else {
		update["$set"].(bson.M)["offline_since"] = time.Now()
		update["$unset"] = bson.M{"online_since": ""}
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) GetOnlineDriversInArea(lat, lng, radiusKm float64, serviceType string) ([]*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_online":    true,
		"is_available": true,
		"is_verified":  true,
		"is_deleted":   bson.M{"$ne": true},
		"current_location": bson.M{
			"$geoWithin": bson.M{
				"$centerSphere": []interface{}{
					[]float64{lng, lat},
					radiusKm / 6378.1,
				},
			},
		},
	}

	if serviceType != "" {
		filter["service_types"] = bson.M{"$in": []string{serviceType}}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "rating", Value: -1}, {Key: "last_active", Value: -1}}).
		SetLimit(50)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []*models.DriverInfo
	if err = cursor.All(ctx, &drivers); err != nil {
		return nil, err
	}

	return drivers, nil
}

func (r *driverRepository) UpdateAvailability(userID string, isAvailable bool, maxDistance int, serviceTypes []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_available": isAvailable,
			"last_active":  time.Now(),
		},
	}

	if maxDistance > 0 {
		update["$set"].(bson.M)["preferences.max_distance"] = maxDistance
	}

	if len(serviceTypes) > 0 {
		update["$set"].(bson.M)["service_types"] = serviceTypes
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) GetAvailableDrivers(serviceType string, lat, lng float64, radiusKm float64, limit int) ([]*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_online":    true,
		"is_available": true,
		"is_verified":  true,
		"is_deleted":   bson.M{"$ne": true},
	}

	if serviceType != "" {
		filter["service_types"] = bson.M{"$in": []string{serviceType}}
	}

	if lat != 0 && lng != 0 && radiusKm > 0 {
		filter["current_location"] = bson.M{
			"$geoWithin": bson.M{
				"$centerSphere": []interface{}{
					[]float64{lng, lat},
					radiusKm / 6378.1,
				},
			},
		}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "rating", Value: -1}, {Key: "last_active", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []*models.DriverInfo
	if err = cursor.All(ctx, &drivers); err != nil {
		return nil, err
	}

	return drivers, nil
}

// Location Management

// UpdateLocation updates driver's location in the database
func (r *driverRepository) UpdateLocation(userID string, location *models.Location) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Ensure coordinates array has at least 2 elements (longitude, latitude)
	var coordinates []float64
	if len(location.Coordinates) >= 2 {
		coordinates = []float64{location.Coordinates[0], location.Coordinates[1]}
	} else {
		// If coordinates are not properly set, return an error
		return fmt.Errorf("invalid coordinates: must contain at least [longitude, latitude]")
	}

	update := bson.M{
		"$set": bson.M{
			"current_location": bson.M{
				"type":        "Point",
				"coordinates": coordinates,
			},
			"location":    location,
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) GetDriverLocation(userID string) (*models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Location models.Location `bson:"location"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"user_id":    userObjectID,
		"is_deleted": bson.M{"$ne": true},
	}, options.FindOne().SetProjection(bson.M{"location": 1})).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result.Location, nil
}

func (r *driverRepository) GetNearbyDrivers(lat, lng, radiusKm float64, limit int) ([]*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_online":  true,
		"is_deleted": bson.M{"$ne": true},
		"current_location": bson.M{
			"$geoWithin": bson.M{
				"$centerSphere": []interface{}{
					[]float64{lng, lat},
					radiusKm / 6378.1,
				},
			},
		},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "last_active", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []*models.DriverInfo
	if err = cursor.All(ctx, &drivers); err != nil {
		return nil, err
	}

	return drivers, nil
}

// Vehicle Management

func (r *driverRepository) AddVehicle(userID string, vehicle *models.VehicleDetails) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	vehicle.CreatedAt = time.Now()
	vehicle.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"vehicle":     vehicle,
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) UpdateVehicle(userID string, vehicle *models.VehicleDetails) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	vehicle.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"vehicle":     vehicle,
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) GetDriverVehicle(userID string) (*models.VehicleDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Vehicle models.VehicleDetails `bson:"vehicle"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"user_id":    userObjectID,
		"is_deleted": bson.M{"$ne": true},
	}, options.FindOne().SetProjection(bson.M{"vehicle": 1})).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result.Vehicle, nil
}

func (r *driverRepository) AddVehiclePhoto(userID string, photo *models.VehiclePhoto) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	photo.ID = primitive.NewObjectID()
	photo.UploadedAt = time.Now()

	update := bson.M{
		"$push": bson.M{
			"vehicle.photos": photo,
		},
		"$set": bson.M{
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) DeleteVehiclePhoto(userID, photoID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	photoObjectID, err := primitive.ObjectIDFromHex(photoID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$pull": bson.M{
			"vehicle.photos": bson.M{"_id": photoObjectID},
		},
		"$set": bson.M{
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

// Documents Management

func (r *driverRepository) AddDocument(userID string, document *models.VerificationDoc) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	document.ID = primitive.NewObjectID()
	document.UploadedAt = time.Now()

	update := bson.M{
		"$push": bson.M{
			"documents": document,
		},
		"$set": bson.M{
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) GetDocuments(userID string) ([]models.VerificationDoc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Documents []models.VerificationDoc `bson:"documents"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"user_id":    userObjectID,
		"is_deleted": bson.M{"$ne": true},
	}, options.FindOne().SetProjection(bson.M{"documents": 1})).Decode(&result)

	if err != nil {
		return nil, err
	}

	return result.Documents, nil
}

func (r *driverRepository) UpdateDocumentStatus(userID, documentID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	docObjectID, err := primitive.ObjectIDFromHex(documentID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"documents.$.status":      status,
			"documents.$.verified_at": time.Now(),
			"last_active":             time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{
		"user_id":      userObjectID,
		"documents.id": docObjectID,
	}, update)
	return err
}

func (r *driverRepository) DeleteDocument(userID, documentID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	docObjectID, err := primitive.ObjectIDFromHex(documentID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$pull": bson.M{
			"documents": bson.M{"id": docObjectID},
		},
		"$set": bson.M{
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

// Earnings & Finance (Simplified implementations - would need proper earnings collections)

func (r *driverRepository) GetEarnings(userID string, page, limit int) ([]*models.Earning, int64, error) {
	// This would typically query a separate earnings collection
	// For now, returning empty results
	return []*models.Earning{}, 0, nil
}

func (r *driverRepository) GetEarningsByPeriod(userID, period string, startDate, endDate time.Time) ([]*models.Earning, error) {
	// This would typically query a separate earnings collection
	return []*models.Earning{}, nil
}

func (r *driverRepository) AddEarning(userID string, earning *models.Earning) error {
	// This would typically insert into a separate earnings collection
	return nil
}

func (r *driverRepository) GetPayouts(userID string, page, limit int) ([]*models.Payout, int64, error) {
	// This would typically query a separate payouts collection
	return []*models.Payout{}, 0, nil
}

func (r *driverRepository) AddPayout(userID string, payout *models.Payout) error {
	// This would typically insert into a separate payouts collection
	return nil
}

// Statistics (Simplified implementations)

func (r *driverRepository) GetDriverStats(userID string) (*models.DriverStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Stats models.DriverStats `bson:"stats"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"user_id":    userObjectID,
		"is_deleted": bson.M{"$ne": true},
	}, options.FindOne().SetProjection(bson.M{"stats": 1})).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result.Stats, nil
}

func (r *driverRepository) UpdateDriverStats(userID string, stats *models.DriverStats) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"stats":       stats,
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) GetPerformanceMetrics(userID, period string) (*models.PerformanceMetrics, error) {
	// This would calculate metrics based on rides, earnings, etc.
	return &models.PerformanceMetrics{}, nil
}

func (r *driverRepository) GetRatingSummary(userID string) (*models.RatingSummary, error) {
	// This would aggregate rating data from rides
	return &models.RatingSummary{}, nil
}

// Working Hours & Schedule

func (r *driverRepository) GetWorkingHours(userID string) (*models.WorkingHours, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var result struct {
		WorkingHours models.WorkingHours `bson:"working_hours"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"user_id":    userObjectID,
		"is_deleted": bson.M{"$ne": true},
	}, options.FindOne().SetProjection(bson.M{"working_hours": 1})).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result.WorkingHours, nil
}

func (r *driverRepository) UpdateWorkingHours(userID string, workingHours *models.WorkingHours) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"working_hours": workingHours,
			"last_active":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) GetDriverSchedule(userID, date string) (*models.DriverSchedule, error) {
	// This would typically query a separate schedules collection
	return &models.DriverSchedule{}, nil
}

func (r *driverRepository) AddBreak(userID string, breakInfo *models.BreakInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"current_break": breakInfo,
			"status":        "on_break",
			"is_available":  false,
			"last_active":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) EndBreak(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$unset": bson.M{
			"current_break": "",
		},
		"$set": bson.M{
			"status":      "available",
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

// Preferences

func (r *driverRepository) GetDriverPreferences(userID string) (*models.DriverPreferences, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Preferences models.DriverPreferences `bson:"preferences"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"user_id":    userObjectID,
		"is_deleted": bson.M{"$ne": true},
	}, options.FindOne().SetProjection(bson.M{"preferences": 1})).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result.Preferences, nil
}

func (r *driverRepository) UpdateDriverPreferences(userID string, preferences *models.DriverPreferences) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"preferences": preferences,
			"last_active": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) UpdateServiceAreas(userID string, serviceAreas []models.ServiceArea) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"service_areas": serviceAreas,
			"last_active":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

func (r *driverRepository) UpdateServiceTypes(userID string, serviceTypes []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"service_types": serviceTypes,
			"last_active":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

// Verification

func (r *driverRepository) GetVerificationStatus(userID string) (*models.VerificationStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var result struct {
		Verification models.VerificationStatus `bson:"verification"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"user_id":    userObjectID,
		"is_deleted": bson.M{"$ne": true},
	}, options.FindOne().SetProjection(bson.M{"verification": 1})).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result.Verification, nil
}

func (r *driverRepository) UpdateVerificationStatus(userID string, status *models.VerificationStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"verification": status,
			"is_verified":  status.IsFullyVerified,
			"last_active":  time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userObjectID}, update)
	return err
}

// Support

func (r *driverRepository) CreateSupportTicket(userID string, ticket *models.SupportTicket) (*models.SupportTicket, error) {
	// This would typically use a separate support tickets collection
	ticket.ID = primitive.NewObjectID()
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()
	return ticket, nil
}

func (r *driverRepository) GetSupportTickets(userID string, page, limit int) ([]*models.SupportTicket, int64, error) {
	// This would typically query a separate support tickets collection
	return []*models.SupportTicket{}, 0, nil
}

func (r *driverRepository) UpdateSupportTicket(ticketID, message string) (*models.SupportTicket, error) {
	// This would typically update in a separate support tickets collection
	return &models.SupportTicket{}, nil
}

// Bulk Operations

func (r *driverRepository) BulkUpdateDriverStats(updates []models.DriverStatsBulkUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var operations []mongo.WriteModel
	for _, update := range updates {
		userObjectID, err := primitive.ObjectIDFromHex(update.UserID)
		if err != nil {
			continue
		}

		updateModel := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"user_id": userObjectID}).
			SetUpdate(bson.M{
				"$inc": update.StatsUpdate,
				"$set": bson.M{"last_active": time.Now()},
			})

		operations = append(operations, updateModel)
	}

	if len(operations) == 0 {
		return nil
	}

	_, err := r.collection.BulkWrite(ctx, operations)
	return err
}

func (r *driverRepository) GetDriversByServiceType(serviceType string, limit int) ([]*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"service_types": bson.M{"$in": []string{serviceType}},
		"is_verified":   true,
		"is_deleted":    bson.M{"$ne": true},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "rating", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []*models.DriverInfo
	if err = cursor.All(ctx, &drivers); err != nil {
		return nil, err
	}

	return drivers, nil
}

// Advanced Queries

func (r *driverRepository) SearchDrivers(query string, filter DriverFilter) ([]*models.DriverInfo, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build text search filter
	mongoFilter := bson.M{
		"$text":      bson.M{"$search": query},
		"is_deleted": bson.M{"$ne": true},
	}

	// Apply additional filters (reuse logic from List method)
	if filter.Status != "" {
		mongoFilter["status"] = filter.Status
	}
	// ... other filters

	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, err
	}

	skip := (filter.Page - 1) * filter.Limit
	opts := options.Find().
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}, {Key: "rating", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(filter.Limit))

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var drivers []*models.DriverInfo
	if err = cursor.All(ctx, &drivers); err != nil {
		return nil, 0, err
	}

	return drivers, total, nil
}

func (r *driverRepository) GetTopRatedDrivers(limit int) ([]*models.DriverInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_verified": true,
		"is_deleted":  bson.M{"$ne": true},
		"rating":      bson.M{"$gte": 4.0},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "rating", Value: -1}, {Key: "total_rides", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []*models.DriverInfo
	if err = cursor.All(ctx, &drivers); err != nil {
		return nil, err
	}

	return drivers, nil
}

func (r *driverRepository) GetActiveDriversCount() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_online":  true,
		"is_deleted": bson.M{"$ne": true},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	return count, err
}
