package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideStatus string

const (
	RideStatusPending         RideStatus = "pending"
	RideStatusFareNegotiation RideStatus = "fare_negotiation"
	RideStatusAccepted        RideStatus = "accepted"
	RideStatusDriverEnRoute   RideStatus = "driver_en_route"
	RideStatusDriverArrived   RideStatus = "driver_arrived"
	RideStatusStarted         RideStatus = "started"
	RideStatusInProgress      RideStatus = "in_progress"
	RideStatusCompleted       RideStatus = "completed"
	RideStatusCancelled       RideStatus = "cancelled"
	RideStatusDisputed        RideStatus = "disputed"
)

type RideType string

const (
	RideTypeInstant   RideType = "instant"
	RideTypeScheduled RideType = "scheduled"
	RideTypeCourier   RideType = "courier"
	RideTypeFreight   RideType = "freight"
)

type CancellationReason string

const (
	CancelByPassenger      CancellationReason = "passenger_cancelled"
	CancelByDriver         CancellationReason = "driver_cancelled"
	CancelBySystem         CancellationReason = "system_cancelled"
	CancelByAdmin          CancellationReason = "admin_cancelled"
	CancelNoDriver         CancellationReason = "no_driver_available"
	CancelPaymentFailed    CancellationReason = "payment_failed"
	CancelWeatherCondition CancellationReason = "weather_condition"
	CancelVehicleIssue     CancellationReason = "vehicle_issue"
	CancelEmergency        CancellationReason = "emergency"
)

type Ride struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// Participants
	PassengerID primitive.ObjectID  `json:"passenger_id" bson:"passenger_id"`
	DriverID    *primitive.ObjectID `json:"driver_id,omitempty" bson:"driver_id,omitempty"`

	// Ride Information
	Type        RideType    `json:"type" bson:"type" validate:"required"`
	Status      RideStatus  `json:"status" bson:"status"`
	ServiceType ServiceType `json:"service_type" bson:"service_type"`

	// Location Details
	PickupLocation  RideLocation   `json:"pickup_location" bson:"pickup_location"`
	DropoffLocation RideLocation   `json:"dropoff_location" bson:"dropoff_location"`
	Waypoints       []RideLocation `json:"waypoints" bson:"waypoints"`

	// Fare Information
	FareDetails FareDetails `json:"fare_details" bson:"fare_details"`

	// Timing
	RequestedAt time.Time  `json:"requested_at" bson:"requested_at"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" bson:"scheduled_at,omitempty"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty" bson:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" bson:"cancelled_at,omitempty"`

	// Route Information
	Route             []Location `json:"route" bson:"route"`
	EstimatedDistance float64    `json:"estimated_distance" bson:"estimated_distance"` // km
	ActualDistance    float64    `json:"actual_distance" bson:"actual_distance"`       // km
	EstimatedDuration int        `json:"estimated_duration" bson:"estimated_duration"` // minutes
	ActualDuration    int        `json:"actual_duration" bson:"actual_duration"`       // minutes
	EstimatedArrival  *time.Time `json:"estimated_arrival,omitempty" bson:"estimated_arrival,omitempty"`

	// Requirements & Preferences
	Requirements RideRequirements `json:"requirements" bson:"requirements"`
	Preferences  RidePreferences  `json:"preferences" bson:"preferences"`

	// Payment Information
	PaymentMethodID string `json:"payment_method_id" bson:"payment_method_id"`
	PaymentStatus   string `json:"payment_status" bson:"payment_status"`
	TransactionID   string `json:"transaction_id" bson:"transaction_id"`

	// Additional Information
	Notes               string             `json:"notes" bson:"notes"`
	SpecialInstructions string             `json:"special_instructions" bson:"special_instructions"`
	CancellationReason  CancellationReason `json:"cancellation_reason,omitempty" bson:"cancellation_reason,omitempty"`
	CancellationNotes   string             `json:"cancellation_notes,omitempty" bson:"cancellation_notes,omitempty"`

	// Tracking & Sharing
	ShareCode       string `json:"share_code" bson:"share_code"`
	IsShared        bool   `json:"is_shared" bson:"is_shared"`
	TrackingEnabled bool   `json:"tracking_enabled" bson:"tracking_enabled"`

	// Real-time Data
	DriverLocation     *Location  `json:"driver_location,omitempty" bson:"driver_location,omitempty"`
	PassengerLocation  *Location  `json:"passenger_location,omitempty" bson:"passenger_location,omitempty"`
	LastLocationUpdate *time.Time `json:"last_location_update,omitempty" bson:"last_location_update,omitempty"`

	// Communication
	ChatEnabled    bool                `json:"chat_enabled" bson:"chat_enabled"`
	ConversationID *primitive.ObjectID `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`
	LastMessageAt  *time.Time          `json:"last_message_at,omitempty" bson:"last_message_at,omitempty"`

	// Rating & Feedback
	PassengerRating *Rating `json:"passenger_rating,omitempty" bson:"passenger_rating,omitempty"`
	DriverRating    *Rating `json:"driver_rating,omitempty" bson:"driver_rating,omitempty"`

	// System Information
	Platform       string    `json:"platform" bson:"platform"` // ios, android, web
	AppVersion     string    `json:"app_version" bson:"app_version"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" bson:"updated_at"`
	CompletionCode string    `json:"completion_code" bson:"completion_code"`
}

type RideLocation struct {
	Type         string    `json:"type" bson:"type"`
	Coordinates  []float64 `json:"coordinates" bson:"coordinates" validate:"coordinates"`
	Address      string    `json:"address" bson:"address" validate:"required"`
	Name         string    `json:"name" bson:"name"`
	PlaceID      string    `json:"place_id" bson:"place_id"`
	City         string    `json:"city" bson:"city"`
	State        string    `json:"state" bson:"state"`
	Country      string    `json:"country" bson:"country"`
	PostalCode   string    `json:"postal_code" bson:"postal_code"`
	Floor        string    `json:"floor" bson:"floor"`
	Building     string    `json:"building" bson:"building"`
	Instructions string    `json:"instructions" bson:"instructions"`
}

type RideRequirements struct {
	// Passenger Requirements
	PassengerCount       int  `json:"passenger_count" bson:"passenger_count" validate:"gte=1,lte=8"`
	ChildSeatCount       int  `json:"child_seat_count" bson:"child_seat_count"`
	LuggageCount         int  `json:"luggage_count" bson:"luggage_count"`
	WheelchairAccessible bool `json:"wheelchair_accessible" bson:"wheelchair_accessible"`

	// Vehicle Requirements
	VehicleType     VehicleType `json:"vehicle_type" bson:"vehicle_type"`
	AirConditioning bool        `json:"air_conditioning" bson:"air_conditioning"`
	MinimumRating   float64     `json:"minimum_rating" bson:"minimum_rating"`

	// Special Requirements
	PetsAllowed    bool   `json:"pets_allowed" bson:"pets_allowed"`
	SmokingAllowed bool   `json:"smoking_allowed" bson:"smoking_allowed"`
	QuietRide      bool   `json:"quiet_ride" bson:"quiet_ride"`
	SpecialNeeds   string `json:"special_needs" bson:"special_needs"`
	HelperRequired bool   `json:"helper_required" bson:"helper_required"`

	// Driver Requirements
	FemaleDriverOnly   bool   `json:"female_driver_only" bson:"female_driver_only"`
	VerifiedDriverOnly bool   `json:"verified_driver_only" bson:"verified_driver_only"`
	LanguagePreference string `json:"language_preference" bson:"language_preference"`

	// Package Requirements (for courier/freight)
	PackageDetails *PackageDetails `json:"package_details,omitempty" bson:"package_details,omitempty"`
}

type RidePreferences struct {
	// Route Preferences
	AvoidTolls         bool     `json:"avoid_tolls" bson:"avoid_tolls"`
	AvoidHighways      bool     `json:"avoid_highways" bson:"avoid_highways"`
	PreferFastestRoute bool     `json:"prefer_fastest_route" bson:"prefer_fastest_route"`
	PreferredRoute     []string `json:"preferred_route" bson:"preferred_route"`

	// Comfort Preferences
	Temperature       int    `json:"temperature" bson:"temperature"`
	MusicType         string `json:"music_type" bson:"music_type"`
	ConversationLevel string `json:"conversation_level" bson:"conversation_level"` // chatty, normal, quiet

	// Safety Preferences
	ShareRideDetails  bool     `json:"share_ride_details" bson:"share_ride_details"`
	EmergencyContacts []string `json:"emergency_contacts" bson:"emergency_contacts"`
	SafetyFeatures    []string `json:"safety_features" bson:"safety_features"`

	// Service Preferences
	WaitingTime     int  `json:"waiting_time" bson:"waiting_time"` // minutes
	FlexiblePickup  bool `json:"flexible_pickup" bson:"flexible_pickup"`
	FlexibleDropoff bool `json:"flexible_dropoff" bson:"flexible_dropoff"`
}

type PackageDetails struct {
	// Physical Properties
	Weight     float64 `json:"weight" bson:"weight" validate:"gte=0,lte=1000"` // kg
	Length     float64 `json:"length" bson:"length"`                           // cm
	Width      float64 `json:"width" bson:"width"`                             // cm
	Height     float64 `json:"height" bson:"height"`                           // cm
	Dimensions string  `json:"dimensions" bson:"dimensions"`                   // "20x30x40 cm"

	// Content Information
	Description string  `json:"description" bson:"description" validate:"required"`
	Category    string  `json:"category" bson:"category"`
	Value       float64 `json:"value" bson:"value"`
	Currency    string  `json:"currency" bson:"currency"`

	// Special Handling
	Fragile           bool     `json:"fragile" bson:"fragile"`
	Perishable        bool     `json:"perishable" bson:"perishable"`
	Hazardous         bool     `json:"hazardous" bson:"hazardous"`
	RequiresSignature bool     `json:"requires_signature" bson:"requires_signature"`
	SpecialHandling   []string `json:"special_handling" bson:"special_handling"`

	// Recipient Information
	RecipientName        string `json:"recipient_name" bson:"recipient_name" validate:"required"`
	RecipientPhone       string `json:"recipient_phone" bson:"recipient_phone" validate:"required,phone"`
	RecipientEmail       string `json:"recipient_email" bson:"recipient_email" validate:"email"`
	DeliveryInstructions string `json:"delivery_instructions" bson:"delivery_instructions"`

	// Photos
	Photos []PackagePhoto `json:"photos" bson:"photos"`

	// Insurance
	InsuranceRequired bool    `json:"insurance_required" bson:"insurance_required"`
	InsuranceValue    float64 `json:"insurance_value" bson:"insurance_value"`
	InsuranceCost     float64 `json:"insurance_cost" bson:"insurance_cost"`

	// Tracking
	TrackingCode  string     `json:"tracking_code" bson:"tracking_code"`
	DeliveryProof string     `json:"delivery_proof" bson:"delivery_proof"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`
	DeliveredBy   string     `json:"delivered_by" bson:"delivered_by"`
	SignatureURL  string     `json:"signature_url" bson:"signature_url"`
}

type PackagePhoto struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	URL         string             `json:"url" bson:"url"`
	Type        string             `json:"type" bson:"type"` // before_pickup, after_pickup, before_delivery, after_delivery
	Description string             `json:"description" bson:"description"`
	TakenAt     time.Time          `json:"taken_at" bson:"taken_at"`
}
