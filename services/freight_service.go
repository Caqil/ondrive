package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FreightService interface defines freight-related business logic
type FreightService interface {
	// Freight Service Management
	CreateFreightRequest(userID string, req interface{}) (*models.FreightRequest, error)
	GetFreightRequests(userID string, page, limit int, status, cargoType, vehicleType string) ([]*models.FreightRequest, int64, error)
	GetFreightRequest(userID string, requestID primitive.ObjectID) (*models.FreightRequest, error)
	UpdateFreightRequest(userID string, requestID primitive.ObjectID, req interface{}) (*models.FreightRequest, error)
	CancelFreightRequest(userID string, requestID primitive.ObjectID) error

	// Cargo Management
	SetCargoDetails(userID string, requestID primitive.ObjectID, cargo *models.FreightCargo) (*models.FreightRequest, error)
	GetCargoDetails(userID string, requestID primitive.ObjectID) (*models.FreightCargo, error)
	UpdateCargoDetails(userID string, requestID primitive.ObjectID, cargo *models.FreightCargo) (*models.FreightRequest, error)
	UploadCargoPhotos(userID string, requestID primitive.ObjectID, files []*multipart.FileHeader, photoType, description string) ([]*models.CargoPhoto, error)
	GetCargoPhotos(userID string, requestID primitive.ObjectID, photoType string) ([]*models.CargoPhoto, error)

	// Vehicle Requirements
	GetFreightVehicleTypes(cargoType, weight, volume string) ([]*VehicleTypeOption, error)
	SetVehicleRequirements(userID string, requestID primitive.ObjectID, requirements *models.FreightVehicleRequirements) (*models.FreightRequest, error)
	GetVehicleRequirements(userID string, requestID primitive.ObjectID) (*models.FreightVehicleRequirements, error)
	CheckVehicleAvailability(vehicleType, location, pickupTime, weight, volume string) (*VehicleAvailability, error)

	// Loading & Unloading
	GetLoadingOptions(cargoType, weight, dimensions string) ([]*LoadingOption, error)
	RequestLoadingAssistance(userID string, requestID primitive.ObjectID, req interface{}) (*LoadingAssistance, error)
	RequestUnloadingAssistance(userID string, requestID primitive.ObjectID, req interface{}) (*LoadingAssistance, error)
	GetEquipmentNeeded(userID string, requestID primitive.ObjectID) (*EquipmentRequirements, error)

	// Freight Tracking
	TrackFreightDelivery(userID string, requestID primitive.ObjectID) (*FreightTracking, error)
	ConfirmLoading(userID string, requestID primitive.ObjectID, req interface{}) (*LoadingConfirmation, error)
	ConfirmFreightDelivery(userID string, requestID primitive.ObjectID, req interface{}) (*DeliveryConfirmation, error)
	GetMilestoneUpdates(userID string, requestID primitive.ObjectID) ([]*MilestoneUpdate, error)
	AddMilestoneUpdate(userID string, requestID primitive.ObjectID, req interface{}) (*MilestoneUpdate, error)

	// Route & Distance
	CalculateFreightRoute(req interface{}) (*FreightRoute, error)
	OptimizeRoute(userID string, requestID primitive.ObjectID) (*OptimizedRoute, error)
	CalculateDistance(userID string, requestID primitive.ObjectID) (*DistanceCalculation, error)
	EstimateFuelCost(userID string, requestID primitive.ObjectID) (*FuelCostEstimate, error)

	// Freight Pricing
	EstimateFreightPrice(req interface{}) (*FreightPriceEstimate, error)
	GetPricingFactors(cargoType, vehicleType, serviceType string) (*PricingFactors, error)
	GetWeightBasedPricing(weight, cargoType, vehicleType string) (*WeightBasedPricing, error)
	GetDistanceBasedPricing(distance, vehicleType, serviceType string) (*DistanceBasedPricing, error)

	// Multi-Stop Delivery
	AddDeliveryStop(userID string, requestID primitive.ObjectID, req interface{}) (*models.FreightStop, error)
	GetDeliveryStops(userID string, requestID primitive.ObjectID) ([]*models.FreightStop, error)
	UpdateDeliveryStop(userID string, requestID, stopID primitive.ObjectID, req interface{}) (*models.FreightStop, error)
	RemoveDeliveryStop(userID string, requestID, stopID primitive.ObjectID) error
	OptimizeDeliveryStops(userID string, requestID primitive.ObjectID) ([]*models.FreightStop, error)

	// History & Analytics
	GetFreightHistory(userID string, page, limit int, startDate, endDate string) ([]*models.FreightRequest, int64, error)
	GetFreightAnalytics(userID string, period string) (*FreightAnalytics, error)
	GetCostBreakdown(userID, freightID, period string) (*CostBreakdown, error)

	// Special Services
	GetTemperatureControlledOptions(cargoType, temperature string) ([]*TemperatureControlledOption, error)
	GetHazardousMaterialsGuidelines(materialType, vehicleType string) (*HazardousMaterialsGuidelines, error)
	GetOversizedCargoOptions(weight, dimensions, cargoType string) ([]*OversizedCargoOption, error)
}

// Response types for freight service
type VehicleTypeOption struct {
	VehicleType  models.FreightVehicleType `json:"vehicle_type"`
	Name         string                    `json:"name"`
	Description  string                    `json:"description"`
	MaxWeight    float64                   `json:"max_weight"`
	MaxVolume    float64                   `json:"max_volume"`
	Features     []string                  `json:"features"`
	PriceRange   PriceRange                `json:"price_range"`
	Availability string                    `json:"availability"`
}

type VehicleAvailability struct {
	VehicleType     models.FreightVehicleType `json:"vehicle_type"`
	Available       bool                      `json:"available"`
	AvailableCount  int                       `json:"available_count"`
	NextAvailable   *time.Time                `json:"next_available,omitempty"`
	NearbyLocations []LocationAvailability    `json:"nearby_locations"`
	PriceEstimate   float64                   `json:"price_estimate"`
}

type LocationAvailability struct {
	Location  models.Location `json:"location"`
	Distance  float64         `json:"distance"`
	Available int             `json:"available"`
}

type LoadingOption struct {
	LoadingType       models.LoadingType `json:"loading_type"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	RequiredEquipment []string           `json:"required_equipment"`
	RequiredPersonnel int                `json:"required_personnel"`
	EstimatedTime     int                `json:"estimated_time"`
	Cost              float64            `json:"cost"`
	Suitable          bool               `json:"suitable"`
}

type LoadingAssistance struct {
	ID                primitive.ObjectID    `json:"id"`
	RequestID         primitive.ObjectID    `json:"request_id"`
	AssistanceType    string                `json:"assistance_type"`
	Status            string                `json:"status"`
	ScheduledTime     *time.Time            `json:"scheduled_time,omitempty"`
	Personnel         []AssistancePersonnel `json:"personnel"`
	Equipment         []AssistanceEquipment `json:"equipment"`
	Cost              float64               `json:"cost"`
	EstimatedDuration int                   `json:"estimated_duration"`
	SpecialNotes      string                `json:"special_notes"`
	CreatedAt         time.Time             `json:"created_at"`
}

type AssistancePersonnel struct {
	Name           string   `json:"name"`
	Role           string   `json:"role"`
	Certifications []string `json:"certifications"`
	Experience     int      `json:"experience"`
}

type AssistanceEquipment struct {
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	Capacity  float64 `json:"capacity"`
	Available bool    `json:"available"`
	Cost      float64 `json:"cost"`
}

type EquipmentRequirements struct {
	RequestID          primitive.ObjectID `json:"request_id"`
	LoadingEquipment   []EquipmentItem    `json:"loading_equipment"`
	UnloadingEquipment []EquipmentItem    `json:"unloading_equipment"`
	TransportEquipment []EquipmentItem    `json:"transport_equipment"`
	SafetyEquipment    []EquipmentItem    `json:"safety_equipment"`
	TotalCost          float64            `json:"total_cost"`
}

type EquipmentItem struct {
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Required  bool    `json:"required"`
	Cost      float64 `json:"cost"`
	Available bool    `json:"available"`
	Provider  string  `json:"provider"`
}

type FreightTracking struct {
	RequestID        primitive.ObjectID   `json:"request_id"`
	Status           models.FreightStatus `json:"status"`
	CurrentLocation  *models.Location     `json:"current_location,omitempty"`
	Progress         float64              `json:"progress"`
	EstimatedArrival *time.Time           `json:"estimated_arrival,omitempty"`
	Milestones       []*MilestoneUpdate   `json:"milestones"`
	RoutePoints      []RoutePoint         `json:"route_points"`
	CarrierInfo      *CarrierInfo         `json:"carrier_info,omitempty"`
	LastUpdate       time.Time            `json:"last_update"`
}

type MilestoneUpdate struct {
	ID            primitive.ObjectID     `json:"id"`
	RequestID     primitive.ObjectID     `json:"request_id"`
	MilestoneType string                 `json:"milestone_type"`
	Status        string                 `json:"status"`
	Location      models.Location        `json:"location"`
	Timestamp     time.Time              `json:"timestamp"`
	Notes         string                 `json:"notes"`
	Photos        []string               `json:"photos"`
	UpdatedBy     primitive.ObjectID     `json:"updated_by"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
}

type RoutePoint struct {
	Location  models.Location `json:"location"`
	Timestamp time.Time       `json:"timestamp"`
	Speed     float64         `json:"speed"`
	Heading   float64         `json:"heading"`
	Status    string          `json:"status"`
}

type CarrierInfo struct {
	CarrierID   primitive.ObjectID `json:"carrier_id"`
	Name        string             `json:"name"`
	Phone       string             `json:"phone"`
	VehicleInfo VehicleInfo        `json:"vehicle_info"`
	DriverInfo  DriverInfo         `json:"driver_info"`
}

type VehicleInfo struct {
	PlateNumber string `json:"plate_number"`
	Make        string `json:"make"`
	Model       string `json:"model"`
	Year        int    `json:"year"`
	Color       string `json:"color"`
}

type DriverInfo struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	License string  `json:"license"`
	Rating  float64 `json:"rating"`
}

type LoadingConfirmation struct {
	RequestID          primitive.ObjectID `json:"request_id"`
	LoadingStartedAt   time.Time          `json:"loading_started_at"`
	LoadingCompletedAt time.Time          `json:"loading_completed_at"`
	WeightConfirmation float64            `json:"weight_confirmation"`
	Photos             []string           `json:"photos"`
	Notes              string             `json:"notes"`
	ConfirmedBy        primitive.ObjectID `json:"confirmed_by"`
	Signature          string             `json:"signature"`
	CreatedAt          time.Time          `json:"created_at"`
}

type DeliveryConfirmation struct {
	RequestID            primitive.ObjectID `json:"request_id"`
	UnloadingStartedAt   time.Time          `json:"unloading_started_at"`
	UnloadingCompletedAt time.Time          `json:"unloading_completed_at"`
	DeliveredAt          time.Time          `json:"delivered_at"`
	RecipientName        string             `json:"recipient_name"`
	RecipientSignature   string             `json:"recipient_signature"`
	Photos               []string           `json:"photos"`
	Notes                string             `json:"notes"`
	Condition            string             `json:"condition"`
	ConfirmedBy          primitive.ObjectID `json:"confirmed_by"`
	CreatedAt            time.Time          `json:"created_at"`
}

type FreightRoute struct {
	PickupLocation          models.FreightLocation `json:"pickup_location"`
	DeliveryLocation        models.FreightLocation `json:"delivery_location"`
	Stops                   []*models.FreightStop  `json:"stops"`
	TotalDistance           float64                `json:"total_distance"`
	TotalDuration           int                    `json:"total_duration"`
	RoutePoints             []models.Location      `json:"route_points"`
	TollCosts               float64                `json:"toll_costs"`
	FuelCost                float64                `json:"fuel_cost"`
	EstimatedCost           float64                `json:"estimated_cost"`
	OptimizationSuggestions []string               `json:"optimization_suggestions"`
}

type OptimizedRoute struct {
	OriginalRoute      FreightRoute          `json:"original_route"`
	OptimizedRoute     FreightRoute          `json:"optimized_route"`
	Savings            RouteSavings          `json:"savings"`
	OptimizedStops     []*models.FreightStop `json:"optimized_stops"`
	RecommendedChanges []string              `json:"recommended_changes"`
}

type RouteSavings struct {
	DistanceSaved float64 `json:"distance_saved"`
	TimeSaved     int     `json:"time_saved"`
	CostSaved     float64 `json:"cost_saved"`
	FuelSaved     float64 `json:"fuel_saved"`
}

type DistanceCalculation struct {
	RequestID         primitive.ObjectID `json:"request_id"`
	TotalDistance     float64            `json:"total_distance"`
	DistanceBySegment []DistanceSegment  `json:"distance_by_segment"`
	EstimatedDuration int                `json:"estimated_duration"`
	RouteType         string             `json:"route_type"`
	CalculatedAt      time.Time          `json:"calculated_at"`
}

type DistanceSegment struct {
	From        models.Location `json:"from"`
	To          models.Location `json:"to"`
	Distance    float64         `json:"distance"`
	Duration    int             `json:"duration"`
	SegmentType string          `json:"segment_type"`
}

type FuelCostEstimate struct {
	RequestID       primitive.ObjectID `json:"request_id"`
	TotalDistance   float64            `json:"total_distance"`
	VehicleType     string             `json:"vehicle_type"`
	FuelConsumption float64            `json:"fuel_consumption"`
	FuelPrice       float64            `json:"fuel_price"`
	TotalFuelCost   float64            `json:"total_fuel_cost"`
	FuelCostPerKm   float64            `json:"fuel_cost_per_km"`
	EstimatedLiters float64            `json:"estimated_liters"`
	Currency        string             `json:"currency"`
	CalculatedAt    time.Time          `json:"calculated_at"`
}

type FreightPriceEstimate struct {
	RequestID          primitive.ObjectID   `json:"request_id"`
	BaseFare           float64              `json:"base_fare"`
	WeightCost         float64              `json:"weight_cost"`
	DistanceCost       float64              `json:"distance_cost"`
	VolumeCost         float64              `json:"volume_cost"`
	LoadingCost        float64              `json:"loading_cost"`
	UnloadingCost      float64              `json:"unloading_cost"`
	InsuranceCost      float64              `json:"insurance_cost"`
	SpecialServiceCost float64              `json:"special_service_cost"`
	TaxAmount          float64              `json:"tax_amount"`
	TotalCost          float64              `json:"total_cost"`
	Currency           string               `json:"currency"`
	ValidUntil         time.Time            `json:"valid_until"`
	PriceBreakdown     []PriceBreakdownItem `json:"price_breakdown"`
	PricingFactors     PricingFactors       `json:"pricing_factors"`
}

type PriceBreakdownItem struct {
	Item        string  `json:"item"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Rate        float64 `json:"rate"`
	Amount      float64 `json:"amount"`
}

type PricingFactors struct {
	CargoType     string          `json:"cargo_type"`
	VehicleType   string          `json:"vehicle_type"`
	ServiceType   string          `json:"service_type"`
	WeightRate    float64         `json:"weight_rate"`
	VolumeRate    float64         `json:"volume_rate"`
	DistanceRate  float64         `json:"distance_rate"`
	BaseRate      float64         `json:"base_rate"`
	LoadingRate   float64         `json:"loading_rate"`
	UnloadingRate float64         `json:"unloading_rate"`
	InsuranceRate float64         `json:"insurance_rate"`
	TaxRate       float64         `json:"tax_rate"`
	Surcharges    []SurchargeItem `json:"surcharges"`
	Discounts     []DiscountItem  `json:"discounts"`
	LastUpdated   time.Time       `json:"last_updated"`
}

type SurchargeItem struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Rate        float64 `json:"rate"`
	RateType    string  `json:"rate_type"` // percentage, fixed
	Applicable  bool    `json:"applicable"`
}

type DiscountItem struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Rate        float64 `json:"rate"`
	RateType    string  `json:"rate_type"` // percentage, fixed
	Applicable  bool    `json:"applicable"`
}

type WeightBasedPricing struct {
	WeightRange   WeightRange     `json:"weight_range"`
	Rate          float64         `json:"rate"`
	RateType      string          `json:"rate_type"`
	MinimumCharge float64         `json:"minimum_charge"`
	Surcharges    []SurchargeItem `json:"surcharges"`
	EstimatedCost float64         `json:"estimated_cost"`
}

type WeightRange struct {
	MinWeight float64 `json:"min_weight"`
	MaxWeight float64 `json:"max_weight"`
	Unit      string  `json:"unit"`
}

type DistanceBasedPricing struct {
	DistanceRange DistanceRange   `json:"distance_range"`
	Rate          float64         `json:"rate"`
	RateType      string          `json:"rate_type"`
	MinimumCharge float64         `json:"minimum_charge"`
	Surcharges    []SurchargeItem `json:"surcharges"`
	EstimatedCost float64         `json:"estimated_cost"`
}

type DistanceRange struct {
	MinDistance float64 `json:"min_distance"`
	MaxDistance float64 `json:"max_distance"`
	Unit        string  `json:"unit"`
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

type TemperatureControlledOption struct {
	OptionType       string                  `json:"option_type"`
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	TemperatureRange models.TemperatureRange `json:"temperature_range"`
	VehicleTypes     []string                `json:"vehicle_types"`
	Features         []string                `json:"features"`
	AdditionalCost   float64                 `json:"additional_cost"`
	Available        bool                    `json:"available"`
}

type HazardousMaterialsGuidelines struct {
	MaterialType         string              `json:"material_type"`
	Classifications      []string            `json:"classifications"`
	Requirements         []SafetyRequirement `json:"requirements"`
	Restrictions         []string            `json:"restrictions"`
	Documentation        []string            `json:"documentation"`
	VehicleRequirements  VehicleRequirement  `json:"vehicle_requirements"`
	HandlingInstructions []string            `json:"handling_instructions"`
	EmergencyProcedures  []string            `json:"emergency_procedures"`
}

type SafetyRequirement struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Mandatory   bool   `json:"mandatory"`
}

type VehicleRequirement struct {
	VehicleType        string   `json:"vehicle_type"`
	SpecialEquipment   []string `json:"special_equipment"`
	Certifications     []string `json:"certifications"`
	AdditionalFeatures []string `json:"additional_features"`
}

type OversizedCargoOption struct {
	OptionType       string              `json:"option_type"`
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	MaxDimensions    CargoDimensions     `json:"max_dimensions"`
	MaxWeight        float64             `json:"max_weight"`
	VehicleTypes     []string            `json:"vehicle_types"`
	SpecialEquipment []string            `json:"special_equipment"`
	Permits          []PermitRequirement `json:"permits"`
	AdditionalCost   float64             `json:"additional_cost"`
	LeadTime         int                 `json:"lead_time"` // days
	Available        bool                `json:"available"`
}

type CargoDimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"`
}

type PermitRequirement struct {
	Type           string  `json:"type"`
	Description    string  `json:"description"`
	Cost           float64 `json:"cost"`
	ProcessingTime int     `json:"processing_time"` // days
	Required       bool    `json:"required"`
}

type PriceRange struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Currency string  `json:"currency"`
}

// freightService implements FreightService interface
type freightService struct {
	freightRepo     repositories.FreightRepository
	userRepo        repositories.UserRepository
	locationService LocationService
	uploadService   UploadService
	logger          utils.Logger
}

// NewFreightService creates a new freight service
func NewFreightService(
	freightRepo repositories.FreightRepository,
	userRepo repositories.UserRepository,
	locationService LocationService,
	uploadService UploadService,
) FreightService {
	return &freightService{
		freightRepo:     freightRepo,
		userRepo:        userRepo,
		locationService: locationService,
		uploadService:   uploadService,
		logger:          utils.ServiceLogger("freight"),
	}
}

// Freight Service Management

func (s *freightService) CreateFreightRequest(userID string, reqInterface interface{}) (*models.FreightRequest, error) {
	// Type assertion to get the actual request structure
	req, ok := reqInterface.(*CreateFreightRequestReq)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate user exists
	_, err = s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create freight request
	freightRequest := &models.FreightRequest{
		ID:                    primitive.NewObjectID(),
		Status:                models.FreightStatusPending,
		Type:                  req.Type,
		Priority:              req.Priority,
		ShipperID:             userObjID,
		Cargo:                 req.Cargo,
		VehicleRequirements:   req.VehicleRequirements,
		PickupLocation:        req.PickupLocation,
		DeliveryLocation:      req.DeliveryLocation,
		FlexibleScheduling:    req.FlexibleScheduling,
		LoadingRequirements:   req.LoadingRequirements,
		UnloadingRequirements: req.UnloadingRequirements,
		PaymentTerms:          req.PaymentTerms,
		SpecialInstructions:   req.SpecialInstructions,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Parse scheduled times if provided
	if req.ScheduledPickupTime != nil {
		if pickupTime, err := time.Parse(time.RFC3339, *req.ScheduledPickupTime); err == nil {
			freightRequest.ScheduledPickupTime = &pickupTime
		}
	}

	if req.ScheduledDeliveryTime != nil {
		if deliveryTime, err := time.Parse(time.RFC3339, *req.ScheduledDeliveryTime); err == nil {
			freightRequest.ScheduledDeliveryTime = &deliveryTime
		}
	}

	// Calculate initial fare estimate
	fareEstimate, err := s.calculateInitialFare(freightRequest)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to calculate initial fare")
	} else {
		freightRequest.Fare = *fareEstimate
	}

	// Save to repository
	createdRequest, err := s.freightRepo.Create(freightRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create freight request: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID).
		Str("request_id", createdRequest.ID.Hex()).
		Str("cargo_type", string(createdRequest.Cargo.Type)).
		Msg("Freight request created successfully")

	return createdRequest, nil
}

func (s *freightService) GetFreightRequests(userID string, page, limit int, status, cargoType, vehicleType string) ([]*models.FreightRequest, int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	// Build filter
	filter := map[string]interface{}{
		"shipper_id": userObjID,
	}

	if status != "" {
		filter["status"] = status
	}
	if cargoType != "" {
		filter["cargo.type"] = cargoType
	}
	if vehicleType != "" {
		filter["vehicle_requirements.vehicle_type"] = vehicleType
	}

	return s.freightRepo.GetByFilter(filter, page, limit)
}

func (s *freightService) GetFreightRequest(userID string, requestID primitive.ObjectID) (*models.FreightRequest, error) {
	request, err := s.freightRepo.GetByID(requestID)
	if err != nil {
		return nil, fmt.Errorf("freight request not found: %w", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if user has access to this request
	if request.ShipperID != userObjID && (request.CarrierID == nil || *request.CarrierID != userObjID) {
		return nil, errors.New("unauthorized access to freight request")
	}

	return request, nil
}

func (s *freightService) UpdateFreightRequest(userID string, requestID primitive.ObjectID, reqInterface interface{}) (*models.FreightRequest, error) {
	// Get existing request
	request, err := s.GetFreightRequest(userID, requestID)
	if err != nil {
		return nil, err
	}

	// Check if request can be updated
	if request.Status != models.FreightStatusPending && request.Status != models.FreightStatusQuoted {
		return nil, errors.New("cannot update freight request in current status")
	}

	// Type assertion
	req, ok := reqInterface.(*UpdateFreightRequestReq)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	// Update fields
	if req.Priority != "" {
		request.Priority = req.Priority
	}
	if req.ScheduledPickupTime != nil {
		if pickupTime, err := time.Parse(time.RFC3339, *req.ScheduledPickupTime); err == nil {
			request.ScheduledPickupTime = &pickupTime
		}
	}
	if req.ScheduledDeliveryTime != nil {
		if deliveryTime, err := time.Parse(time.RFC3339, *req.ScheduledDeliveryTime); err == nil {
			request.ScheduledDeliveryTime = &deliveryTime
		}
	}
	if req.FlexibleScheduling != nil {
		request.FlexibleScheduling = *req.FlexibleScheduling
	}
	if req.LoadingRequirements != nil {
		request.LoadingRequirements = *req.LoadingRequirements
	}
	if req.UnloadingRequirements != nil {
		request.UnloadingRequirements = *req.UnloadingRequirements
	}
	if req.PaymentTerms != "" {
		request.PaymentTerms = req.PaymentTerms
	}
	if req.SpecialInstructions != "" {
		request.SpecialInstructions = req.SpecialInstructions
	}

	request.UpdatedAt = time.Now()

	// Update in repository
	updatedRequest, err := s.freightRepo.Update(request)
	if err != nil {
		return nil, fmt.Errorf("failed to update freight request: %w", err)
	}

	return updatedRequest, nil
}

func (s *freightService) CancelFreightRequest(userID string, requestID primitive.ObjectID) error {
	// Get existing request
	request, err := s.GetFreightRequest(userID, requestID)
	if err != nil {
		return err
	}

	// Check if request can be cancelled
	if request.Status == models.FreightStatusDelivered || request.Status == models.FreightStatusCancelled {
		return errors.New("cannot cancel freight request in current status")
	}

	// Update status
	request.Status = models.FreightStatusCancelled
	request.UpdatedAt = time.Now()

	// Save changes
	_, err = s.freightRepo.Update(request)
	if err != nil {
		return fmt.Errorf("failed to cancel freight request: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID).
		Str("request_id", requestID.Hex()).
		Msg("Freight request cancelled successfully")

	return nil
}

// Cargo Management

func (s *freightService) SetCargoDetails(userID string, requestID primitive.ObjectID, cargo *models.FreightCargo) (*models.FreightRequest, error) {
	request, err := s.GetFreightRequest(userID, requestID)
	if err != nil {
		return nil, err
	}

	// Update cargo details
	request.Cargo = *cargo
	request.UpdatedAt = time.Now()

	// Recalculate fare if needed
	if fareEstimate, err := s.calculateInitialFare(request); err == nil {
		request.Fare = *fareEstimate
	}

	// Save changes
	updatedRequest, err := s.freightRepo.Update(request)
	if err != nil {
		return nil, fmt.Errorf("failed to update cargo details: %w", err)
	}

	return updatedRequest, nil
}

func (s *freightService) GetCargoDetails(userID string, requestID primitive.ObjectID) (*models.FreightCargo, error) {
	request, err := s.GetFreightRequest(userID, requestID)
	if err != nil {
		return nil, err
	}

	return &request.Cargo, nil
}

func (s *freightService) UpdateCargoDetails(userID string, requestID primitive.ObjectID, cargo *models.FreightCargo) (*models.FreightRequest, error) {
	return s.SetCargoDetails(userID, requestID, cargo)
}

func (s *freightService) UploadCargoPhotos(userID string, requestID primitive.ObjectID, files []*multipart.FileHeader, photoType, description string) ([]*models.CargoPhoto, error) {
	// Verify user has access to the request
	_, err := s.GetFreightRequest(userID, requestID)
	if err != nil {
		return nil, err
	}

	var photos []*models.CargoPhoto

	for _, file := range files {
		// Upload file
		uploadResult, err := s.uploadService.UploadFile(file, fmt.Sprintf("freight/%s/cargo", requestID.Hex()))
		if err != nil {
			s.logger.Error().Err(err).Str("filename", file.Filename).Msg("Failed to upload cargo photo")
			continue
		}

		userObjID, _ := primitive.ObjectIDFromHex(userID)

		photo := &models.CargoPhoto{
			ID:          primitive.NewObjectID(),
			URL:         uploadResult.URL,
			Type:        photoType,
			Description: description,
			TakenAt:     time.Now(),
			TakenBy:     userObjID,
		}

		photos = append(photos, photo)
	}

	// Save photos to request
	err = s.freightRepo.AddCargoPhotos(requestID, photos)
	if err != nil {
		return nil, fmt.Errorf("failed to save cargo photos: %w", err)
	}

	return photos, nil
}

func (s *freightService) GetCargoPhotos(userID string, requestID primitive.ObjectID, photoType string) ([]*models.CargoPhoto, error) {
	// Verify user has access to the request
	_, err := s.GetFreightRequest(userID, requestID)
	if err != nil {
		return nil, err
	}

	return s.freightRepo.GetCargoPhotos(requestID, photoType)
}

// Helper method to calculate initial fare
func (s *freightService) calculateInitialFare(request *models.FreightRequest) (*models.FreightFare, error) {
	// Calculate distance
	distance, _, err := s.locationService.CalculateRoute(
		models.Location{
			Latitude:  request.PickupLocation.Coordinates[1],
			Longitude: request.PickupLocation.Coordinates[0],
		},
		models.Location{
			Latitude:  request.DeliveryLocation.Coordinates[1],
			Longitude: request.DeliveryLocation.Coordinates[0],
		},
	)
	if err != nil {
		return nil, err
	}

	// Calculate base fare components
	baseFare := 50.0    // Base freight fee
	weightRate := 2.0   // Per kg
	distanceRate := 1.5 // Per km
	volumeRate := 10.0  // Per m³

	weightCost := request.Cargo.TotalWeight * weightRate
	distanceCost := distance * distanceRate
	volumeCost := request.Cargo.TotalVolume * volumeRate

	// Loading/unloading fees
	loadingFee := 20.0
	unloadingFee := 20.0
	if request.LoadingRequirements.LoadingMethod == models.LoadingTypeCrane {
		loadingFee = 100.0
	}

	// Insurance
	insuranceFee := 0.0
	if request.Cargo.TotalValue > 0 {
		insuranceFee = request.Cargo.TotalValue * 0.01 // 1% of cargo value
	}

	// Calculate tax
	subtotal := baseFare + weightCost + distanceCost + volumeCost + loadingFee + unloadingFee + insuranceFee
	taxAmount := subtotal * 0.1 // 10% tax

	totalAmount := subtotal + taxAmount

	return &models.FreightFare{
		BaseFare:     baseFare,
		WeightRate:   weightRate,
		DistanceRate: distanceRate,
		VolumeRate:   volumeRate,
		LoadingFee:   loadingFee,
		UnloadingFee: unloadingFee,
		InsuranceFee: insuranceFee,
		TaxAmount:    taxAmount,
		SubTotal:     subtotal,
		TotalAmount:  totalAmount,
		Currency:     "USD",
	}, nil
}

// Vehicle Requirements methods (simplified implementations)
func (s *freightService) GetFreightVehicleTypes(cargoType, weight, volume string) ([]*VehicleTypeOption, error) {
	// Implementation would return vehicle types based on cargo requirements
	return []*VehicleTypeOption{
		{
			VehicleType:  models.FreightVehicleSmallTruck,
			Name:         "Small Truck",
			Description:  "1-3 tons capacity",
			MaxWeight:    3000,
			MaxVolume:    15,
			Features:     []string{"hydraulic tailgate", "GPS tracking"},
			PriceRange:   PriceRange{Min: 50, Max: 150, Currency: "USD"},
			Availability: "available",
		},
	}, nil
}

func (s *freightService) SetVehicleRequirements(userID string, requestID primitive.ObjectID, requirements *models.FreightVehicleRequirements) (*models.FreightRequest, error) {
	request, err := s.GetFreightRequest(userID, requestID)
	if err != nil {
		return nil, err
	}

	request.VehicleRequirements = *requirements
	request.UpdatedAt = time.Now()

	return s.freightRepo.Update(request)
}

func (s *freightService) GetVehicleRequirements(userID string, requestID primitive.ObjectID) (*models.FreightVehicleRequirements, error) {
	request, err := s.GetFreightRequest(userID, requestID)
	if err != nil {
		return nil, err
	}

	return &request.VehicleRequirements, nil
}

func (s *freightService) CheckVehicleAvailability(vehicleType, location, pickupTime, weight, volume string) (*VehicleAvailability, error) {
	// Simplified implementation
	return &VehicleAvailability{
		VehicleType:     models.FreightVehicleType(vehicleType),
		Available:       true,
		AvailableCount:  5,
		PriceEstimate:   100.0,
		NearbyLocations: []LocationAvailability{},
	}, nil
}

// Placeholder implementations for remaining methods...
// (These would follow similar patterns to the implemented methods above)

func (s *freightService) GetLoadingOptions(cargoType, weight, dimensions string) ([]*LoadingOption, error) {
	return []*LoadingOption{}, nil
}

func (s *freightService) RequestLoadingAssistance(userID string, requestID primitive.ObjectID, req interface{}) (*LoadingAssistance, error) {
	return &LoadingAssistance{}, nil
}

func (s *freightService) RequestUnloadingAssistance(userID string, requestID primitive.ObjectID, req interface{}) (*LoadingAssistance, error) {
	return &LoadingAssistance{}, nil
}

func (s *freightService) GetEquipmentNeeded(userID string, requestID primitive.ObjectID) (*EquipmentRequirements, error) {
	return &EquipmentRequirements{}, nil
}

func (s *freightService) TrackFreightDelivery(userID string, requestID primitive.ObjectID) (*FreightTracking, error) {
	return &FreightTracking{}, nil
}

func (s *freightService) ConfirmLoading(userID string, requestID primitive.ObjectID, req interface{}) (*LoadingConfirmation, error) {
	return &LoadingConfirmation{}, nil
}

func (s *freightService) ConfirmFreightDelivery(userID string, requestID primitive.ObjectID, req interface{}) (*DeliveryConfirmation, error) {
	return &DeliveryConfirmation{}, nil
}

func (s *freightService) GetMilestoneUpdates(userID string, requestID primitive.ObjectID) ([]*MilestoneUpdate, error) {
	return []*MilestoneUpdate{}, nil
}

func (s *freightService) AddMilestoneUpdate(userID string, requestID primitive.ObjectID, req interface{}) (*MilestoneUpdate, error) {
	return &MilestoneUpdate{}, nil
}

func (s *freightService) CalculateFreightRoute(req interface{}) (*FreightRoute, error) {
	return &FreightRoute{}, nil
}

func (s *freightService) OptimizeRoute(userID string, requestID primitive.ObjectID) (*OptimizedRoute, error) {
	return &OptimizedRoute{}, nil
}

func (s *freightService) CalculateDistance(userID string, requestID primitive.ObjectID) (*DistanceCalculation, error) {
	return &DistanceCalculation{}, nil
}

func (s *freightService) EstimateFuelCost(userID string, requestID primitive.ObjectID) (*FuelCostEstimate, error) {
	return &FuelCostEstimate{}, nil
}

func (s *freightService) EstimateFreightPrice(req interface{}) (*FreightPriceEstimate, error) {
	return &FreightPriceEstimate{}, nil
}

func (s *freightService) GetPricingFactors(cargoType, vehicleType, serviceType string) (*PricingFactors, error) {
	return &PricingFactors{}, nil
}

func (s *freightService) GetWeightBasedPricing(weight, cargoType, vehicleType string) (*WeightBasedPricing, error) {
	return &WeightBasedPricing{}, nil
}

func (s *freightService) GetDistanceBasedPricing(distance, vehicleType, serviceType string) (*DistanceBasedPricing, error) {
	return &DistanceBasedPricing{}, nil
}

func (s *freightService) AddDeliveryStop(userID string, requestID primitive.ObjectID, req interface{}) (*models.FreightStop, error) {
	return &models.FreightStop{}, nil
}

func (s *freightService) GetDeliveryStops(userID string, requestID primitive.ObjectID) ([]*models.FreightStop, error) {
	return []*models.FreightStop{}, nil
}

func (s *freightService) UpdateDeliveryStop(userID string, requestID, stopID primitive.ObjectID, req interface{}) (*models.FreightStop, error) {
	return &models.FreightStop{}, nil
}

func (s *freightService) RemoveDeliveryStop(userID string, requestID, stopID primitive.ObjectID) error {
	return nil
}

func (s *freightService) OptimizeDeliveryStops(userID string, requestID primitive.ObjectID) ([]*models.FreightStop, error) {
	return []*models.FreightStop{}, nil
}

func (s *freightService) GetFreightHistory(userID string, page, limit int, startDate, endDate string) ([]*models.FreightRequest, int64, error) {
	return []*models.FreightRequest{}, 0, nil
}

func (s *freightService) GetFreightAnalytics(userID string, period string) (*FreightAnalytics, error) {
	return &FreightAnalytics{}, nil
}

func (s *freightService) GetCostBreakdown(userID, freightID, period string) (*CostBreakdown, error) {
	return &CostBreakdown{}, nil
}

func (s *freightService) GetTemperatureControlledOptions(cargoType, temperature string) ([]*TemperatureControlledOption, error) {
	return []*TemperatureControlledOption{}, nil
}

func (s *freightService) GetHazardousMaterialsGuidelines(materialType, vehicleType string) (*HazardousMaterialsGuidelines, error) {
	return &HazardousMaterialsGuidelines{}, nil
}

func (s *freightService) GetOversizedCargoOptions(weight, dimensions, cargoType string) ([]*OversizedCargoOption, error) {
	return []*OversizedCargoOption{}, nil
}

// Additional request types that would be imported from controller
type CreateFreightRequestReq struct {
	Type                  string                            `json:"type"`
	Priority              string                            `json:"priority"`
	Cargo                 models.FreightCargo               `json:"cargo"`
	VehicleRequirements   models.FreightVehicleRequirements `json:"vehicle_requirements"`
	PickupLocation        models.FreightLocation            `json:"pickup_location"`
	DeliveryLocation      models.FreightLocation            `json:"delivery_location"`
	ScheduledPickupTime   *string                           `json:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *string                           `json:"scheduled_delivery_time,omitempty"`
	FlexibleScheduling    bool                              `json:"flexible_scheduling"`
	LoadingRequirements   models.LoadingRequirements        `json:"loading_requirements"`
	UnloadingRequirements models.LoadingRequirements        `json:"unloading_requirements"`
	PaymentTerms          string                            `json:"payment_terms"`
	SpecialInstructions   string                            `json:"special_instructions"`
	RequiresInsurance     bool                              `json:"requires_insurance"`
	InsuranceValue        float64                           `json:"insurance_value"`
}

type UpdateFreightRequestReq struct {
	Priority              string                      `json:"priority,omitempty"`
	ScheduledPickupTime   *string                     `json:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *string                     `json:"scheduled_delivery_time,omitempty"`
	FlexibleScheduling    *bool                       `json:"flexible_scheduling,omitempty"`
	LoadingRequirements   *models.LoadingRequirements `json:"loading_requirements,omitempty"`
	UnloadingRequirements *models.LoadingRequirements `json:"unloading_requirements,omitempty"`
	PaymentTerms          string                      `json:"payment_terms,omitempty"`
	SpecialInstructions   string                      `json:"special_instructions,omitempty"`
}
