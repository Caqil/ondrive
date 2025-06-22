package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeBike       VehicleType = "bike"
	VehicleTypeTruck      VehicleType = "truck"
	VehicleTypeVan        VehicleType = "van"
	VehicleTypeSUV        VehicleType = "suv"
	VehicleTypePickup     VehicleType = "pickup"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
)

type ServiceType string

const (
	ServiceTypeRide    ServiceType = "ride"
	ServiceTypeCourier ServiceType = "courier"
	ServiceTypeFreight ServiceType = "freight"
)

type DriverStatus string

const (
	DriverStatusOffline     DriverStatus = "offline"
	DriverStatusOnline      DriverStatus = "online"
	DriverStatusBusy        DriverStatus = "busy"
	DriverStatusBreak       DriverStatus = "break"
	DriverStatusUnavailable DriverStatus = "unavailable"
)

type DriverInfo struct {
	// Vehicle Information
	VehicleType    VehicleType    `json:"vehicle_type" bson:"vehicle_type" validate:"vehicle_type"`
	VehicleDetails VehicleDetails `json:"vehicle_details" bson:"vehicle_details"`

	// Service Configuration
	ServiceTypes  []ServiceType `json:"service_types" bson:"service_types"`
	MaxWeightKg   int           `json:"max_weight_kg" bson:"max_weight_kg" validate:"gte=0,lte=1000"`
	MaxPassengers int           `json:"max_passengers" bson:"max_passengers" validate:"gte=1,lte=8"`

	// Status & Availability
	Status        DriverStatus        `json:"status" bson:"status"`
	IsOnline      bool                `json:"is_online" bson:"is_online"`
	IsAvailable   bool                `json:"is_available" bson:"is_available"`
	CurrentRideID *primitive.ObjectID `json:"current_ride_id,omitempty" bson:"current_ride_id,omitempty"`

	// Performance Metrics
	Rating      DriverRating      `json:"rating" bson:"rating"`
	Earnings    DriverEarnings    `json:"earnings" bson:"earnings"`
	Performance DriverPerformance `json:"performance" bson:"performance"`

	// Work Configuration
	WorkingHours WorkingHours      `json:"working_hours" bson:"working_hours"`
	Preferences  DriverPreferences `json:"preferences" bson:"preferences"`
	ServiceAreas []ServiceArea     `json:"service_areas" bson:"service_areas"`

	// Account Information
	LicenseNumber   string     `json:"license_number" bson:"license_number"`
	LicenseExpiry   *time.Time `json:"license_expiry" bson:"license_expiry"`
	InsuranceExpiry *time.Time `json:"insurance_expiry" bson:"insurance_expiry"`

	// Timestamps
	OnlineSince *time.Time `json:"online_since,omitempty" bson:"online_since,omitempty"`
	LastActive  time.Time  `json:"last_active" bson:"last_active"`
	JoinedAt    time.Time  `json:"joined_at" bson:"joined_at"`
}

type VehicleDetails struct {
	Make         string `json:"make" bson:"make" validate:"required"`
	Model        string `json:"model" bson:"model" validate:"required"`
	Year         int    `json:"year" bson:"year" validate:"gte=1990,lte=2030"`
	Color        string `json:"color" bson:"color" validate:"required"`
	LicensePlate string `json:"license_plate" bson:"license_plate" validate:"required"`
	VIN          string `json:"vin" bson:"vin"`

	// Photos
	PhotoURL string         `json:"photo_url" bson:"photo_url"`
	Photos   []VehiclePhoto `json:"photos" bson:"photos"`

	// Specifications
	FuelType        string `json:"fuel_type" bson:"fuel_type"`
	Transmission    string `json:"transmission" bson:"transmission"`
	Seats           int    `json:"seats" bson:"seats" validate:"gte=1,lte=50"`
	AirConditioning bool   `json:"air_conditioning" bson:"air_conditioning"`

	// Features
	Features  []string `json:"features" bson:"features"`
	Amenities []string `json:"amenities" bson:"amenities"`

	// Verification
	IsVerified bool       `json:"is_verified" bson:"is_verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty" bson:"verified_at,omitempty"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type VehiclePhoto struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	URL         string             `json:"url" bson:"url"`
	Type        string             `json:"type" bson:"type"` // front, back, side, interior, documents
	Description string             `json:"description" bson:"description"`
	IsPrimary   bool               `json:"is_primary" bson:"is_primary"`
	UploadedAt  time.Time          `json:"uploaded_at" bson:"uploaded_at"`
}

type DriverRating struct {
	Average      float64    `json:"average" bson:"average"`
	TotalRides   int        `json:"total_rides" bson:"total_rides"`
	TotalRating  float64    `json:"total_rating" bson:"total_rating"`
	FiveStars    int        `json:"five_stars" bson:"five_stars"`
	FourStars    int        `json:"four_stars" bson:"four_stars"`
	ThreeStars   int        `json:"three_stars" bson:"three_stars"`
	TwoStars     int        `json:"two_stars" bson:"two_stars"`
	OneStar      int        `json:"one_star" bson:"one_star"`
	LastRatingAt *time.Time `json:"last_rating_at,omitempty" bson:"last_rating_at,omitempty"`
}

type DriverEarnings struct {
	// Total Earnings
	TotalEarnings float64 `json:"total_earnings" bson:"total_earnings"`
	NetEarnings   float64 `json:"net_earnings" bson:"net_earnings"`
	GrossEarnings float64 `json:"gross_earnings" bson:"gross_earnings"`

	// Period Earnings
	Today     float64 `json:"today" bson:"today"`
	ThisWeek  float64 `json:"this_week" bson:"this_week"`
	ThisMonth float64 `json:"this_month" bson:"this_month"`
	LastMonth float64 `json:"last_month" bson:"last_month"`

	// Commission & Fees
	TotalCommission float64 `json:"total_commission" bson:"total_commission"`
	TotalFees       float64 `json:"total_fees" bson:"total_fees"`
	CommissionRate  float64 `json:"commission_rate" bson:"commission_rate"`

	// Payout Information
	PendingPayout   float64    `json:"pending_payout" bson:"pending_payout"`
	LastPayoutAt    *time.Time `json:"last_payout_at,omitempty" bson:"last_payout_at,omitempty"`
	NextPayoutDate  *time.Time `json:"next_payout_date,omitempty" bson:"next_payout_date,omitempty"`
	PayoutFrequency string     `json:"payout_frequency" bson:"payout_frequency"`

	// Statistics
	AvgEarningsPerRide float64 `json:"avg_earnings_per_ride" bson:"avg_earnings_per_ride"`
	AvgEarningsPerHour float64 `json:"avg_earnings_per_hour" bson:"avg_earnings_per_hour"`
	TotalTips          float64 `json:"total_tips" bson:"total_tips"`
	TotalBonuses       float64 `json:"total_bonuses" bson:"total_bonuses"`
}

type DriverPerformance struct {
	// Acceptance & Completion
	AcceptanceRate   float64 `json:"acceptance_rate" bson:"acceptance_rate"`
	CompletionRate   float64 `json:"completion_rate" bson:"completion_rate"`
	CancellationRate float64 `json:"cancellation_rate" bson:"cancellation_rate"`

	// Time Metrics
	AvgPickupTime   float64 `json:"avg_pickup_time" bson:"avg_pickup_time"`     // minutes
	AvgDeliveryTime float64 `json:"avg_delivery_time" bson:"avg_delivery_time"` // minutes
	OnTimeRate      float64 `json:"on_time_rate" bson:"on_time_rate"`

	// Ride Metrics
	TotalRides     int `json:"total_rides" bson:"total_rides"`
	CompletedRides int `json:"completed_rides" bson:"completed_rides"`
	CancelledRides int `json:"cancelled_rides" bson:"cancelled_rides"`

	// Online Time
	TotalOnlineHours float64 `json:"total_online_hours" bson:"total_online_hours"`
	ActiveHours      float64 `json:"active_hours" bson:"active_hours"`
	UtilizationRate  float64 `json:"utilization_rate" bson:"utilization_rate"`

	// Customer Satisfaction
	CustomerRating  float64 `json:"customer_rating" bson:"customer_rating"`
	ComplaintCount  int     `json:"complaint_count" bson:"complaint_count"`
	ComplimentCount int     `json:"compliment_count" bson:"compliment_count"`

	// Period Performance
	ThisWeekPerformance  PeriodPerformance `json:"this_week_performance" bson:"this_week_performance"`
	ThisMonthPerformance PeriodPerformance `json:"this_month_performance" bson:"this_month_performance"`
}

type PeriodPerformance struct {
	Rides          int     `json:"rides" bson:"rides"`
	Hours          float64 `json:"hours" bson:"hours"`
	Earnings       float64 `json:"earnings" bson:"earnings"`
	Rating         float64 `json:"rating" bson:"rating"`
	AcceptanceRate float64 `json:"acceptance_rate" bson:"acceptance_rate"`
	CompletionRate float64 `json:"completion_rate" bson:"completion_rate"`
}

type WorkingHours struct {
	// Schedule
	Monday    DaySchedule `json:"monday" bson:"monday"`
	Tuesday   DaySchedule `json:"tuesday" bson:"tuesday"`
	Wednesday DaySchedule `json:"wednesday" bson:"wednesday"`
	Thursday  DaySchedule `json:"thursday" bson:"thursday"`
	Friday    DaySchedule `json:"friday" bson:"friday"`
	Saturday  DaySchedule `json:"saturday" bson:"saturday"`
	Sunday    DaySchedule `json:"sunday" bson:"sunday"`

	// Flexibility
	IsFlexible            bool `json:"is_flexible" bson:"is_flexible"`
	MaxHoursPerDay        int  `json:"max_hours_per_day" bson:"max_hours_per_day"`
	MaxHoursPerWeek       int  `json:"max_hours_per_week" bson:"max_hours_per_week"`
	MinHoursBetweenShifts int  `json:"min_hours_between_shifts" bson:"min_hours_between_shifts"`

	// Break Configuration
	BreakDuration      int `json:"break_duration" bson:"break_duration"` // minutes
	MaxContinuousHours int `json:"max_continuous_hours" bson:"max_continuous_hours"`

	// Current Status
	CurrentlyWorking bool       `json:"currently_working" bson:"currently_working"`
	ShiftStartTime   *time.Time `json:"shift_start_time,omitempty" bson:"shift_start_time,omitempty"`
	LastBreakTime    *time.Time `json:"last_break_time,omitempty" bson:"last_break_time,omitempty"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type DaySchedule struct {
	IsWorking bool    `json:"is_working" bson:"is_working"`
	StartTime string  `json:"start_time" bson:"start_time"` // "09:00"
	EndTime   string  `json:"end_time" bson:"end_time"`     // "17:00"
	Breaks    []Break `json:"breaks" bson:"breaks"`
}

type Break struct {
	StartTime string `json:"start_time" bson:"start_time"`
	EndTime   string `json:"end_time" bson:"end_time"`
	Type      string `json:"type" bson:"type"` // lunch, coffee, rest
}

type DriverPreferences struct {
	// Ride Preferences
	AutoAcceptRides bool     `json:"auto_accept_rides" bson:"auto_accept_rides"`
	MaxDistance     float64  `json:"max_distance" bson:"max_distance"` // km
	MinFareAmount   float64  `json:"min_fare_amount" bson:"min_fare_amount"`
	PreferredAreas  []string `json:"preferred_areas" bson:"preferred_areas"`
	AvoidAreas      []string `json:"avoid_areas" bson:"avoid_areas"`

	// Service Preferences
	AcceptCashPayments bool `json:"accept_cash_payments" bson:"accept_cash_payments"`
	AcceptCardPayments bool `json:"accept_card_payments" bson:"accept_card_payments"`
	AcceptPets         bool `json:"accept_pets" bson:"accept_pets"`
	AcceptLuggage      bool `json:"accept_luggage" bson:"accept_luggage"`
	AcceptSmoking      bool `json:"accept_smoking" bson:"accept_smoking"`
	AcceptChildren     bool `json:"accept_children" bson:"accept_children"`
	AcceptFoodDelivery bool `json:"accept_food_delivery" bson:"accept_food_delivery"`

	// Passenger Preferences
	PreferredGender          string `json:"preferred_gender" bson:"preferred_gender"`
	MaxPassengers            int    `json:"max_passengers" bson:"max_passengers"`
	RequirePhoneVerification bool   `json:"require_phone_verification" bson:"require_phone_verification"`

	// Notification Preferences
	RideRequestNotifications bool `json:"ride_request_notifications" bson:"ride_request_notifications"`
	PaymentNotifications     bool `json:"payment_notifications" bson:"payment_notifications"`
	PromoNotifications       bool `json:"promo_notifications" bson:"promo_notifications"`

	// Route Preferences
	AvoidTolls         bool `json:"avoid_tolls" bson:"avoid_tolls"`
	AvoidHighways      bool `json:"avoid_highways" bson:"avoid_highways"`
	PreferFastestRoute bool `json:"prefer_fastest_route" bson:"prefer_fastest_route"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type ServiceArea struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Type        string             `json:"type" bson:"type"`               // city, zone, custom
	Coordinates [][]float64        `json:"coordinates" bson:"coordinates"` // polygon coordinates
	Center      Location           `json:"center" bson:"center"`
	Radius      float64            `json:"radius" bson:"radius"` // km (for circular areas)
	IsActive    bool               `json:"is_active" bson:"is_active"`
	Priority    int                `json:"priority" bson:"priority"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}
