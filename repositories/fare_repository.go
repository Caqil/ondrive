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

// FareRepository interface defines fare-related database operations
type FareRepository interface {
	// Rate Cards
	GetRateCard(serviceType, vehicleType string, location models.Location) (*models.RateCard, error)
	GetRateCards(serviceType, vehicleType, city, country string) ([]*models.RateCard, error)
	GetRateCardByType(serviceType, city string) (*models.RateCard, error)
	CreateRateCard(rateCard *models.RateCard) (*models.RateCard, error)
	UpdateRateCard(rateCard *models.RateCard) (*models.RateCard, error)
	DeleteRateCard(id primitive.ObjectID) error

	// Fare Details
	SaveFareDetails(rideID primitive.ObjectID, fareDetails *models.FareDetails) error
	GetFareDetails(rideID primitive.ObjectID) (*models.FareDetails, error)
	UpdateFareDetails(rideID primitive.ObjectID, fareDetails *models.FareDetails) error

	// Fare Negotiations
	CreateNegotiation(negotiation *models.FareNegotiation) (*models.FareNegotiation, error)
	GetNegotiation(negotiationID primitive.ObjectID) (*models.FareNegotiation, error)
	GetActiveNegotiation(rideID primitive.ObjectID) (*models.FareNegotiation, error)
	UpdateNegotiation(negotiation *models.FareNegotiation) (*models.FareNegotiation, error)
	GetNegotiationHistory(rideID primitive.ObjectID, userID primitive.ObjectID) ([]*models.FareNegotiation, error)
	ExpireNegotiations() error

	// Promotional Codes
	GetPromoCode(code string) (*models.PromotionalRate, error)
	GetActivePromoCodes(userID string) ([]*models.PromotionalRate, error)
	CreatePromoCode(promo *models.PromotionalRate) (*models.PromotionalRate, error)
	UpdatePromoCode(promo *models.PromotionalRate) (*models.PromotionalRate, error)
	HasUserUsedPromo(userID, promoCode string) (bool, error)
	RecordPromoUsage(userID, promoCode string, rideID primitive.ObjectID) error

	// Fare History & Analytics
	GetFareHistoryByUser(userID string, page, limit int, filters map[string]interface{}) ([]*models.FareDetails, int64, error)
	GetFareStatistics(userID string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetMarketRates(serviceType, city, country string) (*models.MarketRate, error)
	UpdateMarketRates(marketRate *models.MarketRate) error

	// Dynamic Pricing
	GetDynamicPricing(location models.Location, serviceType string) (*models.DynamicPricing, error)
	SaveDynamicPricing(pricing *models.DynamicPricing) error
	UpdateDynamicPricing(pricing *models.DynamicPricing) error

	// Fare Rules
	GetFareRules(serviceType, city, country string) ([]*FareRule, error)
	CreateFareRule(rule *FareRule) (*FareRule, error)
	UpdateFareRule(rule *FareRule) (*FareRule, error)
	DeleteFareRule(id primitive.ObjectID) error
}

// FareRule represents fare calculation rules
type FareRule struct {
	ID          primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Name        string                 `json:"name" bson:"name"`
	Description string                 `json:"description" bson:"description"`
	RuleType    string                 `json:"rule_type" bson:"rule_type"`
	ServiceType string                 `json:"service_type" bson:"service_type"`
	City        string                 `json:"city" bson:"city"`
	Country     string                 `json:"country" bson:"country"`
	Conditions  map[string]interface{} `json:"conditions" bson:"conditions"`
	Actions     map[string]interface{} `json:"actions" bson:"actions"`
	Priority    int                    `json:"priority" bson:"priority"`
	IsActive    bool                   `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" bson:"updated_at"`
}

// PromoUsage tracks promo code usage
type PromoUsage struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id"`
	PromoCode string             `json:"promo_code" bson:"promo_code"`
	RideID    primitive.ObjectID `json:"ride_id" bson:"ride_id"`
	UsedAt    time.Time          `json:"used_at" bson:"used_at"`
}

// fareRepository implements FareRepository interface
type fareRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

// NewFareRepository creates a new fare repository
func NewFareRepository(db *mongo.Database, logger utils.Logger) FareRepository {
	return &fareRepository{
		db:     db,
		logger: logger,
	}
}

// Rate Cards

func (r *fareRepository) GetRateCard(serviceType, vehicleType string, location models.Location) (*models.RateCard, error) {
	collection := r.db.Collection("rate_cards")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"service_type": serviceType,
		"vehicle_type": vehicleType,
		"is_active":    true,
	}

	// Add location-based filtering (simplified - would need geospatial queries in production)
	if location.Latitude != 0 && location.Longitude != 0 {
		// For now, just filter by city/country if available
		// In production, you'd use geospatial queries to find the appropriate rate card
	}

	var rateCard models.RateCard
	err := collection.FindOne(ctx, filter).Decode(&rateCard)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no rate card found for service type %s and vehicle type %s", serviceType, vehicleType)
		}
		r.logger.Error().Err(err).Msg("Failed to get rate card")
		return nil, fmt.Errorf("failed to get rate card: %w", err)
	}

	return &rateCard, nil
}

func (r *fareRepository) GetRateCards(serviceType, vehicleType, city, country string) ([]*models.RateCard, error) {
	collection := r.db.Collection("rate_cards")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"is_active": true}

	if serviceType != "" {
		filter["service_type"] = serviceType
	}
	if vehicleType != "" {
		filter["vehicle_type"] = vehicleType
	}
	if city != "" {
		filter["city"] = city
	}
	if country != "" {
		filter["country"] = country
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get rate cards")
		return nil, fmt.Errorf("failed to get rate cards: %w", err)
	}
	defer cursor.Close(ctx)

	var rateCards []*models.RateCard
	if err := cursor.All(ctx, &rateCards); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode rate cards")
		return nil, fmt.Errorf("failed to decode rate cards: %w", err)
	}

	return rateCards, nil
}

func (r *fareRepository) GetRateCardByType(serviceType, city string) (*models.RateCard, error) {
	collection := r.db.Collection("rate_cards")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"service_type": serviceType,
		"is_active":    true,
	}

	if city != "" {
		filter["city"] = city
	}

	var rateCard models.RateCard
	err := collection.FindOne(ctx, filter).Decode(&rateCard)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no rate card found for service type %s", serviceType)
		}
		r.logger.Error().Err(err).Msg("Failed to get rate card by type")
		return nil, fmt.Errorf("failed to get rate card: %w", err)
	}

	return &rateCard, nil
}

func (r *fareRepository) CreateRateCard(rateCard *models.RateCard) (*models.RateCard, error) {
	collection := r.db.Collection("rate_cards")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rateCard.CreatedAt = time.Now()
	rateCard.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, rateCard)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create rate card")
		return nil, fmt.Errorf("failed to create rate card: %w", err)
	}

	rateCard.ID = result.InsertedID.(primitive.ObjectID)
	return rateCard, nil
}

func (r *fareRepository) UpdateRateCard(rateCard *models.RateCard) (*models.RateCard, error) {
	collection := r.db.Collection("rate_cards")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rateCard.UpdatedAt = time.Now()

	filter := bson.M{"_id": rateCard.ID}
	update := bson.M{"$set": rateCard}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("rate_card_id", rateCard.ID.Hex()).Msg("Failed to update rate card")
		return nil, fmt.Errorf("failed to update rate card: %w", err)
	}

	return rateCard, nil
}

func (r *fareRepository) DeleteRateCard(id primitive.ObjectID) error {
	collection := r.db.Collection("rate_cards")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("rate_card_id", id.Hex()).Msg("Failed to delete rate card")
		return fmt.Errorf("failed to delete rate card: %w", err)
	}

	return nil
}

// Fare Details

func (r *fareRepository) SaveFareDetails(rideID primitive.ObjectID, fareDetails *models.FareDetails) error {
	collection := r.db.Collection("fare_details")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add ride ID to fare details
	fareDetailsDoc := bson.M{
		"ride_id":      rideID,
		"fare_details": fareDetails,
		"created_at":   time.Now(),
		"updated_at":   time.Now(),
	}

	_, err := collection.InsertOne(ctx, fareDetailsDoc)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID.Hex()).Msg("Failed to save fare details")
		return fmt.Errorf("failed to save fare details: %w", err)
	}

	return nil
}

func (r *fareRepository) GetFareDetails(rideID primitive.ObjectID) (*models.FareDetails, error) {
	collection := r.db.Collection("fare_details")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"ride_id": rideID}

	var result struct {
		FareDetails *models.FareDetails `bson:"fare_details"`
	}

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("fare details not found for ride %s", rideID.Hex())
		}
		r.logger.Error().Err(err).Str("ride_id", rideID.Hex()).Msg("Failed to get fare details")
		return nil, fmt.Errorf("failed to get fare details: %w", err)
	}

	return result.FareDetails, nil
}

func (r *fareRepository) UpdateFareDetails(rideID primitive.ObjectID, fareDetails *models.FareDetails) error {
	collection := r.db.Collection("fare_details")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"ride_id": rideID}
	update := bson.M{
		"$set": bson.M{
			"fare_details": fareDetails,
			"updated_at":   time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID.Hex()).Msg("Failed to update fare details")
		return fmt.Errorf("failed to update fare details: %w", err)
	}

	return nil
}

// Fare Negotiations

func (r *fareRepository) CreateNegotiation(negotiation *models.FareNegotiation) (*models.FareNegotiation, error) {
	collection := r.db.Collection("fare_negotiations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, negotiation)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create fare negotiation")
		return nil, fmt.Errorf("failed to create fare negotiation: %w", err)
	}

	negotiation.ID = result.InsertedID.(primitive.ObjectID)
	return negotiation, nil
}

func (r *fareRepository) GetNegotiation(negotiationID primitive.ObjectID) (*models.FareNegotiation, error) {
	collection := r.db.Collection("fare_negotiations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var negotiation models.FareNegotiation
	err := collection.FindOne(ctx, bson.M{"_id": negotiationID}).Decode(&negotiation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("negotiation not found")
		}
		r.logger.Error().Err(err).Str("negotiation_id", negotiationID.Hex()).Msg("Failed to get negotiation")
		return nil, fmt.Errorf("failed to get negotiation: %w", err)
	}

	return &negotiation, nil
}

func (r *fareRepository) GetActiveNegotiation(rideID primitive.ObjectID) (*models.FareNegotiation, error) {
	collection := r.db.Collection("fare_negotiations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"ride_id": rideID,
		"status": bson.M{
			"$in": []models.FareStatus{
				models.FareStatusPending,
				models.FareStatusProposed,
				models.FareStatusCountered,
			},
		},
	}

	var negotiation models.FareNegotiation
	err := collection.FindOne(ctx, filter).Decode(&negotiation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No active negotiation found
		}
		r.logger.Error().Err(err).Str("ride_id", rideID.Hex()).Msg("Failed to get active negotiation")
		return nil, fmt.Errorf("failed to get active negotiation: %w", err)
	}

	return &negotiation, nil
}

func (r *fareRepository) UpdateNegotiation(negotiation *models.FareNegotiation) (*models.FareNegotiation, error) {
	collection := r.db.Collection("fare_negotiations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": negotiation.ID}
	update := bson.M{"$set": negotiation}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("negotiation_id", negotiation.ID.Hex()).Msg("Failed to update negotiation")
		return nil, fmt.Errorf("failed to update negotiation: %w", err)
	}

	return negotiation, nil
}

func (r *fareRepository) GetNegotiationHistory(rideID primitive.ObjectID, userID primitive.ObjectID) ([]*models.FareNegotiation, error) {
	collection := r.db.Collection("fare_negotiations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"ride_id": rideID,
		"$or": []bson.M{
			{"passenger_id": userID},
			{"driver_id": userID},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("ride_id", rideID.Hex()).Msg("Failed to get negotiation history")
		return nil, fmt.Errorf("failed to get negotiation history: %w", err)
	}
	defer cursor.Close(ctx)

	var negotiations []*models.FareNegotiation
	if err := cursor.All(ctx, &negotiations); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode negotiations")
		return nil, fmt.Errorf("failed to decode negotiations: %w", err)
	}

	return negotiations, nil
}

func (r *fareRepository) ExpireNegotiations() error {
	collection := r.db.Collection("fare_negotiations")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
		"status": bson.M{
			"$in": []models.FareStatus{
				models.FareStatusPending,
				models.FareStatusProposed,
				models.FareStatusCountered,
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     models.FareStatusExpired,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to expire negotiations")
		return fmt.Errorf("failed to expire negotiations: %w", err)
	}

	r.logger.Info().Int64("count", result.ModifiedCount).Msg("Expired fare negotiations")
	return nil
}

// Promotional Codes

func (r *fareRepository) GetPromoCode(code string) (*models.PromotionalRate, error) {
	collection := r.db.Collection("promotional_rates")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"code":       code,
		"is_active":  true,
		"valid_from": bson.M{"$lte": time.Now()},
		"valid_to":   bson.M{"$gte": time.Now()},
	}

	var promo models.PromotionalRate
	err := collection.FindOne(ctx, filter).Decode(&promo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("promo code not found or expired")
		}
		r.logger.Error().Err(err).Str("code", code).Msg("Failed to get promo code")
		return nil, fmt.Errorf("failed to get promo code: %w", err)
	}

	return &promo, nil
}

func (r *fareRepository) GetActivePromoCodes(userID string) ([]*models.PromotionalRate, error) {
	collection := r.db.Collection("promotional_rates")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_active":  true,
		"valid_from": bson.M{"$lte": time.Now()},
		"valid_to":   bson.M{"$gte": time.Now()},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get active promo codes")
		return nil, fmt.Errorf("failed to get active promo codes: %w", err)
	}
	defer cursor.Close(ctx)

	var promoCodes []*models.PromotionalRate
	if err := cursor.All(ctx, &promoCodes); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode promo codes")
		return nil, fmt.Errorf("failed to decode promo codes: %w", err)
	}

	return promoCodes, nil
}

func (r *fareRepository) CreatePromoCode(promo *models.PromotionalRate) (*models.PromotionalRate, error) {
	collection := r.db.Collection("promotional_rates")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, promo)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create promo code")
		return nil, fmt.Errorf("failed to create promo code: %w", err)
	}

	promo.ID = result.InsertedID.(primitive.ObjectID)
	return promo, nil
}

func (r *fareRepository) UpdatePromoCode(promo *models.PromotionalRate) (*models.PromotionalRate, error) {
	collection := r.db.Collection("promotional_rates")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": promo.ID}
	update := bson.M{"$set": promo}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("promo_id", promo.ID.Hex()).Msg("Failed to update promo code")
		return nil, fmt.Errorf("failed to update promo code: %w", err)
	}

	return promo, nil
}

func (r *fareRepository) HasUserUsedPromo(userID, promoCode string) (bool, error) {
	collection := r.db.Collection("promo_usage")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"promo_code": promoCode,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("promo_code", promoCode).Msg("Failed to check promo usage")
		return false, fmt.Errorf("failed to check promo usage: %w", err)
	}

	return count > 0, nil
}

func (r *fareRepository) RecordPromoUsage(userID, promoCode string, rideID primitive.ObjectID) error {
	collection := r.db.Collection("promo_usage")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usage := PromoUsage{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		PromoCode: promoCode,
		RideID:    rideID,
		UsedAt:    time.Now(),
	}

	_, err := collection.InsertOne(ctx, usage)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Str("promo_code", promoCode).Msg("Failed to record promo usage")
		return fmt.Errorf("failed to record promo usage: %w", err)
	}

	return nil
}

// Additional methods would continue with similar patterns...

// Placeholder implementations for remaining interface methods
func (r *fareRepository) GetFareHistoryByUser(userID string, page, limit int, filters map[string]interface{}) ([]*models.FareDetails, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *fareRepository) GetFareStatistics(userID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *fareRepository) GetMarketRates(serviceType, city, country string) (*models.MarketRate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *fareRepository) UpdateMarketRates(marketRate *models.MarketRate) error {
	return fmt.Errorf("not implemented")
}

func (r *fareRepository) GetDynamicPricing(location models.Location, serviceType string) (*models.DynamicPricing, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *fareRepository) SaveDynamicPricing(pricing *models.DynamicPricing) error {
	return fmt.Errorf("not implemented")
}

func (r *fareRepository) UpdateDynamicPricing(pricing *models.DynamicPricing) error {
	return fmt.Errorf("not implemented")
}

func (r *fareRepository) GetFareRules(serviceType, city, country string) ([]*FareRule, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *fareRepository) CreateFareRule(rule *FareRule) (*FareRule, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *fareRepository) UpdateFareRule(rule *FareRule) (*FareRule, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *fareRepository) DeleteFareRule(id primitive.ObjectID) error {
	return fmt.Errorf("not implemented")
}
