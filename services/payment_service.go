package services

import (
	"errors"
	"fmt"
	"time"

	"ondrive/models"
	"ondrive/repositories"
	"ondrive/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentService interface defines payment-related business logic
type PaymentService interface {
	// Payment Methods
	GetPaymentMethods(userID string) ([]*models.PaymentMethod, error)
	AddPaymentMethod(userID string, req interface{}) (*models.PaymentMethod, error)
	UpdatePaymentMethod(userID string, methodID primitive.ObjectID, req interface{}) (*models.PaymentMethod, error)
	DeletePaymentMethod(userID string, methodID primitive.ObjectID) error
	SetDefaultPaymentMethod(userID string, methodID primitive.ObjectID) error

	// Wallet Management
	GetWallet(userID string) (*models.Wallet, error)
	AddMoneyToWallet(userID string, amount float64, paymentMethodID, currency string) (*models.Transaction, error)
	WithdrawFromWallet(userID string, amount float64, bankAccount, currency, reason string) (*models.Transaction, error)
	GetWalletTransactions(userID string, page, limit int, transactionType string) ([]*models.Transaction, int64, error)

	// Payment Processing
	ProcessPayment(userID string, req interface{}) (*models.Transaction, error)
	PayForRide(userID string, rideID primitive.ObjectID, paymentMethodID string, tipAmount float64) (*models.Transaction, error)
	GetRidePaymentStatus(userID string, rideID primitive.ObjectID) (*RidePaymentStatus, error)

	// Payment History
	GetPaymentHistory(userID string, page, limit int, transactionType, status, startDate, endDate string) ([]*models.Transaction, int64, error)
	GetPaymentReceipt(userID string, paymentID primitive.ObjectID) (*PaymentReceipt, error)
	SendReceiptEmail(userID string, paymentID primitive.ObjectID, email string) error

	// Refunds & Disputes
	RequestRefund(userID string, req interface{}) (*models.Refund, error)
	GetRefunds(userID string, page, limit int, status string) ([]*models.Refund, int64, error)
	GetRefundStatus(userID string, refundID primitive.ObjectID) (*models.Refund, error)
	CreateDispute(userID string, req interface{}) (*models.PaymentDispute, error)
	GetDisputes(userID string, page, limit int, status string) ([]*models.PaymentDispute, int64, error)

	// Billing & Invoices
	GetInvoices(userID string, page, limit int, status string) ([]*Invoice, int64, error)
	GetInvoice(userID string, invoiceID primitive.ObjectID) (*Invoice, error)
	PayInvoice(userID string, invoiceID primitive.ObjectID, paymentMethodID string) (*models.Transaction, error)

	// Cards & Banking
	VerifyCard(userID string, req interface{}) (*CardVerificationResult, error)
	GetSupportedBanks(country string) ([]*SupportedBank, error)
	AddBankAccount(userID string, req interface{}) (*models.BankDetails, error)
	GetBankAccounts(userID string) ([]*models.BankDetails, error)

	// Analytics
	GetSpendingSummary(userID string, period, category string) (*SpendingSummary, error)
	GetMonthlyReport(userID string, year, month int) (*MonthlyReport, error)

	// Webhooks
	HandleStripeWebhook(body []byte, signature string) error
	HandlePayPalWebhook(body []byte, headers map[string]string) error

	// Promo Codes & Credits
	GetPromoCodes(userID string) ([]*PromoCode, error)
	ApplyPromoCode(userID, code string, amount float64) (*PromoCodeResult, error)
	GetCredits(userID string) (*UserCredits, error)
	ApplyCredits(userID string, amount float64, creditType string) (*CreditResult, error)

	// Internal methods
	DebitWallet(userID string, amount float64, description string, transactionType models.TransactionType) (*models.Transaction, error)
	CreditWallet(userID string, amount float64, description string, transactionType models.TransactionType) (*models.Transaction, error)
	ValidatePaymentMethod(userID string, paymentMethodID string) (*models.PaymentMethod, error)
}

// Response types for payment service
type RidePaymentStatus struct {
	RideID        primitive.ObjectID       `json:"ride_id"`
	PaymentStatus models.PaymentStatus     `json:"payment_status"`
	Amount        float64                  `json:"amount"`
	Currency      string                   `json:"currency"`
	PaymentMethod models.PaymentMethodType `json:"payment_method"`
	TransactionID *primitive.ObjectID      `json:"transaction_id,omitempty"`
	LastUpdated   time.Time                `json:"last_updated"`
	FailureReason string                   `json:"failure_reason,omitempty"`
	TipAmount     float64                  `json:"tip_amount"`
	TotalAmount   float64                  `json:"total_amount"`
}

type PaymentReceipt struct {
	TransactionID primitive.ObjectID          `json:"transaction_id"`
	ReceiptNumber string                      `json:"receipt_number"`
	Date          time.Time                   `json:"date"`
	Amount        float64                     `json:"amount"`
	Currency      string                      `json:"currency"`
	PaymentMethod models.PaymentMethodType    `json:"payment_method"`
	Description   string                      `json:"description"`
	Breakdown     models.TransactionBreakdown `json:"breakdown"`
	Status        models.PaymentStatus        `json:"status"`
	RideID        *primitive.ObjectID         `json:"ride_id,omitempty"`
	UserInfo      UserReceiptInfo             `json:"user_info"`
	ServiceInfo   ServiceReceiptInfo          `json:"service_info"`
}

type UserReceiptInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type ServiceReceiptInfo struct {
	ServiceName  string `json:"service_name"`
	Website      string `json:"website"`
	SupportEmail string `json:"support_email"`
	TaxID        string `json:"tax_id"`
}

type CardVerificationResult struct {
	Valid       bool   `json:"valid"`
	CardType    string `json:"card_type"`
	LastFour    string `json:"last_four"`
	Bank        string `json:"bank"`
	Country     string `json:"country"`
	Fingerprint string `json:"fingerprint"`
}

type SupportedBank struct {
	BankCode    string `json:"bank_code"`
	BankName    string `json:"bank_name"`
	Country     string `json:"country"`
	LogoURL     string `json:"logo_url"`
	IsSupported bool   `json:"is_supported"`
}

type SpendingSummary struct {
	UserID             string                  `json:"user_id"`
	Period             string                  `json:"period"`
	TotalSpent         float64                 `json:"total_spent"`
	TotalTransactions  int64                   `json:"total_transactions"`
	AverageTransaction float64                 `json:"average_transaction"`
	CategoryBreakdown  []CategorySpending      `json:"category_breakdown"`
	MethodBreakdown    []PaymentMethodSpending `json:"method_breakdown"`
	MonthlyTrend       []MonthlySpending       `json:"monthly_trend"`
	TopCategories      []CategorySpending      `json:"top_categories"`
	Currency           string                  `json:"currency"`
}

type CategorySpending struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type PaymentMethodSpending struct {
	Method     models.PaymentMethodType `json:"method"`
	Amount     float64                  `json:"amount"`
	Count      int64                    `json:"count"`
	Percentage float64                  `json:"percentage"`
}

type MonthlySpending struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
}

type MonthlyReport struct {
	UserID             string                  `json:"user_id"`
	Year               int                     `json:"year"`
	Month              int                     `json:"month"`
	TotalSpent         float64                 `json:"total_spent"`
	TotalEarned        float64                 `json:"total_earned"`
	NetAmount          float64                 `json:"net_amount"`
	TransactionCount   int64                   `json:"transaction_count"`
	CategoryBreakdown  []CategorySpending      `json:"category_breakdown"`
	DailyBreakdown     []DailySpending         `json:"daily_breakdown"`
	PaymentMethodUsage []PaymentMethodSpending `json:"payment_method_usage"`
	TopMerchants       []MerchantSpending      `json:"top_merchants"`
	Currency           string                  `json:"currency"`
}

type DailySpending struct {
	Day    int     `json:"day"`
	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
}

type MerchantSpending struct {
	Merchant string  `json:"merchant"`
	Amount   float64 `json:"amount"`
	Count    int64   `json:"count"`
}

type PromoCode struct {
	Code          string    `json:"code"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	MinAmount     float64   `json:"min_amount"`
	MaxDiscount   float64   `json:"max_discount"`
	UsageLimit    int       `json:"usage_limit"`
	UsedCount     int       `json:"used_count"`
	ExpiresAt     time.Time `json:"expires_at"`
	IsActive      bool      `json:"is_active"`
	ApplicableFor []string  `json:"applicable_for"`
}

type PromoCodeResult struct {
	Valid          bool    `json:"valid"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalAmount    float64 `json:"final_amount"`
	Message        string  `json:"message"`
	PromoCode      string  `json:"promo_code"`
}

type UserCredits struct {
	UserID           string             `json:"user_id"`
	TotalCredits     float64            `json:"total_credits"`
	AvailableCredits float64            `json:"available_credits"`
	PendingCredits   float64            `json:"pending_credits"`
	CreditTypes      []CreditTypeAmount `json:"credit_types"`
	ExpiringCredits  []ExpiringCredit   `json:"expiring_credits"`
	Currency         string             `json:"currency"`
}

type CreditTypeAmount struct {
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

type ExpiringCredit struct {
	Amount    float64   `json:"amount"`
	ExpiresAt time.Time `json:"expires_at"`
	Type      string    `json:"type"`
}

type CreditResult struct {
	Applied          bool    `json:"applied"`
	AmountApplied    float64 `json:"amount_applied"`
	RemainingCredits float64 `json:"remaining_credits"`
	Message          string  `json:"message"`
}

type Invoice struct {
	ID            primitive.ObjectID `json:"id"`
	InvoiceNumber string             `json:"invoice_number"`
	UserID        primitive.ObjectID `json:"user_id"`
	Amount        float64            `json:"amount"`
	Currency      string             `json:"currency"`
	Status        string             `json:"status"`
	DueDate       time.Time          `json:"due_date"`
	IssuedDate    time.Time          `json:"issued_date"`
	PaidDate      *time.Time         `json:"paid_date,omitempty"`
	Description   string             `json:"description"`
	LineItems     []InvoiceLineItem  `json:"line_items"`
	TaxAmount     float64            `json:"tax_amount"`
	TotalAmount   float64            `json:"total_amount"`
}

type InvoiceLineItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

// paymentService implements PaymentService interface
type paymentService struct {
	paymentRepo  repositories.PaymentRepository
	userRepo     repositories.UserRepository
	rideRepo     repositories.RideRepository
	emailService EmailService
	logger       utils.Logger
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	userRepo repositories.UserRepository,
	rideRepo repositories.RideRepository,
	emailService EmailService,
) PaymentService {
	return &paymentService{
		paymentRepo:  paymentRepo,
		userRepo:     userRepo,
		rideRepo:     rideRepo,
		emailService: emailService,
		logger:       utils.ServiceLogger("payment"),
	}
}

// Payment Methods

func (s *paymentService) GetPaymentMethods(userID string) ([]*models.PaymentMethod, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return s.paymentRepo.GetPaymentMethodsByUserID(userObjID)
}

func (s *paymentService) AddPaymentMethod(userID string, reqInterface interface{}) (*models.PaymentMethod, error) {
	req, ok := reqInterface.(*AddPaymentMethodRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate user exists
	_, err = s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create payment method
	paymentMethod := &models.PaymentMethod{
		ID:            primitive.NewObjectID(),
		UserID:        userObjID,
		Type:          req.Type,
		CardDetails:   req.CardDetails,
		BankDetails:   req.BankDetails,
		WalletDetails: req.WalletDetails,
		IsDefault:     req.IsDefault,
		IsActive:      true,
		Nickname:      req.Nickname,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Validate payment method details
	if err := s.validatePaymentMethodDetails(paymentMethod); err != nil {
		return nil, fmt.Errorf("invalid payment method: %w", err)
	}

	// If this is the first payment method or set as default, make it default
	existingMethods, _ := s.paymentRepo.GetPaymentMethodsByUserID(userObjID)
	if len(existingMethods) == 0 || req.IsDefault {
		// Set all other methods as non-default
		for _, method := range existingMethods {
			if method.IsDefault {
				method.IsDefault = false
				s.paymentRepo.UpdatePaymentMethod(method)
			}
		}
		paymentMethod.IsDefault = true
	}

	// Save payment method
	createdMethod, err := s.paymentRepo.CreatePaymentMethod(paymentMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID).
		Str("method_id", createdMethod.ID.Hex()).
		Str("type", string(createdMethod.Type)).
		Msg("Payment method added successfully")

	return createdMethod, nil
}

func (s *paymentService) UpdatePaymentMethod(userID string, methodID primitive.ObjectID, reqInterface interface{}) (*models.PaymentMethod, error) {
	req, ok := reqInterface.(*UpdatePaymentMethodRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	// Get existing payment method
	method, err := s.paymentRepo.GetPaymentMethodByID(methodID)
	if err != nil {
		return nil, fmt.Errorf("payment method not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)
	if method.UserID != userObjID {
		return nil, errors.New("unauthorized access to payment method")
	}

	// Update fields
	if req.Nickname != "" {
		method.Nickname = req.Nickname
	}
	if req.IsActive != nil {
		method.IsActive = *req.IsActive
	}
	if req.BillingAddress != nil {
		if method.CardDetails != nil {
			method.CardDetails.BillingAddress = *req.BillingAddress
		}
	}
	if req.ExpiryMonth != nil && req.ExpiryYear != nil && method.CardDetails != nil {
		method.CardDetails.ExpiryMonth = *req.ExpiryMonth
		method.CardDetails.ExpiryYear = *req.ExpiryYear
	}

	method.UpdatedAt = time.Now()

	return s.paymentRepo.UpdatePaymentMethod(method)
}

func (s *paymentService) DeletePaymentMethod(userID string, methodID primitive.ObjectID) error {
	// Get existing payment method
	method, err := s.paymentRepo.GetPaymentMethodByID(methodID)
	if err != nil {
		return fmt.Errorf("payment method not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)
	if method.UserID != userObjID {
		return errors.New("unauthorized access to payment method")
	}

	// Check if this is the only payment method
	methods, err := s.paymentRepo.GetPaymentMethodsByUserID(userObjID)
	if err != nil {
		return fmt.Errorf("failed to get payment methods: %w", err)
	}

	if len(methods) == 1 {
		return errors.New("cannot delete the only payment method")
	}

	// If this is the default method, set another as default
	if method.IsDefault {
		for _, m := range methods {
			if m.ID != methodID && m.IsActive {
				m.IsDefault = true
				s.paymentRepo.UpdatePaymentMethod(m)
				break
			}
		}
	}

	return s.paymentRepo.DeletePaymentMethod(methodID)
}

func (s *paymentService) SetDefaultPaymentMethod(userID string, methodID primitive.ObjectID) error {
	// Get existing payment method
	method, err := s.paymentRepo.GetPaymentMethodByID(methodID)
	if err != nil {
		return fmt.Errorf("payment method not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)
	if method.UserID != userObjID {
		return errors.New("unauthorized access to payment method")
	}

	// Set all other methods as non-default
	methods, err := s.paymentRepo.GetPaymentMethodsByUserID(userObjID)
	if err != nil {
		return fmt.Errorf("failed to get payment methods: %w", err)
	}

	for _, m := range methods {
		if m.IsDefault {
			m.IsDefault = false
			s.paymentRepo.UpdatePaymentMethod(m)
		}
	}

	// Set this method as default
	method.IsDefault = true
	_, err = s.paymentRepo.UpdatePaymentMethod(method)
	return err
}

// Wallet Management

func (s *paymentService) GetWallet(userID string) (*models.Wallet, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	wallet, err := s.paymentRepo.GetWalletByUserID(userObjID)
	if err != nil {
		// Create wallet if it doesn't exist
		wallet = &models.Wallet{
			ID:        primitive.NewObjectID(),
			UserID:    userObjID,
			Balance:   0.0,
			Currency:  "USD",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		wallet, err = s.paymentRepo.CreateWallet(wallet)
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet: %w", err)
		}
	}

	return wallet, nil
}

func (s *paymentService) AddMoneyToWallet(userID string, amount float64, paymentMethodID, currency string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	// Validate payment method
	paymentMethod, err := s.ValidatePaymentMethod(userID, paymentMethodID)
	if err != nil {
		return nil, err
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)

	// Create transaction
	transaction := &models.Transaction{
		ID:              primitive.NewObjectID(),
		UserID:          userObjID,
		Type:            models.TransactionTypeWalletTopup,
		Status:          models.PaymentStatusPending,
		Amount:          amount,
		Currency:        currency,
		Description:     fmt.Sprintf("Add money to wallet: $%.2f", amount),
		PaymentMethodID: &paymentMethod.ID,
		PaymentMethod:   paymentMethod.Type,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Process payment through external provider
	err = s.processExternalPayment(transaction, paymentMethod)
	if err != nil {
		transaction.Status = models.PaymentStatusFailed
		transaction.FailureReason = err.Error()
		s.paymentRepo.CreateTransaction(transaction)
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	// Update transaction status
	transaction.Status = models.PaymentStatusCompleted
	transaction.ProcessedAt = &[]time.Time{time.Now()}[0]

	// Save transaction
	savedTransaction, err := s.paymentRepo.CreateTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// Update wallet balance
	_, err = s.CreditWallet(userID, amount, transaction.Description, models.TransactionTypeWalletTopup)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Float64("amount", amount).Msg("Failed to credit wallet")
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID).
		Float64("amount", amount).
		Str("transaction_id", savedTransaction.ID.Hex()).
		Msg("Money added to wallet successfully")

	return savedTransaction, nil
}

func (s *paymentService) WithdrawFromWallet(userID string, amount float64, bankAccount, currency, reason string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	// Check wallet balance
	wallet, err := s.GetWallet(userID)
	if err != nil {
		return nil, err
	}

	if wallet.Balance < amount {
		return nil, errors.New("insufficient wallet balance")
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)

	// Create transaction
	transaction := &models.Transaction{
		ID:          primitive.NewObjectID(),
		UserID:      userObjID,
		Type:        models.TransactionTypeWalletWithdraw,
		Status:      models.PaymentStatusPending,
		Amount:      amount,
		Currency:    currency,
		Description: fmt.Sprintf("Withdraw from wallet: $%.2f - %s", amount, reason),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Debit wallet first
	_, err = s.DebitWallet(userID, amount, transaction.Description, models.TransactionTypeWalletWithdraw)
	if err != nil {
		return nil, fmt.Errorf("failed to debit wallet: %w", err)
	}

	// Process withdrawal through external provider
	err = s.processExternalWithdrawal(transaction, bankAccount)
	if err != nil {
		// Refund wallet on failure
		s.CreditWallet(userID, amount, "Withdrawal failed - refund", models.TransactionTypeWalletTopup)
		transaction.Status = models.PaymentStatusFailed
		transaction.FailureReason = err.Error()
		s.paymentRepo.CreateTransaction(transaction)
		return nil, fmt.Errorf("withdrawal failed: %w", err)
	}

	// Update transaction status
	transaction.Status = models.PaymentStatusProcessing
	transaction.ProcessedAt = &[]time.Time{time.Now()}[0]

	// Save transaction
	savedTransaction, err := s.paymentRepo.CreateTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID).
		Float64("amount", amount).
		Str("transaction_id", savedTransaction.ID.Hex()).
		Msg("Withdrawal initiated successfully")

	return savedTransaction, nil
}

func (s *paymentService) GetWalletTransactions(userID string, page, limit int, transactionType string) ([]*models.Transaction, int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	filter := map[string]interface{}{
		"user_id": userObjID,
		"type": map[string]interface{}{
			"$in": []models.TransactionType{
				models.TransactionTypeWalletTopup,
				models.TransactionTypeWalletWithdraw,
			},
		},
	}

	if transactionType != "" {
		filter["type"] = models.TransactionType(transactionType)
	}

	return s.paymentRepo.GetTransactionsByFilter(filter, page, limit)
}

// Payment Processing

func (s *paymentService) ProcessPayment(userID string, reqInterface interface{}) (*models.Transaction, error) {
	req, ok := reqInterface.(*ProcessPaymentRequest)
	if !ok {
		return nil, errors.New("invalid request type")
	}

	// Validate payment method
	paymentMethod, err := s.ValidatePaymentMethod(userID, req.PaymentMethodID)
	if err != nil {
		return nil, err
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)

	// Create transaction
	transaction := &models.Transaction{
		ID:              primitive.NewObjectID(),
		UserID:          userObjID,
		Type:            req.Type,
		Status:          models.PaymentStatusPending,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Description:     req.Description,
		PaymentMethodID: &paymentMethod.ID,
		PaymentMethod:   paymentMethod.Type,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if req.RideID != nil {
		rideObjID, err := primitive.ObjectIDFromHex(*req.RideID)
		if err == nil {
			transaction.RideID = &rideObjID
		}
	}

	// Process payment based on method type
	switch paymentMethod.Type {
	case models.PaymentTypeWallet:
		err = s.processWalletPayment(transaction)
	case models.PaymentTypeCard:
		err = s.processCardPayment(transaction, paymentMethod)
	default:
		err = s.processExternalPayment(transaction, paymentMethod)
	}

	if err != nil {
		transaction.Status = models.PaymentStatusFailed
		transaction.FailureReason = err.Error()
		s.paymentRepo.CreateTransaction(transaction)
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	// Update transaction status
	transaction.Status = models.PaymentStatusCompleted
	transaction.ProcessedAt = &[]time.Time{time.Now()}[0]

	// Save transaction
	savedTransaction, err := s.paymentRepo.CreateTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID).
		Float64("amount", req.Amount).
		Str("transaction_id", savedTransaction.ID.Hex()).
		Msg("Payment processed successfully")

	return savedTransaction, nil
}

func (s *paymentService) PayForRide(userID string, rideID primitive.ObjectID, paymentMethodID string, tipAmount float64) (*models.Transaction, error) {
	// Get ride details
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, fmt.Errorf("ride not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)
	if ride.PassengerID != userObjID {
		return nil, errors.New("unauthorized access to ride")
	}

	// Calculate total amount (fare + tip)
	totalAmount := ride.Fare.FinalFare + tipAmount

	// Create payment request
	paymentReq := &ProcessPaymentRequest{
		Amount:          totalAmount,
		Currency:        ride.Fare.Currency,
		PaymentMethodID: paymentMethodID,
		Description:     fmt.Sprintf("Payment for ride %s", rideID.Hex()),
		Type:            models.TransactionTypeRidePayment,
		RideID:          &[]string{rideID.Hex()}[0],
		Metadata: map[string]interface{}{
			"ride_fare":  ride.Fare.FinalFare,
			"tip_amount": tipAmount,
		},
	}

	transaction, err := s.ProcessPayment(userID, paymentReq)
	if err != nil {
		return nil, err
	}

	// Update ride payment status
	// This would typically update the ride status to paid

	return transaction, nil
}

func (s *paymentService) GetRidePaymentStatus(userID string, rideID primitive.ObjectID) (*RidePaymentStatus, error) {
	// Get ride details
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, fmt.Errorf("ride not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)
	if ride.PassengerID != userObjID && ride.DriverID != userObjID {
		return nil, errors.New("unauthorized access to ride")
	}

	// Get payment transaction for this ride
	filter := map[string]interface{}{
		"ride_id": rideID,
		"type":    models.TransactionTypeRidePayment,
	}

	transactions, _, err := s.paymentRepo.GetTransactionsByFilter(filter, 1, 1)
	if err != nil || len(transactions) == 0 {
		return &RidePaymentStatus{
			RideID:        rideID,
			PaymentStatus: models.PaymentStatusPending,
			Amount:        ride.Fare.FinalFare,
			Currency:      ride.Fare.Currency,
			LastUpdated:   ride.UpdatedAt,
		}, nil
	}

	transaction := transactions[0]

	return &RidePaymentStatus{
		RideID:        rideID,
		PaymentStatus: transaction.Status,
		Amount:        ride.Fare.FinalFare,
		Currency:      ride.Fare.Currency,
		PaymentMethod: transaction.PaymentMethod,
		TransactionID: &transaction.ID,
		LastUpdated:   transaction.UpdatedAt,
		FailureReason: transaction.FailureReason,
		TipAmount:     getTipAmount(transaction.Metadata),
		TotalAmount:   transaction.Amount,
	}, nil
}

// Helper methods

func (s *paymentService) validatePaymentMethodDetails(method *models.PaymentMethod) error {
	switch method.Type {
	case models.PaymentTypeCard:
		if method.CardDetails == nil {
			return errors.New("card details required for card payment method")
		}
		// Additional card validation logic
	case models.PaymentTypeBankTransfer:
		if method.BankDetails == nil {
			return errors.New("bank details required for bank transfer payment method")
		}
		// Additional bank validation logic
	case models.PaymentTypeWallet:
		// Wallet validation if needed
	}
	return nil
}

func (s *paymentService) ValidatePaymentMethod(userID string, paymentMethodID string) (*models.PaymentMethod, error) {
	methodObjID, err := primitive.ObjectIDFromHex(paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method ID: %w", err)
	}

	method, err := s.paymentRepo.GetPaymentMethodByID(methodObjID)
	if err != nil {
		return nil, fmt.Errorf("payment method not found: %w", err)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)
	if method.UserID != userObjID {
		return nil, errors.New("unauthorized access to payment method")
	}

	if !method.IsActive {
		return nil, errors.New("payment method is not active")
	}

	return method, nil
}

func (s *paymentService) DebitWallet(userID string, amount float64, description string, transactionType models.TransactionType) (*models.Transaction, error) {
	wallet, err := s.GetWallet(userID)
	if err != nil {
		return nil, err
	}

	if wallet.Balance < amount {
		return nil, errors.New("insufficient wallet balance")
	}

	// Update wallet balance
	wallet.Balance -= amount
	wallet.UpdatedAt = time.Now()

	_, err = s.paymentRepo.UpdateWallet(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Create transaction record
	userObjID, _ := primitive.ObjectIDFromHex(userID)
	transaction := &models.Transaction{
		ID:            primitive.NewObjectID(),
		UserID:        userObjID,
		Type:          transactionType,
		Status:        models.PaymentStatusCompleted,
		Amount:        -amount, // Negative for debit
		Currency:      wallet.Currency,
		Description:   description,
		PaymentMethod: models.PaymentTypeWallet,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return s.paymentRepo.CreateTransaction(transaction)
}

func (s *paymentService) CreditWallet(userID string, amount float64, description string, transactionType models.TransactionType) (*models.Transaction, error) {
	wallet, err := s.GetWallet(userID)
	if err != nil {
		return nil, err
	}

	// Update wallet balance
	wallet.Balance += amount
	wallet.UpdatedAt = time.Now()

	_, err = s.paymentRepo.UpdateWallet(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Create transaction record
	userObjID, _ := primitive.ObjectIDFromHex(userID)
	transaction := &models.Transaction{
		ID:            primitive.NewObjectID(),
		UserID:        userObjID,
		Type:          transactionType,
		Status:        models.PaymentStatusCompleted,
		Amount:        amount, // Positive for credit
		Currency:      wallet.Currency,
		Description:   description,
		PaymentMethod: models.PaymentTypeWallet,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return s.paymentRepo.CreateTransaction(transaction)
}

// Placeholder implementations for remaining methods...
// These would be implemented with full business logic

func (s *paymentService) processWalletPayment(transaction *models.Transaction) error {
	// Debit wallet for payment
	userID := transaction.UserID.Hex()
	_, err := s.DebitWallet(userID, transaction.Amount, transaction.Description, transaction.Type)
	return err
}

func (s *paymentService) processCardPayment(transaction *models.Transaction, method *models.PaymentMethod) error {
	// Process card payment through payment gateway
	// This would integrate with Stripe, PayPal, etc.
	return nil // Simplified implementation
}

func (s *paymentService) processExternalPayment(transaction *models.Transaction, method *models.PaymentMethod) error {
	// Process payment through external provider based on method type
	// This would integrate with various payment gateways
	return nil // Simplified implementation
}

func (s *paymentService) processExternalWithdrawal(transaction *models.Transaction, bankAccount string) error {
	// Process withdrawal to bank account
	// This would integrate with banking APIs
	return nil // Simplified implementation
}

func getTipAmount(metadata map[string]interface{}) float64 {
	if metadata != nil {
		if tip, ok := metadata["tip_amount"]; ok {
			if tipFloat, ok := tip.(float64); ok {
				return tipFloat
			}
		}
	}
	return 0.0
}

// Placeholder implementations for remaining interface methods
func (s *paymentService) GetPaymentHistory(userID string, page, limit int, transactionType, status, startDate, endDate string) ([]*models.Transaction, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (s *paymentService) GetPaymentReceipt(userID string, paymentID primitive.ObjectID) (*PaymentReceipt, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) SendReceiptEmail(userID string, paymentID primitive.ObjectID, email string) error {
	return errors.New("not implemented")
}

func (s *paymentService) RequestRefund(userID string, req interface{}) (*models.Refund, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetRefunds(userID string, page, limit int, status string) ([]*models.Refund, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (s *paymentService) GetRefundStatus(userID string, refundID primitive.ObjectID) (*models.Refund, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) CreateDispute(userID string, req interface{}) (*models.PaymentDispute, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetDisputes(userID string, page, limit int, status string) ([]*models.PaymentDispute, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (s *paymentService) GetInvoices(userID string, page, limit int, status string) ([]*Invoice, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (s *paymentService) GetInvoice(userID string, invoiceID primitive.ObjectID) (*Invoice, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) PayInvoice(userID string, invoiceID primitive.ObjectID, paymentMethodID string) (*models.Transaction, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) VerifyCard(userID string, req interface{}) (*CardVerificationResult, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetSupportedBanks(country string) ([]*SupportedBank, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) AddBankAccount(userID string, req interface{}) (*models.BankDetails, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetBankAccounts(userID string) ([]*models.BankDetails, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetSpendingSummary(userID string, period, category string) (*SpendingSummary, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetMonthlyReport(userID string, year, month int) (*MonthlyReport, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) HandleStripeWebhook(body []byte, signature string) error {
	return errors.New("not implemented")
}

func (s *paymentService) HandlePayPalWebhook(body []byte, headers map[string]string) error {
	return errors.New("not implemented")
}

func (s *paymentService) GetPromoCodes(userID string) ([]*PromoCode, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) ApplyPromoCode(userID, code string, amount float64) (*PromoCodeResult, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetCredits(userID string) (*UserCredits, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) ApplyCredits(userID string, amount float64, creditType string) (*CreditResult, error) {
	return nil, errors.New("not implemented")
}

// Additional request types (matching controller requests)
type AddPaymentMethodRequest struct {
	Type          models.PaymentMethodType `json:"type"`
	CardDetails   *models.CardDetails      `json:"card_details,omitempty"`
	BankDetails   *models.BankDetails      `json:"bank_details,omitempty"`
	WalletDetails *models.WalletDetails    `json:"wallet_details,omitempty"`
	IsDefault     bool                     `json:"is_default"`
	Nickname      string                   `json:"nickname,omitempty"`
}

type UpdatePaymentMethodRequest struct {
	Nickname       string          `json:"nickname,omitempty"`
	IsActive       *bool           `json:"is_active,omitempty"`
	ExpiryMonth    *int            `json:"expiry_month,omitempty"`
	ExpiryYear     *int            `json:"expiry_year,omitempty"`
	BillingAddress *models.Address `json:"billing_address,omitempty"`
}

type ProcessPaymentRequest struct {
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency,omitempty"`
	PaymentMethodID string                 `json:"payment_method_id"`
	Description     string                 `json:"description"`
	Type            models.TransactionType `json:"type"`
	RideID          *string                `json:"ride_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// EmailService interface for sending emails
type EmailService interface {
	SendReceiptEmail(to, subject string, receiptData interface{}) error
}

// RideRepository interface for ride operations
type RideRepository interface {
	GetByID(id primitive.ObjectID) (*models.Ride, error)
}

// Simplified models for compatibility
type Ride struct {
	ID          primitive.ObjectID `json:"id"`
	PassengerID primitive.ObjectID `json:"passenger_id"`
	DriverID    primitive.ObjectID `json:"driver_id"`
	Fare        RideFare           `json:"fare"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type RideFare struct {
	FinalFare float64 `json:"final_fare"`
	Currency  string  `json:"currency"`
}
