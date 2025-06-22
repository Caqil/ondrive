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

// CourierRepository interface for courier data operations
type CourierRepository interface {
	Create(request *models.CourierRequest) (*models.CourierRequest, error)
	GetByID(id primitive.ObjectID) (*models.CourierRequest, error)
	GetByTrackingCode(trackingCode string) (*models.CourierRequest, error)
	Update(id primitive.ObjectID, request *models.CourierRequest) (*models.CourierRequest, error)
	Delete(id primitive.ObjectID) error
	GetByFilter(filter map[string]interface{}, page, limit int) ([]*models.CourierRequest, int64, error)
	GetBySenderID(senderID primitive.ObjectID, page, limit int) ([]*models.CourierRequest, int64, error)
	GetByCourierID(courierID primitive.ObjectID, page, limit int) ([]*models.CourierRequest, int64, error)
	GetByRecipient(phone, email string, page, limit int) ([]*models.CourierRequest, int64, error)
	GetActiveRequests(courierID primitive.ObjectID) ([]*models.CourierRequest, error)
	GetPendingRequests(location models.Location, radius float64) ([]*models.CourierRequest, error)
	UpdateStatus(id primitive.ObjectID, status models.CourierStatus) error
	AddTrackingEvent(id primitive.ObjectID, event models.TrackingEvent) error
	UpdatePackageDetails(id primitive.ObjectID, packageDetails models.CourierPackage) error
	UpdateRecipient(id primitive.ObjectID, recipient models.CourierRecipient) error
	AddPackagePhotos(id primitive.ObjectID, photos []string) error
	UpdateDeliveryProof(id primitive.ObjectID, proof *models.DeliveryProof) error
	AddInsurance(id primitive.ObjectID, insurance *models.CourierInsurance) error
	GetCourierStats(courierID primitive.ObjectID) (*CourierStats, error)
	GetSenderStats(senderID primitive.ObjectID) (*SenderStats, error)
	MarkAsCompleted(id primitive.ObjectID, completedAt time.Time) error
	Cancel(id primitive.ObjectID, reason string, cancelledBy primitive.ObjectID) error
	AssignCourier(id, courierID primitive.ObjectID) error
	UnassignCourier(id primitive.ObjectID) error
}

// SavedAddressRepository interface for saved addresses
type SavedAddressRepository interface {
	Create(address *models.SavedAddress) (*models.SavedAddress, error)
	GetByUserID(userID primitive.ObjectID) ([]*models.SavedAddress, error)
	GetByID(id primitive.ObjectID) (*models.SavedAddress, error)
	Update(id primitive.ObjectID, address *models.SavedAddress) (*models.SavedAddress, error)
	Delete(id primitive.ObjectID) error
}

// CourierIssueRepository interface for courier issues and claims
type CourierIssueRepository interface {
	CreateIssue(issue *models.CourierIssue) (*models.CourierIssue, error)
	GetIssuesByRequestID(requestID primitive.ObjectID) ([]*models.CourierIssue, error)
	CreateClaim(claim *models.CourierClaim) (*models.CourierClaim, error)
	GetClaimsByRequestID(requestID primitive.ObjectID) ([]*models.CourierClaim, error)
	UpdateIssueStatus(issueID primitive.ObjectID, status string) error
	UpdateClaimStatus(claimID primitive.ObjectID, status string) error
}

// Stats types
type CourierStats struct {
	TotalDeliveries     int64   `json:"total_deliveries"`
	CompletedDeliveries int64   `json:"completed_deliveries"`
	CancelledDeliveries int64   `json:"cancelled_deliveries"`
	SuccessRate         float64 `json:"success_rate"`
	AverageRating       float64 `json:"average_rating"`
	TotalEarnings       float64 `json:"total_earnings"`
	ThisMonthEarnings   float64 `json:"this_month_earnings"`
}

type SenderStats struct {
	TotalRequests     int64   `json:"total_requests"`
	CompletedRequests int64   `json:"completed_requests"`
	CancelledRequests int64   `json:"cancelled_requests"`
	TotalSpent        float64 `json:"total_spent"`
	ThisMonthSpent    float64 `json:"this_month_spent"`
}

// Implementation
type courierRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

type savedAddressRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

type courierIssueRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

// Constructors
func NewCourierRepository(db *mongo.Database, logger utils.Logger) CourierRepository {
	return &courierRepository{
		db:     db,
		logger: logger,
	}
}

func NewSavedAddressRepository(db *mongo.Database, logger utils.Logger) SavedAddressRepository {
	return &savedAddressRepository{
		db:     db,
		logger: logger,
	}
}

func NewCourierIssueRepository(db *mongo.Database, logger utils.Logger) CourierIssueRepository {
	return &courierIssueRepository{
		db:     db,
		logger: logger,
	}
}

// CourierRepository Implementation

func (r *courierRepository) Create(request *models.CourierRequest) (*models.CourierRequest, error) {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, request)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", request.ID.Hex()).Msg("Failed to create courier request")
		return nil, fmt.Errorf("failed to create courier request: %w", err)
	}

	r.logger.Info().Str("request_id", request.ID.Hex()).Msg("Courier request created successfully")
	return request, nil
}

func (r *courierRepository) GetByID(id primitive.ObjectID) (*models.CourierRequest, error) {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var request models.CourierRequest
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("courier request not found")
		}
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to get courier request")
		return nil, fmt.Errorf("failed to get courier request: %w", err)
	}

	return &request, nil
}

func (r *courierRepository) GetByTrackingCode(trackingCode string) (*models.CourierRequest, error) {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var request models.CourierRequest
	err := collection.FindOne(ctx, bson.M{"tracking_code": trackingCode}).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("courier request not found")
		}
		r.logger.Error().Err(err).Str("tracking_code", trackingCode).Msg("Failed to get courier request by tracking code")
		return nil, fmt.Errorf("failed to get courier request: %w", err)
	}

	return &request, nil
}

func (r *courierRepository) Update(id primitive.ObjectID, request *models.CourierRequest) (*models.CourierRequest, error) {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(ctx, bson.M{"_id": id}, request)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to update courier request")
		return nil, fmt.Errorf("failed to update courier request: %w", err)
	}

	return request, nil
}

func (r *courierRepository) Delete(id primitive.ObjectID) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to delete courier request")
		return fmt.Errorf("failed to delete courier request: %w", err)
	}

	return nil
}

func (r *courierRepository) GetByFilter(filter map[string]interface{}, page, limit int) ([]*models.CourierRequest, int64, error) {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count courier requests")
		return nil, 0, fmt.Errorf("failed to count courier requests: %w", err)
	}

	// Build options
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created_at": -1})

	// Pagination
	if page > 0 && limit > 0 {
		skip := (page - 1) * limit
		findOptions.SetSkip(int64(skip))
		findOptions.SetLimit(int64(limit))
	}

	// Find requests
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to find courier requests")
		return nil, 0, fmt.Errorf("failed to find courier requests: %w", err)
	}
	defer cursor.Close(ctx)

	var requests []*models.CourierRequest
	for cursor.Next(ctx) {
		var request models.CourierRequest
		if err := cursor.Decode(&request); err != nil {
			r.logger.Error().Err(err).Msg("Failed to decode courier request")
			continue
		}
		requests = append(requests, &request)
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error().Err(err).Msg("Cursor error while listing courier requests")
		return nil, 0, fmt.Errorf("cursor error: %w", err)
	}

	return requests, total, nil
}

func (r *courierRepository) GetBySenderID(senderID primitive.ObjectID, page, limit int) ([]*models.CourierRequest, int64, error) {
	filter := bson.M{"sender_id": senderID}
	return r.GetByFilter(filter, page, limit)
}

func (r *courierRepository) GetByCourierID(courierID primitive.ObjectID, page, limit int) ([]*models.CourierRequest, int64, error) {
	filter := bson.M{"courier_id": courierID}
	return r.GetByFilter(filter, page, limit)
}

func (r *courierRepository) GetByRecipient(phone, email string, page, limit int) ([]*models.CourierRequest, int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"recipient.phone": phone},
			{"recipient.email": email},
		},
	}
	return r.GetByFilter(filter, page, limit)
}

func (r *courierRepository) GetActiveRequests(courierID primitive.ObjectID) ([]*models.CourierRequest, error) {
	filter := bson.M{
		"courier_id": courierID,
		"status": bson.M{
			"$in": []models.CourierStatus{
				models.CourierStatusAccepted,
				models.CourierStatusPickedUp,
				models.CourierStatusInTransit,
			},
		},
	}

	requests, _, err := r.GetByFilter(filter, 0, 0)
	return requests, err
}

func (r *courierRepository) GetPendingRequests(location models.Location, radius float64) ([]*models.CourierRequest, error) {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build geospatial query for nearby requests
	filter := bson.M{
		"status": models.CourierStatusPending,
		"pickup_location.coordinates": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": location.Coordinates,
				},
				"$maxDistance": radius * 1000, // Convert km to meters
			},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to find pending courier requests")
		return nil, fmt.Errorf("failed to find pending requests: %w", err)
	}
	defer cursor.Close(ctx)

	var requests []*models.CourierRequest
	for cursor.Next(ctx) {
		var request models.CourierRequest
		if err := cursor.Decode(&request); err != nil {
			r.logger.Error().Err(err).Msg("Failed to decode courier request")
			continue
		}
		requests = append(requests, &request)
	}

	return requests, nil
}

func (r *courierRepository) UpdateStatus(id primitive.ObjectID, status models.CourierStatus) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to update courier request status")
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (r *courierRepository) AddTrackingEvent(id primitive.ObjectID, event models.TrackingEvent) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$push": bson.M{
			"tracking_history": event,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to add tracking event")
		return fmt.Errorf("failed to add tracking event: %w", err)
	}

	return nil
}

func (r *courierRepository) UpdatePackageDetails(id primitive.ObjectID, packageDetails models.CourierPackage) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"package":    packageDetails,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to update package details")
		return fmt.Errorf("failed to update package details: %w", err)
	}

	return nil
}

func (r *courierRepository) UpdateRecipient(id primitive.ObjectID, recipient models.CourierRecipient) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"recipient":  recipient,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to update recipient")
		return fmt.Errorf("failed to update recipient: %w", err)
	}

	return nil
}

func (r *courierRepository) AddPackagePhotos(id primitive.ObjectID, photos []string) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$push": bson.M{
			"package.photos": bson.M{"$each": photos},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to add package photos")
		return fmt.Errorf("failed to add package photos: %w", err)
	}

	return nil
}

func (r *courierRepository) UpdateDeliveryProof(id primitive.ObjectID, proof *models.DeliveryProof) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"delivery_proof": proof,
			"updated_at":     time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to update delivery proof")
		return fmt.Errorf("failed to update delivery proof: %w", err)
	}

	return nil
}

func (r *courierRepository) AddInsurance(id primitive.ObjectID, insurance *models.CourierInsurance) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"insurance":  insurance,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to add insurance")
		return fmt.Errorf("failed to add insurance: %w", err)
	}

	return nil
}

func (r *courierRepository) GetCourierStats(courierID primitive.ObjectID) (*CourierStats, error) {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation pipeline to calculate stats
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"courier_id": courierID,
			},
		},
		{
			"$group": bson.M{
				"_id":              nil,
				"total_deliveries": bson.M{"$sum": 1},
				"completed_deliveries": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", models.CourierStatusDelivered}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"cancelled_deliveries": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", models.CourierStatusCancelled}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"total_earnings": bson.M{"$sum": "$fare.courier_earnings"},
				"average_rating": bson.M{"$avg": "$courier_rating.rating"},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error().Err(err).Str("courier_id", courierID.Hex()).Msg("Failed to get courier stats")
		return nil, fmt.Errorf("failed to get courier stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalDeliveries     int64   `bson:"total_deliveries"`
		CompletedDeliveries int64   `bson:"completed_deliveries"`
		CancelledDeliveries int64   `bson:"cancelled_deliveries"`
		TotalEarnings       float64 `bson:"total_earnings"`
		AverageRating       float64 `bson:"average_rating"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode stats: %w", err)
		}
	}

	stats := &CourierStats{
		TotalDeliveries:     result.TotalDeliveries,
		CompletedDeliveries: result.CompletedDeliveries,
		CancelledDeliveries: result.CancelledDeliveries,
		TotalEarnings:       result.TotalEarnings,
		AverageRating:       result.AverageRating,
	}

	if result.TotalDeliveries > 0 {
		stats.SuccessRate = float64(result.CompletedDeliveries) / float64(result.TotalDeliveries) * 100
	}

	// Get this month's earnings
	thisMonth := time.Now().AddDate(0, 0, -30)
	monthFilter := bson.M{
		"courier_id": courierID,
		"status":     models.CourierStatusDelivered,
		"delivered_at": bson.M{
			"$gte": thisMonth,
		},
	}

	monthPipeline := []bson.M{
		{"$match": monthFilter},
		{"$group": bson.M{
			"_id":              nil,
			"monthly_earnings": bson.M{"$sum": "$fare.courier_earnings"},
		}},
	}

	monthCursor, err := collection.Aggregate(ctx, monthPipeline)
	if err == nil {
		defer monthCursor.Close(ctx)
		var monthResult struct {
			MonthlyEarnings float64 `bson:"monthly_earnings"`
		}
		if monthCursor.Next(ctx) {
			monthCursor.Decode(&monthResult)
			stats.ThisMonthEarnings = monthResult.MonthlyEarnings
		}
	}

	return stats, nil
}

func (r *courierRepository) GetSenderStats(senderID primitive.ObjectID) (*SenderStats, error) {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation pipeline to calculate sender stats
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"sender_id": senderID,
			},
		},
		{
			"$group": bson.M{
				"_id":            nil,
				"total_requests": bson.M{"$sum": 1},
				"completed_requests": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", models.CourierStatusDelivered}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"cancelled_requests": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", models.CourierStatusCancelled}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"total_spent": bson.M{"$sum": "$fare.total_amount"},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error().Err(err).Str("sender_id", senderID.Hex()).Msg("Failed to get sender stats")
		return nil, fmt.Errorf("failed to get sender stats: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalRequests     int64   `bson:"total_requests"`
		CompletedRequests int64   `bson:"completed_requests"`
		CancelledRequests int64   `bson:"cancelled_requests"`
		TotalSpent        float64 `bson:"total_spent"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode sender stats: %w", err)
		}
	}

	stats := &SenderStats{
		TotalRequests:     result.TotalRequests,
		CompletedRequests: result.CompletedRequests,
		CancelledRequests: result.CancelledRequests,
		TotalSpent:        result.TotalSpent,
	}

	// Get this month's spending
	thisMonth := time.Now().AddDate(0, 0, -30)
	monthFilter := bson.M{
		"sender_id": senderID,
		"status":    models.CourierStatusDelivered,
		"delivered_at": bson.M{
			"$gte": thisMonth,
		},
	}

	monthPipeline := []bson.M{
		{"$match": monthFilter},
		{"$group": bson.M{
			"_id":           nil,
			"monthly_spent": bson.M{"$sum": "$fare.total_amount"},
		}},
	}

	monthCursor, err := collection.Aggregate(ctx, monthPipeline)
	if err == nil {
		defer monthCursor.Close(ctx)
		var monthResult struct {
			MonthlySpent float64 `bson:"monthly_spent"`
		}
		if monthCursor.Next(ctx) {
			monthCursor.Decode(&monthResult)
			stats.ThisMonthSpent = monthResult.MonthlySpent
		}
	}

	return stats, nil
}

func (r *courierRepository) MarkAsCompleted(id primitive.ObjectID, completedAt time.Time) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":       models.CourierStatusDelivered,
			"delivered_at": completedAt,
			"updated_at":   time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to mark courier request as completed")
		return fmt.Errorf("failed to mark as completed: %w", err)
	}

	return nil
}

func (r *courierRepository) Cancel(id primitive.ObjectID, reason string, cancelledBy primitive.ObjectID) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":              models.CourierStatusCancelled,
			"cancellation_reason": reason,
			"cancelled_by":        cancelledBy,
			"cancelled_at":        time.Now(),
			"updated_at":          time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to cancel courier request")
		return fmt.Errorf("failed to cancel request: %w", err)
	}

	return nil
}

func (r *courierRepository) AssignCourier(id, courierID primitive.ObjectID) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"courier_id": courierID,
			"status":     models.CourierStatusAccepted,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to assign courier")
		return fmt.Errorf("failed to assign courier: %w", err)
	}

	return nil
}

func (r *courierRepository) UnassignCourier(id primitive.ObjectID) error {
	collection := r.db.Collection("courier_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$unset": bson.M{
			"courier_id": "",
		},
		"$set": bson.M{
			"status":     models.CourierStatusPending,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to unassign courier")
		return fmt.Errorf("failed to unassign courier: %w", err)
	}

	return nil
}

// SavedAddressRepository Implementation

func (r *savedAddressRepository) Create(address *models.SavedAddress) (*models.SavedAddress, error) {
	collection := r.db.Collection("saved_addresses")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, address)
	if err != nil {
		r.logger.Error().Err(err).Str("address_id", address.ID.Hex()).Msg("Failed to create saved address")
		return nil, fmt.Errorf("failed to create saved address: %w", err)
	}

	return address, nil
}

func (r *savedAddressRepository) GetByUserID(userID primitive.ObjectID) ([]*models.SavedAddress, error) {
	collection := r.db.Collection("saved_addresses")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get saved addresses")
		return nil, fmt.Errorf("failed to get saved addresses: %w", err)
	}
	defer cursor.Close(ctx)

	var addresses []*models.SavedAddress
	for cursor.Next(ctx) {
		var address models.SavedAddress
		if err := cursor.Decode(&address); err != nil {
			r.logger.Error().Err(err).Msg("Failed to decode saved address")
			continue
		}
		addresses = append(addresses, &address)
	}

	return addresses, nil
}

func (r *savedAddressRepository) GetByID(id primitive.ObjectID) (*models.SavedAddress, error) {
	collection := r.db.Collection("saved_addresses")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var address models.SavedAddress
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&address)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("address not found")
		}
		r.logger.Error().Err(err).Str("address_id", id.Hex()).Msg("Failed to get saved address")
		return nil, fmt.Errorf("failed to get saved address: %w", err)
	}

	return &address, nil
}

func (r *savedAddressRepository) Update(id primitive.ObjectID, address *models.SavedAddress) (*models.SavedAddress, error) {
	collection := r.db.Collection("saved_addresses")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	address.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(ctx, bson.M{"_id": id}, address)
	if err != nil {
		r.logger.Error().Err(err).Str("address_id", id.Hex()).Msg("Failed to update saved address")
		return nil, fmt.Errorf("failed to update saved address: %w", err)
	}

	return address, nil
}

func (r *savedAddressRepository) Delete(id primitive.ObjectID) error {
	collection := r.db.Collection("saved_addresses")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		r.logger.Error().Err(err).Str("address_id", id.Hex()).Msg("Failed to delete saved address")
		return fmt.Errorf("failed to delete saved address: %w", err)
	}

	return nil
}

// CourierIssueRepository Implementation (placeholder - would need proper models)

func (r *courierIssueRepository) CreateIssue(issue *models.CourierIssue) (*models.CourierIssue, error) {
	// Implementation would depend on CourierIssue model structure
	return nil, fmt.Errorf("not implemented")
}

func (r *courierIssueRepository) GetIssuesByRequestID(requestID primitive.ObjectID) ([]*models.CourierIssue, error) {
	// Implementation would depend on CourierIssue model structure
	return nil, fmt.Errorf("not implemented")
}

func (r *courierIssueRepository) CreateClaim(claim *models.CourierClaim) (*models.CourierClaim, error) {
	// Implementation would depend on CourierClaim model structure
	return nil, fmt.Errorf("not implemented")
}

func (r *courierIssueRepository) GetClaimsByRequestID(requestID primitive.ObjectID) ([]*models.CourierClaim, error) {
	// Implementation would depend on CourierClaim model structure
	return nil, fmt.Errorf("not implemented")
}

func (r *courierIssueRepository) UpdateIssueStatus(issueID primitive.ObjectID, status string) error {
	// Implementation would depend on CourierIssue model structure
	return fmt.Errorf("not implemented")
}

func (r *courierIssueRepository) UpdateClaimStatus(claimID primitive.ObjectID, status string) error {
	// Implementation would depend on CourierClaim model structure
	return fmt.Errorf("not implemented")
}
