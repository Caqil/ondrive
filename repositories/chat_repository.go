package repositories

import (
	"context"
	"fmt"
	"time"

	"ondrive/models"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// QuickReplyRepository interface for quick reply operations
type QuickReplyRepository interface {
	Create(quickReply *models.QuickReplyTemplate) (*models.QuickReplyTemplate, error)
	GetByID(id string) (*models.QuickReplyTemplate, error)
	GetByUserID(userID primitive.ObjectID) ([]*models.QuickReplyTemplate, error)
	Update(quickReply *models.QuickReplyTemplate) (*models.QuickReplyTemplate, error)
	SoftDelete(id string) error
	IncrementUsage(id string) error
	GetByCategory(userID primitive.ObjectID, category string) ([]*models.QuickReplyTemplate, error)
}

// MessageTemplateRepository interface for message template operations
type MessageTemplateRepository interface {
	Create(template *models.MessageTemplate) (*models.MessageTemplate, error)
	GetByID(id string) (*models.MessageTemplate, error)
	GetActive() ([]*models.MessageTemplate, error)
	GetByCategory(category string) ([]*models.MessageTemplate, error)
	Update(template *models.MessageTemplate) (*models.MessageTemplate, error)
	SoftDelete(id string) error
	GetSystemTemplates() ([]*models.MessageTemplate, error)
	IncrementUsage(id string) error
}

// ChatSettingsRepository interface for chat settings operations
type ChatSettingsRepository interface {
	Create(settings *models.ChatSettings) (*models.ChatSettings, error)
	GetByUserID(userID primitive.ObjectID) (*models.ChatSettings, error)
	Update(settings *models.ChatSettings) (*models.ChatSettings, error)
	UpdateOnlineStatus(userID primitive.ObjectID, isOnline bool) error
	UpdateLastSeen(userID primitive.ObjectID, lastSeen time.Time) error
	AddBlockedUser(userID, blockedUserID primitive.ObjectID) error
	RemoveBlockedUser(userID, blockedUserID primitive.ObjectID) error
	IsUserBlocked(userID, targetUserID primitive.ObjectID) (bool, error)
}

// QuickReplyRepository Implementation

type quickReplyRepository struct {
	collection *mongo.Collection
	logger     utils.Logger
}

func NewQuickReplyRepository(db *mongo.Database, logger utils.Logger) QuickReplyRepository {
	return &quickReplyRepository{
		collection: db.Collection("quick_replies"),
		logger:     logger,
	}
}

func (r *quickReplyRepository) Create(quickReply *models.QuickReplyTemplate) (*models.QuickReplyTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	quickReply.ID = primitive.NewObjectID()
	quickReply.CreatedAt = time.Now()
	quickReply.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, quickReply)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create quick reply")
		return nil, fmt.Errorf("failed to create quick reply: %w", err)
	}

	return quickReply, nil
}

func (r *quickReplyRepository) GetByID(id string) (*models.QuickReplyTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var quickReply models.QuickReplyTemplate
	err = r.collection.FindOne(ctx, bson.M{
		"_id":       objectID,
		"is_active": true,
	}).Decode(&quickReply)

	if err != nil {
		r.logger.Error().Err(err).Str("quick_reply_id", id).Msg("Failed to get quick reply")
		return nil, fmt.Errorf("failed to get quick reply: %w", err)
	}

	return &quickReply, nil
}

func (r *quickReplyRepository) GetByUserID(userID primitive.ObjectID) ([]*models.QuickReplyTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":   userID,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "usage_count", Value: -1}}))
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get user quick replies")
		return nil, fmt.Errorf("failed to get user quick replies: %w", err)
	}
	defer cursor.Close(ctx)

	var quickReplies []*models.QuickReplyTemplate
	if err = cursor.All(ctx, &quickReplies); err != nil {
		return nil, err
	}

	return quickReplies, nil
}

func (r *quickReplyRepository) Update(quickReply *models.QuickReplyTemplate) (*models.QuickReplyTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	quickReply.UpdatedAt = time.Now()

	filter := bson.M{"_id": quickReply.ID}
	update := bson.M{"$set": quickReply}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("quick_reply_id", quickReply.ID.Hex()).Msg("Failed to update quick reply")
		return nil, fmt.Errorf("failed to update quick reply: %w", err)
	}

	return quickReply, nil
}

func (r *quickReplyRepository) SoftDelete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("quick_reply_id", id).Msg("Failed to delete quick reply")
		return fmt.Errorf("failed to delete quick reply: %w", err)
	}

	return nil
}

func (r *quickReplyRepository) IncrementUsage(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$inc": bson.M{
			"usage_count": 1,
		},
		"$set": bson.M{
			"last_used_at": time.Now(),
			"updated_at":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *quickReplyRepository) GetByCategory(userID primitive.ObjectID, category string) ([]*models.QuickReplyTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":   userID,
		"category":  category,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "usage_count", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quickReplies []*models.QuickReplyTemplate
	if err = cursor.All(ctx, &quickReplies); err != nil {
		return nil, err
	}

	return quickReplies, nil
}

// MessageTemplateRepository Implementation

type messageTemplateRepository struct {
	collection *mongo.Collection
	logger     utils.Logger
}

func NewMessageTemplateRepository(db *mongo.Database, logger utils.Logger) MessageTemplateRepository {
	return &messageTemplateRepository{
		collection: db.Collection("message_templates"),
		logger:     logger,
	}
}

func (r *messageTemplateRepository) Create(template *models.MessageTemplate) (*models.MessageTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	template.ID = primitive.NewObjectID()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, template)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create message template")
		return nil, fmt.Errorf("failed to create message template: %w", err)
	}

	return template, nil
}

func (r *messageTemplateRepository) GetByID(id string) (*models.MessageTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var template models.MessageTemplate
	err = r.collection.FindOne(ctx, bson.M{
		"_id":       objectID,
		"is_active": true,
	}).Decode(&template)

	if err != nil {
		r.logger.Error().Err(err).Str("template_id", id).Msg("Failed to get message template")
		return nil, fmt.Errorf("failed to get message template: %w", err)
	}

	return &template, nil
}

func (r *messageTemplateRepository) GetActive() ([]*models.MessageTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"is_active": true}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "category", Value: 1}, {Key: "name", Value: 1}}))
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get active message templates")
		return nil, fmt.Errorf("failed to get active message templates: %w", err)
	}
	defer cursor.Close(ctx)

	var templates []*models.MessageTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *messageTemplateRepository) GetByCategory(category string) ([]*models.MessageTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"category":  category,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.MessageTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *messageTemplateRepository) Update(template *models.MessageTemplate) (*models.MessageTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	template.UpdatedAt = time.Now()

	filter := bson.M{"_id": template.ID}
	update := bson.M{"$set": template}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("template_id", template.ID.Hex()).Msg("Failed to update message template")
		return nil, fmt.Errorf("failed to update message template: %w", err)
	}

	return template, nil
}

func (r *messageTemplateRepository) SoftDelete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("template_id", id).Msg("Failed to delete message template")
		return fmt.Errorf("failed to delete message template: %w", err)
	}

	return nil
}

func (r *messageTemplateRepository) GetSystemTemplates() ([]*models.MessageTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_system": true,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "category", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.MessageTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *messageTemplateRepository) IncrementUsage(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$inc": bson.M{
			"usage_count": 1,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// ChatSettingsRepository Implementation

type chatSettingsRepository struct {
	collection *mongo.Collection
	logger     utils.Logger
}

func NewChatSettingsRepository(db *mongo.Database, logger utils.Logger) ChatSettingsRepository {
	return &chatSettingsRepository{
		collection: db.Collection("chat_settings"),
		logger:     logger,
	}
}

func (r *chatSettingsRepository) Create(settings *models.ChatSettings) (*models.ChatSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	settings.ID = primitive.NewObjectID()
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, settings)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create chat settings")
		return nil, fmt.Errorf("failed to create chat settings: %w", err)
	}

	return settings, nil
}

func (r *chatSettingsRepository) GetByUserID(userID primitive.ObjectID) (*models.ChatSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var settings models.ChatSettings
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&settings)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get chat settings")
		return nil, fmt.Errorf("failed to get chat settings: %w", err)
	}

	return &settings, nil
}

func (r *chatSettingsRepository) Update(settings *models.ChatSettings) (*models.ChatSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	settings.UpdatedAt = time.Now()

	filter := bson.M{"_id": settings.ID}
	update := bson.M{"$set": settings}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("settings_id", settings.ID.Hex()).Msg("Failed to update chat settings")
		return nil, fmt.Errorf("failed to update chat settings: %w", err)
	}

	return settings, nil
}

func (r *chatSettingsRepository) UpdateOnlineStatus(userID primitive.ObjectID, isOnline bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_online":  isOnline,
			"updated_at": time.Now(),
		},
	}

	if !isOnline {
		update["$set"].(bson.M)["last_seen"] = time.Now()
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to update online status")
		return fmt.Errorf("failed to update online status: %w", err)
	}

	return nil
}

func (r *chatSettingsRepository) UpdateLastSeen(userID primitive.ObjectID, lastSeen time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"last_seen":  lastSeen,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	return err
}

func (r *chatSettingsRepository) AddBlockedUser(userID, blockedUserID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$addToSet": bson.M{
			"blocked_users": blockedUserID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	return err
}

func (r *chatSettingsRepository) RemoveBlockedUser(userID, blockedUserID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$pull": bson.M{
			"blocked_users": blockedUserID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	return err
}

func (r *chatSettingsRepository) IsUserBlocked(userID, targetUserID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.M{
		"user_id":       userID,
		"blocked_users": targetUserID,
	})

	return count > 0, err
}
