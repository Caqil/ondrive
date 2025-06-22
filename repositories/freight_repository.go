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

// FreightRepository interface defines freight-related database operations
type FreightRepository interface {
	// Basic CRUD operations
	Create(request *models.FreightRequest) (*models.FreightRequest, error)
	GetByID(id primitive.ObjectID) (*models.FreightRequest, error)
	Update(request *models.FreightRequest) (*models.FreightRequest, error)
	Delete(id primitive.ObjectID) error
	GetByFilter(filter map[string]interface{}, page, limit int) ([]*models.FreightRequest, int64, error)

	// Cargo management
	AddCargoPhotos(requestID primitive.ObjectID, photos []*models.CargoPhoto) error
	GetCargoPhotos(requestID primitive.ObjectID, photoType string) ([]*models.CargoPhoto, error)
	UpdateCargoDetails(requestID primitive.ObjectID, cargo *models.FreightCargo) error

	// Multi-stop delivery
	AddStop(requestID primitive.ObjectID, stop *models.FreightStop) (*models.FreightStop, error)
	UpdateStop(requestID, stopID primitive.ObjectID, stop *models.FreightStop) (*models.FreightStop, error)
	RemoveStop(requestID, stopID primitive.ObjectID) error
	GetStops(requestID primitive.ObjectID) ([]*models.FreightStop, error)
	ReorderStops(requestID primitive.ObjectID, stops []*models.FreightStop) error

	// Tracking and milestones
	AddMilestone(requestID primitive.ObjectID, milestone *MilestoneUpdate) (*MilestoneUpdate, error)
	GetMilestones(requestID primitive.ObjectID) ([]*MilestoneUpdate, error)
	UpdateStatus(requestID primitive.ObjectID, status models.FreightStatus) error
	UpdateLocation(requestID primitive.ObjectID, location *models.Location) error

	// Documentation
	AddDocument(requestID primitive.ObjectID, document *models.FreightDocument) error
	GetDocuments(requestID primitive.ObjectID, docType string) ([]*models.FreightDocument, error)
	RemoveDocument(requestID, documentID primitive.ObjectID) error

	// Analytics and reporting
	GetFreightStats(userID string, startDate, endDate time.Time) (*FreightStats, error)
	GetAnalytics(userID string, period string) (*FreightAnalytics, error)
	GetCostBreakdown(userID string, filters map[string]interface{}) (*CostBreakdown, error)

	// Search and filtering
	SearchRequests(query string, filters map[string]interface{}, page, limit int) ([]*models.FreightRequest, int64, error)
	GetRequestsByStatus(status models.FreightStatus, page, limit int) ([]*models.FreightRequest, int64, error)
	GetRequestsByCargoType(cargoType models.CargoType, page, limit int) ([]*models.FreightRequest, int64, error)
	GetRequestsByVehicleType(vehicleType models.FreightVehicleType, page, limit int) ([]*models.FreightRequest, int64, error)

	// Performance optimization
	CreateIndexes() error
	CleanupOldRequests(retentionDays int) error
	ArchiveCompletedRequests(archiveAfterDays int) error
}

// Supporting types for repository operations
type MilestoneUpdate struct {
	ID            primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	RequestID     primitive.ObjectID     `json:"request_id" bson:"request_id"`
	MilestoneType string                 `json:"milestone_type" bson:"milestone_type"`
	Status        string                 `json:"status" bson:"status"`
	Location      models.Location        `json:"location" bson:"location"`
	Timestamp     time.Time              `json:"timestamp" bson:"timestamp"`
	Notes         string                 `json:"notes" bson:"notes"`
	Photos        []string               `json:"photos" bson:"photos"`
	UpdatedBy     primitive.ObjectID     `json:"updated_by" bson:"updated_by"`
	Metadata      map[string]interface{} `json:"metadata" bson:"metadata"`
	CreatedAt     time.Time              `json:"created_at" bson:"created_at"`
}

type FreightStats struct {
	UserID            string           `json:"user_id" bson:"user_id"`
	TotalRequests     int64            `json:"total_requests" bson:"total_requests"`
	CompletedRequests int64            `json:"completed_requests" bson:"completed_requests"`
	CancelledRequests int64            `json:"cancelled_requests" bson:"cancelled_requests"`
	TotalValue        float64          `json:"total_value" bson:"total_value"`
	TotalWeight       float64          `json:"total_weight" bson:"total_weight"`
	TotalVolume       float64          `json:"total_volume" bson:"total_volume"`
	TotalDistance     float64          `json:"total_distance" bson:"total_distance"`
	TotalCost         float64          `json:"total_cost" bson:"total_cost"`
	AverageCost       float64          `json:"average_cost" bson:"average_cost"`
	SuccessRate       float64          `json:"success_rate" bson:"success_rate"`
	OnTimeRate        float64          `json:"on_time_rate" bson:"on_time_rate"`
	CargoBreakdown    map[string]int64 `json:"cargo_breakdown" bson:"cargo_breakdown"`
	VehicleBreakdown  map[string]int64 `json:"vehicle_breakdown" bson:"vehicle_breakdown"`
	RouteBreakdown    map[string]int64 `json:"route_breakdown" bson:"route_breakdown"`
	MonthlyTrends     []MonthlyTrend   `json:"monthly_trends" bson:"monthly_trends"`
}

type MonthlyTrend struct {
	Month       string  `json:"month" bson:"month"`
	Count       int64   `json:"count" bson:"count"`
	TotalCost   float64 `json:"total_cost" bson:"total_cost"`
	AverageCost float64 `json:"average_cost" bson:"average_cost"`
	TotalWeight float64 `json:"total_weight" bson:"total_weight"`
}

type FreightAnalytics struct {
	UserID               string             `json:"user_id"`
	Period               string             `json:"period"`
	TotalRequests        int64              `json:"total_requests"`
	CompletedRequests    int64              `json:"completed_requests"`
	CancelledRequests    int64              `json:"cancelled_requests"`
	TotalSpent           float64            `json:"total_spent"`
	AverageCost          float64            `json:"average_cost"`
	TotalWeight          float64            `json:"total_weight"`
	TotalDistance        float64            `json:"total_distance"`
	MostUsedVehicleType  string             `json:"most_used_vehicle_type"`
	MostShippedCargoType string             `json:"most_shipped_cargo_type"`
	TopRoutes            []RouteStatistic   `json:"top_routes"`
	CostTrends           []CostTrend        `json:"cost_trends"`
	PerformanceMetrics   PerformanceMetrics `json:"performance_metrics"`
}

type RouteStatistic struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Count     int64   `json:"count"`
	TotalCost float64 `json:"total_cost"`
	AvgCost   float64 `json:"avg_cost"`
}

type CostTrend struct {
	Period    string  `json:"period"`
	TotalCost float64 `json:"total_cost"`
	Count     int64   `json:"count"`
	AvgCost   float64 `json:"avg_cost"`
}

type PerformanceMetrics struct {
	OnTimeDeliveryRate   float64 `json:"on_time_delivery_rate"`
	AverageDeliveryTime  int     `json:"average_delivery_time"`
	CustomerSatisfaction float64 `json:"customer_satisfaction"`
	CarrierRating        float64 `json:"carrier_rating"`
	DamageRate           float64 `json:"damage_rate"`
	CancellationRate     float64 `json:"cancellation_rate"`
}

type CostBreakdown struct {
	Period            string            `json:"period"`
	FreightID         string            `json:"freight_id,omitempty"`
	TotalCost         float64           `json:"total_cost"`
	CostByCategory    []CostCategory    `json:"cost_by_category"`
	CostByVehicleType []VehicleTypeCost `json:"cost_by_vehicle_type"`
	CostByCargoType   []CargoTypeCost   `json:"cost_by_cargo_type"`
	CostByRoute       []RouteCost       `json:"cost_by_route"`
	MonthlyTrends     []MonthlyCost     `json:"monthly_trends"`
}

type CostCategory struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type VehicleTypeCost struct {
	VehicleType string  `json:"vehicle_type"`
	Amount      float64 `json:"amount"`
	Count       int64   `json:"count"`
	AvgCost     float64 `json:"avg_cost"`
}

type CargoTypeCost struct {
	CargoType string  `json:"cargo_type"`
	Amount    float64 `json:"amount"`
	Count     int64   `json:"count"`
	AvgCost   float64 `json:"avg_cost"`
}

type RouteCost struct {
	Route   string  `json:"route"`
	Amount  float64 `json:"amount"`
	Count   int64   `json:"count"`
	AvgCost float64 `json:"avg_cost"`
}

type MonthlyCost struct {
	Month   string  `json:"month"`
	Amount  float64 `json:"amount"`
	Count   int64   `json:"count"`
	AvgCost float64 `json:"avg_cost"`
}

// freightRepository implements FreightRepository interface
type freightRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

// NewFreightRepository creates a new freight repository
func NewFreightRepository(db *mongo.Database, logger utils.Logger) FreightRepository {
	repo := &freightRepository{
		db:     db,
		logger: logger,
	}

	// Create indexes for better performance
	go repo.CreateIndexes()

	return repo
}

// Basic CRUD operations

func (r *freightRepository) Create(request *models.FreightRequest) (*models.FreightRequest, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, request)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", request.ID.Hex()).Msg("Failed to create freight request")
		return nil, fmt.Errorf("failed to create freight request: %w", err)
	}

	request.ID = result.InsertedID.(primitive.ObjectID)
	return request, nil
}

func (r *freightRepository) GetByID(id primitive.ObjectID) (*models.FreightRequest, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var request models.FreightRequest
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("freight request not found")
		}
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to get freight request")
		return nil, fmt.Errorf("failed to get freight request: %w", err)
	}

	return &request, nil
}

func (r *freightRepository) Update(request *models.FreightRequest) (*models.FreightRequest, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request.UpdatedAt = time.Now()

	filter := bson.M{"_id": request.ID}
	update := bson.M{"$set": request}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", request.ID.Hex()).Msg("Failed to update freight request")
		return nil, fmt.Errorf("failed to update freight request: %w", err)
	}

	return request, nil
}

func (r *freightRepository) Delete(id primitive.ObjectID) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", id.Hex()).Msg("Failed to delete freight request")
		return fmt.Errorf("failed to delete freight request: %w", err)
	}

	return nil
}

func (r *freightRepository) GetByFilter(filter map[string]interface{}, page, limit int) ([]*models.FreightRequest, int64, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert filter to BSON
	bsonFilter := bson.M{}
	for k, v := range filter {
		bsonFilter[k] = v
	}

	// Count total documents
	total, err := collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count freight requests")
		return nil, 0, fmt.Errorf("failed to count freight requests: %w", err)
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find documents with pagination
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to find freight requests")
		return nil, 0, fmt.Errorf("failed to find freight requests: %w", err)
	}
	defer cursor.Close(ctx)

	var requests []*models.FreightRequest
	if err := cursor.All(ctx, &requests); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode freight requests")
		return nil, 0, fmt.Errorf("failed to decode freight requests: %w", err)
	}

	return requests, total, nil
}

// Cargo management

func (r *freightRepository) AddCargoPhotos(requestID primitive.ObjectID, photos []*models.CargoPhoto) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$push": bson.M{"cargo.photos": bson.M{"$each": photos}},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to add cargo photos")
		return fmt.Errorf("failed to add cargo photos: %w", err)
	}

	return nil
}

func (r *freightRepository) GetCargoPhotos(requestID primitive.ObjectID, photoType string) ([]*models.CargoPhoto, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}

	var request models.FreightRequest
	err := collection.FindOne(ctx, filter).Decode(&request)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to get freight request for photos")
		return nil, fmt.Errorf("failed to get freight request: %w", err)
	}

	// Filter photos by type if specified
	if photoType != "" {
		var filteredPhotos []*models.CargoPhoto
		for i := range request.Cargo.Photos {
			if request.Cargo.Photos[i].Type == photoType {
				filteredPhotos = append(filteredPhotos, &request.Cargo.Photos[i])
			}
		}
		return filteredPhotos, nil
	}

	var photos []*models.CargoPhoto
	for i := range request.Cargo.Photos {
		photos = append(photos, &request.Cargo.Photos[i])
	}
	return photos, nil
}

func (r *freightRepository) UpdateCargoDetails(requestID primitive.ObjectID, cargo *models.FreightCargo) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$set": bson.M{
			"cargo":      cargo,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to update cargo details")
		return fmt.Errorf("failed to update cargo details: %w", err)
	}

	return nil
}

// Multi-stop delivery

func (r *freightRepository) AddStop(requestID primitive.ObjectID, stop *models.FreightStop) (*models.FreightStop, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stop.ID = primitive.NewObjectID()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$push": bson.M{"route_stops": stop},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to add stop")
		return nil, fmt.Errorf("failed to add stop: %w", err)
	}

	return stop, nil
}

func (r *freightRepository) UpdateStop(requestID, stopID primitive.ObjectID, stop *models.FreightStop) (*models.FreightStop, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":             requestID,
		"route_stops._id": stopID,
	}
	update := bson.M{
		"$set": bson.M{
			"route_stops.$": stop,
			"updated_at":    time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Str("stop_id", stopID.Hex()).Msg("Failed to update stop")
		return nil, fmt.Errorf("failed to update stop: %w", err)
	}

	return stop, nil
}

func (r *freightRepository) RemoveStop(requestID, stopID primitive.ObjectID) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$pull": bson.M{"route_stops": bson.M{"_id": stopID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Str("stop_id", stopID.Hex()).Msg("Failed to remove stop")
		return fmt.Errorf("failed to remove stop: %w", err)
	}

	return nil
}

func (r *freightRepository) GetStops(requestID primitive.ObjectID) ([]*models.FreightStop, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}

	var request models.FreightRequest
	err := collection.FindOne(ctx, filter).Decode(&request)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to get freight request for stops")
		return nil, fmt.Errorf("failed to get freight request: %w", err)
	}

	var stops []*models.FreightStop
	for i := range request.RouteStops {
		stops = append(stops, &request.RouteStops[i])
	}
	return stops, nil
}

func (r *freightRepository) ReorderStops(requestID primitive.ObjectID, stops []*models.FreightStop) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$set": bson.M{
			"route_stops": stops,
			"updated_at":  time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to reorder stops")
		return fmt.Errorf("failed to reorder stops: %w", err)
	}

	return nil
}

// Tracking and milestones

func (r *freightRepository) AddMilestone(requestID primitive.ObjectID, milestone *MilestoneUpdate) (*MilestoneUpdate, error) {
	collection := r.db.Collection("freight_milestones")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	milestone.ID = primitive.NewObjectID()
	milestone.RequestID = requestID
	milestone.CreatedAt = time.Now()

	_, err := collection.InsertOne(ctx, milestone)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to add milestone")
		return nil, fmt.Errorf("failed to add milestone: %w", err)
	}

	return milestone, nil
}

func (r *freightRepository) GetMilestones(requestID primitive.ObjectID) ([]*MilestoneUpdate, error) {
	collection := r.db.Collection("freight_milestones")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"request_id": requestID}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to get milestones")
		return nil, fmt.Errorf("failed to get milestones: %w", err)
	}
	defer cursor.Close(ctx)

	var milestones []*MilestoneUpdate
	if err := cursor.All(ctx, &milestones); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode milestones")
		return nil, fmt.Errorf("failed to decode milestones: %w", err)
	}

	return milestones, nil
}

func (r *freightRepository) UpdateStatus(requestID primitive.ObjectID, status models.FreightStatus) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to update status")
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (r *freightRepository) UpdateLocation(requestID primitive.ObjectID, location *models.Location) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$set": bson.M{
			"current_location": location,
			"updated_at":       time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to update location")
		return fmt.Errorf("failed to update location: %w", err)
	}

	return nil
}

// Documentation

func (r *freightRepository) AddDocument(requestID primitive.ObjectID, document *models.FreightDocument) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$push": bson.M{"documents": document},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to add document")
		return fmt.Errorf("failed to add document: %w", err)
	}

	return nil
}

func (r *freightRepository) GetDocuments(requestID primitive.ObjectID, docType string) ([]*models.FreightDocument, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}

	var request models.FreightRequest
	err := collection.FindOne(ctx, filter).Decode(&request)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Msg("Failed to get freight request for documents")
		return nil, fmt.Errorf("failed to get freight request: %w", err)
	}

	// Filter documents by type if specified
	if docType != "" {
		var filteredDocs []*models.FreightDocument
		for i := range request.Documents {
			if request.Documents[i].Type == docType {
				filteredDocs = append(filteredDocs, &request.Documents[i])
			}
		}
		return filteredDocs, nil
	}

	var docs []*models.FreightDocument
	for i := range request.Documents {
		docs = append(docs, &request.Documents[i])
	}
	return docs, nil
}

func (r *freightRepository) RemoveDocument(requestID, documentID primitive.ObjectID) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": requestID}
	update := bson.M{
		"$pull": bson.M{"documents": bson.M{"_id": documentID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("request_id", requestID.Hex()).Str("doc_id", documentID.Hex()).Msg("Failed to remove document")
		return fmt.Errorf("failed to remove document: %w", err)
	}

	return nil
}

// Analytics and reporting

func (r *freightRepository) GetFreightStats(userID string, startDate, endDate time.Time) (*FreightStats, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Build aggregation pipeline
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"shipper_id": userObjID,
			"created_at": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":            nil,
			"total_requests": bson.M{"$sum": 1},
			"completed_requests": bson.M{"$sum": bson.M{"$cond": bson.A{
				bson.M{"$eq": bson.A{"$status", models.FreightStatusDelivered}},
				1,
				0,
			}}},
			"cancelled_requests": bson.M{"$sum": bson.M{"$cond": bson.A{
				bson.M{"$eq": bson.A{"$status", models.FreightStatusCancelled}},
				1,
				0,
			}}},
			"total_value":  bson.M{"$sum": "$cargo.total_value"},
			"total_weight": bson.M{"$sum": "$cargo.total_weight"},
			"total_volume": bson.M{"$sum": "$cargo.total_volume"},
			"total_cost":   bson.M{"$sum": "$fare.total_amount"},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get freight stats")
		return nil, fmt.Errorf("failed to get freight stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	if len(results) == 0 {
		return &FreightStats{UserID: userID}, nil
	}

	result := results[0]
	stats := &FreightStats{
		UserID:            userID,
		TotalRequests:     getInt64FromBSON(result, "total_requests"),
		CompletedRequests: getInt64FromBSON(result, "completed_requests"),
		CancelledRequests: getInt64FromBSON(result, "cancelled_requests"),
		TotalValue:        getFloat64FromBSON(result, "total_value"),
		TotalWeight:       getFloat64FromBSON(result, "total_weight"),
		TotalVolume:       getFloat64FromBSON(result, "total_volume"),
		TotalCost:         getFloat64FromBSON(result, "total_cost"),
	}

	// Calculate derived metrics
	if stats.TotalRequests > 0 {
		stats.AverageCost = stats.TotalCost / float64(stats.TotalRequests)
		stats.SuccessRate = float64(stats.CompletedRequests) / float64(stats.TotalRequests) * 100
	}

	return stats, nil
}

func (r *freightRepository) GetAnalytics(userID string, period string) (*FreightAnalytics, error) {
	// Simplified implementation - would build complex aggregation pipeline
	stats, err := r.GetFreightStats(userID, time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		return nil, err
	}

	return &FreightAnalytics{
		UserID:            userID,
		Period:            period,
		TotalRequests:     stats.TotalRequests,
		CompletedRequests: stats.CompletedRequests,
		CancelledRequests: stats.CancelledRequests,
		TotalSpent:        stats.TotalCost,
		AverageCost:       stats.AverageCost,
		TotalWeight:       stats.TotalWeight,
		// Other fields would be calculated similarly
	}, nil
}

func (r *freightRepository) GetCostBreakdown(userID string, filters map[string]interface{}) (*CostBreakdown, error) {
	// Simplified implementation
	return &CostBreakdown{
		TotalCost: 0,
		// Would implement proper aggregation
	}, nil
}

// Search and filtering

func (r *freightRepository) SearchRequests(query string, filters map[string]interface{}, page, limit int) ([]*models.FreightRequest, int64, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build search filter
	searchFilter := bson.M{}

	// Add text search if query provided
	if query != "" {
		searchFilter["$text"] = bson.M{"$search": query}
	}

	// Add additional filters
	for k, v := range filters {
		searchFilter[k] = v
	}

	return r.findWithPagination(ctx, collection, searchFilter, page, limit)
}

func (r *freightRepository) GetRequestsByStatus(status models.FreightStatus, page, limit int) ([]*models.FreightRequest, int64, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"status": status}
	return r.findWithPagination(ctx, collection, filter, page, limit)
}

func (r *freightRepository) GetRequestsByCargoType(cargoType models.CargoType, page, limit int) ([]*models.FreightRequest, int64, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"cargo.type": cargoType}
	return r.findWithPagination(ctx, collection, filter, page, limit)
}

func (r *freightRepository) GetRequestsByVehicleType(vehicleType models.FreightVehicleType, page, limit int) ([]*models.FreightRequest, int64, error) {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"vehicle_requirements.vehicle_type": vehicleType}
	return r.findWithPagination(ctx, collection, filter, page, limit)
}

// Performance optimization

func (r *freightRepository) CreateIndexes() error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "shipper_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "carrier_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "cargo.type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "vehicle_requirements.vehicle_type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
		},
		{
			Keys: bson.D{
				{Key: "pickup_location.coordinates", Value: "2dsphere"},
			},
		},
		{
			Keys: bson.D{
				{Key: "delivery_location.coordinates", Value: "2dsphere"},
			},
		},
		{
			Keys: bson.D{
				{Key: "shipper_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create indexes")
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Create text index for search
	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "cargo.description", Value: "text"},
			{Key: "special_instructions", Value: "text"},
			{Key: "pickup_location.address", Value: "text"},
			{Key: "delivery_location.address", Value: "text"},
		},
	}

	_, err = collection.Indexes().CreateOne(ctx, textIndex)
	if err != nil {
		r.logger.Warn().Err(err).Msg("Failed to create text index")
	}

	r.logger.Info().Msg("Created indexes for freight_requests collection")
	return nil
}

func (r *freightRepository) CleanupOldRequests(retentionDays int) error {
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	filter := bson.M{
		"created_at": bson.M{"$lt": cutoffDate},
		"status": bson.M{"$in": []models.FreightStatus{
			models.FreightStatusCancelled,
			models.FreightStatusDelivered,
		}},
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup old requests")
		return fmt.Errorf("failed to cleanup old requests: %w", err)
	}

	r.logger.Info().Int64("deleted_count", result.DeletedCount).Msg("Cleaned up old freight requests")
	return nil
}

func (r *freightRepository) ArchiveCompletedRequests(archiveAfterDays int) error {
	// Implementation would move completed requests to archive collection
	// For now, just mark them as archived
	collection := r.db.Collection("freight_requests")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoffDate := time.Now().AddDate(0, 0, -archiveAfterDays)
	filter := bson.M{
		"delivered_at": bson.M{"$lt": cutoffDate},
		"status":       models.FreightStatusDelivered,
		"archived":     bson.M{"$ne": true},
	}

	update := bson.M{"$set": bson.M{"archived": true, "archived_at": time.Now()}}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to archive completed requests")
		return fmt.Errorf("failed to archive completed requests: %w", err)
	}

	r.logger.Info().Int64("archived_count", result.ModifiedCount).Msg("Archived completed freight requests")
	return nil
}

// Helper methods

func (r *freightRepository) findWithPagination(ctx context.Context, collection *mongo.Collection, filter bson.M, page, limit int) ([]*models.FreightRequest, int64, error) {
	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find documents with pagination
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var requests []*models.FreightRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, 0, fmt.Errorf("failed to decode documents: %w", err)
	}

	return requests, total, nil
}

// Helper functions for BSON value extraction
func getInt64FromBSON(doc bson.M, key string) int64 {
	if val, ok := doc[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int32:
			return int64(v)
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

func getFloat64FromBSON(doc bson.M, key string) float64 {
	if val, ok := doc[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return 0.0
}
