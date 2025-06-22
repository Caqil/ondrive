package repositories

import (
	"context"
	"time"

	"ondrive/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageRepository interface {
	Create(message *models.Message) (*models.Message, error)
	GetByID(id string) (*models.Message, error)
	GetConversationMessages(conversationID string, page, limit int) ([]*models.Message, int64, error)
	GetLatestMessages(conversationID string, limit int) ([]*models.Message, error)
	Update(message *models.Message) (*models.Message, error)
	SoftDelete(id, deletedBy string) error
	MarkAsRead(messageID string, userID primitive.ObjectID) error
	MarkAsDelivered(messageID string) error
	GetUnreadCount(conversationID, userID string) (int64, error)
	GetMessagesByType(conversationID string, messageType models.MessageType, page, limit int) ([]*models.Message, int64, error)
	GetMessagesAfterTime(conversationID string, timestamp time.Time) ([]*models.Message, error)
	SearchMessages(conversationID, query string, page, limit int) ([]*models.Message, int64, error)
	GetMediaMessages(conversationID string, page, limit int) ([]*models.Message, int64, error)
	BulkMarkAsRead(conversationID, userID string) error
	GetMessageReactions(messageID string) ([]models.MessageReaction, error)
	AddMessageReaction(messageID string, reaction models.MessageReaction) error
	RemoveMessageReaction(messageID, userID string) error
	GetQuotedMessage(messageID string) (*models.Message, error)
}

type messageRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewMessageRepository(db *mongo.Database) MessageRepository {
	return &messageRepository{
		collection: db.Collection("messages"),
		db:         db,
	}
}

func (r *messageRepository) Create(message *models.Message) (*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message.ID = primitive.NewObjectID()
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (r *messageRepository) GetByID(id string) (*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var message models.Message
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&message)

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *messageRepository) GetConversationMessages(conversationID string, page, limit int) ([]*models.Message, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"is_deleted":      bson.M{"$ne": true},
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find with pagination (latest first)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) GetLatestMessages(conversationID string, limit int) ([]*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"is_deleted":      bson.M{"$ne": true},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) Update(message *models.Message) (*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message.UpdatedAt = time.Now()

	filter := bson.M{"_id": message.ID}
	update := bson.M{"$set": message}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (r *messageRepository) SoftDelete(id, deletedBy string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	deletedByObjectID, err := primitive.ObjectIDFromHex(deletedBy)
	if err != nil {
		return err
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"deleted_at": &now,
			"deleted_by": &deletedByObjectID,
			"updated_at": now,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *messageRepository) MarkAsRead(messageID string, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}

	now := time.Now()
	readStatus := models.MessageReadStatus{
		UserID: userID,
		ReadAt: now,
	}

	// Add to read_by array if not already exists
	update := bson.M{
		"$addToSet": bson.M{
			"read_by": readStatus,
		},
		"$set": bson.M{
			"read_at":    &now,
			"updated_at": now,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *messageRepository) MarkAsDelivered(messageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":       models.MessageStatusDelivered,
			"delivered_at": &now,
			"updated_at":   now,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *messageRepository) GetUnreadCount(conversationID, userID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return 0, err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"sender_id":       bson.M{"$ne": userObjectID},
		"is_deleted":      bson.M{"$ne": true},
		"read_by.user_id": bson.M{"$ne": userObjectID},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	return count, err
}

func (r *messageRepository) GetMessagesByType(conversationID string, messageType models.MessageType, page, limit int) ([]*models.Message, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"type":            messageType,
		"is_deleted":      bson.M{"$ne": true},
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) GetMessagesAfterTime(conversationID string, timestamp time.Time) ([]*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"created_at":      bson.M{"$gt": timestamp},
		"is_deleted":      bson.M{"$ne": true},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) SearchMessages(conversationID, query string, page, limit int) ([]*models.Message, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"content":         bson.M{"$regex": query, "$options": "i"},
		"is_deleted":      bson.M{"$ne": true},
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) GetMediaMessages(conversationID string, page, limit int) ([]*models.Message, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"type": bson.M{
			"$in": []models.MessageType{
				models.MessageTypeImage,
				models.MessageTypeAudio,
				models.MessageTypeFile,
			},
		},
		"is_deleted": bson.M{"$ne": true},
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) BulkMarkAsRead(conversationID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversationObjectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	readStatus := models.MessageReadStatus{
		UserID: userObjectID,
		ReadAt: now,
	}

	filter := bson.M{
		"conversation_id": conversationObjectID,
		"sender_id":       bson.M{"$ne": userObjectID},
		"is_deleted":      bson.M{"$ne": true},
		"read_by.user_id": bson.M{"$ne": userObjectID},
	}

	update := bson.M{
		"$addToSet": bson.M{
			"read_by": readStatus,
		},
		"$set": bson.M{
			"read_at":    &now,
			"updated_at": now,
		},
	}

	_, err = r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *messageRepository) GetMessageReactions(messageID string) ([]models.MessageReaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, err
	}

	var message models.Message
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&message)
	if err != nil {
		return nil, err
	}

	return message.Reactions, nil
}

func (r *messageRepository) AddMessageReaction(messageID string, reaction models.MessageReaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}

	reaction.ReactedAt = time.Now()

	// Remove existing reaction from same user first
	filter := bson.M{"_id": objectID}
	pullUpdate := bson.M{
		"$pull": bson.M{
			"reactions": bson.M{"user_id": reaction.UserID},
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, pullUpdate)
	if err != nil {
		return err
	}

	// Add new reaction
	pushUpdate := bson.M{
		"$push": bson.M{
			"reactions": reaction,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, pushUpdate)
	return err
}

func (r *messageRepository) RemoveMessageReaction(messageID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$pull": bson.M{
			"reactions": bson.M{"user_id": userObjectID},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *messageRepository) GetQuotedMessage(messageID string) (*models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"reply_to_id": objectID,
				"is_deleted":  bson.M{"$ne": true},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "messages",
				"localField":   "reply_to_id",
				"foreignField": "_id",
				"as":           "quoted_message",
			},
		},
		{
			"$unwind": "$quoted_message",
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		QuotedMessage models.Message `bson:"quoted_message"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		return &result.QuotedMessage, nil
	}

	return nil, mongo.ErrNoDocuments
}
