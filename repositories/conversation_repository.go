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

type ConversationRepository interface {
	Create(conversation *models.Conversation) (*models.Conversation, error)
	GetByID(id string) (*models.Conversation, error)
	GetByRideID(rideID string) (*models.Conversation, error)
	GetUserConversations(userID string, page, limit int) ([]*models.Conversation, int64, error)
	Update(conversation *models.Conversation) (*models.Conversation, error)
	UpdateLastMessage(conversationID string, message *models.Message) error
	SoftDelete(id string) error
	AddParticipant(conversationID string, participant models.ConversationParticipant) error
	RemoveParticipant(conversationID, userID string) error
	UpdateParticipantLastRead(conversationID, userID string, timestamp time.Time) error
	GetActiveConversationsByUser(userID string) ([]*models.Conversation, error)
	GetConversationParticipants(conversationID string) ([]models.ConversationParticipant, error)
	ArchiveConversation(conversationID string) error
	UnarchiveConversation(conversationID string) error
}

type conversationRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewConversationRepository(db *mongo.Database) ConversationRepository {
	return &conversationRepository{
		collection: db.Collection("conversations"),
		db:         db,
	}
}

func (r *conversationRepository) Create(conversation *models.Conversation) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversation.ID = primitive.NewObjectID()
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, conversation)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *conversationRepository) GetByID(id string) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var conversation models.Conversation
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"is_active":  true,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&conversation)

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *conversationRepository) GetByRideID(rideID string) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rideObjectID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		return nil, err
	}

	var conversation models.Conversation
	err = r.collection.FindOne(ctx, bson.M{
		"ride_id":    rideObjectID,
		"is_active":  true,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&conversation)

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *conversationRepository) GetUserConversations(userID string, page, limit int) ([]*models.Conversation, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"participants.user_id": userObjectID,
		"participants.left_at": nil,
		"is_active":            true,
		"is_deleted":           bson.M{"$ne": true},
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
		SetSort(bson.D{{Key: "last_message_at", Value: -1}, {Key: "updated_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var conversations []*models.Conversation
	if err = cursor.All(ctx, &conversations); err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

func (r *conversationRepository) Update(conversation *models.Conversation) (*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversation.UpdatedAt = time.Now()

	filter := bson.M{"_id": conversation.ID}
	update := bson.M{"$set": conversation}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *conversationRepository) UpdateLastMessage(conversationID string, message *models.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"last_message":    message,
			"last_message_at": message.CreatedAt,
			"updated_at":      time.Now(),
		},
		"$inc": bson.M{
			"message_count": 1,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *conversationRepository) SoftDelete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *conversationRepository) AddParticipant(conversationID string, participant models.ConversationParticipant) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}

	participant.JoinedAt = time.Now()

	update := bson.M{
		"$push": bson.M{
			"participants": participant,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *conversationRepository) RemoveParticipant(conversationID, userID string) error {
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

	update := bson.M{
		"$set": bson.M{
			"participants.$.left_at": time.Now(),
			"updated_at":             time.Now(),
		},
	}

	filter := bson.M{
		"_id":                  conversationObjectID,
		"participants.user_id": userObjectID,
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *conversationRepository) UpdateParticipantLastRead(conversationID, userID string, timestamp time.Time) error {
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

	update := bson.M{
		"$set": bson.M{
			"participants.$.last_read_at": timestamp,
			"participants.$.unread_count": 0,
			"updated_at":                  time.Now(),
		},
	}

	filter := bson.M{
		"_id":                  conversationObjectID,
		"participants.user_id": userObjectID,
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *conversationRepository) GetActiveConversationsByUser(userID string) ([]*models.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"participants.user_id": userObjectID,
		"participants.left_at": nil,
		"is_active":            true,
		"is_deleted":           bson.M{"$ne": true},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conversations []*models.Conversation
	if err = cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *conversationRepository) GetConversationParticipants(conversationID string) ([]models.ConversationParticipant, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return nil, err
	}

	var conversation models.Conversation
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&conversation)
	if err != nil {
		return nil, err
	}

	return conversation.Participants, nil
}

func (r *conversationRepository) ArchiveConversation(conversationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_archived": true,
			"updated_at":  time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *conversationRepository) UnarchiveConversation(conversationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_archived": false,
			"updated_at":  time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}
