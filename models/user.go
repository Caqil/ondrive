package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	RolePassenger UserRole = "passenger"
	RoleDriver    UserRole = "driver"
	RoleAdmin     UserRole = "admin"
)


type Gender string

const (
	GenderMale           Gender = "male"
	GenderFemale         Gender = "female"
	GenderOther          Gender = "other"
	GenderPreferNotToSay Gender = "prefer_not_to_say"
)

type UserStatsBulkUpdate struct {
	UserID string    `json:"user_id" bson:"user_id"`
	Stats  UserStats `json:"stats" bson:"stats"`
}
type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Phone        string             `json:"phone" bson:"phone" validate:"required,phone"`
	Email        string             `json:"email" bson:"email" validate:"email"`
	PasswordHash string             `json:"-" bson:"password_hash"`
	Role         UserRole           `json:"role" bson:"role" validate:"required,role"`

	// Profile Information
	Profile      UserProfile      `json:"profile" bson:"profile"`
	Verification UserVerification `json:"verification" bson:"verification"`
	Location     *Location        `json:"location,omitempty" bson:"location,omitempty"`
	Settings     UserSettings     `json:"settings" bson:"settings"`

	// Driver-specific information (only for drivers)
	DriverInfo   *DriverInfo `json:"driver_info,omitempty" bson:"driver_info,omitempty"`
	DistanceUnit string      `json:"distance_unit" bson:"distance_unit"` // km, miles
	// Emergency Contacts
	EmergencyContacts []EmergencyContact `json:"emergency_contacts" bson:"emergency_contacts"`

	// Favorite Places
	FavoritePlaces []FavoritePlace `json:"favorite_places" bson:"favorite_places"`

	// Account Status
	IsActive    bool `json:"is_active" bson:"is_active"`
	IsDeleted   bool `json:"is_deleted" bson:"is_deleted"`
	IsSuspended bool `json:"is_suspended" bson:"is_suspended"`
	IsVerified  bool `json:"is_verified" bson:"is_verified"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`

	// Device Information
	DeviceTokens []DeviceToken `json:"device_tokens" bson:"device_tokens"`

	// Statistics
	Stats UserStats `json:"stats" bson:"stats"`
}

type UserProfile struct {
	FirstName    string     `json:"first_name" bson:"first_name" validate:"required,min=2,max=50"`
	LastName     string     `json:"last_name" bson:"last_name" validate:"required,min=2,max=50"`
	FullName     string     `json:"full_name" bson:"full_name"`
	AvatarURL    string     `json:"avatar_url" bson:"avatar_url"`
	DateOfBirth  *time.Time `json:"date_of_birth,omitempty" bson:"date_of_birth,omitempty"`
	Gender       Gender     `json:"gender" bson:"gender" validate:"gender"`
	Language     string     `json:"language" bson:"language"`
	CountryCode  string     `json:"country_code" bson:"country_code"`
	City         string     `json:"city" bson:"city"`
	State        string     `json:"state" bson:"state"`
	Country      string     `json:"country" bson:"country"`
	PostalCode   string     `json:"postal_code" bson:"postal_code"`
	Bio          string     `json:"bio" bson:"bio" validate:"max=500"`
	ProfileViews int        `json:"profile_views" bson:"profile_views"`
}

type UserVerification struct {
	PhoneVerified    bool                `json:"phone_verified" bson:"phone_verified"`
	EmailVerified    bool                `json:"email_verified" bson:"email_verified"`
	IdentityVerified bool                `json:"identity_verified" bson:"identity_verified"`
	Documents        []VerificationDoc   `json:"documents" bson:"documents"`
	VerifiedAt       *time.Time          `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	VerifiedBy       *primitive.ObjectID `json:"verified_by,omitempty" bson:"verified_by,omitempty"`
	RejectionReason  string              `json:"rejection_reason,omitempty" bson:"rejection_reason,omitempty"`
}

type VerificationDoc struct {
	ID         primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Type       string              `json:"type" bson:"type" validate:"required"`
	URL        string              `json:"url" bson:"url" validate:"required,url"`
	Status     VerificationStatus  `json:"status" bson:"status"`
	UploadedAt time.Time           `json:"uploaded_at" bson:"uploaded_at"`
	ReviewedAt *time.Time          `json:"reviewed_at,omitempty" bson:"reviewed_at,omitempty"`
	ReviewedBy *primitive.ObjectID `json:"reviewed_by,omitempty" bson:"reviewed_by,omitempty"`
	Notes      string              `json:"notes" bson:"notes"`
	ExpiryDate *time.Time          `json:"expiry_date,omitempty" bson:"expiry_date,omitempty"`
}

type UserSettings struct {
	// Notification Preferences
	NotificationsEnabled     bool `json:"notifications_enabled" bson:"notifications_enabled"`
	PushNotifications        bool `json:"push_notifications" bson:"push_notifications"`
	EmailNotifications       bool `json:"email_notifications" bson:"email_notifications"`
	SMSNotifications         bool `json:"sms_notifications" bson:"sms_notifications"`
	RideNotifications        bool `json:"ride_notifications" bson:"ride_notifications"`
	PromotionalNotifications bool `json:"promotional_notifications" bson:"promotional_notifications"`

	// App Preferences
	Language         string `json:"language" bson:"language"`
	Currency         string `json:"currency" bson:"currency"`
	Theme            string `json:"theme" bson:"theme"`
	AutoAcceptRides  bool   `json:"auto_accept_rides" bson:"auto_accept_rides"`
	ShowOnlineStatus bool   `json:"show_online_status" bson:"show_online_status"`
	DistanceUnit     string `json:"distance_unit" bson:"distance_unit"` // km, miles
	// Privacy Settings
	PrivacySettings PrivacySettings `json:"privacy_settings" bson:"privacy_settings"`

	// Location Settings
	ShareLocation       bool   `json:"share_location" bson:"share_location"`
	LocationAccuracy    string `json:"location_accuracy" bson:"location_accuracy"`
	SaveLocationHistory bool   `json:"save_location_history" bson:"save_location_history"`

	// Payment Settings
	DefaultPaymentMethod string `json:"default_payment_method" bson:"default_payment_method"`
	AutoPayEnabled       bool   `json:"auto_pay_enabled" bson:"auto_pay_enabled"`
	SavePaymentMethods   bool   `json:"save_payment_methods" bson:"save_payment_methods"`
}

type PrivacySettings struct {
	ShareLocation      bool `json:"share_location" bson:"share_location"`
	ShowPhoneNumber    bool `json:"show_phone_number" bson:"show_phone_number"`
	AllowCalls         bool `json:"allow_calls" bson:"allow_calls"`
	ShowLastSeen       bool `json:"show_last_seen" bson:"show_last_seen"`
	ShowProfilePicture bool `json:"show_profile_picture" bson:"show_profile_picture"`
	AllowMessages      bool `json:"allow_messages" bson:"allow_messages"`
	ShowOnlineStatus   bool `json:"show_online_status" bson:"show_online_status"`
}

type Location struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates" validate:"coordinates"`
	Address     string    `json:"address" bson:"address"`
	Name        string    `json:"name" bson:"name"`
	City        string    `json:"city" bson:"city"`
	State       string    `json:"state" bson:"state"`
	Country     string    `json:"country" bson:"country"`
	PostalCode  string    `json:"postal_code" bson:"postal_code"`
	PlaceID     string    `json:"place_id" bson:"place_id"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	Accuracy    float64   `json:"accuracy" bson:"accuracy"`
	Heading     float64   `json:"heading" bson:"heading"`
	Speed       float64   `json:"speed" bson:"speed"`
}

type EmergencyContact struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email        string             `json:"email"`
	Name         string             `json:"name" bson:"name" validate:"required"`
	Phone        string             `json:"phone" bson:"phone" validate:"required,phone"`
	Relationship string             `json:"relationship" bson:"relationship"`
	IsPrimary    bool               `json:"is_primary" bson:"is_primary"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
}

type FavoritePlace struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name" validate:"required"`
	Address    string             `json:"address" bson:"address" validate:"required"`
	Location   Location           `json:"location" bson:"location"`
	Type       string             `json:"type" bson:"type"` // home, work, other
	Icon       string             `json:"icon" bson:"icon"`
	IsDefault  bool               `json:"is_default" bson:"is_default"`
	UsageCount int                `json:"usage_count" bson:"usage_count"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

type DeviceToken struct {
	Token      string    `json:"token" bson:"token"`
	Platform   string    `json:"platform" bson:"platform"` // ios, android, web
	DeviceID   string    `json:"device_id" bson:"device_id"`
	AppVersion string    `json:"app_version" bson:"app_version"`
	IsActive   bool      `json:"is_active" bson:"is_active"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

type UserStats struct {
	TotalRides          int     `json:"total_rides" bson:"total_rides"`
	CompletedRides      int     `json:"completed_rides" bson:"completed_rides"`
	CancelledRides      int     `json:"cancelled_rides" bson:"cancelled_rides"`
	AverageRating       float64 `json:"average_rating" bson:"average_rating"`
	TotalRatings        int     `json:"total_ratings" bson:"total_ratings"`
	TotalSpent          float64 `json:"total_spent" bson:"total_spent"`
	TotalSaved          float64 `json:"total_saved" bson:"total_saved"`
	CancellationRate    float64 `json:"cancellation_rate" bson:"cancellation_rate"`
	OnTimeRate          float64 `json:"on_time_rate" bson:"on_time_rate"`
	FavoriteServiceType string  `json:"favorite_service_type" bson:"favorite_service_type"`
	JoinedDaysAgo       int     `json:"joined_days_ago" bson:"joined_days_ago"`
}

type UpdateProfileRequest struct {
	FirstName   *string    `json:"first_name,omitempty"`
	LastName    *string    `json:"last_name,omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Gender      *string    `json:"gender,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	Language    *string    `json:"language,omitempty"`
	City        *string    `json:"city,omitempty"`
	State       *string    `json:"state,omitempty"`
	Country     *string    `json:"country,omitempty"`
	PostalCode  *string    `json:"postal_code,omitempty"`
}

type DeleteProfileRequest struct {
	Password string `json:"password" binding:"required"`
	Reason   string `json:"reason"`
}

type UpdateLocationRequest struct {
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	Country    string  `json:"country"`
	PostalCode string  `json:"postal_code"`
}

type UpdateSettingsRequest struct {
	Language     *string `json:"language,omitempty"`
	Currency     *string `json:"currency,omitempty"`
	Theme        *string `json:"theme,omitempty"`
	DistanceUnit *string `json:"distance_unit,omitempty"`
}

type UpdatePrivacySettingsRequest struct {
	ShowLastSeen         *bool `json:"show_last_seen,omitempty"`
	ShowOnlineStatus     *bool `json:"show_online_status,omitempty"`
	ShowPhoneNumber      *bool `json:"show_phone_number,omitempty"`
	AllowLocationSharing *bool `json:"allow_location_sharing,omitempty"`
}

type UpdateNotificationSettingsRequest struct {
	NotificationsEnabled     *bool `json:"notifications_enabled,omitempty"`
	PushNotifications        *bool `json:"push_notifications,omitempty"`
	EmailNotifications       *bool `json:"email_notifications,omitempty"`
	SMSNotifications         *bool `json:"sms_notifications,omitempty"`
	RideNotifications        *bool `json:"ride_notifications,omitempty"`
	PromotionalNotifications *bool `json:"promotional_notifications,omitempty"`
}

type AddEmergencyContactRequest struct {
	Name         string `json:"name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Relationship string `json:"relationship" binding:"required"`
	Email        string `json:"email"`
}

type UpdateEmergencyContactRequest struct {
	Name         *string `json:"name,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	Relationship *string `json:"relationship,omitempty"`
	Email        *string `json:"email,omitempty"`
}

type AddFavoritePlaceRequest struct {
	Name       string  `json:"name" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	Icon       string  `json:"icon"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	Country    string  `json:"country"`
	PostalCode string  `json:"postal_code"`
}

type UpdateFavoritePlaceRequest struct {
	Name       *string  `json:"name,omitempty"`
	Type       *string  `json:"type,omitempty"`
	Icon       *string  `json:"icon,omitempty"`
	Latitude   *float64 `json:"latitude,omitempty"`
	Longitude  *float64 `json:"longitude,omitempty"`
	Address    *string  `json:"address,omitempty"`
	City       *string  `json:"city,omitempty"`
	State      *string  `json:"state,omitempty"`
	Country    *string  `json:"country,omitempty"`
	PostalCode *string  `json:"postal_code,omitempty"`
}
