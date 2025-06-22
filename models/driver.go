package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Driver struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	DriverInfo *DriverInfo        `json:"driver_info" bson:"driver_info"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}
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
	AvoidTolls         bool      `json:"avoid_tolls" bson:"avoid_tolls"`
	AvoidHighways      bool      `json:"avoid_highways" bson:"avoid_highways"`
	PreferFastestRoute bool      `json:"prefer_fastest_route" bson:"prefer_fastest_route"`
	WaitingTime        int       `json:"waiting_time"`
	FlexiblePickup     bool      `json:"flexible_pickup"`
	FlexibleDropoff    bool      `json:"flexible_dropoff"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at"`
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

// DriverStats represents comprehensive driver statistics
type DriverStats struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	// Ride Statistics
	TotalRides       int     `json:"total_rides" bson:"total_rides"`
	CompletedRides   int     `json:"completed_rides" bson:"completed_rides"`
	CancelledRides   int     `json:"cancelled_rides" bson:"cancelled_rides"`
	AcceptanceRate   float64 `json:"acceptance_rate" bson:"acceptance_rate"`
	CancellationRate float64 `json:"cancellation_rate" bson:"cancellation_rate"`
	CompletionRate   float64 `json:"completion_rate" bson:"completion_rate"`

	// Rating Statistics
	AverageRating    float64 `json:"average_rating" bson:"average_rating"`
	TotalRatings     int     `json:"total_ratings" bson:"total_ratings"`
	FiveStarRatings  int     `json:"five_star_ratings" bson:"five_star_ratings"`
	FourStarRatings  int     `json:"four_star_ratings" bson:"four_star_ratings"`
	ThreeStarRatings int     `json:"three_star_ratings" bson:"three_star_ratings"`
	TwoStarRatings   int     `json:"two_star_ratings" bson:"two_star_ratings"`
	OneStarRatings   int     `json:"one_star_ratings" bson:"one_star_ratings"`

	// Financial Statistics
	TotalEarnings    float64 `json:"total_earnings" bson:"total_earnings"`
	TotalWithdrawn   float64 `json:"total_withdrawn" bson:"total_withdrawn"`
	AvailableBalance float64 `json:"available_balance" bson:"available_balance"`
	TotalTips        float64 `json:"total_tips" bson:"total_tips"`
	TotalBonuses     float64 `json:"total_bonuses" bson:"total_bonuses"`

	// Time Statistics
	TotalOnlineHours   float64 `json:"total_online_hours" bson:"total_online_hours"`
	TotalActiveHours   float64 `json:"total_active_hours" bson:"total_active_hours"`
	UtilizationRate    float64 `json:"utilization_rate" bson:"utilization_rate"`
	AvgRidesPerHour    float64 `json:"avg_rides_per_hour" bson:"avg_rides_per_hour"`
	AvgEarningsPerHour float64 `json:"avg_earnings_per_hour" bson:"avg_earnings_per_hour"`
	AvgEarningsPerRide float64 `json:"avg_earnings_per_ride" bson:"avg_earnings_per_ride"`

	// Distance Statistics
	TotalDistance   float64 `json:"total_distance" bson:"total_distance"`
	AvgRideDistance float64 `json:"avg_ride_distance" bson:"avg_ride_distance"`
	AvgPickupTime   float64 `json:"avg_pickup_time" bson:"avg_pickup_time"`
	AvgDeliveryTime float64 `json:"avg_delivery_time" bson:"avg_delivery_time"`

	// Period Statistics
	TodaysRides     int     `json:"todays_rides" bson:"todays_rides"`
	TodaysEarnings  float64 `json:"todays_earnings" bson:"todays_earnings"`
	WeeklyRides     int     `json:"weekly_rides" bson:"weekly_rides"`
	WeeklyEarnings  float64 `json:"weekly_earnings" bson:"weekly_earnings"`
	MonthlyRides    int     `json:"monthly_rides" bson:"monthly_rides"`
	MonthlyEarnings float64 `json:"monthly_earnings" bson:"monthly_earnings"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// PerformanceMetrics represents driver performance metrics for a specific period
type PerformanceMetrics struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	Period string             `json:"period" bson:"period"` // daily, weekly, monthly

	// Performance Scores
	OverallScore     float64 `json:"overall_score" bson:"overall_score"`
	QualityScore     float64 `json:"quality_score" bson:"quality_score"`
	EfficiencyScore  float64 `json:"efficiency_score" bson:"efficiency_score"`
	ReliabilityScore float64 `json:"reliability_score" bson:"reliability_score"`

	// Key Metrics
	AcceptanceRate   float64 `json:"acceptance_rate" bson:"acceptance_rate"`
	CancellationRate float64 `json:"cancellation_rate" bson:"cancellation_rate"`
	OnTimeRate       float64 `json:"on_time_rate" bson:"on_time_rate"`
	CustomerRating   float64 `json:"customer_rating" bson:"customer_rating"`
	UtilizationRate  float64 `json:"utilization_rate" bson:"utilization_rate"`

	// Comparisons
	ComparedToAverage float64 `json:"compared_to_average" bson:"compared_to_average"`
	Ranking           int     `json:"ranking" bson:"ranking"`
	TotalDrivers      int     `json:"total_drivers" bson:"total_drivers"`

	// Achievements
	Achievements []string `json:"achievements" bson:"achievements"`
	Badges       []string `json:"badges" bson:"badges"`

	// Areas for Improvement
	ImprovementAreas []string `json:"improvement_areas" bson:"improvement_areas"`

	StartDate time.Time `json:"start_date" bson:"start_date"`
	EndDate   time.Time `json:"end_date" bson:"end_date"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type RatingComment struct {
	RideID    primitive.ObjectID `json:"ride_id" bson:"ride_id"`
	Rating    float64            `json:"rating" bson:"rating"`
	Comment   string             `json:"comment" bson:"comment"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// Earning represents a single earning record
type Earning struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	RideID primitive.ObjectID `json:"ride_id" bson:"ride_id"`

	// Amounts
	Amount     float64 `json:"amount" bson:"amount"`
	BaseAmount float64 `json:"base_amount" bson:"base_amount"`
	Distance   float64 `json:"distance" bson:"distance"`
	Duration   float64 `json:"duration" bson:"duration"`
	Tips       float64 `json:"tips" bson:"tips"`
	Bonuses    float64 `json:"bonuses" bson:"bonuses"`
	Deductions float64 `json:"deductions" bson:"deductions"`
	Commission float64 `json:"commission" bson:"commission"`
	NetAmount  float64 `json:"net_amount" bson:"net_amount"`

	// Ride Details
	ServiceType string    `json:"service_type" bson:"service_type"`
	RideType    string    `json:"ride_type" bson:"ride_type"`
	PickupTime  time.Time `json:"pickup_time" bson:"pickup_time"`
	DropoffTime time.Time `json:"dropoff_time" bson:"dropoff_time"`
	WaitTime    float64   `json:"wait_time" bson:"wait_time"`

	// Location
	PickupLocation  Location `json:"pickup_location" bson:"pickup_location"`
	DropoffLocation Location `json:"dropoff_location" bson:"dropoff_location"`

	// Status
	Status     string     `json:"status" bson:"status"` // completed, disputed, adjusted
	PayoutDate *time.Time `json:"payout_date,omitempty" bson:"payout_date,omitempty"`
	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
}

// Payout represents a payout transaction
type Payout struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	// Amount Details
	Amount        float64 `json:"amount" bson:"amount"`
	ProcessingFee float64 `json:"processing_fee" bson:"processing_fee"`
	NetAmount     float64 `json:"net_amount" bson:"net_amount"`

	// Payout Details
	Method        string `json:"method" bson:"method"` // bank_transfer, digital_wallet, etc.
	BankAccount   string `json:"bank_account,omitempty" bson:"bank_account,omitempty"`
	WalletAddress string `json:"wallet_address,omitempty" bson:"wallet_address,omitempty"`
	TransactionID string `json:"transaction_id,omitempty" bson:"transaction_id,omitempty"`

	// Status
	Status        string     `json:"status" bson:"status"` // pending, processing, completed, failed
	StatusMessage string     `json:"status_message,omitempty" bson:"status_message,omitempty"`
	RequestedAt   time.Time  `json:"requested_at" bson:"requested_at"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty" bson:"completed_at,omitempty"`

	// Period Covered
	EarningsFrom time.Time `json:"earnings_from" bson:"earnings_from"`
	EarningsTo   time.Time `json:"earnings_to" bson:"earnings_to"`
}

// DriverSchedule represents a driver's schedule for a specific day
type DriverSchedule struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	Date   time.Time          `json:"date" bson:"date"`

	// Schedule
	StartTime    *time.Time  `json:"start_time,omitempty" bson:"start_time,omitempty"`
	EndTime      *time.Time  `json:"end_time,omitempty" bson:"end_time,omitempty"`
	IsWorking    bool        `json:"is_working" bson:"is_working"`
	IsOnBreak    bool        `json:"is_on_break" bson:"is_on_break"`
	CurrentBreak *BreakInfo  `json:"current_break,omitempty" bson:"current_break,omitempty"`
	Breaks       []BreakInfo `json:"breaks" bson:"breaks"`

	// Planned vs Actual
	PlannedHours float64 `json:"planned_hours" bson:"planned_hours"`
	ActualHours  float64 `json:"actual_hours" bson:"actual_hours"`

	// Performance
	RidesCompleted int     `json:"rides_completed" bson:"rides_completed"`
	EarningsToday  float64 `json:"earnings_today" bson:"earnings_today"`
	OnlineHours    float64 `json:"online_hours" bson:"online_hours"`
	ActiveHours    float64 `json:"active_hours" bson:"active_hours"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// BreakInfo represents information about a driver's break
type BreakInfo struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StartTime time.Time          `json:"start_time" bson:"start_time"`
	EndTime   *time.Time         `json:"end_time,omitempty" bson:"end_time,omitempty"`
	Duration  time.Duration      `json:"duration" bson:"duration"`
	Reason    string             `json:"reason" bson:"reason"`
	Type      string             `json:"type" bson:"type"` // scheduled, emergency, lunch, rest
	IsActive  bool               `json:"is_active" bson:"is_active"`
}

// VerificationStatus represents the verification status for a driver
type VerificationStatus struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	// Overall Status
	IsFullyVerified bool       `json:"is_fully_verified" bson:"is_fully_verified"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty" bson:"verified_at,omitempty"`

	// Document Verification
	DriverLicenseVerified         bool       `json:"driver_license_verified" bson:"driver_license_verified"`
	DriverLicenseVerifiedAt       *time.Time `json:"driver_license_verified_at,omitempty" bson:"driver_license_verified_at,omitempty"`
	VehicleRegistrationVerified   bool       `json:"vehicle_registration_verified" bson:"vehicle_registration_verified"`
	VehicleRegistrationVerifiedAt *time.Time `json:"vehicle_registration_verified_at,omitempty" bson:"vehicle_registration_verified_at,omitempty"`
	InsuranceVerified             bool       `json:"insurance_verified" bson:"insurance_verified"`
	InsuranceVerifiedAt           *time.Time `json:"insurance_verified_at,omitempty" bson:"insurance_verified_at,omitempty"`

	// Background Check
	BackgroundCheckCompleted   bool       `json:"background_check_completed" bson:"background_check_completed"`
	BackgroundCheckStatus      string     `json:"background_check_status" bson:"background_check_status"`
	BackgroundCheckCompletedAt *time.Time `json:"background_check_completed_at,omitempty" bson:"background_check_completed_at,omitempty"`

	// Vehicle Inspection
	VehicleInspectionCompleted   bool       `json:"vehicle_inspection_completed" bson:"vehicle_inspection_completed"`
	VehicleInspectionStatus      string     `json:"vehicle_inspection_status" bson:"vehicle_inspection_status"`
	VehicleInspectionCompletedAt *time.Time `json:"vehicle_inspection_completed_at,omitempty" bson:"vehicle_inspection_completed_at,omitempty"`

	// Medical Check
	MedicalCheckCompleted   bool       `json:"medical_check_completed" bson:"medical_check_completed"`
	MedicalCheckStatus      string     `json:"medical_check_status" bson:"medical_check_status"`
	MedicalCheckCompletedAt *time.Time `json:"medical_check_completed_at,omitempty" bson:"medical_check_completed_at,omitempty"`

	// Training
	TrainingCompleted   bool       `json:"training_completed" bson:"training_completed"`
	TrainingCompletedAt *time.Time `json:"training_completed_at,omitempty" bson:"training_completed_at,omitempty"`

	// Rejection Information
	RejectionReason string     `json:"rejection_reason,omitempty" bson:"rejection_reason,omitempty"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty" bson:"rejected_at,omitempty"`

	// Re-verification
	RequiresReverification bool       `json:"requires_reverification" bson:"requires_reverification"`
	NextVerificationDate   *time.Time `json:"next_verification_date,omitempty" bson:"next_verification_date,omitempty"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// SupportTicket represents a support ticket created by a driver
type SupportTicket struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	// Ticket Details
	Subject     string `json:"subject" bson:"subject"`
	Description string `json:"description" bson:"description"`
	Category    string `json:"category" bson:"category"` // technical, payment, vehicle, ride, general
	Priority    string `json:"priority" bson:"priority"` // low, medium, high, urgent
	Status      string `json:"status" bson:"status"`     // open, in_progress, resolved, closed

	// Assignment
	AssignedTo *primitive.ObjectID `json:"assigned_to,omitempty" bson:"assigned_to,omitempty"`
	AssignedAt *time.Time          `json:"assigned_at,omitempty" bson:"assigned_at,omitempty"`

	// Resolution
	Resolution string              `json:"resolution,omitempty" bson:"resolution,omitempty"`
	ResolvedAt *time.Time          `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
	ResolvedBy *primitive.ObjectID `json:"resolved_by,omitempty" bson:"resolved_by,omitempty"`

	// Communication
	Messages    []TicketMessage `json:"messages" bson:"messages"`
	LastMessage *TicketMessage  `json:"last_message,omitempty" bson:"last_message,omitempty"`

	// Attachments
	Attachments []string `json:"attachments" bson:"attachments"`

	// Feedback
	SatisfactionRating *int   `json:"satisfaction_rating,omitempty" bson:"satisfaction_rating,omitempty"`
	Feedback           string `json:"feedback,omitempty" bson:"feedback,omitempty"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type TicketMessage struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SenderID   primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	SenderType string             `json:"sender_type" bson:"sender_type"` // user, admin, system
	Message    string             `json:"message" bson:"message"`
	IsInternal bool               `json:"is_internal" bson:"is_internal"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
}

// NotificationSettings represents driver notification preferences
type NotificationSettings struct {
	RideRequests      bool `json:"ride_requests" bson:"ride_requests"`
	RideUpdates       bool `json:"ride_updates" bson:"ride_updates"`
	PaymentUpdates    bool `json:"payment_updates" bson:"payment_updates"`
	EarningsReports   bool `json:"earnings_reports" bson:"earnings_reports"`
	PerformanceAlerts bool `json:"performance_alerts" bson:"performance_alerts"`
	PromoOffers       bool `json:"promo_offers" bson:"promo_offers"`
	SystemUpdates     bool `json:"system_updates" bson:"system_updates"`
	EmergencyAlerts   bool `json:"emergency_alerts" bson:"emergency_alerts"`

	// Delivery Method
	PushNotifications  bool `json:"push_notifications" bson:"push_notifications"`
	SMSNotifications   bool `json:"sms_notifications" bson:"sms_notifications"`
	EmailNotifications bool `json:"email_notifications" bson:"email_notifications"`

	// Timing
	QuietHoursStart      string `json:"quiet_hours_start" bson:"quiet_hours_start"`
	QuietHoursEnd        string `json:"quiet_hours_end" bson:"quiet_hours_end"`
	WeekendNotifications bool   `json:"weekend_notifications" bson:"weekend_notifications"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// DriverStatsBulkUpdate represents bulk updates for driver statistics
type DriverStatsBulkUpdate struct {
	UserID      string                 `json:"user_id" bson:"user_id"`
	StatsUpdate map[string]interface{} `json:"stats_update" bson:"stats_update"`
}
