package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SavedAddress represents a user's saved address for courier services
type SavedAddress struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	Label       string             `json:"label" bson:"label"` // e.g., "Home", "Office", "Mom's House"
	Name        string             `json:"name" bson:"name"`   // Contact name for this address
	Phone       string             `json:"phone" bson:"phone"`
	Address     string             `json:"address" bson:"address"`
	City        string             `json:"city" bson:"city"`
	State       string             `json:"state" bson:"state"`
	Country     string             `json:"country" bson:"country"`
	PostalCode  string             `json:"postal_code" bson:"postal_code"`
	Location    Location           `json:"location" bson:"location"`
	IsDefault   bool               `json:"is_default" bson:"is_default"`
	AddressType string             `json:"address_type" bson:"address_type"` // home, office, other

	// Additional details
	Apartment string `json:"apartment" bson:"apartment"`
	Building  string `json:"building" bson:"building"`
	Floor     string `json:"floor" bson:"floor"`
	Landmark  string `json:"landmark" bson:"landmark"`

	// Delivery instructions
	DeliveryInstructions string `json:"delivery_instructions" bson:"delivery_instructions"`
	AccessCode           string `json:"access_code" bson:"access_code"`

	// Metadata
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// CourierIssue represents an issue reported for a courier request
type CourierIssue struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RequestID   primitive.ObjectID `json:"request_id" bson:"request_id"`
	ReporterID  primitive.ObjectID `json:"reporter_id" bson:"reporter_id"`
	IssueType   string             `json:"issue_type" bson:"issue_type"` // package_damage, delay, lost_package, wrong_delivery, etc.
	Priority    string             `json:"priority" bson:"priority"`     // low, medium, high, critical
	Status      string             `json:"status" bson:"status"`         // open, in_progress, resolved, closed
	Subject     string             `json:"subject" bson:"subject"`
	Description string             `json:"description" bson:"description"`

	// Evidence
	Photos    []string `json:"photos" bson:"photos"`
	Documents []string `json:"documents" bson:"documents"`

	// Resolution
	Resolution string              `json:"resolution" bson:"resolution"`
	ResolvedBy *primitive.ObjectID `json:"resolved_by,omitempty" bson:"resolved_by,omitempty"`
	ResolvedAt *time.Time          `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`

	// Communication
	ConversationID *primitive.ObjectID `json:"conversation_id,omitempty" bson:"conversation_id,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// CourierClaim represents a claim filed for a courier request
type CourierClaim struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RequestID  primitive.ObjectID `json:"request_id" bson:"request_id"`
	ClaimantID primitive.ObjectID `json:"claimant_id" bson:"claimant_id"`
	ClaimType  string             `json:"claim_type" bson:"claim_type"` // damage, loss, delay, etc.
	Status     string             `json:"status" bson:"status"`         // pending, approved, rejected, paid

	// Claim Details
	Amount      float64 `json:"amount" bson:"amount"`
	Currency    string  `json:"currency" bson:"currency"`
	Description string  `json:"description" bson:"description"`

	// Evidence
	Evidence []string `json:"evidence" bson:"evidence"` // Photo URLs, document URLs

	// Processing
	ProcessedBy    *primitive.ObjectID `json:"processed_by,omitempty" bson:"processed_by,omitempty"`
	ProcessedAt    *time.Time          `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	PaymentMethod  string              `json:"payment_method" bson:"payment_method"`
	PaymentDetails map[string]string   `json:"payment_details" bson:"payment_details"`

	// Rejection details
	RejectionReason string `json:"rejection_reason" bson:"rejection_reason"`

	// Metadata
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}


// PricingService interface (placeholder for dependency injection)
type PricingService interface {
	CalculatePrice(request interface{}) (float64, error)
	GetPricingRules() (interface{}, error)
	ValidatePromoCode(code string, amount float64) (float64, error)
}

// NotificationService interface (placeholder for dependency injection)
type NotificationService interface {
	SendSMS(phone, message string) error
	SendEmail(email, subject, message string) error
	SendPushNotification(userID primitive.ObjectID, message string) error
	SendPushNotificationByContact(phone, message string) error
	NotifyRequestAccepted(senderID, requestID, courierID primitive.ObjectID) error
	NotifyRequestCancelled(userID, requestID primitive.ObjectID, reason string) error
	NotifyPickupConfirmed(senderID, requestID primitive.ObjectID) error
	NotifyDeliveryCompleted(senderID, requestID primitive.ObjectID) error
}
