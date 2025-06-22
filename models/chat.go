package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageType string

const (
	MessageTypeText       MessageType = "text"
	MessageTypeImage      MessageType = "image"
	MessageTypeAudio      MessageType = "audio"
	MessageTypeLocation   MessageType = "location"
	MessageTypeFile       MessageType = "file"
	MessageTypeSystem     MessageType = "system"
	MessageTypeQuickReply MessageType = "quick_reply"
)

type ConversationType string

const (
	ConversationTypeRide      ConversationType = "ride"
	ConversationTypeSupport   ConversationType = "support"
	ConversationTypeEmergency ConversationType = "emergency"
	ConversationTypeDirect    ConversationType = "direct"
	ConversationTypeGroup     ConversationType = "group"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

type Conversation struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// Conversation Details
	Type        ConversationType `json:"type" bson:"type"`
	Title       string           `json:"title" bson:"title"`
	Description string           `json:"description" bson:"description"`

	// Participants
	Participants []ConversationParticipant `json:"participants" bson:"participants"`
	CreatedBy    primitive.ObjectID        `json:"created_by" bson:"created_by"`

	// Related Entities
	RideID          *primitive.ObjectID `json:"ride_id,omitempty" bson:"ride_id,omitempty"`
	SupportTicketID *primitive.ObjectID `json:"support_ticket_id,omitempty" bson:"support_ticket_id,omitempty"`

	// Last Message Info
	LastMessage   *Message   `json:"last_message,omitempty" bson:"last_message,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty" bson:"last_message_at,omitempty"`
	MessageCount  int        `json:"message_count" bson:"message_count"`

	// Status & Settings
	IsActive   bool `json:"is_active" bson:"is_active"`
	IsArchived bool `json:"is_archived" bson:"is_archived"`
	IsBlocked  bool `json:"is_blocked" bson:"is_blocked"`
	IsMuted    bool `json:"is_muted" bson:"is_muted"`

	// Auto-close settings (for ride conversations)
	AutoCloseAfter time.Duration `json:"auto_close_after" bson:"auto_close_after"`
	ClosedAt       *time.Time    `json:"closed_at,omitempty" bson:"closed_at,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata" bson:"metadata"`
	Tags     []string               `json:"tags" bson:"tags"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type ConversationParticipant struct {
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	Role        string             `json:"role" bson:"role"` // owner, member, admin
	JoinedAt    time.Time          `json:"joined_at" bson:"joined_at"`
	LastReadAt  *time.Time         `json:"last_read_at,omitempty" bson:"last_read_at,omitempty"`
	UnreadCount int                `json:"unread_count" bson:"unread_count"`
	IsMuted     bool               `json:"is_muted" bson:"is_muted"`
	IsBlocked   bool               `json:"is_blocked" bson:"is_blocked"`
	LeftAt      *time.Time         `json:"left_at,omitempty" bson:"left_at,omitempty"`
}

type Message struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ConversationID primitive.ObjectID `json:"conversation_id" bson:"conversation_id"`

	// Sender Information
	SenderID   primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	SenderRole string             `json:"sender_role" bson:"sender_role"`

	// Message Content
	Type         MessageType `json:"type" bson:"type"`
	Content      string      `json:"content" bson:"content"`
	MediaURL     string      `json:"media_url" bson:"media_url"`
	ThumbnailURL string      `json:"thumbnail_url" bson:"thumbnail_url"`
	FileName     string      `json:"file_name" bson:"file_name"`
	FileSize     int64       `json:"file_size" bson:"file_size"`
	Duration     int         `json:"duration" bson:"duration"` // for audio/video in seconds

	// Location Data (for location messages)
	Location *MessageLocation `json:"location,omitempty" bson:"location,omitempty"`

	// Reply Information
	ReplyToID     *primitive.ObjectID `json:"reply_to_id,omitempty" bson:"reply_to_id,omitempty"`
	QuotedMessage *Message            `json:"quoted_message,omitempty" bson:"quoted_message,omitempty"`

	// Status & Tracking
	Status      MessageStatus       `json:"status" bson:"status"`
	DeliveredAt *time.Time          `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`
	ReadAt      *time.Time          `json:"read_at,omitempty" bson:"read_at,omitempty"`
	ReadBy      []MessageReadStatus `json:"read_by" bson:"read_by"`

	// Editing & Deletion
	IsEdited        bool                `json:"is_edited" bson:"is_edited"`
	EditedAt        *time.Time          `json:"edited_at,omitempty" bson:"edited_at,omitempty"`
	OriginalContent string              `json:"original_content" bson:"original_content"`
	IsDeleted       bool                `json:"is_deleted" bson:"is_deleted"`
	DeletedAt       *time.Time          `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DeletedBy       *primitive.ObjectID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`

	// Quick Replies (for system messages)
	QuickReplies []QuickReply `json:"quick_replies" bson:"quick_replies"`

	// System Message Data
	SystemData *SystemMessageData `json:"system_data,omitempty" bson:"system_data,omitempty"`

	// Reactions
	Reactions []MessageReaction `json:"reactions" bson:"reactions"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata" bson:"metadata"`
	Platform string                 `json:"platform" bson:"platform"` // ios, android, web

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type MessageLocation struct {
	Latitude       float64    `json:"latitude" bson:"latitude"`
	Longitude      float64    `json:"longitude" bson:"longitude"`
	Address        string     `json:"address" bson:"address"`
	PlaceName      string     `json:"place_name" bson:"place_name"`
	Accuracy       float64    `json:"accuracy" bson:"accuracy"`
	IsLiveLocation bool       `json:"is_live_location" bson:"is_live_location"`
	LiveUntil      *time.Time `json:"live_until,omitempty" bson:"live_until,omitempty"`
}

type MessageReadStatus struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	ReadAt time.Time          `json:"read_at" bson:"read_at"`
}

type QuickReply struct {
	ID      string `json:"id" bson:"id"`
	Text    string `json:"text" bson:"text"`
	Payload string `json:"payload" bson:"payload"`
	Icon    string `json:"icon" bson:"icon"`
}

type SystemMessageData struct {
	EventType  string                 `json:"event_type" bson:"event_type"`
	EventData  map[string]interface{} `json:"event_data" bson:"event_data"`
	TemplateID string                 `json:"template_id" bson:"template_id"`
	Variables  map[string]string      `json:"variables" bson:"variables"`
}

type MessageReaction struct {
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Emoji     string             `json:"emoji" bson:"emoji"`
	ReactedAt time.Time          `json:"reacted_at" bson:"reacted_at"`
}

type ChatSettings struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	// General Settings
	IsOnline         bool       `json:"is_online" bson:"is_online"`
	LastSeen         *time.Time `json:"last_seen,omitempty" bson:"last_seen,omitempty"`
	ShowLastSeen     bool       `json:"show_last_seen" bson:"show_last_seen"`
	ShowOnlineStatus bool       `json:"show_online_status" bson:"show_online_status"`

	// Message Settings
	ReadReceipts      bool `json:"read_receipts" bson:"read_receipts"`
	TypingIndicator   bool `json:"typing_indicator" bson:"typing_indicator"`
	AutoDownloadMedia bool `json:"auto_download_media" bson:"auto_download_media"`

	// Notification Settings
	MessageNotifications bool `json:"message_notifications" bson:"message_notifications"`
	SoundEnabled         bool `json:"sound_enabled" bson:"sound_enabled"`
	VibrationEnabled     bool `json:"vibration_enabled" bson:"vibration_enabled"`

	// Privacy Settings
	AllowMessagesFrom   string `json:"allow_messages_from" bson:"allow_messages_from"` // everyone, contacts, nobody
	BlockUnknownSenders bool   `json:"block_unknown_senders" bson:"block_unknown_senders"`

	// Blocked Users
	BlockedUsers []primitive.ObjectID `json:"blocked_users" bson:"blocked_users"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type MessageTemplate struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name"`
	Category   string             `json:"category" bson:"category"`
	Type       MessageType        `json:"type" bson:"type"`
	Content    string             `json:"content" bson:"content"`
	Variables  []string           `json:"variables" bson:"variables"`
	Language   string             `json:"language" bson:"language"`
	IsActive   bool               `json:"is_active" bson:"is_active"`
	IsSystem   bool               `json:"is_system" bson:"is_system"`
	UsageCount int                `json:"usage_count" bson:"usage_count"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

type QuickReplyTemplate struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	Text       string             `json:"text" bson:"text"`
	Category   string             `json:"category" bson:"category"`
	UsageCount int                `json:"usage_count" bson:"usage_count"`
	LastUsedAt *time.Time         `json:"last_used_at,omitempty" bson:"last_used_at,omitempty"`
	IsActive   bool               `json:"is_active" bson:"is_active"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}
