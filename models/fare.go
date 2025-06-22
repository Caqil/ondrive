package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FareCalculationType string

const (
	FareTypeDistance FareCalculationType = "distance"
	FareTypeTime     FareCalculationType = "time"
	FareTypeFlat     FareCalculationType = "flat"
	FareTypeCustom   FareCalculationType = "custom"
)

type FareStatus string

const (
	FareStatusPending   FareStatus = "pending"
	FareStatusProposed  FareStatus = "proposed"
	FareStatusCountered FareStatus = "countered"
	FareStatusAccepted  FareStatus = "accepted"
	FareStatusRejected  FareStatus = "rejected"
	FareStatusExpired   FareStatus = "expired"
)

type FareDetails struct {
	// Basic Fare Information
	ProposedFare float64    `json:"proposed_fare" bson:"proposed_fare"`
	FinalFare    float64    `json:"final_fare" bson:"final_fare"`
	Currency     string     `json:"currency" bson:"currency"`
	Status       FareStatus `json:"status" bson:"status"`

	// Fare Breakdown
	BaseFare        float64 `json:"base_fare" bson:"base_fare"`
	DistanceFare    float64 `json:"distance_fare" bson:"distance_fare"`
	TimeFare        float64 `json:"time_fare" bson:"time_fare"`
	SurgeFare       float64 `json:"surge_fare" bson:"surge_fare"`
	PeakHourFare    float64 `json:"peak_hour_fare" bson:"peak_hour_fare"`
	ServiceFee      float64 `json:"service_fee" bson:"service_fee"`
	TollsFee        float64 `json:"tolls_fee" bson:"tolls_fee"`
	WaitingFee      float64 `json:"waiting_fee" bson:"waiting_fee"`
	CancellationFee float64 `json:"cancellation_fee" bson:"cancellation_fee"`

	// Discounts & Tips
	DiscountAmount float64 `json:"discount_amount" bson:"discount_amount"`
	PromoCode      string  `json:"promo_code" bson:"promo_code"`
	TipAmount      float64 `json:"tip_amount" bson:"tip_amount"`
	CreditApplied  float64 `json:"credit_applied" bson:"credit_applied"`

	// Commission & Fees
	Commission     float64 `json:"commission" bson:"commission"`
	CommissionRate float64 `json:"commission_rate" bson:"commission_rate"`
	DriverEarnings float64 `json:"driver_earnings" bson:"driver_earnings"`
	PlatformFee    float64 `json:"platform_fee" bson:"platform_fee"`

	// Negotiation History
	Negotiations     []FareOffer `json:"negotiations" bson:"negotiations"`
	NegotiationCount int         `json:"negotiation_count" bson:"negotiation_count"`

	// Calculation Details
	CalculationType FareCalculationType `json:"calculation_type" bson:"calculation_type"`
	CalculatedAt    time.Time           `json:"calculated_at" bson:"calculated_at"`
	RateCard        RateCard            `json:"rate_card" bson:"rate_card"`

	// Market Information
	MarketRate    MarketRate     `json:"market_rate" bson:"market_rate"`
	SuggestedFare float64        `json:"suggested_fare" bson:"suggested_fare"`
	PriceRange    FarePriceRange `json:"price_range" bson:"price_range"`

	// Special Pricing
	DynamicPricing  DynamicPricing   `json:"dynamic_pricing" bson:"dynamic_pricing"`
	PromotionalRate *PromotionalRate `json:"promotional_rate,omitempty" bson:"promotional_rate,omitempty"`
}

type FareOffer struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OfferBy     primitive.ObjectID `json:"offer_by" bson:"offer_by"`
	OfferTo     primitive.ObjectID `json:"offer_to" bson:"offer_to"`
	Amount      float64            `json:"amount" bson:"amount"`
	Message     string             `json:"message" bson:"message"`
	OfferedAt   time.Time          `json:"offered_at" bson:"offered_at"`
	ExpiresAt   time.Time          `json:"expires_at" bson:"expires_at"`
	IsAccepted  bool               `json:"is_accepted" bson:"is_accepted"`
	IsRejected  bool               `json:"is_rejected" bson:"is_rejected"`
	IsCountered bool               `json:"is_countered" bson:"is_countered"`
	IsExpired   bool               `json:"is_expired" bson:"is_expired"`
	Response    string             `json:"response" bson:"response"`
	ResponseAt  *time.Time         `json:"response_at,omitempty" bson:"response_at,omitempty"`
}

type RateCard struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	ServiceType ServiceType        `json:"service_type" bson:"service_type"`
	VehicleType VehicleType        `json:"vehicle_type" bson:"vehicle_type"`
	City        string             `json:"city" bson:"city"`
	Country     string             `json:"country" bson:"country"`

	// Base Rates
	BaseFare    float64 `json:"base_fare" bson:"base_fare"`
	MinimumFare float64 `json:"minimum_fare" bson:"minimum_fare"`
	MaximumFare float64 `json:"maximum_fare" bson:"maximum_fare"`

	// Distance & Time Rates
	PerKmRate       float64 `json:"per_km_rate" bson:"per_km_rate"`
	PerMinuteRate   float64 `json:"per_minute_rate" bson:"per_minute_rate"`
	FreeWaitingTime int     `json:"free_waiting_time" bson:"free_waiting_time"` // minutes
	WaitingTimeRate float64 `json:"waiting_time_rate" bson:"waiting_time_rate"`

	// Special Rates
	NightSurcharge   float64 `json:"night_surcharge" bson:"night_surcharge"`
	WeekendSurcharge float64 `json:"weekend_surcharge" bson:"weekend_surcharge"`
	HolidaySurcharge float64 `json:"holiday_surcharge" bson:"holiday_surcharge"`
	AirportSurcharge float64 `json:"airport_surcharge" bson:"airport_surcharge"`

	// Commission & Fees
	CommissionRate  float64 `json:"commission_rate" bson:"commission_rate"`
	ServiceFeeRate  float64 `json:"service_fee_rate" bson:"service_fee_rate"`
	CancellationFee float64 `json:"cancellation_fee" bson:"cancellation_fee"`

	// Time-based Rates
	PeakHours []PeakHour `json:"peak_hours" bson:"peak_hours"`

	// Validity
	ValidFrom time.Time `json:"valid_from" bson:"valid_from"`
	ValidTo   time.Time `json:"valid_to" bson:"valid_to"`
	IsActive  bool      `json:"is_active" bson:"is_active"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type PeakHour struct {
	DayOfWeek   int     `json:"day_of_week" bson:"day_of_week"` // 0=Sunday, 1=Monday, etc.
	StartTime   string  `json:"start_time" bson:"start_time"`   // "08:00"
	EndTime     string  `json:"end_time" bson:"end_time"`       // "10:00"
	Multiplier  float64 `json:"multiplier" bson:"multiplier"`   // 1.5 for 50% increase
	Description string  `json:"description" bson:"description"`
}

type MarketRate struct {
	AverageRate    float64   `json:"average_rate" bson:"average_rate"`
	MinRate        float64   `json:"min_rate" bson:"min_rate"`
	MaxRate        float64   `json:"max_rate" bson:"max_rate"`
	SampleSize     int       `json:"sample_size" bson:"sample_size"`
	LastUpdated    time.Time `json:"last_updated" bson:"last_updated"`
	Confidence     float64   `json:"confidence" bson:"confidence"`
	TrendDirection string    `json:"trend_direction" bson:"trend_direction"` // up, down, stable
}

type FarePriceRange struct {
	MinPrice    float64 `json:"min_price" bson:"min_price"`
	MaxPrice    float64 `json:"max_price" bson:"max_price"`
	Recommended float64 `json:"recommended" bson:"recommended"`
	Confidence  float64 `json:"confidence" bson:"confidence"`
}

type DynamicPricing struct {
	IsActive      bool      `json:"is_active" bson:"is_active"`
	Multiplier    float64   `json:"multiplier" bson:"multiplier"`
	Reason        string    `json:"reason" bson:"reason"`
	DemandLevel   string    `json:"demand_level" bson:"demand_level"` // low, medium, high, very_high
	SupplyLevel   string    `json:"supply_level" bson:"supply_level"` // low, medium, high
	WeatherImpact float64   `json:"weather_impact" bson:"weather_impact"`
	EventImpact   float64   `json:"event_impact" bson:"event_impact"`
	TrafficImpact float64   `json:"traffic_impact" bson:"traffic_impact"`
	CalculatedAt  time.Time `json:"calculated_at" bson:"calculated_at"`
	ExpiresAt     time.Time `json:"expires_at" bson:"expires_at"`
}

type PromotionalRate struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name            string             `json:"name" bson:"name"`
	Code            string             `json:"code" bson:"code"`
	DiscountType    string             `json:"discount_type" bson:"discount_type"` // percentage, fixed, free_ride
	DiscountValue   float64            `json:"discount_value" bson:"discount_value"`
	MaxDiscount     float64            `json:"max_discount" bson:"max_discount"`
	MinFareAmount   float64            `json:"min_fare_amount" bson:"min_fare_amount"`
	UsageLimit      int                `json:"usage_limit" bson:"usage_limit"`
	UsageCount      int                `json:"usage_count" bson:"usage_count"`
	ValidFrom       time.Time          `json:"valid_from" bson:"valid_from"`
	ValidTo         time.Time          `json:"valid_to" bson:"valid_to"`
	ApplicableAreas []string           `json:"applicable_areas" bson:"applicable_areas"`
	FirstRideOnly   bool               `json:"first_ride_only" bson:"first_ride_only"`
	IsActive        bool               `json:"is_active" bson:"is_active"`
}

// FareEstimate represents a fare estimate for a potential ride
type FareEstimate struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ServiceType ServiceType        `json:"service_type" bson:"service_type"`
	VehicleType VehicleType        `json:"vehicle_type" bson:"vehicle_type"`

	// Route Information
	Distance float64 `json:"distance" bson:"distance"` // in kilometers
	Duration int     `json:"duration" bson:"duration"` // in minutes

	// Fare Breakdown
	BaseFare     float64 `json:"base_fare" bson:"base_fare"`
	DistanceFare float64 `json:"distance_fare" bson:"distance_fare"`
	TimeFare     float64 `json:"time_fare" bson:"time_fare"`
	SurgeFare    float64 `json:"surge_fare" bson:"surge_fare"`
	ServiceFee   float64 `json:"service_fee" bson:"service_fee"`
	TollsFee     float64 `json:"tolls_fee" bson:"tolls_fee"`

	// Discounts & Promotions
	DiscountAmount float64 `json:"discount_amount" bson:"discount_amount"`
	PromoCode      string  `json:"promo_code" bson:"promo_code"`

	// Totals
	Subtotal float64 `json:"subtotal" bson:"subtotal"`
	Total    float64 `json:"total" bson:"total"`
	Currency string  `json:"currency" bson:"currency"`

	// Validity
	EstimatedAt time.Time `json:"estimated_at" bson:"estimated_at"`
	ValidUntil  time.Time `json:"valid_until" bson:"valid_until"`

	// Additional Information
	Notes    []string `json:"notes,omitempty" bson:"notes,omitempty"`
	Warnings []string `json:"warnings,omitempty" bson:"warnings,omitempty"`

	// Price Range & Confidence
	PriceRange FarePriceRange `json:"price_range" bson:"price_range"`
	Confidence float64        `json:"confidence" bson:"confidence"`
}

// Additional response types
type FareComparison struct {
	ServiceOptions []ServiceOption `json:"service_options"`
	Recommended    string          `json:"recommended"`
	PriceRange     struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"price_range"`
}

type ServiceOption struct {
	ServiceType   ServiceType `json:"service_type"`
	VehicleType   VehicleType `json:"vehicle_type"`
	EstimatedFare float64     `json:"estimated_fare"`
	Duration      int         `json:"duration"`
	Availability  string      `json:"availability"`
}

type SuggestedFare struct {
	Amount     float64  `json:"amount"`
	Confidence float64  `json:"confidence"`
	Factors    []string `json:"factors"`
	PriceRange struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"price_range"`
	Description string `json:"description"`
}

type FareRule struct {
	ID          primitive.ObjectID     `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	RuleType    string                 `json:"rule_type"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     map[string]interface{} `json:"actions"`
	IsActive    bool                   `json:"is_active"`
}

type FareStatistics struct {
	TotalRides    int64     `json:"total_rides"`
	TotalSpent    float64   `json:"total_spent"`
	AverageFare   float64   `json:"average_fare"`
	HighestFare   float64   `json:"highest_fare"`
	LowestFare    float64   `json:"lowest_fare"`
	TotalSavings  float64   `json:"total_savings"`
	DiscountUsage int64     `json:"discount_usage"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
}

type FareTrend struct {
	Date        time.Time `json:"date"`
	AverageFare float64   `json:"average_fare"`
	RideCount   int64     `json:"ride_count"`
	DemandLevel string    `json:"demand_level"`
}

type DiscountCode struct {
	Code         string    `json:"code"`
	Description  string    `json:"description"`
	DiscountType string    `json:"discount_type"`
	Value        float64   `json:"value"`
	MaxDiscount  float64   `json:"max_discount"`
	MinFare      float64   `json:"min_fare"`
	ExpiresAt    time.Time `json:"expires_at"`
	UsageLeft    int       `json:"usage_left"`
}

type DiscountResult struct {
	OriginalFare   float64 `json:"original_fare"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalFare      float64 `json:"final_fare"`
	PromoCode      string  `json:"promo_code"`
	Savings        float64 `json:"savings"`
}

type CommissionRate struct {
	ServiceType string  `json:"service_type"`
	Rate        float64 `json:"rate"`
	Minimum     float64 `json:"minimum"`
	Maximum     float64 `json:"maximum"`
	City        string  `json:"city"`
}

type ServiceFee struct {
	ServiceType string  `json:"service_type"`
	Fee         float64 `json:"fee"`
	FeeType     string  `json:"fee_type"` // fixed, percentage
	City        string  `json:"city"`
}

type FareBreakdown struct {
	RideID         primitive.ObjectID `json:"ride_id"`
	BaseFare       float64            `json:"base_fare"`
	DistanceFare   float64            `json:"distance_fare"`
	TimeFare       float64            `json:"time_fare"`
	WaitingFee     float64            `json:"waiting_fee"`
	TollsFee       float64            `json:"tolls_fee"`
	SurgeFare      float64            `json:"surge_fare"`
	ServiceFee     float64            `json:"service_fee"`
	DiscountAmount float64            `json:"discount_amount"`
	TipAmount      float64            `json:"tip_amount"`
	TotalFare      float64            `json:"total_fare"`
	DriverEarnings float64            `json:"driver_earnings"`
	Commission     float64            `json:"commission"`
	PlatformFee    float64            `json:"platform_fee"`
	Currency       string             `json:"currency"`
}

// Fare Negotiation Models
type FareNegotiation struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RideID      primitive.ObjectID `json:"ride_id" bson:"ride_id"`
	PassengerID primitive.ObjectID `json:"passenger_id" bson:"passenger_id"`
	DriverID    primitive.ObjectID `json:"driver_id" bson:"driver_id"`

	// Negotiation Details
	Status      FareStatus `json:"status" bson:"status"`
	InitialFare float64    `json:"initial_fare" bson:"initial_fare"`
	FinalFare   float64    `json:"final_fare" bson:"final_fare"`
	Currency    string     `json:"currency" bson:"currency"`

	// Offers History
	Offers      []FareOffer `json:"offers" bson:"offers"`
	TotalOffers int         `json:"total_offers" bson:"total_offers"`

	// Timing
	StartedAt   time.Time  `json:"started_at" bson:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at" bson:"expires_at"`

	// Results
	AcceptedBy      *primitive.ObjectID `json:"accepted_by,omitempty" bson:"accepted_by,omitempty"`
	RejectedBy      *primitive.ObjectID `json:"rejected_by,omitempty" bson:"rejected_by,omitempty"`
	RejectionReason string              `json:"rejection_reason" bson:"rejection_reason"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
