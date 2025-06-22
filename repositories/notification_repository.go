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

type NotificationRepository interface {
	Create(notification *models.Notification) (*models.Notification, error)
	GetByID(id string) (*models.Notification, error)
	GetUserNotifications(userID string, page, limit int) ([]*models.Notification, int64, error)
	GetNotificationsByType(userID string, notificationType models.NotificationType, page, limit int) ([]*models.Notification, int64, error)
	GetNotificationsByCategory(userID string, category string, page, limit int) ([]*models.Notification, int64, error)
	Update(notification *models.Notification) (*models.Notification, error)
	MarkAsRead(id, userID string) error
	MarkAllAsRead(userID string) error
	Delete(id string) error
	ClearAllNotifications(userID string) error
	GetUnreadCount(userID string) (int64, error)
	GetNotificationHistory(userID string, days int, page, limit int) ([]*models.Notification, int64, error)
	GetNotificationsByStatus(userID string, status models.NotificationStatus, page, limit int) ([]*models.Notification, int64, error)
	BulkInsert(notifications []*models.Notification) error
	GetDeliveryStatus(id string) (*models.Notification, error)
	UpdateDeliveryResult(id string, result models.DeliveryResult) error
	GetAnalytics(userID string, startDate, endDate time.Time) (*models.NotificationAnalytics, error)
	GetEngagementStats(templateID string, startDate, endDate time.Time) (*models.EngagementStats, error)
}

type NotificationTemplateRepository interface {
	Create(template *models.NotificationTemplate) (*models.NotificationTemplate, error)
	GetByID(id string) (*models.NotificationTemplate, error)
	GetByType(notificationType models.NotificationType) ([]*models.NotificationTemplate, error)
	GetByCategory(category string) ([]*models.NotificationTemplate, error)
	GetActive() ([]*models.NotificationTemplate, error)
	Update(template *models.NotificationTemplate) (*models.NotificationTemplate, error)
	Delete(id string) error
}

type NotificationPreferenceRepository interface {
	Create(preferences *models.NotificationPreferences) (*models.NotificationPreferences, error)
	GetByUserID(userID primitive.ObjectID) (*models.NotificationPreferences, error)
	Update(preferences *models.NotificationPreferences) (*models.NotificationPreferences, error)
	Delete(userID primitive.ObjectID) error
}

type DeviceTokenRepository interface {
	Create(token *models.DeviceToken) (*models.DeviceToken, error)
	GetByUserID(userID primitive.ObjectID) ([]*models.DeviceToken, error)
	GetByToken(token string) (*models.DeviceToken, error)
	Update(token *models.DeviceToken) (*models.DeviceToken, error)
	Delete(id string) error
	DeleteByToken(token string) error
	GetActiveTokensByUserID(userID primitive.ObjectID) ([]*models.DeviceToken, error)
}

// Implementation
type notificationRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
	logger     utils.Logger
}

func NewNotificationRepository(db *mongo.Database, logger utils.Logger) NotificationRepository {
	return &notificationRepository{
		collection: db.Collection("notifications"),
		db:         db,
		logger:     logger,
	}
}

func (r *notificationRepository) Create(notification *models.Notification) (*models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	notification.ID = primitive.NewObjectID()
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, notification)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create notification")
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

func (r *notificationRepository) GetByID(id string) (*models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var notification models.Notification
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&notification)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, err
	}

	return &notification, nil
}

func (r *notificationRepository) GetUserNotifications(userID string, page, limit int) ([]*models.Notification, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"recipient_id": userObjectID,
		"is_deleted":   bson.M{"$ne": true},
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

	var notifications []*models.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *notificationRepository) GetNotificationsByType(userID string, notificationType models.NotificationType, page, limit int) ([]*models.Notification, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"recipient_id": userObjectID,
		"type":         notificationType,
		"is_deleted":   bson.M{"$ne": true},
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

	var notifications []*models.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *notificationRepository) GetNotificationsByCategory(userID string, category string, page, limit int) ([]*models.Notification, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"recipient_id": userObjectID,
		"category":     category,
		"is_deleted":   bson.M{"$ne": true},
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

	var notifications []*models.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *notificationRepository) Update(notification *models.Notification) (*models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	notification.UpdatedAt = time.Now()

	filter := bson.M{"_id": notification.ID}
	update := bson.M{"$set": notification}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("notification_id", notification.ID.Hex()).Msg("Failed to update notification")
		return nil, fmt.Errorf("failed to update notification: %w", err)
	}

	return notification, nil
}

func (r *notificationRepository) MarkAsRead(id, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_read":    true,
			"read_at":    &now,
			"updated_at": now,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("notification_id", id).Msg("Failed to mark notification as read")
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

func (r *notificationRepository) MarkAllAsRead(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	filter := bson.M{
		"recipient_id": userObjectID,
		"is_read":      false,
		"is_deleted":   bson.M{"$ne": true},
	}

	update := bson.M{
		"$set": bson.M{
			"is_read":    true,
			"read_at":    &now,
			"updated_at": now,
		},
	}

	_, err = r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to mark all notifications as read")
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

func (r *notificationRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("notification_id", id).Msg("Failed to delete notification")
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}

func (r *notificationRepository) ClearAllNotifications(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"recipient_id": userObjectID,
		"is_deleted":   bson.M{"$ne": true},
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to clear all notifications")
		return fmt.Errorf("failed to clear all notifications: %w", err)
	}

	return nil
}

func (r *notificationRepository) GetUnreadCount(userID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	filter := bson.M{
		"recipient_id": userObjectID,
		"is_read":      false,
		"is_deleted":   bson.M{"$ne": true},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get unread count")
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

func (r *notificationRepository) GetNotificationHistory(userID string, days int, page, limit int) ([]*models.Notification, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	startDate := time.Now().AddDate(0, 0, -days)
	filter := bson.M{
		"recipient_id": userObjectID,
		"created_at":   bson.M{"$gte": startDate},
		"is_deleted":   bson.M{"$ne": true},
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

	var notifications []*models.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *notificationRepository) GetNotificationsByStatus(userID string, status models.NotificationStatus, page, limit int) ([]*models.Notification, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{
		"recipient_id": userObjectID,
		"status":       status,
		"is_deleted":   bson.M{"$ne": true},
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

	var notifications []*models.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *notificationRepository) BulkInsert(notifications []*models.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	documents := make([]interface{}, len(notifications))
	for i, notification := range notifications {
		notification.ID = primitive.NewObjectID()
		notification.CreatedAt = time.Now()
		notification.UpdatedAt = time.Now()
		documents[i] = notification
	}

	_, err := r.collection.InsertMany(ctx, documents)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to bulk insert notifications")
		return fmt.Errorf("failed to bulk insert notifications: %w", err)
	}

	return nil
}

func (r *notificationRepository) GetDeliveryStatus(id string) (*models.Notification, error) {
	return r.GetByID(id)
}

func (r *notificationRepository) UpdateDeliveryResult(id string, result models.DeliveryResult) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$push": bson.M{
			"delivery_results": result,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("notification_id", id).Msg("Failed to update delivery result")
		return fmt.Errorf("failed to update delivery result: %w", err)
	}

	return nil
}

func (r *notificationRepository) GetAnalytics(userID string, startDate, endDate time.Time) (*models.NotificationAnalytics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"recipient_id": userObjectID,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
				"is_deleted": bson.M{"$ne": true},
			},
		},
		{
			"$group": bson.M{
				"_id":        nil,
				"total_sent": bson.M{"$sum": 1},
				"total_delivered": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$status", models.NotificationStatusDelivered}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"total_read": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$is_read", true}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"total_clicked": bson.M{
					"$sum": "$click_count",
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalSent      int64 `bson:"total_sent"`
		TotalDelivered int64 `bson:"total_delivered"`
		TotalRead      int64 `bson:"total_read"`
		TotalClicked   int64 `bson:"total_clicked"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	}

	analytics := &models.NotificationAnalytics{
		TotalSent:      result.TotalSent,
		TotalDelivered: result.TotalDelivered,
		TotalRead:      result.TotalRead,
		TotalClicked:   result.TotalClicked,
	}

	// Calculate rates
	if analytics.TotalSent > 0 {
		analytics.DeliveryRate = float64(analytics.TotalDelivered) / float64(analytics.TotalSent)
		analytics.ReadRate = float64(analytics.TotalRead) / float64(analytics.TotalSent)
		analytics.ClickRate = float64(analytics.TotalClicked) / float64(analytics.TotalSent)
	}

	return analytics, nil
}

func (r *notificationRepository) GetEngagementStats(templateID string, startDate, endDate time.Time) (*models.EngagementStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"template_id": templateID,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
				"is_deleted": bson.M{"$ne": true},
			},
		},
		{
			"$group": bson.M{
				"_id":        nil,
				"total_sent": bson.M{"$sum": 1},
				"total_read": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$is_read", true}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"total_clicked": bson.M{
					"$sum": "$click_count",
				},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalSent    int64 `bson:"total_sent"`
		TotalRead    int64 `bson:"total_read"`
		TotalClicked int64 `bson:"total_clicked"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	}

	stats := &models.EngagementStats{
		TemplateID:   templateID,
		TotalSent:    result.TotalSent,
		TotalRead:    result.TotalRead,
		TotalClicked: result.TotalClicked,
	}

	// Calculate rates
	if stats.TotalSent > 0 {
		stats.ReadRate = float64(stats.TotalRead) / float64(stats.TotalSent)
		stats.ClickRate = float64(stats.TotalClicked) / float64(stats.TotalSent)
	}

	return stats, nil
}

// Template Repository Implementation
type notificationTemplateRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
	logger     utils.Logger
}

func NewNotificationTemplateRepository(db *mongo.Database, logger utils.Logger) NotificationTemplateRepository {
	return &notificationTemplateRepository{
		collection: db.Collection("notification_templates"),
		db:         db,
		logger:     logger,
	}
}

func (r *notificationTemplateRepository) Create(template *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	template.ID = primitive.NewObjectID()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, template)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create notification template")
		return nil, fmt.Errorf("failed to create notification template: %w", err)
	}

	return template, nil
}

func (r *notificationTemplateRepository) GetByID(id string) (*models.NotificationTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var template models.NotificationTemplate
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objectID,
		"is_active":  true,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&template)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("notification template not found")
		}
		return nil, err
	}

	return &template, nil
}

func (r *notificationTemplateRepository) GetByType(notificationType models.NotificationType) ([]*models.NotificationTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"type":       notificationType,
		"is_active":  true,
		"is_deleted": bson.M{"$ne": true},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.NotificationTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *notificationTemplateRepository) GetByCategory(category string) ([]*models.NotificationTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"category":   category,
		"is_active":  true,
		"is_deleted": bson.M{"$ne": true},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.NotificationTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *notificationTemplateRepository) GetActive() ([]*models.NotificationTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"is_active":  true,
		"is_deleted": bson.M{"$ne": true},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.NotificationTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (r *notificationTemplateRepository) Update(template *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	template.UpdatedAt = time.Now()

	filter := bson.M{"_id": template.ID}
	update := bson.M{"$set": template}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("template_id", template.ID.Hex()).Msg("Failed to update notification template")
		return nil, fmt.Errorf("failed to update notification template: %w", err)
	}

	return template, nil
}

func (r *notificationTemplateRepository) Delete(id string) error {
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
	if err != nil {
		r.logger.Error().Err(err).Str("template_id", id).Msg("Failed to delete notification template")
		return fmt.Errorf("failed to delete notification template: %w", err)
	}

	return nil
}

// Preference Repository Implementation
type notificationPreferenceRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
	logger     utils.Logger
}

func NewNotificationPreferenceRepository(db *mongo.Database, logger utils.Logger) NotificationPreferenceRepository {
	return &notificationPreferenceRepository{
		collection: db.Collection("notification_preferences"),
		db:         db,
		logger:     logger,
	}
}

func (r *notificationPreferenceRepository) Create(preferences *models.NotificationPreferences) (*models.NotificationPreferences, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	preferences.ID = primitive.NewObjectID()
	preferences.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, preferences)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create notification preferences")
		return nil, fmt.Errorf("failed to create notification preferences: %w", err)
	}

	return preferences, nil
}

func (r *notificationPreferenceRepository) GetByUserID(userID primitive.ObjectID) (*models.NotificationPreferences, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var preferences models.NotificationPreferences
	err := r.collection.FindOne(ctx, bson.M{
		"user_id": userID,
	}).Decode(&preferences)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Return nil if not found (will create default)
		}
		return nil, err
	}

	return &preferences, nil
}

func (r *notificationPreferenceRepository) Update(preferences *models.NotificationPreferences) (*models.NotificationPreferences, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	preferences.UpdatedAt = time.Now()

	filter := bson.M{"user_id": preferences.UserID}
	update := bson.M{"$set": preferences}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", preferences.UserID.Hex()).Msg("Failed to update notification preferences")
		return nil, fmt.Errorf("failed to update notification preferences: %w", err)
	}

	return preferences, nil
}

func (r *notificationPreferenceRepository) Delete(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to delete notification preferences")
		return fmt.Errorf("failed to delete notification preferences: %w", err)
	}

	return nil
}

// Device Token Repository Implementation
type deviceTokenRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
	logger     utils.Logger
}

func NewDeviceTokenRepository(db *mongo.Database, logger utils.Logger) DeviceTokenRepository {
	return &deviceTokenRepository{
		collection: db.Collection("device_tokens"),
		db:         db,
		logger:     logger,
	}
}

func (r *deviceTokenRepository) Create(token *models.DeviceToken) (*models.DeviceToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token.Token = primitive.NewObjectID().Hex()
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to create device token")
		return nil, fmt.Errorf("failed to create device token: %w", err)
	}

	return token, nil
}

func (r *deviceTokenRepository) GetByUserID(userID primitive.ObjectID) ([]*models.DeviceToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"is_deleted": bson.M{"$ne": true},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []*models.DeviceToken
	if err = cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *deviceTokenRepository) GetByToken(token string) (*models.DeviceToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var deviceToken models.DeviceToken
	err := r.collection.FindOne(ctx, bson.M{
		"token":      token,
		"is_deleted": bson.M{"$ne": true},
	}).Decode(&deviceToken)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("device token not found")
		}
		return nil, err
	}

	return &deviceToken, nil
}

func (r *deviceTokenRepository) Update(token *models.DeviceToken) (*models.DeviceToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token.UpdatedAt = time.Now()

	filter := bson.M{"_id": token.Token}
	update := bson.M{"$set": token}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("token_id", token.Token).Msg("Failed to update device token")
		return nil, fmt.Errorf("failed to update device token: %w", err)
	}

	return token, nil
}

func (r *deviceTokenRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("token_id", id).Msg("Failed to delete device token")
		return fmt.Errorf("failed to delete device token: %w", err)
	}

	return nil
}

func (r *deviceTokenRepository) DeleteByToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"token": token}, update)
	if err != nil {
		r.logger.Error().Err(err).Str("token", token).Msg("Failed to delete device token by token")
		return fmt.Errorf("failed to delete device token by token: %w", err)
	}

	return nil
}

func (r *deviceTokenRepository) GetActiveTokensByUserID(userID primitive.ObjectID) ([]*models.DeviceToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"is_active":  true,
		"is_deleted": bson.M{"$ne": true},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []*models.DeviceToken
	if err = cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}
