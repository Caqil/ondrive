package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CourierStatus string

const (
	CourierStatusPending   CourierStatus = "pending"
	CourierStatusAccepted  CourierStatus = "accepted"
	CourierStatusPickedUp  CourierStatus = "picked_up"
	CourierStatusInTransit CourierStatus = "in_transit"
	CourierStatusDelivered CourierStatus = "delivered"
	CourierStatusCancelled CourierStatus = "cancelled"
	CourierStatusReturned  CourierStatus = "returned"
	CourierStatusFailed    CourierStatus = "failed"
)

type PackageCategory string

const (
	PackageCategoryDocuments   PackageCategory = "documents"
	PackageCategoryElectronics PackageCategory = "electronics"
	PackageCategoryClothing    PackageCategory = "clothing"
	PackageCategoryFood        PackageCategory = "food"
	PackageCategoryMedicine    PackageCategory = "medicine"
	PackageCategoryFragile     PackageCategory = "fragile"
	PackageCategoryPersonal    PackageCategory = "personal"
	PackageCategoryOther       PackageCategory = "other"
)

type DeliverySpeed string

const (
	DeliverySpeedStandard DeliverySpeed = "standard" // Same day
	DeliverySpeedExpress  DeliverySpeed = "express"  // 2-4 hours
	DeliverySpeedPriority DeliverySpeed = "priority" // 1-2 hours
	DeliverySpeedUrgent   DeliverySpeed = "urgent"   // Within 1 hour
)

type CourierRequest struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// Request Information
	Status   CourierStatus `json:"status" bson:"status"`
	Type     string        `json:"type" bson:"type"` // pickup_delivery, send_package
	Priority DeliverySpeed `json:"priority" bson:"priority"`

	// Participants
	SenderID  primitive.ObjectID  `json:"sender_id" bson:"sender_id"`
	CourierID *primitive.ObjectID `json:"courier_id,omitempty" bson:"courier_id,omitempty"`

	// Package Information
	Package CourierPackage `json:"package" bson:"package"`

	// Location Information
	PickupLocation   CourierLocation `json:"pickup_location" bson:"pickup_location"`
	DeliveryLocation CourierLocation `json:"delivery_location" bson:"delivery_location"`

	// Recipient Information
	Recipient CourierRecipient `json:"recipient" bson:"recipient"`

	// Scheduling
	ScheduledPickupTime   *time.Time  `json:"scheduled_pickup_time,omitempty" bson:"scheduled_pickup_time,omitempty"`
	ScheduledDeliveryTime *time.Time  `json:"scheduled_delivery_time,omitempty" bson:"scheduled_delivery_time,omitempty"`
	FlexibleTiming        bool        `json:"flexible_timing" bson:"flexible_timing"`
	TimeWindow            *TimeWindow `json:"time_window,omitempty" bson:"time_window,omitempty"`

	// Actual Timing
	PickedUpAt  *time.Time `json:"picked_up_at,omitempty" bson:"picked_up_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`

	// Pricing
	Fare            CourierFare `json:"fare" bson:"fare"`
	PaymentMethodID string      `json:"payment_method_id" bson:"payment_method_id"`
	PaymentStatus   string      `json:"payment_status" bson:"payment_status"`

	// Tracking
	TrackingCode     string          `json:"tracking_code" bson:"tracking_code"`
	TrackingHistory  []TrackingEvent `json:"tracking_history" bson:"tracking_history"`
	CurrentLocation  *Location       `json:"current_location,omitempty" bson:"current_location,omitempty"`
	EstimatedArrival *time.Time      `json:"estimated_arrival,omitempty" bson:"estimated_arrival,omitempty"`

	// Communication
	SpecialInstructions string              `json:"special_instructions" bson:"special_instructions"`
	ConversationID      *primitive.ObjectID `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`

	// Proof of Delivery
	DeliveryProof *DeliveryProof `json:"delivery_proof,omitempty" bson:"delivery_proof,omitempty"`

	// Insurance & Safety
	Insurance *CourierInsurance `json:"insurance,omitempty" bson:"insurance,omitempty"`

	// Cancellation
	CancellationReason string              `json:"cancellation_reason" bson:"cancellation_reason"`
	CancelledBy        *primitive.ObjectID `json:"cancelled_by,omitempty" bson:"cancelled_by,omitempty"`
	CancelledAt        *time.Time          `json:"cancelled_at,omitempty" bson:"cancelled_at,omitempty"`

	// Rating & Feedback
	SenderRating  *Rating `json:"sender_rating,omitempty" bson:"sender_rating,omitempty"`
	CourierRating *Rating `json:"courier_rating,omitempty" bson:"courier_rating,omitempty"`

	// Metadata
	Platform  string    `json:"platform" bson:"platform"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type CourierPackage struct {
	// Basic Information
	Description string          `json:"description" bson:"description" validate:"required"`
	Category    PackageCategory `json:"category" bson:"category"`
	Quantity    int             `json:"quantity" bson:"quantity" validate:"gte=1"`

	// Physical Properties
	Weight     float64           `json:"weight" bson:"weight" validate:"gte=0.1,lte=20"` // kg (max 20kg for courier)
	Dimensions PackageDimensions `json:"dimensions" bson:"dimensions"`
	Volume     float64           `json:"volume" bson:"volume"` // calculated volume in cm³

	// Content Details
	Contents []PackageItem `json:"contents" bson:"contents"`
	Value    float64       `json:"value" bson:"value"`
	Currency string        `json:"currency" bson:"currency"`

	// Special Requirements
	IsFragile         bool     `json:"is_fragile" bson:"is_fragile"`
	IsPerishable      bool     `json:"is_perishable" bson:"is_perishable"`
	RequiresColdChain bool     `json:"requires_cold_chain" bson:"requires_cold_chain"`
	RequiresUpright   bool     `json:"requires_upright" bson:"requires_upright"`
	SpecialHandling   []string `json:"special_handling" bson:"special_handling"`

	// Documentation
	Photos          []PackagePhoto `json:"photos" bson:"photos"`
	BarcodeID       string         `json:"barcode_id" bson:"barcode_id"`
	ReferenceNumber string         `json:"reference_number" bson:"reference_number"`

	// Restrictions
	AgeRestricted     bool `json:"age_restricted" bson:"age_restricted"`
	RequiresSignature bool `json:"requires_signature" bson:"requires_signature"`
	RequiresID        bool `json:"requires_id" bson:"requires_id"`

	// Packaging
	PackagingType     string `json:"packaging_type" bson:"packaging_type"`
	PackagingMaterial string `json:"packaging_material" bson:"packaging_material"`
	IsOwnPackaging    bool   `json:"is_own_packaging" bson:"is_own_packaging"`
}

type PackageDimensions struct {
	Length float64 `json:"length" bson:"length"` // cm
	Width  float64 `json:"width" bson:"width"`   // cm
	Height float64 `json:"height" bson:"height"` // cm
}

type PackageItem struct {
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	Quantity    int     `json:"quantity" bson:"quantity"`
	Value       float64 `json:"value" bson:"value"`
	Weight      float64 `json:"weight" bson:"weight"`
	Category    string  `json:"category" bson:"category"`
}

type CourierLocation struct {
	// Geographic Location
	Coordinates []float64 `json:"coordinates" bson:"coordinates" validate:"coordinates"`
	Address     string    `json:"address" bson:"address" validate:"required"`
	Name        string    `json:"name" bson:"name"`
	PlaceID     string    `json:"place_id" bson:"place_id"`

	// Detailed Address
	BuildingName       string `json:"building_name" bson:"building_name"`
	Floor              string `json:"floor" bson:"floor"`
	RoomNumber         string `json:"room_number" bson:"room_number"`
	Landmark           string `json:"landmark" bson:"landmark"`
	AccessInstructions string `json:"access_instructions" bson:"access_instructions"`

	// Contact Information
	ContactPerson  string `json:"contact_person" bson:"contact_person"`
	ContactPhone   string `json:"contact_phone" bson:"contact_phone" validate:"phone"`
	AlternatePhone string `json:"alternate_phone" bson:"alternate_phone"`

	// Business Information
	BusinessName  string `json:"business_name" bson:"business_name"`
	BusinessType  string `json:"business_type" bson:"business_type"`
	BusinessHours string `json:"business_hours" bson:"business_hours"`

	// Access & Security
	RequiresAppointment bool   `json:"requires_appointment" bson:"requires_appointment"`
	SecurityCode        string `json:"security_code" bson:"security_code"`
	GateCode            string `json:"gate_code" bson:"gate_code"`
	ParkingInstructions string `json:"parking_instructions" bson:"parking_instructions"`

	// Location Type
	Type         string `json:"type" bson:"type"` // home, office, shop, warehouse, other
	IsCommercial bool   `json:"is_commercial" bson:"is_commercial"`
	HasElevator  bool   `json:"has_elevator" bson:"has_elevator"`
	HasParking   bool   `json:"has_parking" bson:"has_parking"`
}

type CourierRecipient struct {
	// Personal Information
	Name  string `json:"name" bson:"name" validate:"required"`
	Phone string `json:"phone" bson:"phone" validate:"required,phone"`
	Email string `json:"email" bson:"email" validate:"email"`

	// Identification (for age-restricted items)
	IDType   string `json:"id_type" bson:"id_type"`
	IDNumber string `json:"id_number" bson:"id_number"`

	// Delivery Preferences
	PreferredDeliveryTime string   `json:"preferred_delivery_time" bson:"preferred_delivery_time"`
	AvailableTimeSlots    []string `json:"available_time_slots" bson:"available_time_slots"`
	DeliveryInstructions  string   `json:"delivery_instructions" bson:"delivery_instructions"`

	// Alternative Recipients
	AlternateRecipients []AlternateRecipient `json:"alternate_recipients" bson:"alternate_recipients"`

	// Communication Preferences
	SMSNotifications   bool `json:"sms_notifications" bson:"sms_notifications"`
	EmailNotifications bool `json:"email_notifications" bson:"email_notifications"`
	CallNotifications  bool `json:"call_notifications" bson:"call_notifications"`

	// Special Requirements
	RequiresSignature    bool `json:"requires_signature" bson:"requires_signature"`
	RequiresPhoto        bool `json:"requires_photo" bson:"requires_photo"`
	RequiresID           bool `json:"requires_id" bson:"requires_id"`
	CanLeaveWithNeighbor bool `json:"can_leave_with_neighbor" bson:"can_leave_with_neighbor"`
	CanLeaveAtDoor       bool `json:"can_leave_at_door" bson:"can_leave_at_door"`

	// Relationship to Sender
	RelationshipToSender string `json:"relationship_to_sender" bson:"relationship_to_sender"`
}

type AlternateRecipient struct {
	Name         string `json:"name" bson:"name"`
	Phone        string `json:"phone" bson:"phone"`
	Relationship string `json:"relationship" bson:"relationship"`
	CanAccept    bool   `json:"can_accept" bson:"can_accept"`
}

type TimeWindow struct {
	StartTime time.Time `json:"start_time" bson:"start_time"`
	EndTime   time.Time `json:"end_time" bson:"end_time"`
	Type      string    `json:"type" bson:"type"` // morning, afternoon, evening, specific
}

type CourierFare struct {
	// Base Pricing
	BaseFare     float64 `json:"base_fare" bson:"base_fare"`
	WeightFare   float64 `json:"weight_fare" bson:"weight_fare"`
	DistanceFare float64 `json:"distance_fare" bson:"distance_fare"`
	UrgencyFare  float64 `json:"urgency_fare" bson:"urgency_fare"`

	// Special Service Fees
	FragileHandlingFee   float64 `json:"fragile_handling_fee" bson:"fragile_handling_fee"`
	ColdChainFee         float64 `json:"cold_chain_fee" bson:"cold_chain_fee"`
	SignatureRequiredFee float64 `json:"signature_required_fee" bson:"signature_required_fee"`
	WaitingTimeFee       float64 `json:"waiting_time_fee" bson:"waiting_time_fee"`

	// Additional Charges
	FuelSurcharge     float64 `json:"fuel_surcharge" bson:"fuel_surcharge"`
	PeakHourSurcharge float64 `json:"peak_hour_surcharge" bson:"peak_hour_surcharge"`
	WeekendSurcharge  float64 `json:"weekend_surcharge" bson:"weekend_surcharge"`

	// Insurance & Protection
	InsuranceFee float64 `json:"insurance_fee" bson:"insurance_fee"`

	// Discounts
	DiscountAmount float64 `json:"discount_amount" bson:"discount_amount"`
	PromoCode      string  `json:"promo_code" bson:"promo_code"`

	// Total & Commission
	SubTotal        float64 `json:"sub_total" bson:"sub_total"`
	TaxAmount       float64 `json:"tax_amount" bson:"tax_amount"`
	TotalAmount     float64 `json:"total_amount" bson:"total_amount"`
	CourierEarnings float64 `json:"courier_earnings" bson:"courier_earnings"`
	Commission      float64 `json:"commission" bson:"commission"`
	CommissionRate  float64 `json:"commission_rate" bson:"commission_rate"`

	Currency string `json:"currency" bson:"currency"`
}

type TrackingEvent struct {
	ID          primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Status      CourierStatus       `json:"status" bson:"status"`
	Description string              `json:"description" bson:"description"`
	Location    *Location           `json:"location,omitempty" bson:"location,omitempty"`
	Timestamp   time.Time           `json:"timestamp" bson:"timestamp"`
	CreatedBy   *primitive.ObjectID `json:"created_by,omitempty" bson:"created_by,omitempty"`
	Notes       string              `json:"notes" bson:"notes"`
	Photos      []string            `json:"photos" bson:"photos"`
	IsPublic    bool                `json:"is_public" bson:"is_public"`
}

type DeliveryProof struct {
	// Proof Type
	Type string `json:"type" bson:"type"` // signature, photo, otp, id_verification

	// Signature Proof
	SignatureURL string `json:"signature_url" bson:"signature_url"`
	SignedBy     string `json:"signed_by" bson:"signed_by"`

	// Photo Proof
	DeliveryPhotos []string `json:"delivery_photos" bson:"delivery_photos"`
	RecipientPhoto string   `json:"recipient_photo" bson:"recipient_photo"`
	LocationPhoto  string   `json:"location_photo" bson:"location_photo"`

	// OTP Verification
	OTPCode       string     `json:"otp_code" bson:"otp_code"`
	OTPVerifiedAt *time.Time `json:"otp_verified_at,omitempty" bson:"otp_verified_at,omitempty"`

	// ID Verification
	IDType     string `json:"id_type" bson:"id_type"`
	IDNumber   string `json:"id_number" bson:"id_number"`
	IDPhotoURL string `json:"id_photo_url" bson:"id_photo_url"`

	// Delivery Details
	DeliveredTo      string             `json:"delivered_to" bson:"delivered_to"`
	DeliveredBy      primitive.ObjectID `json:"delivered_by" bson:"delivered_by"`
	DeliveryTime     time.Time          `json:"delivery_time" bson:"delivery_time"`
	DeliveryLocation Location           `json:"delivery_location" bson:"delivery_location"`

	// Additional Information
	Condition         string `json:"condition" bson:"condition"` // good, damaged, partial
	RecipientFeedback string `json:"recipient_feedback" bson:"recipient_feedback"`
	DeliveryNotes     string `json:"delivery_notes" bson:"delivery_notes"`

	// Verification Status
	IsVerified bool                `json:"is_verified" bson:"is_verified"`
	VerifiedAt *time.Time          `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	VerifiedBy *primitive.ObjectID `json:"verified_by,omitempty" bson:"verified_by,omitempty"`
}

type CourierInsurance struct {
	IsRequired     bool      `json:"is_required" bson:"is_required"`
	CoverageAmount float64   `json:"coverage_amount" bson:"coverage_amount"`
	Premium        float64   `json:"premium" bson:"premium"`
	Currency       string    `json:"currency" bson:"currency"`
	PolicyNumber   string    `json:"policy_number" bson:"policy_number"`
	ProviderName   string    `json:"provider_name" bson:"provider_name"`
	CoverageType   string    `json:"coverage_type" bson:"coverage_type"` // basic, comprehensive
	Deductible     float64   `json:"deductible" bson:"deductible"`
	ValidFrom      time.Time `json:"valid_from" bson:"valid_from"`
	ValidTo        time.Time `json:"valid_to" bson:"valid_to"`

	// Claim Information
	ClaimProcess   string `json:"claim_process" bson:"claim_process"`
	ClaimDeadline  int    `json:"claim_deadline" bson:"claim_deadline"` // days
	SupportContact string `json:"support_contact" bson:"support_contact"`

	// Exclusions
	Exclusions []string `json:"exclusions" bson:"exclusions"`
	Terms      string   `json:"terms" bson:"terms"`
}
