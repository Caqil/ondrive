package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ondrive/models"
	"ondrive/utils"
)

// ConnectionType defines the type of WebSocket connection
type ConnectionType string

const (
	ConnectionTypeGeneral         ConnectionType = "general"
	ConnectionTypeRide            ConnectionType = "ride"
	ConnectionTypeDriverLocation  ConnectionType = "driver_location"
	ConnectionTypeFareNegotiation ConnectionType = "fare_negotiation"
	ConnectionTypeChat            ConnectionType = "chat"
	ConnectionTypeAdminDashboard  ConnectionType = "admin_dashboard"
	ConnectionTypeAdminTracking   ConnectionType = "admin_tracking"
	ConnectionTypeEmergency       ConnectionType = "emergency"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by connection type
	clients map[ConnectionType]map[*Client]bool

	// User ID to client mapping for quick lookup
	userClients map[string]map[ConnectionType]*Client

	// Ride ID to clients mapping (for ride-specific events)
	rideClients map[string][]*Client

	// Chat conversation ID to clients mapping
	chatClients map[string][]*Client

	// Driver location subscribers (admin/tracking)
	locationSubscribers map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to all clients of a type
	broadcast chan BroadcastMessage

	// Send message to specific user
	sendToUser chan UserMessage

	// Send message to ride participants
	sendToRide chan RideMessage

	// Send message to chat participants
	sendToChat chan ChatMessage

	// Send location update to subscribers
	sendLocationUpdate chan LocationUpdate

	// Emergency broadcast (highest priority)
	emergencyBroadcast chan EmergencyMessage

	// Connection statistics
	stats HubStats

	// Mutex for thread safety
	mutex sync.RWMutex

	// Message handlers by event type
	handlers map[models.WSEventType]EventHandler

	// Rate limiting
	rateLimiter *RateLimiter

	// Logger
	logger utils.Logger

	// Configuration
	config HubConfig
}

type HubConfig struct {
	MaxConnections         int           `json:"max_connections"`
	MaxConnectionsPerUser  int           `json:"max_connections_per_user"`
	PingInterval           time.Duration `json:"ping_interval"`
	PongTimeout            time.Duration `json:"pong_timeout"`
	WriteTimeout           time.Duration `json:"write_timeout"`
	ReadTimeout            time.Duration `json:"read_timeout"`
	MaxMessageSize         int64         `json:"max_message_size"`
	EnableRateLimit        bool          `json:"enable_rate_limit"`
	RateLimitMessages      int           `json:"rate_limit_messages"`
	RateLimitWindow        time.Duration `json:"rate_limit_window"`
	EnableCompression      bool          `json:"enable_compression"`
	BroadcastChannelBuffer int           `json:"broadcast_channel_buffer"`
}

type HubStats struct {
	TotalConnections  int64                  `json:"total_connections"`
	ActiveConnections int                    `json:"active_connections"`
	MessagesSent      int64                  `json:"messages_sent"`
	MessagesReceived  int64                  `json:"messages_received"`
	ErrorCount        int64                  `json:"error_count"`
	LastActivity      time.Time              `json:"last_activity"`
	ConnectionsByType map[ConnectionType]int `json:"connections_by_type"`
	OnlineUsers       int                    `json:"online_users"`
	OnlineDrivers     int                    `json:"online_drivers"`
	ActiveRides       int                    `json:"active_rides"`
	ActiveChats       int                    `json:"active_chats"`
}

type BroadcastMessage struct {
	Type        ConnectionType `json:"type"`
	Message     WSMessage      `json:"message"`
	ExcludeUser string         `json:"exclude_user,omitempty"`
}

type UserMessage struct {
	UserID         string         `json:"user_id"`
	ConnectionType ConnectionType `json:"connection_type"`
	Message        WSMessage      `json:"message"`
}

type RideMessage struct {
	RideID      string    `json:"ride_id"`
	Message     WSMessage `json:"message"`
	ExcludeUser string    `json:"exclude_user,omitempty"`
}

type ChatMessage struct {
	ConversationID string    `json:"conversation_id"`
	Message        WSMessage `json:"message"`
	ExcludeUser    string    `json:"exclude_user,omitempty"`
}

type LocationUpdate struct {
	DriverID string          `json:"driver_id"`
	Location models.Location `json:"location"`
	RideID   string          `json:"ride_id,omitempty"`
}

type EmergencyMessage struct {
	UserID   string    `json:"user_id"`
	RideID   string    `json:"ride_id,omitempty"`
	Message  WSMessage `json:"message"`
	Priority int       `json:"priority"` // 1=highest, 5=lowest
}

type EventHandler func(*Client, WSMessage) error

// NewHub creates a new WebSocket hub
func NewHub(config HubConfig, logger utils.Logger) *Hub {
	// Set default configuration
	if config.MaxConnections == 0 {
		config.MaxConnections = 10000
	}
	if config.MaxConnectionsPerUser == 0 {
		config.MaxConnectionsPerUser = 5
	}
	if config.PingInterval == 0 {
		config.PingInterval = 54 * time.Second
	}
	if config.PongTimeout == 0 {
		config.PongTimeout = 60 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 60 * time.Second
	}
	if config.MaxMessageSize == 0 {
		config.MaxMessageSize = 1024 * 1024 // 1MB
	}
	if config.RateLimitMessages == 0 {
		config.RateLimitMessages = 100
	}
	if config.RateLimitWindow == 0 {
		config.RateLimitWindow = time.Minute
	}
	if config.BroadcastChannelBuffer == 0 {
		config.BroadcastChannelBuffer = 1000
	}

	hub := &Hub{
		clients:             make(map[ConnectionType]map[*Client]bool),
		userClients:         make(map[string]map[ConnectionType]*Client),
		rideClients:         make(map[string][]*Client),
		chatClients:         make(map[string][]*Client),
		locationSubscribers: make(map[*Client]bool),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		broadcast:           make(chan BroadcastMessage, config.BroadcastChannelBuffer),
		sendToUser:          make(chan UserMessage, config.BroadcastChannelBuffer),
		sendToRide:          make(chan RideMessage, config.BroadcastChannelBuffer),
		sendToChat:          make(chan ChatMessage, config.BroadcastChannelBuffer),
		sendLocationUpdate:  make(chan LocationUpdate, config.BroadcastChannelBuffer),
		emergencyBroadcast:  make(chan EmergencyMessage, 100),
		handlers:            make(map[models.WSEventType]EventHandler),
		config:              config,
		logger:              logger,
		stats: HubStats{
			ConnectionsByType: make(map[ConnectionType]int),
		},
	}

	// Initialize connection type maps
	for _, connType := range []ConnectionType{
		ConnectionTypeGeneral, ConnectionTypeRide, ConnectionTypeDriverLocation,
		ConnectionTypeFareNegotiation, ConnectionTypeChat, ConnectionTypeAdminDashboard,
		ConnectionTypeAdminTracking, ConnectionTypeEmergency,
	} {
		hub.clients[connType] = make(map[*Client]bool)
	}

	// Initialize rate limiter if enabled
	if config.EnableRateLimit {
		hub.rateLimiter = NewRateLimiter(config.RateLimitMessages, config.RateLimitWindow)
	}

	return hub
}

// Run starts the hub and handles all WebSocket operations
func (h *Hub) Run() {
	h.logger.Info().Msg("WebSocket hub started")

	// Start periodic cleanup
	go h.periodicCleanup()

	// Start statistics update
	go h.updateStats()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case broadcastMsg := <-h.broadcast:
			h.broadcastToType(broadcastMsg)

		case userMsg := <-h.sendToUser:
			h.sendToSpecificUser(userMsg)

		case rideMsg := <-h.sendToRide:
			h.sendToRideParticipants(rideMsg)

		case chatMsg := <-h.sendToChat:
			h.sendToChatParticipants(chatMsg)

		case locationUpdate := <-h.sendLocationUpdate:
			h.broadcastLocationUpdate(locationUpdate)

		case emergencyMsg := <-h.emergencyBroadcast:
			h.handleEmergencyBroadcast(emergencyMsg)
		}
	}
}

// RegisterHandler registers an event handler for a specific event type
func (h *Hub) RegisterHandler(eventType models.WSEventType, handler EventHandler) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.handlers[eventType] = handler
}

// RegisterClient registers a new client connection
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Check maximum connections
	if h.getTotalConnections() >= h.config.MaxConnections {
		h.logger.Warn().
			Str("user_id", client.UserID).
			Msg("Maximum connections reached, rejecting new connection")
		client.Close()
		return
	}

	// Check maximum connections per user
	if userConnections, exists := h.userClients[client.UserID]; exists {
		if len(userConnections) >= h.config.MaxConnectionsPerUser {
			h.logger.Warn().
				Str("user_id", client.UserID).
				Msg("Maximum connections per user reached")
			client.Close()
			return
		}
	}

	// Register client
	if h.clients[client.Type] == nil {
		h.clients[client.Type] = make(map[*Client]bool)
	}
	h.clients[client.Type][client] = true

	// Register in user clients map
	if h.userClients[client.UserID] == nil {
		h.userClients[client.UserID] = make(map[ConnectionType]*Client)
	}
	h.userClients[client.UserID][client.Type] = client

	// Register for specific contexts
	switch client.Type {
	case ConnectionTypeRide:
		if client.RideID != "" {
			h.rideClients[client.RideID] = append(h.rideClients[client.RideID], client)
		}
	case ConnectionTypeChat:
		if client.ConversationID != "" {
			h.chatClients[client.ConversationID] = append(h.chatClients[client.ConversationID], client)
		}
	case ConnectionTypeDriverLocation, ConnectionTypeAdminTracking:
		h.locationSubscribers[client] = true
	}

	// Update statistics
	h.stats.TotalConnections++
	h.stats.ConnectionsByType[client.Type]++
	h.stats.LastActivity = time.Now()

	h.logger.Info().
		Str("user_id", client.UserID).
		Str("connection_type", string(client.Type)).
		Str("ride_id", client.RideID).
		Int("total_connections", h.getTotalConnections()).
		Msg("Client connected")

	// Send welcome message
	welcomeMsg := WSMessage{
		Type:      models.EventSystemMessage,
		Data:      WelcomeData{Message: "Connected successfully", ServerTime: time.Now()},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
		UserID:    client.UserID,
	}
	client.Send(welcomeMsg)
}

// UnregisterClient removes a client connection
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Remove from clients map
	if clients, exists := h.clients[client.Type]; exists {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			h.stats.ConnectionsByType[client.Type]--
		}
	}

	// Remove from user clients map
	if userConnections, exists := h.userClients[client.UserID]; exists {
		delete(userConnections, client.Type)
		if len(userConnections) == 0 {
			delete(h.userClients, client.UserID)
		}
	}

	// Remove from specific contexts
	switch client.Type {
	case ConnectionTypeRide:
		if client.RideID != "" {
			h.removeClientFromRide(client, client.RideID)
		}
	case ConnectionTypeChat:
		if client.ConversationID != "" {
			h.removeClientFromChat(client, client.ConversationID)
		}
	case ConnectionTypeDriverLocation, ConnectionTypeAdminTracking:
		delete(h.locationSubscribers, client)
	}

	// Close client connection
	client.Close()

	h.logger.Info().
		Str("user_id", client.UserID).
		Str("connection_type", string(client.Type)).
		Int("total_connections", h.getTotalConnections()).
		Msg("Client disconnected")
}

// broadcastToType broadcasts a message to all clients of a specific type
func (h *Hub) broadcastToType(broadcastMsg BroadcastMessage) {
	h.mutex.RLock()
	clients := h.clients[broadcastMsg.Type]
	h.mutex.RUnlock()

	if clients == nil {
		return
	}

	messageBytes, err := json.Marshal(broadcastMsg.Message)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal broadcast message")
		return
	}

	for client := range clients {
		if broadcastMsg.ExcludeUser != "" && client.UserID == broadcastMsg.ExcludeUser {
			continue
		}

		select {
		case client.Send <- messageBytes:
			h.stats.MessagesSent++
		default:
			h.logger.Warn().
				Str("user_id", client.UserID).
				Msg("Client send channel full, removing client")
			go h.unregisterClient(client)
		}
	}
}

// SendToSpecificUser sends a message to a specific user
func (h *Hub) sendToSpecificUser(userMsg UserMessage) {
	h.mutex.RLock()
	userConnections := h.userClients[userMsg.UserID]
	h.mutex.RUnlock()

	if userConnections == nil {
		h.logger.Debug().
			Str("user_id", userMsg.UserID).
			Msg("User not connected")
		return
	}

	client, exists := userConnections[userMsg.ConnectionType]
	if !exists {
		h.logger.Debug().
			Str("user_id", userMsg.UserID).
			Str("connection_type", string(userMsg.ConnectionType)).
			Msg("User not connected to specified type")
		return
	}

	client.Send(userMsg.Message)
	h.stats.MessagesSent++
}

// sendToRideParticipants sends a message to all participants in a ride
func (h *Hub) sendToRideParticipants(rideMsg RideMessage) {
	h.mutex.RLock()
	clients := h.rideClients[rideMsg.RideID]
	h.mutex.RUnlock()

	if clients == nil {
		return
	}

	messageBytes, err := json.Marshal(rideMsg.Message)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal ride message")
		return
	}

	for _, client := range clients {
		if rideMsg.ExcludeUser != "" && client.UserID == rideMsg.ExcludeUser {
			continue
		}

		select {
		case client.Send <- messageBytes:
			h.stats.MessagesSent++
		default:
			h.logger.Warn().
				Str("user_id", client.UserID).
				Str("ride_id", rideMsg.RideID).
				Msg("Ride client send channel full")
		}
	}
}

// sendToChatParticipants sends a message to all participants in a chat
func (h *Hub) sendToChatParticipants(chatMsg ChatMessage) {
	h.mutex.RLock()
	clients := h.chatClients[chatMsg.ConversationID]
	h.mutex.RUnlock()

	if clients == nil {
		return
	}

	messageBytes, err := json.Marshal(chatMsg.Message)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal chat message")
		return
	}

	for _, client := range clients {
		if chatMsg.ExcludeUser != "" && client.UserID == chatMsg.ExcludeUser {
			continue
		}

		select {
		case client.Send <- messageBytes:
			h.stats.MessagesSent++
		default:
			h.logger.Warn().
				Str("user_id", client.UserID).
				Str("conversation_id", chatMsg.ConversationID).
				Msg("Chat client send channel full")
		}
	}
}

// broadcastLocationUpdate broadcasts location updates to subscribers
func (h *Hub) broadcastLocationUpdate(locationUpdate LocationUpdate) {
	h.mutex.RLock()
	subscribers := h.locationSubscribers
	h.mutex.RUnlock()

	message := WSMessage{
		Type: models.EventLocationUpdate,
		Data: models.LocationUpdateEvent{
			UserID:    locationUpdate.DriverID,
			RideID:    locationUpdate.RideID,
			Location:  locationUpdate.Location,
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
		MessageID: utils.GenerateID(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal location update")
		return
	}

	for client := range subscribers {
		select {
		case client.Send <- messageBytes:
			h.stats.MessagesSent++
		default:
			h.logger.Warn().
				Str("user_id", client.UserID).
				Msg("Location subscriber send channel full")
		}
	}
}

// handleEmergencyBroadcast handles emergency messages with highest priority
func (h *Hub) handleEmergencyBroadcast(emergencyMsg EmergencyMessage) {
	h.logger.Warn().
		Str("user_id", emergencyMsg.UserID).
		Str("ride_id", emergencyMsg.RideID).
		Int("priority", emergencyMsg.Priority).
		Msg("Emergency broadcast")

	// Send to admin connections
	h.broadcastToType(BroadcastMessage{
		Type:    ConnectionTypeAdminDashboard,
		Message: emergencyMsg.Message,
	})

	// Send to relevant ride participants if ride_id provided
	if emergencyMsg.RideID != "" {
		h.sendToRideParticipants(RideMessage{
			RideID:  emergencyMsg.RideID,
			Message: emergencyMsg.Message,
		})
	}

	// Send to emergency responders or support team
	h.broadcastToType(BroadcastMessage{
		Type:    ConnectionTypeEmergency,
		Message: emergencyMsg.Message,
	})
}

// Helper methods
func (h *Hub) removeClientFromRide(client *Client, rideID string) {
	clients := h.rideClients[rideID]
	for i, c := range clients {
		if c == client {
			h.rideClients[rideID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.rideClients[rideID]) == 0 {
		delete(h.rideClients, rideID)
	}
}

func (h *Hub) removeClientFromChat(client *Client, conversationID string) {
	clients := h.chatClients[conversationID]
	for i, c := range clients {
		if c == client {
			h.chatClients[conversationID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.chatClients[conversationID]) == 0 {
		delete(h.chatClients, conversationID)
	}
}

func (h *Hub) getTotalConnections() int {
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}

// Periodic cleanup of inactive connections
func (h *Hub) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mutex.Lock()
		now := time.Now()

		for connType, clients := range h.clients {
			for client := range clients {
				if now.Sub(client.LastPong) > h.config.PongTimeout*2 {
					h.logger.Info().
						Str("user_id", client.UserID).
						Str("connection_type", string(connType)).
						Msg("Removing inactive client")
					go h.unregisterClient(client)
				}
			}
		}
		h.mutex.Unlock()
	}
}

// Update statistics periodically
func (h *Hub) updateStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.mutex.Lock()
		h.stats.ActiveConnections = h.getTotalConnections()
		h.stats.OnlineUsers = len(h.userClients)

		// Count online drivers
		onlineDrivers := 0
		for _, userConns := range h.userClients {
			if _, hasDriverLocation := userConns[ConnectionTypeDriverLocation]; hasDriverLocation {
				onlineDrivers++
			}
		}
		h.stats.OnlineDrivers = onlineDrivers
		h.stats.ActiveRides = len(h.rideClients)
		h.stats.ActiveChats = len(h.chatClients)
		h.mutex.Unlock()
	}
}

// Public API methods
func (h *Hub) SendToUser(userID string, connectionType ConnectionType, message WSMessage) error {
	select {
	case h.sendToUser <- UserMessage{
		UserID:         userID,
		ConnectionType: connectionType,
		Message:        message,
	}:
		return nil
	default:
		return fmt.Errorf("user message channel full")
	}
}

func (h *Hub) SendToRide(rideID string, message WSMessage, excludeUserID string) error {
	select {
	case h.sendToRide <- RideMessage{
		RideID:      rideID,
		Message:     message,
		ExcludeUser: excludeUserID,
	}:
		return nil
	default:
		return fmt.Errorf("ride message channel full")
	}
}

func (h *Hub) SendToChat(conversationID string, message WSMessage, excludeUserID string) error {
	select {
	case h.sendToChat <- ChatMessage{
		ConversationID: conversationID,
		Message:        message,
		ExcludeUser:    excludeUserID,
	}:
		return nil
	default:
		return fmt.Errorf("chat message channel full")
	}
}

func (h *Hub) BroadcastLocationUpdate(driverID string, location models.Location, rideID string) error {
	select {
	case h.sendLocationUpdate <- LocationUpdate{
		DriverID: driverID,
		Location: location,
		RideID:   rideID,
	}:
		return nil
	default:
		return fmt.Errorf("location update channel full")
	}
}

func (h *Hub) SendEmergencyMessage(userID, rideID string, message WSMessage, priority int) error {
	select {
	case h.emergencyBroadcast <- EmergencyMessage{
		UserID:   userID,
		RideID:   rideID,
		Message:  message,
		Priority: priority,
	}:
		return nil
	default:
		return fmt.Errorf("emergency broadcast channel full")
	}
}

func (h *Hub) GetStats() HubStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.stats
}

func (h *Hub) IsUserOnline(userID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	_, exists := h.userClients[userID]
	return exists
}

func (h *Hub) GetUserConnectionTypes(userID string) []ConnectionType {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	userConnections, exists := h.userClients[userID]
	if !exists {
		return nil
	}

	types := make([]ConnectionType, 0, len(userConnections))
	for connType := range userConnections {
		types = append(types, connType)
	}
	return types
}
