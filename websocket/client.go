package websocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ondrive/models"
	"ondrive/utils"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 1024 * 1024 // 1MB
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
	EnableCompression: true,
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Hub reference
	hub *Hub

	// Buffered channel of outbound messages
	Send chan []byte

	// Client information
	UserID         string          `json:"user_id"`
	UserRole       models.UserRole `json:"user_role"`
	Type           ConnectionType  `json:"type"`
	RideID         string          `json:"ride_id,omitempty"`
	ConversationID string          `json:"conversation_id,omitempty"`
	Platform       string          `json:"platform"` // ios, android, web
	AppVersion     string          `json:"app_version"`
	DeviceID       string          `json:"device_id"`

	// Connection metadata
	ConnectedAt  time.Time              `json:"connected_at"`
	LastPong     time.Time              `json:"last_pong"`
	LastActivity time.Time              `json:"last_activity"`
	MessageCount int64                  `json:"message_count"`
	ErrorCount   int64                  `json:"error_count"`
	Metadata     map[string]interface{} `json:"metadata"`

	// Rate limiting
	rateLimiter *RateLimiter

	// Authentication
	isAuthenticated bool
	authToken       string

	// Location tracking (for drivers)
	lastLocation *models.Location
	isTracking   bool

	// Logger
	logger utils.Logger
}

// ClientInfo contains information about a connected client
type ClientInfo struct {
	UserID         string                 `json:"user_id"`
	UserRole       models.UserRole        `json:"user_role"`
	Type           ConnectionType         `json:"type"`
	RideID         string                 `json:"ride_id,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	Platform       string                 `json:"platform"`
	AppVersion     string                 `json:"app_version"`
	ConnectedAt    time.Time              `json:"connected_at"`
	LastActivity   time.Time              `json:"last_activity"`
	MessageCount   int64                  `json:"message_count"`
	IsOnline       bool                   `json:"is_online"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID string, userRole models.UserRole,
	connType ConnectionType, logger utils.Logger) *Client {

	client := &Client{
		conn:            conn,
		hub:             hub,
		Send:            make(chan []byte, 256),
		UserID:          userID,
		UserRole:        userRole,
		Type:            connType,
		ConnectedAt:     time.Now(),
		LastPong:        time.Now(),
		LastActivity:    time.Now(),
		Metadata:        make(map[string]interface{}),
		isAuthenticated: true, // Assume authenticated if we got here
		logger:          logger,
	}

	// Initialize rate limiter if enabled
	if hub.config.EnableRateLimit {
		client.rateLimiter = NewRateLimiter(hub.config.RateLimitMessages, hub.config.RateLimitWindow)
	}

	return client
}

// HandleConnection handles a new WebSocket connection
func HandleConnection(hub *Hub, w http.ResponseWriter, r *http.Request,
	userID string, userRole models.UserRole, connType ConnectionType) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.logger.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	client := NewClient(hub, conn, userID, userRole, connType, hub.logger)

	// Extract additional information from request
	client.Platform = r.Header.Get("X-Platform")
	client.AppVersion = r.Header.Get("X-App-Version")
	client.DeviceID = r.Header.Get("X-Device-ID")
	client.RideID = r.URL.Query().Get("ride_id")
	client.ConversationID = r.URL.Query().Get("conversation_id")

	// Register the client with the hub
	hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.LastPong = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error().
					Err(err).
					Str("user_id", c.UserID).
					Msg("WebSocket connection error")
				c.ErrorCount++
			}
			break
		}

		// Rate limiting
		if c.rateLimiter != nil && !c.rateLimiter.Allow() {
			c.logger.Warn().
				Str("user_id", c.UserID).
				Msg("Rate limit exceeded")
			c.sendError("Rate limit exceeded", "RATE_LIMIT_EXCEEDED")
			continue
		}

		// Update activity
		c.LastActivity = time.Now()
		c.MessageCount++
		c.hub.stats.MessagesReceived++

		// Clean the message
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Parse the message
		var wsMessage WSMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			c.logger.Error().
				Err(err).
				Str("user_id", c.UserID).
				Str("message", string(message)).
				Msg("Failed to parse WebSocket message")
			c.sendError("Invalid message format", "INVALID_MESSAGE")
			c.ErrorCount++
			continue
		}

		// Validate message
		if err := c.validateMessage(wsMessage); err != nil {
			c.logger.Warn().
				Err(err).
				Str("user_id", c.UserID).
				Msg("Message validation failed")
			c.sendError(err.Error(), "VALIDATION_ERROR")
			continue
		}

		// Set message metadata
		wsMessage.UserID = c.UserID
		wsMessage.Timestamp = time.Now()
		if wsMessage.MessageID == "" {
			wsMessage.MessageID = utils.GenerateID()
		}

		// Handle the message
		if err := c.handleMessage(wsMessage); err != nil {
			c.logger.Error().
				Err(err).
				Str("user_id", c.UserID).
				Str("event_type", string(wsMessage.Type)).
				Msg("Failed to handle WebSocket message")
			c.sendError("Failed to process message", "PROCESSING_ERROR")
			c.ErrorCount++
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message WSMessage) error {
	// Check if we have a handler for this event type
	if handler, exists := c.hub.handlers[message.Type]; exists {
		return handler(c, message)
	}

	// Handle default message types
	switch message.Type {
	case models.EventLocationUpdate:
		return c.handleLocationUpdate(message)
	case models.EventFareProposed, models.EventFareAccepted, models.EventFareRejected, models.EventFareCountered:
		return c.handleFareNegotiation(message)
	case models.EventChatMessage:
		return c.handleChatMessage(message)
	case models.EventRideAccepted, models.EventRideStarted, models.EventRideCompleted, models.EventRideCancelled:
		return c.handleRideUpdate(message)
	case models.EventDriverOnline, models.EventDriverOffline:
		return c.handleDriverStatus(message)
	default:
		c.logger.Warn().
			Str("user_id", c.UserID).
			Str("event_type", string(message.Type)).
			Msg("Unknown message type")
		return fmt.Errorf("unknown message type: %s", message.Type)
	}
}

// validateMessage validates incoming WebSocket messages
func (c *Client) validateMessage(message WSMessage) error {
	if message.Type == "" {
		return fmt.Errorf("message type is required")
	}

	if message.Data == nil {
		return fmt.Errorf("message data is required")
	}

	// Role-based validation
	switch c.UserRole {
	case models.RolePassenger:
		if !c.isPassengerAllowed(message.Type) {
			return fmt.Errorf("passengers are not allowed to send %s messages", message.Type)
		}
	case models.RoleDriver:
		if !c.isDriverAllowed(message.Type) {
			return fmt.Errorf("drivers are not allowed to send %s messages", message.Type)
		}
	case models.RoleAdmin:
		// Admins can send any message type
	default:
		return fmt.Errorf("unknown user role: %s", c.UserRole)
	}

	return nil
}

// isPassengerAllowed checks if passengers can send this message type
func (c *Client) isPassengerAllowed(eventType models.WSEventType) bool {
	allowedEvents := []models.WSEventType{
		models.EventFareProposed,
		models.EventFareAccepted,
		models.EventFareRejected,
		models.EventFareCountered,
		models.EventChatMessage,
		models.EventLocationUpdate,
		models.EventRideCancelled,
	}

	for _, allowed := range allowedEvents {
		if eventType == allowed {
			return true
		}
	}
	return false
}

// isDriverAllowed checks if drivers can send this message type
func (c *Client) isDriverAllowed(eventType models.WSEventType) bool {
	allowedEvents := []models.WSEventType{
		models.EventFareAccepted,
		models.EventFareRejected,
		models.EventFareCountered,
		models.EventChatMessage,
		models.EventLocationUpdate,
		models.EventRideAccepted,
		models.EventRideStarted,
		models.EventRideCompleted,
		models.EventRideCancelled,
		models.EventDriverOnline,
		models.EventDriverOffline,
		models.EventDriverArriving,
		models.EventDriverArrived,
	}

	for _, allowed := range allowedEvents {
		if eventType == allowed {
			return true
		}
	}
	return false
}

// Send sends a message to the client
func (c *Client) Send(message WSMessage) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case c.Send <- messageBytes:
		return nil
	default:
		return fmt.Errorf("client send channel full")
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(message, code string) {
	errorMsg := WSMessage{
		Type: "error",
		Data: map[string]interface{}{
			"message": message,
			"code":    code,
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    c.UserID,
	}
	c.Send(errorMsg)
}

// Close closes the client connection
func (c *Client) Close() {
	close(c.Send)
	c.conn.Close()
}

// GetInfo returns client information
func (c *Client) GetInfo() ClientInfo {
	return ClientInfo{
		UserID:         c.UserID,
		UserRole:       c.UserRole,
		Type:           c.Type,
		RideID:         c.RideID,
		ConversationID: c.ConversationID,
		Platform:       c.Platform,
		AppVersion:     c.AppVersion,
		ConnectedAt:    c.ConnectedAt,
		LastActivity:   c.LastActivity,
		MessageCount:   c.MessageCount,
		IsOnline:       true,
		Metadata:       c.Metadata,
	}
}

// SetRideID sets the ride ID for this client
func (c *Client) SetRideID(rideID string) {
	c.RideID = rideID
	// Update hub mapping
	if rideID != "" {
		c.hub.mutex.Lock()
		c.hub.rideClients[rideID] = append(c.hub.rideClients[rideID], c)
		c.hub.mutex.Unlock()
	}
}

// SetConversationID sets the conversation ID for this client
func (c *Client) SetConversationID(conversationID string) {
	c.ConversationID = conversationID
	// Update hub mapping
	if conversationID != "" {
		c.hub.mutex.Lock()
		c.hub.chatClients[conversationID] = append(c.hub.chatClients[conversationID], c)
		c.hub.mutex.Unlock()
	}
}

// UpdateLocation updates the client's last known location
func (c *Client) UpdateLocation(location models.Location) {
	c.lastLocation = &location
	c.isTracking = true
}

// GetLastLocation returns the client's last known location
func (c *Client) GetLastLocation() *models.Location {
	return c.lastLocation
}

// IsTracking returns whether the client is currently being tracked
func (c *Client) IsTracking() bool {
	return c.isTracking
}
