

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	NotificationTypeRide         NotificationType = "ride"
	NotificationTypePayment      NotificationType = "payment"
	NotificationTypePromotion    NotificationType = "promotion"
	NotificationTypeSystem       NotificationType = "system"
	NotificationTypeChat         NotificationType = "chat"
	NotificationTypeRating       NotificationType = "rating"
	NotificationTypeEmergency    NotificationType = "emergency"
	NotificationTypeDocument     NotificationType = "document"
	NotificationTypeEarnings     NotificationType = "earnings"
	NotificationTypeSupport      NotificationType = "support"
	NotificationTypeAccount      NotificationType = "account"
	NotificationTypeWeather      NotificationType = "weather"
	NotificationTypeTraffic      NotificationType = "traffic"
	NotificationTypeMaintenance  NotificationType = "maintenance"
)

type NotificationChannel string

const (
	NotificationChannelPush  NotificationChannel = "push"
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelInApp NotificationChannel = "in_app"
	NotificationChannelCall  NotificationChannel = "call"
)

type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusRead      NotificationStatus = "read"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
	NotificationPriorityEmergency NotificationPriority = "emergency"
)

type Notification struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	
	// Recipient Information
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	UserType    string             `json:"user_type" bson:"user_type"` // passenger, driver, admin
	
	// Notification Content
	Type        NotificationType   `json:"type" bson:"type"`
	Title       string             `json:"title" bson:"title" validate:"required"`
	Body        string             `json:"body" bson:"body" validate:"required"`
	Summary     string             `json:"summary" bson:"summary"`
	
	// Metadata
	Priority    NotificationPriority `json:"priority" bson:"priority"`
	Category    string               `json:"category" bson:"category"`
	Tags        []string             `json:"tags" bson:"tags"`
	
	// Related Entities
	RelatedEntityID   *primitive.ObjectID `json:"related_entity_id,omitempty" bson:"related_entity_id,omitempty"`
	RelatedEntityType string              `json:"related_entity_type" bson:"related_entity_type"` // ride, payment, user
	RideID            *primitive.ObjectID `json:"ride_id,omitempty" bson:"ride_id,omitempty"`
	
	// Action Information
	ActionURL         string                 `json:"action_url" bson:"action_url"`
	ActionText        string                 `json:"action_text" bson:"action_text"`
	DeepLink          string                 `json:"deep_link" bson:"deep_link"`
	ActionData        map[string]interface{} `json:"action_data" bson:"action_data"`
	
	// Media
	ImageURL          string                 `json:"image_url" bson:"image_url"`
	IconURL           string                 `json:"icon_url" bson:"icon_url"`
	SoundURL          string                 `json:"sound_url" bson:"sound_url"`
	
	// Delivery Channels
	Channels          []NotificationChannel  `json:"channels" bson:"channels"`
	DeliveryResults   []DeliveryResult       `json:"delivery_results" bson:"delivery_results"`
	
	// Status & Tracking
	Status            NotificationStatus     `json:"status" bson:"status"`
	IsRead            bool                   `json:"is_read" bson:"is_read"`
	ReadAt            *time.Time             `json:"read_at,omitempty" bson:"read_at,omitempty"`
	IsDeleted         bool                   `json:"is_deleted" bson:"is_deleted"`
	DeletedAt         *time.Time             `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	
	// Scheduling
	ScheduledAt       *time.Time             `json:"scheduled_at,omitempty" bson:"scheduled_at,omitempty"`
	SentAt            *time.Time             `json:"sent_at,omitempty" bson:"sent_at,omitempty"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	
	// Retry Logic
	RetryCount        int                    `json:"retry_count" bson:"retry_count"`
	MaxRetries        int                    `json:"max_retries" bson:"max_retries"`
	LastRetryAt       *time.Time             `json:"last_retry_at,omitempty" bson:"last_retry_at,omitempty"`
	
	// Personalization
	Language          string                 `json:"language" bson:"language"`
	TimeZone          string                 `json:"time_zone" bson:"time_zone"`
	LocalizedContent  map[string]LocalizedContent `json:"localized_content" bson:"localized_content"`
	
	// Analytics
	ClickCount        int                    `json:"click_count" bson:"click_count"`
	ClickedAt         *time.Time             `json:"clicked_at,omitempty" bson:"clicked_at,omitempty"`
	
	// Batch Information
	BatchID           string                 `json:"batch_id" bson:"batch_id"`
	CampaignID        string                 `json:"campaign_id" bson:"campaign_id"`
	
	// Source Information
	SourceService     string                 `json:"source_service" bson:"source_service"`
	SourceEvent       string                 `json:"source_event" bson:"source_event"`
	
	CreatedAt         time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" bson:"updated_at"`
}

type DeliveryResult struct {
	Channel       NotificationChannel `json:"channel" bson:"channel"`
	Status        NotificationStatus  `json:"status" bson:"status"`
	DeliveredAt   *time.Time          `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`
	FailureReason string              `json:"failure_reason" bson:"failure_reason"`
	ExternalID    string              `json:"external_id" bson:"external_id"` // Provider message ID
	Provider      string              `json:"provider" bson:"provider"`       // firebase, twilio, sendgrid
	Cost          float64             `json:"cost" bson:"cost"`
	RetryCount    int                 `json:"retry_count" bson:"retry_count"`
}

type LocalizedContent struct {
	Title    string `json:"title" bson:"title"`
	Body     string `json:"body" bson:"body"`
	Summary  string `json:"summary" bson:"summary"`
	ActionText string `json:"action_text" bson:"action_text"`
}

type NotificationTemplate struct {
	ID              primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name            string               `json:"name" bson:"name" validate:"required"`
	Type            NotificationType     `json:"type" bson:"type"`
	Category        string               `json:"category" bson:"category"`
	
	// Content Template
	TitleTemplate   string               `json:"title_template" bson:"title_template"`
	BodyTemplate    string               `json:"body_template" bson:"body_template"`
	SummaryTemplate string               `json:"summary_template" bson:"summary_template"`
	
	// Variables
	Variables       []TemplateVariable   `json:"variables" bson:"variables"`
	SampleData      map[string]interface{} `json:"sample_data" bson:"sample_data"`
	
	// Default Settings
	DefaultPriority NotificationPriority `json:"default_priority" bson:"default_priority"`
	DefaultChannels []NotificationChannel `json:"default_channels" bson:"default_channels"`
	DefaultExpiry   time.Duration        `json:"default_expiry" bson:"default_expiry"`
	
	// Localization
	SupportedLanguages []string          `json:"supported_languages" bson:"supported_languages"`
	LocalizedTemplates map[string]LocalizedTemplate `json:"localized_templates" bson:"localized_templates"`
	
	// Media
	DefaultImageURL string               `json:"default_image_url" bson:"default_image_url"`
	DefaultIconURL  string               `json:"default_icon_url" bson:"default_icon_url"`
	DefaultSoundURL string               `json:"default_sound_url" bson:"default_sound_url"`
	
	// Targeting
	TargetUserTypes []string             `json:"target_user_types" bson:"target_user_types"`
	TargetPlatforms []string             `json:"target_platforms" bson:"target_platforms"`
	
	// Status & Usage
	IsActive        bool                 `json:"is_active" bson:"is_active"`
	UsageCount      int                  `json:"usage_count" bson:"usage_count"`
	LastUsedAt      *time.Time           `json:"last_used_at,omitempty" bson:"last_used_at,omitempty"`
	
	// Metadata
	Version         int                  `json:"version" bson:"version"`
	CreatedBy       primitive.ObjectID   `json:"created_by" bson:"created_by"`
	
	CreatedAt       time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" bson:"updated_at"`
}

type TemplateVariable struct {
	Name         string      `json:"name" bson:"name"`
	Type         string      `json:"type" bson:"type"` // string, number, date, boolean
	Description  string      `json:"description" bson:"description"`
	DefaultValue interface{} `json:"default_value" bson:"default_value"`
	IsRequired   bool        `json:"is_required" bson:"is_required"`
	Format       string      `json:"format" bson:"format"` // For date/number formatting
}

type LocalizedTemplate struct {
	TitleTemplate   string `json:"title_template" bson:"title_template"`
	BodyTemplate    string `json:"body_template" bson:"body_template"`
	SummaryTemplate string `json:"summary_template" bson:"summary_template"`
	ActionText      string `json:"action_text" bson:"action_text"`
}

type NotificationPreferences struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID                primitive.ObjectID `json:"user_id" bson:"user_id"`
	
	// General Settings
	GlobalNotifications   bool               `json:"global_notifications" bson:"global_notifications"`
	QuietHoursEnabled     bool               `json:"quiet_hours_enabled" bson:"quiet_hours_enabled"`
	QuietHoursStart       string             `json:"quiet_hours_start" bson:"quiet_hours_start"` // "22:00"
	QuietHoursEnd         string             `json:"quiet_hours_end" bson:"quiet_hours_end"`     // "08:00"
	TimeZone              string             `json:"time_zone" bson:"time_zone"`
	
	// Channel Preferences
	PushNotifications     ChannelPreference  `json:"push_notifications" bson:"push_notifications"`
	SMSNotifications      ChannelPreference  `json:"sms_notifications" bson:"sms_notifications"`
	EmailNotifications    ChannelPreference  `json:"email_notifications" bson:"email_notifications"`
	InAppNotifications    ChannelPreference  `json:"in_app_notifications" bson:"in_app_notifications"`
	
	// Type-specific Preferences
	RideNotifications     TypePreference     `json:"ride_notifications" bson:"ride_notifications"`
	PaymentNotifications  TypePreference     `json:"payment_notifications" bson:"payment_notifications"`
	ChatNotifications     TypePreference     `json:"chat_notifications" bson:"chat_notifications"`
	PromotionalNotifications TypePreference  `json:"promotional_notifications" bson:"promotional_notifications"`
	SystemNotifications   TypePreference     `json:"system_notifications" bson:"system_notifications"`
	EmergencyNotifications TypePreference    `json:"emergency_notifications" bson:"emergency_notifications"`
	EarningsNotifications TypePreference     `json:"earnings_notifications" bson:"earnings_notifications"`
	
	// Advanced Settings
	GroupSimilarNotifications bool           `json:"group_similar_notifications" bson:"group_similar_notifications"`
	NotificationSound     string             `json:"notification_sound" bson:"notification_sound"`
	VibrationEnabled      bool               `json:"vibration_enabled" bson:"vibration_enabled"`
	LEDEnabled            bool               `json:"led_enabled" bson:"led_enabled"`
	
	// Frequency Settings
	MaxNotificationsPerHour int              `json:"max_notifications_per_hour" bson:"max_notifications_per_hour"`
	DigestMode            bool               `json:"digest_mode" bson:"digest_mode"`
	DigestFrequency       string             `json:"digest_frequency" bson:"digest_frequency"` // hourly, daily, weekly
	
	UpdatedAt             time.Time          `json:"updated_at" bson:"updated_at"`
}

type ChannelPreference struct {
	Enabled     bool     `json:"enabled" bson:"enabled"`
	Priority    []string `json:"priority" bson:"priority"` // Order of notification types by priority
	QuietHours  bool     `json:"quiet_hours" bson:"quiet_hours"`
	MinPriority NotificationPriority `json:"min_priority" bson:"min_priority"`
}

type TypePreference struct {
	Enabled  bool                    `json:"enabled" bson:"enabled"`
	Channels []NotificationChannel   `json:"channels" bson:"channels"`
	Priority NotificationPriority    `json:"priority" bson:"priority"`
	Sound    string                  `json:"sound" bson:"sound"`
}

type NotificationCampaign struct {
	ID                primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name              string               `json:"name" bson:"name"`
	Description       string               `json:"description" bson:"description"`
	Type              string               `json:"type" bson:"type"` // promotional, informational, transactional
	
	// Content
	TemplateID        *primitive.ObjectID  `json:"template_id,omitempty" bson:"template_id,omitempty"`
	Title             string               `json:"title" bson:"title"`
	Body              string               `json:"body" bson:"body"`
	ImageURL          string               `json:"image_url" bson:"image_url"`
	ActionURL         string               `json:"action_url" bson:"action_url"`
	ActionText        string               `json:"action_text" bson:"action_text"`
	
	// Targeting
	TargetAudience    TargetAudience       `json:"target_audience" bson:"target_audience"`
	SegmentFilters    []SegmentFilter      `json:"segment_filters" bson:"segment_filters"`
	
	// Delivery
	Channels          []NotificationChannel `json:"channels" bson:"channels"`
	Priority          NotificationPriority  `json:"priority" bson:"priority"`
	ScheduledAt       *time.Time           `json:"scheduled_at,omitempty" bson:"scheduled_at,omitempty"`
	ExpiresAt         *time.Time           `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	
	// Rate Limiting
	MaxRecipientsPerHour int               `json:"max_recipients_per_hour" bson:"max_recipients_per_hour"`
	RateLimitEnabled     bool              `json:"rate_limit_enabled" bson:"rate_limit_enabled"`
	
	// Status & Results
	Status            string               `json:"status" bson:"status"` // draft, scheduled, running, paused, completed, cancelled
	TotalRecipients   int                  `json:"total_recipients" bson:"total_recipients"`
	SentCount         int                  `json:"sent_count" bson:"sent_count"`
	DeliveredCount    int                  `json:"delivered_count" bson:"delivered_count"`
	ReadCount         int                  `json:"read_count" bson:"read_count"`
	ClickCount        int                  `json:"click_count" bson:"click_count"`
	FailureCount      int                  `json:"failure_count" bson:"failure_count"`
	
	// A/B Testing
	IsABTest          bool                 `json:"is_ab_test" bson:"is_ab_test"`
	ABTestVariants    []ABTestVariant      `json:"ab_test_variants" bson:"ab_test_variants"`
	
	// Analytics
	Analytics         CampaignAnalytics    `json:"analytics" bson:"analytics"`
	
	CreatedBy         primitive.ObjectID   `json:"created_by" bson:"created_by"`
	CreatedAt         time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at" bson:"updated_at"`
	StartedAt         *time.Time           `json:"started_at,omitempty" bson:"started_at,omitempty"`
	CompletedAt       *time.Time           `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
}

type TargetAudience struct {
	UserTypes       []string `json:"user_types" bson:"user_types"`
	Cities          []string `json:"cities" bson:"cities"`
	Countries       []string `json:"countries" bson:"countries"`
	Languages       []string `json:"languages" bson:"languages"`
	Platforms       []string `json:"platforms" bson:"platforms"`
	MinAppVersion   string   `json:"min_app_version" bson:"min_app_version"`
	RegistrationDateFrom *time.Time `json:"registration_date_from,omitempty" bson:"registration_date_from,omitempty"`
	RegistrationDateTo   *time.Time `json:"registration_date_to,omitempty" bson:"registration_date_to,omitempty"`
	LastActiveFrom  *time.Time `json:"last_active_from,omitempty" bson:"last_active_from,omitempty"`
	LastActiveTo    *time.Time `json:"last_active_to,omitempty" bson:"last_active_to,omitempty"`
}

type SegmentFilter struct {
	Field     string      `json:"field" bson:"field"`
	Operator  string      `json:"operator" bson:"operator"` // equals, not_equals, greater_than, less_than, contains, in, not_in
	Value     interface{} `json:"value" bson:"value"`
	LogicGate string      `json:"logic_gate" bson:"logic_gate"` // and, or
}

type ABTestVariant struct {
	Name          string  `json:"name" bson:"name"`
	Title         string  `json:"title" bson:"title"`
	Body          string  `json:"body" bson:"body"`
	ImageURL      string  `json:"image_url" bson:"image_url"`
	ActionText    string  `json:"action_text" bson:"action_text"`
	TrafficSplit  float64 `json:"traffic_split" bson:"traffic_split"` // 0.0 to 1.0
	SentCount     int     `json:"sent_count" bson:"sent_count"`
	ClickCount    int     `json:"click_count" bson:"click_count"`
	ConversionRate float64 `json:"conversion_rate" bson:"conversion_rate"`
}

type CampaignAnalytics struct {
	DeliveryRate   float64            `json:"delivery_rate" bson:"delivery_rate"`
	OpenRate       float64            `json:"open_rate" bson:"open_rate"`
	ClickRate      float64            `json:"click_rate" bson:"click_rate"`
	ConversionRate float64            `json:"conversion_rate" bson:"conversion_rate"`
	UnsubscribeRate float64           `json:"unsubscribe_rate" bson:"unsubscribe_rate"`
	BounceRate     float64            `json:"bounce_rate" bson:"bounce_rate"`
	Revenue        float64            `json:"revenue" bson:"revenue"`
	ROI            float64            `json:"roi" bson:"roi"`
	
	// Channel-specific analytics
	ChannelMetrics map[string]ChannelMetrics `json:"channel_metrics" bson:"channel_metrics"`
	
	// Engagement over time
	HourlyEngagement  []HourlyMetric `json:"hourly_engagement" bson:"hourly_engagement"`
	DailyEngagement   []DailyMetric  `json:"daily_engagement" bson:"daily_engagement"`
	
	// Geographic analytics
	CountryMetrics    []CountryMetric `json:"country_metrics" bson:"country_metrics"`
	CityMetrics       []CityMetric    `json:"city_metrics" bson:"city_metrics"`
	
	// Device & Platform
	PlatformMetrics   []PlatformMetric `json:"platform_metrics" bson:"platform_metrics"`
	DeviceMetrics     []DeviceMetric   `json:"device_metrics" bson:"device_metrics"`
	
	LastUpdatedAt     time.Time        `json:"last_updated_at" bson:"last_updated_at"`
}

type ChannelMetrics struct {
	Channel       NotificationChannel `json:"channel" bson:"channel"`
	SentCount     int                 `json:"sent_count" bson:"sent_count"`
	DeliveredCount int                `json:"delivered_count" bson:"delivered_count"`
	ReadCount     int                 `json:"read_count" bson:"read_count"`
	ClickCount    int                 `json:"click_count" bson:"click_count"`
	FailureCount  int                 `json:"failure_count" bson:"failure_count"`
	Cost          float64             `json:"cost" bson:"cost"`
	DeliveryRate  float64             `json:"delivery_rate" bson:"delivery_rate"`
	OpenRate      float64             `json:"open_rate" bson:"open_rate"`
	ClickRate     float64             `json:"click_rate" bson:"click_rate"`
}

type HourlyMetric struct {
	Hour       int     `json:"hour" bson:"hour"`
	SentCount  int     `json:"sent_count" bson:"sent_count"`
	ReadCount  int     `json:"read_count" bson:"read_count"`
	ClickCount int     `json:"click_count" bson:"click_count"`
	OpenRate   float64 `json:"open_rate" bson:"open_rate"`
}

type DailyMetric struct {
	Date       time.Time `json:"date" bson:"date"`
	SentCount  int       `json:"sent_count" bson:"sent_count"`
	ReadCount  int       `json:"read_count" bson:"read_count"`
	ClickCount int       `json:"click_count" bson:"click_count"`
	Revenue    float64   `json:"revenue" bson:"revenue"`
}

type CountryMetric struct {
	Country    string  `json:"country" bson:"country"`
	SentCount  int     `json:"sent_count" bson:"sent_count"`
	ReadCount  int     `json:"read_count" bson:"read_count"`
	ClickCount int     `json:"click_count" bson:"click_count"`
	OpenRate   float64 `json:"open_rate" bson:"open_rate"`
}

type CityMetric struct {
	City       string  `json:"city" bson:"city"`
	Country    string  `json:"country" bson:"country"`
	SentCount  int     `json:"sent_count" bson:"sent_count"`
	ReadCount  int     `json:"read_count" bson:"read_count"`
	ClickCount int     `json:"click_count" bson:"click_count"`
	OpenRate   float64 `json:"open_rate" bson:"open_rate"`
}

type PlatformMetric struct {
	Platform   string  `json:"platform" bson:"platform"`
	SentCount  int     `json:"sent_count" bson:"sent_count"`
	ReadCount  int     `json:"read_count" bson:"read_count"`
	ClickCount int     `json:"click_count" bson:"click_count"`
	OpenRate   float64 `json:"open_rate" bson:"open_rate"`
}

type DeviceMetric struct {
	DeviceType string  `json:"device_type" bson:"device_type"`
	SentCount  int     `json:"sent_count" bson:"sent_count"`
	ReadCount  int     `json:"read_count" bson:"read_count"`
	ClickCount int     `json:"click_count" bson:"click_count"`
	OpenRate   float64 `json:"open_rate" bson:"open_rate"`
}