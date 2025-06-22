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

// PaymentRepository interface defines payment-related database operations
type PaymentRepository interface {
	// Payment Methods
	CreatePaymentMethod(method *models.PaymentMethod) (*models.PaymentMethod, error)
	GetPaymentMethodByID(id primitive.ObjectID) (*models.PaymentMethod, error)
	GetPaymentMethodsByUserID(userID primitive.ObjectID) ([]*models.PaymentMethod, error)
	UpdatePaymentMethod(method *models.PaymentMethod) (*models.PaymentMethod, error)
	DeletePaymentMethod(id primitive.ObjectID) error
	GetDefaultPaymentMethod(userID primitive.ObjectID) (*models.PaymentMethod, error)

	// Wallets
	CreateWallet(wallet *models.Wallet) (*models.Wallet, error)
	GetWalletByUserID(userID primitive.ObjectID) (*models.Wallet, error)
	GetWalletByID(id primitive.ObjectID) (*models.Wallet, error)
	UpdateWallet(wallet *models.Wallet) (*models.Wallet, error)
	UpdateWalletBalance(userID primitive.ObjectID, amount float64) error

	// Transactions
	CreateTransaction(transaction *models.Transaction) (*models.Transaction, error)
	GetTransactionByID(id primitive.ObjectID) (*models.Transaction, error)
	GetTransactionsByUserID(userID primitive.ObjectID, page, limit int) ([]*models.Transaction, int64, error)
	GetTransactionsByFilter(filter map[string]interface{}, page, limit int) ([]*models.Transaction, int64, error)
	UpdateTransaction(transaction *models.Transaction) (*models.Transaction, error)
	UpdateTransactionStatus(id primitive.ObjectID, status models.PaymentStatus) error

	// Refunds
	CreateRefund(refund *models.Refund) (*models.Refund, error)
	GetRefundByID(id primitive.ObjectID) (*models.Refund, error)
	GetRefundsByUserID(userID primitive.ObjectID, page, limit int) ([]*models.Refund, int64, error)
	GetRefundsByStatus(status models.RefundStatus, page, limit int) ([]*models.Refund, int64, error)
	UpdateRefund(refund *models.Refund) (*models.Refund, error)
	UpdateRefundStatus(id primitive.ObjectID, status models.RefundStatus) error

	// Payment Disputes
	CreateDispute(dispute *models.PaymentDispute) (*models.PaymentDispute, error)
	GetDisputeByID(id primitive.ObjectID) (*models.PaymentDispute, error)
	GetDisputesByUserID(userID primitive.ObjectID, page, limit int) ([]*models.PaymentDispute, int64, error)
	GetDisputesByTransactionID(transactionID primitive.ObjectID) ([]*models.PaymentDispute, error)
	UpdateDispute(dispute *models.PaymentDispute) (*models.PaymentDispute, error)
	UpdateDisputeStatus(id primitive.ObjectID, status string) error

	// Analytics & Reporting
	GetUserSpendingStats(userID primitive.ObjectID, startDate, endDate time.Time) (*PaymentStats, error)
	GetTransactionsByDateRange(userID primitive.ObjectID, startDate, endDate time.Time) ([]*models.Transaction, error)
	GetSpendingByCategory(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]float64, error)
	GetPaymentMethodUsage(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]int64, error)

	// Search & Filtering
	SearchTransactions(query string, filters map[string]interface{}, page, limit int) ([]*models.Transaction, int64, error)
	GetTransactionsByType(userID primitive.ObjectID, transactionType models.TransactionType, page, limit int) ([]*models.Transaction, int64, error)
	GetTransactionsByStatus(status models.PaymentStatus, page, limit int) ([]*models.Transaction, int64, error)

	// Cleanup & Maintenance
	CleanupExpiredTransactions(retentionDays int) error
	ArchiveOldTransactions(archiveAfterDays int) error
	CreateIndexes() error
}

// Supporting types for repository operations
type PaymentStats struct {
	UserID               primitive.ObjectID `json:"user_id" bson:"user_id"`
	TotalTransactions    int64              `json:"total_transactions" bson:"total_transactions"`
	TotalSpent           float64            `json:"total_spent" bson:"total_spent"`
	TotalRefunded        float64            `json:"total_refunded" bson:"total_refunded"`
	AverageTransaction   float64            `json:"average_transaction" bson:"average_transaction"`
	SuccessfulPayments   int64              `json:"successful_payments" bson:"successful_payments"`
	FailedPayments       int64              `json:"failed_payments" bson:"failed_payments"`
	SuccessRate          float64            `json:"success_rate" bson:"success_rate"`
	TopCategory          string             `json:"top_category" bson:"top_category"`
	TopPaymentMethod     string             `json:"top_payment_method" bson:"top_payment_method"`
	LastTransactionDate  time.Time          `json:"last_transaction_date" bson:"last_transaction_date"`
	CurrentWalletBalance float64            `json:"current_wallet_balance" bson:"current_wallet_balance"`
}

// paymentRepository implements PaymentRepository interface
type paymentRepository struct {
	db     *mongo.Database
	logger utils.Logger
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *mongo.Database, logger utils.Logger) PaymentRepository {
	repo := &paymentRepository{
		db:     db,
		logger: logger,
	}

	// Create indexes for better performance
	go repo.CreateIndexes()

	return repo
}

// Payment Methods

func (r *paymentRepository) CreatePaymentMethod(method *models.PaymentMethod) (*models.PaymentMethod, error) {
	collection := r.db.Collection("payment_methods")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	method.CreatedAt = time.Now()
	method.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, method)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", method.UserID.Hex()).Msg("Failed to create payment method")
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	method.ID = result.InsertedID.(primitive.ObjectID)
	return method, nil
}

func (r *paymentRepository) GetPaymentMethodByID(id primitive.ObjectID) (*models.PaymentMethod, error) {
	collection := r.db.Collection("payment_methods")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var method models.PaymentMethod
	err := collection.FindOne(ctx, bson.M{"_id": id, "is_active": true}).Decode(&method)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment method not found")
		}
		r.logger.Error().Err(err).Str("method_id", id.Hex()).Msg("Failed to get payment method")
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}

	return &method, nil
}

func (r *paymentRepository) GetPaymentMethodsByUserID(userID primitive.ObjectID) ([]*models.PaymentMethod, error) {
	collection := r.db.Collection("payment_methods")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":   userID,
		"is_active": true,
	}

	opts := options.Find().SetSort(bson.D{{Key: "is_default", Value: -1}, {Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get payment methods")
		return nil, fmt.Errorf("failed to get payment methods: %w", err)
	}
	defer cursor.Close(ctx)

	var methods []*models.PaymentMethod
	if err := cursor.All(ctx, &methods); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode payment methods")
		return nil, fmt.Errorf("failed to decode payment methods: %w", err)
	}

	return methods, nil
}

func (r *paymentRepository) UpdatePaymentMethod(method *models.PaymentMethod) (*models.PaymentMethod, error) {
	collection := r.db.Collection("payment_methods")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	method.UpdatedAt = time.Now()

	filter := bson.M{"_id": method.ID}
	update := bson.M{"$set": method}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("method_id", method.ID.Hex()).Msg("Failed to update payment method")
		return nil, fmt.Errorf("failed to update payment method: %w", err)
	}

	return method, nil
}

func (r *paymentRepository) DeletePaymentMethod(id primitive.ObjectID) error {
	collection := r.db.Collection("payment_methods")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("method_id", id.Hex()).Msg("Failed to delete payment method")
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	return nil
}

func (r *paymentRepository) GetDefaultPaymentMethod(userID primitive.ObjectID) (*models.PaymentMethod, error) {
	collection := r.db.Collection("payment_methods")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"is_default": true,
		"is_active":  true,
	}

	var method models.PaymentMethod
	err := collection.FindOne(ctx, filter).Decode(&method)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no default payment method found")
		}
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get default payment method")
		return nil, fmt.Errorf("failed to get default payment method: %w", err)
	}

	return &method, nil
}

// Wallets

func (r *paymentRepository) CreateWallet(wallet *models.Wallet) (*models.Wallet, error) {
	collection := r.db.Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, wallet)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", wallet.UserID.Hex()).Msg("Failed to create wallet")
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	wallet.ID = result.InsertedID.(primitive.ObjectID)
	return wallet, nil
}

func (r *paymentRepository) GetWalletByUserID(userID primitive.ObjectID) (*models.Wallet, error) {
	collection := r.db.Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wallet models.Wallet
	err := collection.FindOne(ctx, bson.M{"user_id": userID, "is_active": true}).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wallet not found")
		}
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get wallet")
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

func (r *paymentRepository) GetWalletByID(id primitive.ObjectID) (*models.Wallet, error) {
	collection := r.db.Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wallet models.Wallet
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wallet not found")
		}
		r.logger.Error().Err(err).Str("wallet_id", id.Hex()).Msg("Failed to get wallet")
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

func (r *paymentRepository) UpdateWallet(wallet *models.Wallet) (*models.Wallet, error) {
	collection := r.db.Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wallet.UpdatedAt = time.Now()

	filter := bson.M{"_id": wallet.ID}
	update := bson.M{"$set": wallet}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("wallet_id", wallet.ID.Hex()).Msg("Failed to update wallet")
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return wallet, nil
}

func (r *paymentRepository) UpdateWalletBalance(userID primitive.ObjectID, amount float64) error {
	collection := r.db.Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$inc": bson.M{"balance": amount},
		"$set": bson.M{"updated_at": time.Now()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Float64("amount", amount).Msg("Failed to update wallet balance")
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	return nil
}

// Transactions

func (r *paymentRepository) CreateTransaction(transaction *models.Transaction) (*models.Transaction, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, transaction)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", transaction.UserID.Hex()).Msg("Failed to create transaction")
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	transaction.ID = result.InsertedID.(primitive.ObjectID)
	return transaction, nil
}

func (r *paymentRepository) GetTransactionByID(id primitive.ObjectID) (*models.Transaction, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction models.Transaction
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("transaction not found")
		}
		r.logger.Error().Err(err).Str("transaction_id", id.Hex()).Msg("Failed to get transaction")
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

func (r *paymentRepository) GetTransactionsByUserID(userID primitive.ObjectID, page, limit int) ([]*models.Transaction, int64, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}

	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to count transactions")
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find documents with pagination
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to find transactions")
		return nil, 0, fmt.Errorf("failed to find transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode transactions")
		return nil, 0, fmt.Errorf("failed to decode transactions: %w", err)
	}

	return transactions, total, nil
}

func (r *paymentRepository) GetTransactionsByFilter(filter map[string]interface{}, page, limit int) ([]*models.Transaction, int64, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert filter to BSON
	bsonFilter := bson.M{}
	for k, v := range filter {
		bsonFilter[k] = v
	}

	// Count total documents
	total, err := collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count transactions")
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find documents with pagination
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to find transactions")
		return nil, 0, fmt.Errorf("failed to find transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode transactions")
		return nil, 0, fmt.Errorf("failed to decode transactions: %w", err)
	}

	return transactions, total, nil
}

func (r *paymentRepository) UpdateTransaction(transaction *models.Transaction) (*models.Transaction, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transaction.UpdatedAt = time.Now()

	filter := bson.M{"_id": transaction.ID}
	update := bson.M{"$set": transaction}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("transaction_id", transaction.ID.Hex()).Msg("Failed to update transaction")
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	return transaction, nil
}

func (r *paymentRepository) UpdateTransactionStatus(id primitive.ObjectID, status models.PaymentStatus) error {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("transaction_id", id.Hex()).Msg("Failed to update transaction status")
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// Refunds

func (r *paymentRepository) CreateRefund(refund *models.Refund) (*models.Refund, error) {
	collection := r.db.Collection("refunds")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	refund.CreatedAt = time.Now()
	refund.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, refund)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", refund.UserID.Hex()).Msg("Failed to create refund")
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	refund.ID = result.InsertedID.(primitive.ObjectID)
	return refund, nil
}

func (r *paymentRepository) GetRefundByID(id primitive.ObjectID) (*models.Refund, error) {
	collection := r.db.Collection("refunds")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var refund models.Refund
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&refund)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("refund not found")
		}
		r.logger.Error().Err(err).Str("refund_id", id.Hex()).Msg("Failed to get refund")
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	return &refund, nil
}

func (r *paymentRepository) GetRefundsByUserID(userID primitive.ObjectID, page, limit int) ([]*models.Refund, int64, error) {
	collection := r.db.Collection("refunds")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}

	return r.findRefundsWithPagination(ctx, collection, filter, page, limit)
}

func (r *paymentRepository) GetRefundsByStatus(status models.RefundStatus, page, limit int) ([]*models.Refund, int64, error) {
	collection := r.db.Collection("refunds")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"status": status}

	return r.findRefundsWithPagination(ctx, collection, filter, page, limit)
}

func (r *paymentRepository) UpdateRefund(refund *models.Refund) (*models.Refund, error) {
	collection := r.db.Collection("refunds")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	refund.UpdatedAt = time.Now()

	filter := bson.M{"_id": refund.ID}
	update := bson.M{"$set": refund}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("refund_id", refund.ID.Hex()).Msg("Failed to update refund")
		return nil, fmt.Errorf("failed to update refund: %w", err)
	}

	return refund, nil
}

func (r *paymentRepository) UpdateRefundStatus(id primitive.ObjectID, status models.RefundStatus) error {
	collection := r.db.Collection("refunds")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("refund_id", id.Hex()).Msg("Failed to update refund status")
		return fmt.Errorf("failed to update refund status: %w", err)
	}

	return nil
}

// Payment Disputes

func (r *paymentRepository) CreateDispute(dispute *models.PaymentDispute) (*models.PaymentDispute, error) {
	collection := r.db.Collection("payment_disputes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dispute.CreatedAt = time.Now()
	dispute.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, dispute)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", dispute.DisputerID.Hex()).Msg("Failed to create dispute")
		return nil, fmt.Errorf("failed to create dispute: %w", err)
	}

	dispute.ID = result.InsertedID.(primitive.ObjectID)
	return dispute, nil
}

func (r *paymentRepository) GetDisputeByID(id primitive.ObjectID) (*models.PaymentDispute, error) {
	collection := r.db.Collection("payment_disputes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var dispute models.PaymentDispute
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&dispute)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("dispute not found")
		}
		r.logger.Error().Err(err).Str("dispute_id", id.Hex()).Msg("Failed to get dispute")
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}

	return &dispute, nil
}

func (r *paymentRepository) GetDisputesByUserID(userID primitive.ObjectID, page, limit int) ([]*models.PaymentDispute, int64, error) {
	collection := r.db.Collection("payment_disputes")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"disputer_id": userID}

	return r.findDisputesWithPagination(ctx, collection, filter, page, limit)
}

func (r *paymentRepository) GetDisputesByTransactionID(transactionID primitive.ObjectID) ([]*models.PaymentDispute, error) {
	collection := r.db.Collection("payment_disputes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"transaction_id": transactionID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("transaction_id", transactionID.Hex()).Msg("Failed to get disputes")
		return nil, fmt.Errorf("failed to get disputes: %w", err)
	}
	defer cursor.Close(ctx)

	var disputes []*models.PaymentDispute
	if err := cursor.All(ctx, &disputes); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode disputes")
		return nil, fmt.Errorf("failed to decode disputes: %w", err)
	}

	return disputes, nil
}

func (r *paymentRepository) UpdateDispute(dispute *models.PaymentDispute) (*models.PaymentDispute, error) {
	collection := r.db.Collection("payment_disputes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dispute.UpdatedAt = time.Now()

	filter := bson.M{"_id": dispute.ID}
	update := bson.M{"$set": dispute}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("dispute_id", dispute.ID.Hex()).Msg("Failed to update dispute")
		return nil, fmt.Errorf("failed to update dispute: %w", err)
	}

	return dispute, nil
}

func (r *paymentRepository) UpdateDisputeStatus(id primitive.ObjectID, status string) error {
	collection := r.db.Collection("payment_disputes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("dispute_id", id.Hex()).Msg("Failed to update dispute status")
		return fmt.Errorf("failed to update dispute status: %w", err)
	}

	return nil
}

// Analytics & Reporting

func (r *paymentRepository) GetUserSpendingStats(userID primitive.ObjectID, startDate, endDate time.Time) (*PaymentStats, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build aggregation pipeline
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"created_at": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":                nil,
			"total_transactions": bson.M{"$sum": 1},
			"total_spent":        bson.M{"$sum": "$amount"},
			"successful_payments": bson.M{"$sum": bson.M{"$cond": bson.A{
				bson.M{"$eq": bson.A{"$status", models.PaymentStatusCompleted}},
				1,
				0,
			}}},
			"failed_payments": bson.M{"$sum": bson.M{"$cond": bson.A{
				bson.M{"$eq": bson.A{"$status", models.PaymentStatusFailed}},
				1,
				0,
			}}},
			"last_transaction": bson.M{"$max": "$created_at"},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get spending stats")
		return nil, fmt.Errorf("failed to get spending stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	if len(results) == 0 {
		return &PaymentStats{UserID: userID}, nil
	}

	result := results[0]
	totalTx := getInt64FromBSON(result, "total_transactions")
	successfulPayments := getInt64FromBSON(result, "successful_payments")
	failedPayments := getInt64FromBSON(result, "failed_payments")
	totalSpent := getFloat64FromBSON(result, "total_spent")

	stats := &PaymentStats{
		UserID:              userID,
		TotalTransactions:   totalTx,
		TotalSpent:          totalSpent,
		SuccessfulPayments:  successfulPayments,
		FailedPayments:      failedPayments,
		LastTransactionDate: getTimeFromBSON(result, "last_transaction"),
	}

	// Calculate derived metrics
	if totalTx > 0 {
		stats.AverageTransaction = totalSpent / float64(totalTx)
		stats.SuccessRate = float64(successfulPayments) / float64(totalTx) * 100
	}

	// Get current wallet balance
	wallet, err := r.GetWalletByUserID(userID)
	if err == nil {
		stats.CurrentWalletBalance = wallet.Balance
	}

	return stats, nil
}

func (r *paymentRepository) GetTransactionsByDateRange(userID primitive.ObjectID, startDate, endDate time.Time) ([]*models.Transaction, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get transactions by date range")
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode transactions")
		return nil, fmt.Errorf("failed to decode transactions: %w", err)
	}

	return transactions, nil
}

func (r *paymentRepository) GetSpendingByCategory(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]float64, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build aggregation pipeline
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"created_at": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
			"status": models.PaymentStatusCompleted,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":    "$type",
			"amount": bson.M{"$sum": "$amount"},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get spending by category")
		return nil, fmt.Errorf("failed to get spending by category: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode category spending: %w", err)
	}

	categorySpending := make(map[string]float64)
	for _, result := range results {
		category := result["_id"].(string)
		amount := getFloat64FromBSON(result, "amount")
		categorySpending[category] = amount
	}

	return categorySpending, nil
}

func (r *paymentRepository) GetPaymentMethodUsage(userID primitive.ObjectID, startDate, endDate time.Time) (map[string]int64, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build aggregation pipeline
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"created_at": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
			"status": models.PaymentStatusCompleted,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$payment_method",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to get payment method usage")
		return nil, fmt.Errorf("failed to get payment method usage: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode payment method usage: %w", err)
	}

	methodUsage := make(map[string]int64)
	for _, result := range results {
		method := result["_id"].(string)
		count := getInt64FromBSON(result, "count")
		methodUsage[method] = count
	}

	return methodUsage, nil
}

// Search & Filtering

func (r *paymentRepository) SearchTransactions(query string, filters map[string]interface{}, page, limit int) ([]*models.Transaction, int64, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build search filter
	searchFilter := bson.M{}

	// Add text search if query provided
	if query != "" {
		searchFilter["$text"] = bson.M{"$search": query}
	}

	// Add additional filters
	for k, v := range filters {
		searchFilter[k] = v
	}

	return r.findTransactionsWithPagination(ctx, collection, searchFilter, page, limit)
}

func (r *paymentRepository) GetTransactionsByType(userID primitive.ObjectID, transactionType models.TransactionType, page, limit int) ([]*models.Transaction, int64, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"type":    transactionType,
	}

	return r.findTransactionsWithPagination(ctx, collection, filter, page, limit)
}

func (r *paymentRepository) GetTransactionsByStatus(status models.PaymentStatus, page, limit int) ([]*models.Transaction, int64, error) {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"status": status}

	return r.findTransactionsWithPagination(ctx, collection, filter, page, limit)
}

// Cleanup & Maintenance

func (r *paymentRepository) CleanupExpiredTransactions(retentionDays int) error {
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	filter := bson.M{
		"created_at": bson.M{"$lt": cutoffDate},
		"status":     models.PaymentStatusFailed,
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup expired transactions")
		return fmt.Errorf("failed to cleanup expired transactions: %w", err)
	}

	r.logger.Info().Int64("deleted_count", result.DeletedCount).Msg("Cleaned up expired transactions")
	return nil
}

func (r *paymentRepository) ArchiveOldTransactions(archiveAfterDays int) error {
	// Implementation would move old transactions to archive collection
	// For now, just mark them as archived
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoffDate := time.Now().AddDate(0, 0, -archiveAfterDays)
	filter := bson.M{
		"created_at": bson.M{"$lt": cutoffDate},
		"status":     models.PaymentStatusCompleted,
		"archived":   bson.M{"$ne": true},
	}

	update := bson.M{"$set": bson.M{"archived": true, "archived_at": time.Now()}}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to archive old transactions")
		return fmt.Errorf("failed to archive old transactions: %w", err)
	}

	r.logger.Info().Int64("archived_count", result.ModifiedCount).Msg("Archived old transactions")
	return nil
}

func (r *paymentRepository) CreateIndexes() error {
	collections := map[string][]mongo.IndexModel{
		"payment_methods": {
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_default", Value: -1}}},
			{Keys: bson.D{{Key: "is_active", Value: 1}}},
			{Keys: bson.D{{Key: "created_at", Value: -1}}},
		},
		"wallets": {
			{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "is_active", Value: 1}}},
		},
		"transactions": {
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
			{Keys: bson.D{{Key: "ride_id", Value: 1}}},
			{Keys: bson.D{{Key: "type", Value: 1}}},
			{Keys: bson.D{{Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "created_at", Value: -1}}},
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}, {Key: "created_at", Value: -1}}},
			{Keys: bson.D{{Key: "external_id", Value: 1}}},
		},
		"refunds": {
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
			{Keys: bson.D{{Key: "transaction_id", Value: 1}}},
			{Keys: bson.D{{Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "created_at", Value: -1}}},
		},
		"payment_disputes": {
			{Keys: bson.D{{Key: "disputer_id", Value: 1}}},
			{Keys: bson.D{{Key: "transaction_id", Value: 1}}},
			{Keys: bson.D{{Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "created_at", Value: -1}}},
		},
	}

	for collectionName, indexes := range collections {
		collection := r.db.Collection(collectionName)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		_, err := collection.Indexes().CreateMany(ctx, indexes)
		cancel()

		if err != nil {
			r.logger.Error().Err(err).Str("collection", collectionName).Msg("Failed to create indexes")
			return fmt.Errorf("failed to create indexes for %s: %w", collectionName, err)
		}
	}

	// Create text index for transaction search
	collection := r.db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "description", Value: "text"},
			{Key: "notes", Value: "text"},
		},
	}

	_, err := collection.Indexes().CreateOne(ctx, textIndex)
	if err != nil {
		r.logger.Warn().Err(err).Msg("Failed to create text index for transactions")
	}

	r.logger.Info().Msg("Created indexes for payment collections")
	return nil
}

// Helper methods

func (r *paymentRepository) findTransactionsWithPagination(ctx context.Context, collection *mongo.Collection, filter bson.M, page, limit int) ([]*models.Transaction, int64, error) {
	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find documents with pagination
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, 0, fmt.Errorf("failed to decode documents: %w", err)
	}

	return transactions, total, nil
}

func (r *paymentRepository) findRefundsWithPagination(ctx context.Context, collection *mongo.Collection, filter bson.M, page, limit int) ([]*models.Refund, int64, error) {
	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find documents with pagination
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var refunds []*models.Refund
	if err := cursor.All(ctx, &refunds); err != nil {
		return nil, 0, fmt.Errorf("failed to decode documents: %w", err)
	}

	return refunds, total, nil
}

func (r *paymentRepository) findDisputesWithPagination(ctx context.Context, collection *mongo.Collection, filter bson.M, page, limit int) ([]*models.PaymentDispute, int64, error) {
	// Count total documents
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Find documents with pagination
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var disputes []*models.PaymentDispute
	if err := cursor.All(ctx, &disputes); err != nil {
		return nil, 0, fmt.Errorf("failed to decode documents: %w", err)
	}

	return disputes, total, nil
}

// Helper functions for BSON value extraction
func getInt64FromBSON(doc bson.M, key string) int64 {
	if val, ok := doc[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int32:
			return int64(v)
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

func getTimeFromBSON(doc bson.M, key string) time.Time {
	if val, ok := doc[key]; ok {
		if t, ok := val.(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}
