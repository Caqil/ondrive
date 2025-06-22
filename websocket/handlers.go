package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ondrive/models"
	"ondrive/utils"
)

// Event handlers for different message types

// handleLocationUpdate processes location update messages
func (c *Client) handleLocationUpdate(message WSMessage) error {
	var locationData LocationUpdateData
	if err := mapToStruct(message.Data, &locationData); err != nil {
		return fmt.Errorf("invalid location update data: %w", err)
	}

	// Validate location data
	if !utils.ValidateCoordinates(locationData.Location.Coordinates[1], locationData.Location.Coordinates[0]) {
		return fmt.Errorf("invalid coordinates")
	}

	// Update client's last known location
	c.UpdateLocation(locationData.Location)

	// Set user ID from client if not provided
	if locationData.UserID == "" {
		locationData.UserID = c.UserID
	}

	// Determine user type based on role
	if locationData.UserType == "" {
		switch c.UserRole {
		case models.RoleDriver:
			locationData.UserType = "driver"
		case models.RolePassenger:
			locationData.UserType = "passenger"
		default:
			locationData.UserType = "user"
		}
	}

	locationData.Timestamp = time.Now()

	// Create location update event
	locationEvent := WSMessage{
		Type:      models.EventLocationUpdate,
		Data:      locationData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    locationData.RideID,
	}

	// If this is a driver and they're in a ride, broadcast to ride participants
	if c.UserRole == models.RoleDriver && c.RideID != "" {
		return c.hub.SendToRide(c.RideID, locationEvent, c.UserID)
	}

	// If this is for location tracking (admin/tracking), broadcast to location subscribers
	if c.Type == ConnectionTypeDriverLocation || c.UserRole == models.RoleDriver {
		return c.hub.BroadcastLocationUpdate(c.UserID, locationData.Location, locationData.RideID)
	}

	return nil
}

// handleFareNegotiation processes fare negotiation messages
func (c *Client) handleFareNegotiation(message WSMessage) error {
	switch message.Type {
	case models.EventFareProposed:
		return c.handleFareProposed(message)
	case models.EventFareAccepted:
		return c.handleFareAccepted(message)
	case models.EventFareRejected:
		return c.handleFareRejected(message)
	case models.EventFareCountered:
		return c.handleFareCountered(message)
	default:
		return fmt.Errorf("unknown fare negotiation event: %s", message.Type)
	}
}

func (c *Client) handleFareProposed(message WSMessage) error {
	var fareData FareProposedData
	if err := mapToStruct(message.Data, &fareData); err != nil {
		return fmt.Errorf("invalid fare proposed data: %w", err)
	}

	// Validate fare amount
	if fareData.Amount <= 0 {
		return fmt.Errorf("fare amount must be positive")
	}

	// Set proposer
	fareData.ProposedBy = c.UserID
	fareData.ExpiresAt = time.Now().Add(10 * time.Minute) // 10 minutes to respond

	// Create fare event
	fareEvent := WSMessage{
		Type:      models.EventFareProposed,
		Data:      fareData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    fareData.RideID,
	}

	// Send to ride participants
	return c.hub.SendToRide(fareData.RideID, fareEvent, c.UserID)
}

func (c *Client) handleFareAccepted(message WSMessage) error {
	var fareData FareAcceptedData
	if err := mapToStruct(message.Data, &fareData); err != nil {
		return fmt.Errorf("invalid fare accepted data: %w", err)
	}

	fareData.AcceptedBy = c.UserID
	fareData.AcceptedAt = time.Now()

	fareEvent := WSMessage{
		Type:      models.EventFareAccepted,
		Data:      fareData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    fareData.RideID,
	}

	return c.hub.SendToRide(fareData.RideID, fareEvent, c.UserID)
}

func (c *Client) handleFareRejected(message WSMessage) error {
	var fareData FareRejectedData
	if err := mapToStruct(message.Data, &fareData); err != nil {
		return fmt.Errorf("invalid fare rejected data: %w", err)
	}

	fareData.RejectedBy = c.UserID
	fareData.RejectedAt = time.Now()

	fareEvent := WSMessage{
		Type:      models.EventFareRejected,
		Data:      fareData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    fareData.RideID,
	}

	return c.hub.SendToRide(fareData.RideID, fareEvent, c.UserID)
}

func (c *Client) handleFareCountered(message WSMessage) error {
	var fareData FareCounteredData
	if err := mapToStruct(message.Data, &fareData); err != nil {
		return fmt.Errorf("invalid fare countered data: %w", err)
	}

	// Validate counter amount
	if fareData.CounterAmount <= 0 {
		return fmt.Errorf("counter amount must be positive")
	}

	fareData.CounteredBy = c.UserID
	fareData.ExpiresAt = time.Now().Add(10 * time.Minute)

	fareEvent := WSMessage{
		Type:      models.EventFareCountered,
		Data:      fareData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    fareData.RideID,
	}

	return c.hub.SendToRide(fareData.RideID, fareEvent, c.UserID)
}

// handleChatMessage processes chat messages
func (c *Client) handleChatMessage(message WSMessage) error {
	var chatData ChatMessageData
	if err := mapToStruct(message.Data, &chatData); err != nil {
		return fmt.Errorf("invalid chat message data: %w", err)
	}

	// Set sender information
	chatData.SenderID = c.UserID
	chatData.SenderType = string(c.UserRole)
	chatData.SentAt = time.Now()
	chatData.MessageID = utils.GenerateID()

	// Validate message content
	if chatData.Content == "" && chatData.MediaURL == "" && chatData.Location == nil {
		return fmt.Errorf("message content, media, or location is required")
	}

	// Create chat event
	chatEvent := WSMessage{
		Type:      models.EventChatMessage,
		Data:      chatData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    chatData.RideID,
	}

	// Send to chat participants
	if chatData.ConversationID != "" {
		return c.hub.SendToChat(chatData.ConversationID, chatEvent, c.UserID)
	}

	// If no conversation ID but ride ID exists, send to ride participants
	if chatData.RideID != "" {
		return c.hub.SendToRide(chatData.RideID, chatEvent, c.UserID)
	}

	return fmt.Errorf("conversation ID or ride ID required for chat messages")
}

// handleRideUpdate processes ride status updates
func (c *Client) handleRideUpdate(message WSMessage) error {
	switch message.Type {
	case models.EventRideAccepted:
		return c.handleRideAccepted(message)
	case models.EventRideStarted:
		return c.handleRideStarted(message)
	case models.EventRideCompleted:
		return c.handleRideCompleted(message)
	case models.EventRideCancelled:
		return c.handleRideCancelled(message)
	default:
		return fmt.Errorf("unknown ride update event: %s", message.Type)
	}
}

func (c *Client) handleRideAccepted(message WSMessage) error {
	// Only drivers can accept rides
	if c.UserRole != models.RoleDriver {
		return fmt.Errorf("only drivers can accept rides")
	}

	var rideData RideAcceptedData
	if err := mapToStruct(message.Data, &rideData); err != nil {
		return fmt.Errorf("invalid ride accepted data: %w", err)
	}

	rideData.DriverID = c.UserID
	rideData.AcceptedAt = time.Now()
	rideData.EstimatedArrival = time.Now().Add(15 * time.Minute) // Default 15 minutes

	// Update client's ride ID
	c.SetRideID(rideData.RideID)

	rideEvent := WSMessage{
		Type:      models.EventRideAccepted,
		Data:      rideData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    rideData.RideID,
	}

	return c.hub.SendToRide(rideData.RideID, rideEvent, c.UserID)
}

func (c *Client) handleRideStarted(message WSMessage) error {
	// Only drivers can start rides
	if c.UserRole != models.RoleDriver {
		return fmt.Errorf("only drivers can start rides")
	}

	var rideData RideStartedData
	if err := mapToStruct(message.Data, &rideData); err != nil {
		return fmt.Errorf("invalid ride started data: %w", err)
	}

	rideData.StartedAt = time.Now()

	rideEvent := WSMessage{
		Type:      models.EventRideStarted,
		Data:      rideData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    rideData.RideID,
	}

	return c.hub.SendToRide(rideData.RideID, rideEvent, c.UserID)
}

func (c *Client) handleRideCompleted(message WSMessage) error {
	// Only drivers can complete rides
	if c.UserRole != models.RoleDriver {
		return fmt.Errorf("only drivers can complete rides")
	}

	var rideData RideCompletedData
	if err := mapToStruct(message.Data, &rideData); err != nil {
		return fmt.Errorf("invalid ride completed data: %w", err)
	}

	rideData.CompletedAt = time.Now()
	rideData.RatingRequest = true // Request rating from both parties

	rideEvent := WSMessage{
		Type:      models.EventRideCompleted,
		Data:      rideData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    rideData.RideID,
	}

	// Clear client's ride ID
	c.SetRideID("")

	return c.hub.SendToRide(rideData.RideID, rideEvent, c.UserID)
}

func (c *Client) handleRideCancelled(message WSMessage) error {
	var rideData RideCancelledData
	if err := mapToStruct(message.Data, &rideData); err != nil {
		return fmt.Errorf("invalid ride cancelled data: %w", err)
	}

	rideData.CancelledBy = c.UserID
	rideData.CancelledAt = time.Now()

	// Validate cancellation reason
	if rideData.Reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}

	rideEvent := WSMessage{
		Type:      models.EventRideCancelled,
		Data:      rideData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    rideData.RideID,
	}

	// Clear client's ride ID
	c.SetRideID("")

	return c.hub.SendToRide(rideData.RideID, rideEvent, c.UserID)
}

// handleDriverStatus processes driver online/offline status changes
func (c *Client) handleDriverStatus(message WSMessage) error {
	// Only drivers can change driver status
	if c.UserRole != models.RoleDriver {
		return fmt.Errorf("only drivers can change driver status")
	}

	switch message.Type {
	case models.EventDriverOnline:
		return c.handleDriverOnline(message)
	case models.EventDriverOffline:
		return c.handleDriverOffline(message)
	default:
		return fmt.Errorf("unknown driver status event: %s", message.Type)
	}
}

func (c *Client) handleDriverOnline(message WSMessage) error {
	var driverData DriverOnlineData
	if err := mapToStruct(message.Data, &driverData); err != nil {
		return fmt.Errorf("invalid driver online data: %w", err)
	}

	driverData.DriverID = c.UserID
	driverData.OnlineSince = time.Now()
	driverData.IsAvailable = true

	// Update client location if provided
	if len(driverData.Location.Coordinates) == 2 {
		c.UpdateLocation(driverData.Location)
	}

	driverEvent := WSMessage{
		Type:      models.EventDriverOnline,
		Data:      driverData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
	}

	// Broadcast to admin dashboard and location subscribers
	c.hub.broadcastToType(BroadcastMessage{
		Type:    ConnectionTypeAdminDashboard,
		Message: driverEvent,
	})

	return c.hub.BroadcastLocationUpdate(c.UserID, driverData.Location, "")
}

func (c *Client) handleDriverOffline(message WSMessage) error {
	var driverData DriverOfflineData
	if err := mapToStruct(message.Data, &driverData); err != nil {
		return fmt.Errorf("invalid driver offline data: %w", err)
	}

	driverData.DriverID = c.UserID
	driverData.OfflineAt = time.Now()

	driverEvent := WSMessage{
		Type:      models.EventDriverOffline,
		Data:      driverData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
	}

	// Broadcast to admin dashboard
	c.hub.broadcastToType(BroadcastMessage{
		Type:    ConnectionTypeAdminDashboard,
		Message: driverEvent,
	})

	return nil
}

// Helper function to map interface{} to struct
func mapToStruct(data interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

// Rate limiter implementation
type RateLimiter struct {
	requests int
	window   time.Duration
	clients  map[string]*clientLimiter
	mutex    sync.RWMutex
}

type clientLimiter struct {
	count  int
	window time.Time
}

func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: requests,
		window:   window,
		clients:  make(map[string]*clientLimiter),
	}
}

func (r *RateLimiter) Allow() bool {
	return true // Simplified for now - implement proper rate limiting logic
}

// Emergency handlers
func (c *Client) handleEmergencyAlert(message WSMessage) error {
	var alertData EmergencyAlertData
	if err := mapToStruct(message.Data, &alertData); err != nil {
		return fmt.Errorf("invalid emergency alert data: %w", err)
	}

	alertData.UserID = c.UserID
	alertData.AlertID = utils.GenerateID()
	alertData.Timestamp = time.Now()
	alertData.Severity = 1 // Critical

	emergencyEvent := WSMessage{
		Type:      models.EventNotification,
		Data:      alertData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    alertData.RideID,
	}

	// Send emergency broadcast with highest priority
	return c.hub.SendEmergencyMessage(c.UserID, alertData.RideID, emergencyEvent, 1)
}

func (c *Client) handleSOSAlert(message WSMessage) error {
	var sosData SOSData
	if err := mapToStruct(message.Data, &sosData); err != nil {
		return fmt.Errorf("invalid SOS data: %w", err)
	}

	sosData.UserID = c.UserID
	sosData.Timestamp = time.Now()

	sosEvent := WSMessage{
		Type:      "sos_alert",
		Data:      sosData,
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
		RideID:    sosData.RideID,
	}

	// Send emergency broadcast with highest priority
	return c.hub.SendEmergencyMessage(c.UserID, sosData.RideID, sosEvent, 1)
}
