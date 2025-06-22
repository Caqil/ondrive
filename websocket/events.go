package websocket

import (
	"time"

	"ondrive/models"
	"ondrive/utils"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      models.WSEventType     `json:"type"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	MessageID string                 `json:"message_id"`
	UserID    string                 `json:"user_id,omitempty"`
	RideID    string                 `json:"ride_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Welcome message data
type WelcomeData struct {
	Message    string    `json:"message"`
	ServerTime time.Time `json:"server_time"`
	Version    string    `json:"version"`
	Features   []string  `json:"features"`
}

// Fare Negotiation Events
type FareProposedData struct {
	RideID        string    `json:"ride_id"`
	ProposedBy    string    `json:"proposed_by"`
	ProposedTo    string    `json:"proposed_to"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Message       string    `json:"message,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	SuggestedFare float64   `json:"suggested_fare,omitempty"`
	MarketRate    float64   `json:"market_rate,omitempty"`
	OfferCount    int       `json:"offer_count"`
}

type FareAcceptedData struct {
	RideID      string    `json:"ride_id"`
	AcceptedBy  string    `json:"accepted_by"`
	FinalAmount float64   `json:"final_amount"`
	Currency    string    `json:"currency"`
	AcceptedAt  time.Time `json:"accepted_at"`
	Message     string    `json:"message,omitempty"`
}

type FareRejectedData struct {
	RideID     string    `json:"ride_id"`
	RejectedBy string    `json:"rejected_by"`
	RejectedAt time.Time `json:"rejected_at"`
	Reason     string    `json:"reason,omitempty"`
	Message    string    `json:"message,omitempty"`
}

type FareCounteredData struct {
	RideID         string    `json:"ride_id"`
	CounteredBy    string    `json:"countered_by"`
	CounteredTo    string    `json:"countered_to"`
	OriginalAmount float64   `json:"original_amount"`
	CounterAmount  float64   `json:"counter_amount"`
	Currency       string    `json:"currency"`
	Message        string    `json:"message,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// Ride Events
type RideRequestedData struct {
	RideID            string                  `json:"ride_id"`
	PassengerID       string                  `json:"passenger_id"`
	PassengerName     string                  `json:"passenger_name"`
	PassengerRating   float64                 `json:"passenger_rating"`
	PickupLocation    models.RideLocation     `json:"pickup_location"`
	DropoffLocation   models.RideLocation     `json:"dropoff_location"`
	ServiceType       models.ServiceType      `json:"service_type"`
	EstimatedDistance float64                 `json:"estimated_distance"`
	EstimatedDuration int                     `json:"estimated_duration"`
	ProposedFare      float64                 `json:"proposed_fare"`
	Currency          string                  `json:"currency"`
	Requirements      models.RideRequirements `json:"requirements"`
	RequestedAt       time.Time               `json:"requested_at"`
	ExpiresAt         time.Time               `json:"expires_at"`
}

type RideAcceptedData struct {
	RideID           string      `json:"ride_id"`
	DriverID         string      `json:"driver_id"`
	DriverName       string      `json:"driver_name"`
	DriverRating     float64     `json:"driver_rating"`
	DriverPhone      string      `json:"driver_phone,omitempty"`
	VehicleInfo      VehicleInfo `json:"vehicle_info"`
	EstimatedArrival time.Time   `json:"estimated_arrival"`
	AcceptedAt       time.Time   `json:"accepted_at"`
	FareAmount       float64     `json:"fare_amount"`
	Currency         string      `json:"currency"`
}

type RideStartedData struct {
	RideID           string            `json:"ride_id"`
	StartedAt        time.Time         `json:"started_at"`
	DriverLocation   models.Location   `json:"driver_location"`
	EstimatedArrival time.Time         `json:"estimated_arrival"`
	Route            []models.Location `json:"route,omitempty"`
	Message          string            `json:"message,omitempty"`
}

type RideCompletedData struct {
	RideID        string      `json:"ride_id"`
	CompletedAt   time.Time   `json:"completed_at"`
	FinalAmount   float64     `json:"final_amount"`
	Currency      string      `json:"currency"`
	Distance      float64     `json:"distance"`
	Duration      int         `json:"duration"`
	PaymentStatus string      `json:"payment_status"`
	RatingRequest bool        `json:"rating_request"`
	Receipt       ReceiptData `json:"receipt,omitempty"`
}

type RideCancelledData struct {
	RideID          string    `json:"ride_id"`
	CancelledBy     string    `json:"cancelled_by"`
	CancelledAt     time.Time `json:"cancelled_at"`
	Reason          string    `json:"reason"`
	Message         string    `json:"message,omitempty"`
	CancellationFee float64   `json:"cancellation_fee,omitempty"`
	RefundAmount    float64   `json:"refund_amount,omitempty"`
}

// Location Events
type LocationUpdateData struct {
	UserID    string          `json:"user_id"`
	UserType  string          `json:"user_type"` // driver, passenger
	RideID    string          `json:"ride_id,omitempty"`
	Location  models.Location `json:"location"`
	Heading   float64         `json:"heading,omitempty"`
	Speed     float64         `json:"speed,omitempty"`
	Accuracy  float64         `json:"accuracy,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	IsMoving  bool            `json:"is_moving"`
	Address   string          `json:"address,omitempty"`
}

type DriverArrivingData struct {
	RideID           string          `json:"ride_id"`
	DriverID         string          `json:"driver_id"`
	DriverLocation   models.Location `json:"driver_location"`
	EstimatedArrival time.Time       `json:"estimated_arrival"`
	Distance         float64         `json:"distance"` // meters
	Message          string          `json:"message,omitempty"`
}

type DriverArrivedData struct {
	RideID         string          `json:"ride_id"`
	DriverID       string          `json:"driver_id"`
	DriverLocation models.Location `json:"driver_location"`
	ArrivedAt      time.Time       `json:"arrived_at"`
	WaitingTime    int             `json:"waiting_time"` // seconds
	Message        string          `json:"message,omitempty"`
}

// Chat Events
type ChatMessageData struct {
	ConversationID string                  `json:"conversation_id"`
	MessageID      string                  `json:"message_id"`
	RideID         string                  `json:"ride_id,omitempty"`
	SenderID       string                  `json:"sender_id"`
	SenderName     string                  `json:"sender_name"`
	SenderType     string                  `json:"sender_type"` // passenger, driver, admin
	MessageType    models.MessageType      `json:"message_type"`
	Content        string                  `json:"content"`
	MediaURL       string                  `json:"media_url,omitempty"`
	Location       *models.MessageLocation `json:"location,omitempty"`
	ReplyToID      string                  `json:"reply_to_id,omitempty"`
	SentAt         time.Time               `json:"sent_at"`
	IsRead         bool                    `json:"is_read"`
	QuickReplies   []models.QuickReply     `json:"quick_replies,omitempty"`
}

type MessageReadData struct {
	ConversationID string    `json:"conversation_id"`
	MessageID      string    `json:"message_id"`
	ReadBy         string    `json:"read_by"`
	ReadAt         time.Time `json:"read_at"`
}

// Driver Status Events
type DriverOnlineData struct {
	DriverID      string               `json:"driver_id"`
	Location      models.Location      `json:"location"`
	ServiceTypes  []models.ServiceType `json:"service_types"`
	VehicleInfo   VehicleInfo          `json:"vehicle_info"`
	OnlineSince   time.Time            `json:"online_since"`
	IsAvailable   bool                 `json:"is_available"`
	CurrentRideID string               `json:"current_ride_id,omitempty"`
}

type DriverOfflineData struct {
	DriverID       string    `json:"driver_id"`
	OfflineAt      time.Time `json:"offline_at"`
	OnlineTime     int       `json:"online_time"` // seconds
	RidesCompleted int       `json:"rides_completed"`
	EarningsToday  float64   `json:"earnings_today"`
	Reason         string    `json:"reason,omitempty"`
}

// Notification Events
type NotificationData struct {
	NotificationID string                      `json:"notification_id"`
	Type           models.NotificationType     `json:"type"`
	Title          string                      `json:"title"`
	Body           string                      `json:"body"`
	Priority       models.NotificationPriority `json:"priority"`
	ActionURL      string                      `json:"action_url,omitempty"`
	ActionText     string                      `json:"action_text,omitempty"`
	ImageURL       string                      `json:"image_url,omitempty"`
	Data           map[string]interface{}      `json:"data,omitempty"`
	CreatedAt      time.Time                   `json:"created_at"`
	ExpiresAt      time.Time                   `json:"expires_at,omitempty"`
}

// Emergency Events
type EmergencyAlertData struct {
	AlertID           string             `json:"alert_id"`
	UserID            string             `json:"user_id"`
	UserName          string             `json:"user_name"`
	UserPhone         string             `json:"user_phone"`
	RideID            string             `json:"ride_id,omitempty"`
	AlertType         string             `json:"alert_type"` // panic, accident, breakdown, medical
	Location          models.Location    `json:"location"`
	Message           string             `json:"message,omitempty"`
	Severity          int                `json:"severity"` // 1=critical, 2=high, 3=medium, 4=low
	Timestamp         time.Time          `json:"timestamp"`
	EmergencyContacts []EmergencyContact `json:"emergency_contacts,omitempty"`
	NearbyDrivers     []string           `json:"nearby_drivers,omitempty"`
}

type SOSData struct {
	UserID         string          `json:"user_id"`
	UserName       string          `json:"user_name"`
	UserPhone      string          `json:"user_phone"`
	RideID         string          `json:"ride_id,omitempty"`
	Location       models.Location `json:"location"`
	Timestamp      time.Time       `json:"timestamp"`
	AudioRecording string          `json:"audio_recording,omitempty"`
	VideoRecording string          `json:"video_recording,omitempty"`
	Photos         []string        `json:"photos,omitempty"`
	EmergencyType  string          `json:"emergency_type"`
	AutoGenerated  bool            `json:"auto_generated"` // Auto-generated vs manual SOS
}

// Admin Events
type AdminAlertData struct {
	AlertType     string                 `json:"alert_type"`
	Title         string                 `json:"title"`
	Message       string                 `json:"message"`
	Severity      string                 `json:"severity"` // info, warning, error, critical
	Data          map[string]interface{} `json:"data"`
	AffectedUsers []string               `json:"affected_users,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	ExpiresAt     time.Time              `json:"expires_at,omitempty"`
}

type SystemStatusData struct {
	Status            string             `json:"status"` // operational, degraded, outage
	Services          map[string]string  `json:"services"`
	Message           string             `json:"message,omitempty"`
	MaintenanceWindow *MaintenanceWindow `json:"maintenance_window,omitempty"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// Payment Events
type PaymentUpdateData struct {
	TransactionID string    `json:"transaction_id"`
	RideID        string    `json:"ride_id,omitempty"`
	PaymentStatus string    `json:"payment_status"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	PaymentMethod string    `json:"payment_method"`
	ProcessedAt   time.Time `json:"processed_at,omitempty"`
	FailureReason string    `json:"failure_reason,omitempty"`
	NextAction    string    `json:"next_action,omitempty"`
}

type RefundProcessedData struct {
	RefundID         string    `json:"refund_id"`
	TransactionID    string    `json:"transaction_id"`
	RideID           string    `json:"ride_id,omitempty"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	Reason           string    `json:"reason"`
	ProcessedAt      time.Time `json:"processed_at"`
	RefundMethod     string    `json:"refund_method"`
	EstimatedArrival string    `json:"estimated_arrival,omitempty"`
}

// Courier Events
type CourierUpdateData struct {
	CourierID        string               `json:"courier_id"`
	Status           models.CourierStatus `json:"status"`
	CourierLocation  models.Location      `json:"courier_location,omitempty"`
	EstimatedArrival time.Time            `json:"estimated_arrival,omitempty"`
	UpdatedAt        time.Time            `json:"updated_at"`
	Message          string               `json:"message,omitempty"`
	Photos           []string             `json:"photos,omitempty"`
}

type PackageDeliveredData struct {
	CourierID        string          `json:"courier_id"`
	DeliveredAt      time.Time       `json:"delivered_at"`
	DeliveredTo      string          `json:"delivered_to"`
	DeliveryLocation models.Location `json:"delivery_location"`
	SignatureURL     string          `json:"signature_url,omitempty"`
	DeliveryPhotos   []string        `json:"delivery_photos,omitempty"`
	RecipientRating  float64         `json:"recipient_rating,omitempty"`
	DeliveryNotes    string          `json:"delivery_notes,omitempty"`
}

// Freight Events
type FreightUpdateData struct {
	FreightID        string               `json:"freight_id"`
	Status           models.FreightStatus `json:"status"`
	CarrierLocation  models.Location      `json:"carrier_location,omitempty"`
	EstimatedArrival time.Time            `json:"estimated_arrival,omitempty"`
	UpdatedAt        time.Time            `json:"updated_at"`
	Message          string               `json:"message,omitempty"`
	Milestone        string               `json:"milestone,omitempty"`
	Photos           []string             `json:"photos,omitempty"`
}

// Helper Data Structures
type VehicleInfo struct {
	Type         string `json:"type"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	Color        string `json:"color"`
	LicensePlate string `json:"license_plate"`
	PhotoURL     string `json:"photo_url,omitempty"`
}

type ReceiptData struct {
	ReceiptID   string  `json:"receipt_id"`
	ReceiptURL  string  `json:"receipt_url"`
	BaseFare    float64 `json:"base_fare"`
	TotalAmount float64 `json:"total_amount"`
	TipAmount   float64 `json:"tip_amount,omitempty"`
	TaxAmount   float64 `json:"tax_amount,omitempty"`
	Currency    string  `json:"currency"`
}

type EmergencyContact struct {
	Name             string `json:"name"`
	Phone            string `json:"phone"`
	Relationship     string `json:"relationship"`
	NotificationSent bool   `json:"notification_sent"`
}

type MaintenanceWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Services  []string  `json:"services"`
	Impact    string    `json:"impact"`
}

// Event Factory Functions
func NewFareProposedEvent(rideID, proposedBy, proposedTo string, amount float64, currency string) WSMessage {
	return WSMessage{
		Type: models.EventFareProposed,
		Data: FareProposedData{
			RideID:     rideID,
			ProposedBy: proposedBy,
			ProposedTo: proposedTo,
			Amount:     amount,
			Currency:   currency,
			ExpiresAt:  time.Now().Add(10 * time.Minute), // 10 minutes to respond
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		RideID:    rideID,
	}
}

func NewRideAcceptedEvent(rideID, driverID, driverName string, vehicleInfo VehicleInfo) WSMessage {
	return WSMessage{
		Type: models.EventRideAccepted,
		Data: RideAcceptedData{
			RideID:           rideID,
			DriverID:         driverID,
			DriverName:       driverName,
			VehicleInfo:      vehicleInfo,
			AcceptedAt:       time.Now(),
			EstimatedArrival: time.Now().Add(15 * time.Minute),
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		RideID:    rideID,
	}
}

func NewLocationUpdateEvent(userID, userType string, location models.Location, rideID string) WSMessage {
	return WSMessage{
		Type: models.EventLocationUpdate,
		Data: LocationUpdateData{
			UserID:    userID,
			UserType:  userType,
			Location:  location,
			RideID:    rideID,
			Timestamp: time.Now(),
			IsMoving:  true,
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    userID,
		RideID:    rideID,
	}
}

func NewChatMessageEvent(conversationID, senderID, senderName, content string, messageType models.MessageType) WSMessage {
	return WSMessage{
		Type: models.EventChatMessage,
		Data: ChatMessageData{
			ConversationID: conversationID,
			MessageID:      utils.GenerateID(),
			SenderID:       senderID,
			SenderName:     senderName,
			MessageType:    messageType,
			Content:        content,
			SentAt:         time.Now(),
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    senderID,
	}
}

func NewEmergencyAlertEvent(userID, userName, userPhone string, alertType string, location models.Location) WSMessage {
	return WSMessage{
		Type: models.EventNotification,
		Data: EmergencyAlertData{
			AlertID:   utils.GenerateID(),
			UserID:    userID,
			UserName:  userName,
			UserPhone: userPhone,
			AlertType: alertType,
			Location:  location,
			Severity:  1, // Critical
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    userID,
	}
}
