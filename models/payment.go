package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentMethodType string

const (
	PaymentTypeCard         PaymentMethodType = "card"
	PaymentTypeWallet       PaymentMethodType = "wallet"
	PaymentTypeCash         PaymentMethodType = "cash"
	PaymentTypeBankTransfer PaymentMethodType = "bank_transfer"
	PaymentTypePayPal       PaymentMethodType = "paypal"
	PaymentTypeApplePay     PaymentMethodType = "apple_pay"
	PaymentTypeGooglePay    PaymentMethodType = "google_pay"
	PaymentTypeCrypto       PaymentMethodType = "crypto"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusDisputed   PaymentStatus = "disputed"
)

type TransactionType string

const (
	TransactionTypeRidePayment    TransactionType = "ride_payment"
	TransactionTypeWalletTopup    TransactionType = "wallet_topup"
	TransactionTypeWalletWithdraw TransactionType = "wallet_withdraw"
	TransactionTypeRefund         TransactionType = "refund"
	TransactionTypeCommission     TransactionType = "commission"
	TransactionTypePayout         TransactionType = "payout"
	TransactionTypeTip            TransactionType = "tip"
	TransactionTypePenalty        TransactionType = "penalty"
	TransactionTypeBonus          TransactionType = "bonus"
)

type RefundStatus string

const (
	RefundStatusRequested RefundStatus = "requested"
	RefundStatusApproved  RefundStatus = "approved"
	RefundStatusRejected  RefundStatus = "rejected"
	RefundStatusProcessed RefundStatus = "processed"
	RefundStatusCompleted RefundStatus = "completed"
)

type PaymentMethod struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	Type   PaymentMethodType  `json:"type" bson:"type"`

	// Card Details (encrypted)
	CardDetails *CardDetails `json:"card_details,omitempty" bson:"card_details,omitempty"`

	// Bank Account Details
	BankDetails *BankDetails `json:"bank_details,omitempty" bson:"bank_details,omitempty"`

	// Digital Wallet Details
	WalletDetails *WalletDetails `json:"wallet_details,omitempty" bson:"wallet_details,omitempty"`

	// External Provider
	ProviderID string `json:"provider_id" bson:"provider_id"` // Stripe customer ID, etc.
	ExternalID string `json:"external_id" bson:"external_id"` // Payment method ID from provider

	// Status & Settings
	IsDefault  bool `json:"is_default" bson:"is_default"`
	IsActive   bool `json:"is_active" bson:"is_active"`
	IsVerified bool `json:"is_verified" bson:"is_verified"`

	// Metadata
	Nickname   string     `json:"nickname" bson:"nickname"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" bson:"last_used_at,omitempty"`
	UsageCount int        `json:"usage_count" bson:"usage_count"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type CardDetails struct {
	// Card Information (masked for security)
	MaskedNumber string `json:"masked_number" bson:"masked_number"` // **** **** **** 1234
	Brand        string `json:"brand" bson:"brand"`                 // visa, mastercard, amex
	ExpiryMonth  int    `json:"expiry_month" bson:"expiry_month"`
	ExpiryYear   int    `json:"expiry_year" bson:"expiry_year"`
	Fingerprint  string `json:"fingerprint" bson:"fingerprint"` // Unique card fingerprint

	// Billing Address
	BillingAddress Address `json:"billing_address" bson:"billing_address"`

	// Verification
	CVVVerified     bool `json:"cvv_verified" bson:"cvv_verified"`
	AddressVerified bool `json:"address_verified" bson:"address_verified"`

	// Card Type
	CardType      string `json:"card_type" bson:"card_type"` // credit, debit, prepaid
	IssuerCountry string `json:"issuer_country" bson:"issuer_country"`
	IssuerBank    string `json:"issuer_bank" bson:"issuer_bank"`
}

type BankDetails struct {
	AccountHolderName string `json:"account_holder_name" bson:"account_holder_name"`
	BankName          string `json:"bank_name" bson:"bank_name"`
	AccountNumber     string `json:"account_number" bson:"account_number"` // Masked
	RoutingNumber     string `json:"routing_number" bson:"routing_number"`
	AccountType       string `json:"account_type" bson:"account_type"` // checking, savings
	Currency          string `json:"currency" bson:"currency"`
	Country           string `json:"country" bson:"country"`
	IBAN              string `json:"iban" bson:"iban"`
	SWIFTCode         string `json:"swift_code" bson:"swift_code"`
}

type WalletDetails struct {
	Balance           float64    `json:"balance" bson:"balance"`
	Currency          string     `json:"currency" bson:"currency"`
	AvailableBalance  float64    `json:"available_balance" bson:"available_balance"`
	PendingBalance    float64    `json:"pending_balance" bson:"pending_balance"`
	FrozenBalance     float64    `json:"frozen_balance" bson:"frozen_balance"`
	LastTransactionAt *time.Time `json:"last_transaction_at,omitempty" bson:"last_transaction_at,omitempty"`

	// Limits
	DailyLimit   float64 `json:"daily_limit" bson:"daily_limit"`
	MonthlyLimit float64 `json:"monthly_limit" bson:"monthly_limit"`
	MinBalance   float64 `json:"min_balance" bson:"min_balance"`
	MaxBalance   float64 `json:"max_balance" bson:"max_balance"`

	// Auto top-up
	AutoTopupEnabled   bool    `json:"auto_topup_enabled" bson:"auto_topup_enabled"`
	AutoTopupAmount    float64 `json:"auto_topup_amount" bson:"auto_topup_amount"`
	AutoTopupThreshold float64 `json:"auto_topup_threshold" bson:"auto_topup_threshold"`
}

type Address struct {
	Street1    string `json:"street1" bson:"street1"`
	Street2    string `json:"street2" bson:"street2"`
	City       string `json:"city" bson:"city"`
	State      string `json:"state" bson:"state"`
	Country    string `json:"country" bson:"country"`
	PostalCode string `json:"postal_code" bson:"postal_code"`
}

type Transaction struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	// Participants
	UserID   primitive.ObjectID  `json:"user_id" bson:"user_id"`
	DriverID *primitive.ObjectID `json:"driver_id,omitempty" bson:"driver_id,omitempty"`
	RideID   *primitive.ObjectID `json:"ride_id,omitempty" bson:"ride_id,omitempty"`

	// Transaction Details
	Type        TransactionType `json:"type" bson:"type"`
	Status      PaymentStatus   `json:"status" bson:"status"`
	Amount      float64         `json:"amount" bson:"amount"`
	Currency    string          `json:"currency" bson:"currency"`
	Description string          `json:"description" bson:"description"`

	// Payment Information
	PaymentMethodID *primitive.ObjectID `json:"payment_method_id,omitempty" bson:"payment_method_id,omitempty"`
	PaymentMethod   PaymentMethodType   `json:"payment_method" bson:"payment_method"`

	// External References
	ExternalID string `json:"external_id" bson:"external_id"` // Stripe payment intent ID
	ProviderID string `json:"provider_id" bson:"provider_id"` // stripe, paypal, etc.
	InvoiceID  string `json:"invoice_id" bson:"invoice_id"`
	ReceiptURL string `json:"receipt_url" bson:"receipt_url"`

	// Breakdown
	Breakdown TransactionBreakdown `json:"breakdown" bson:"breakdown"`

	// Fees & Commission
	PlatformFee   float64 `json:"platform_fee" bson:"platform_fee"`
	ProcessingFee float64 `json:"processing_fee" bson:"processing_fee"`
	Commission    float64 `json:"commission" bson:"commission"`
	TaxAmount     float64 `json:"tax_amount" bson:"tax_amount"`
	NetAmount     float64 `json:"net_amount" bson:"net_amount"`

	// Timing
	ProcessedAt *time.Time `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty" bson:"failed_at,omitempty"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty" bson:"refunded_at,omitempty"`

	// Error Information
	FailureReason string `json:"failure_reason" bson:"failure_reason"`
	FailureCode   string `json:"failure_code" bson:"failure_code"`

	// Refund Information
	RefundAmount float64 `json:"refund_amount" bson:"refund_amount"`
	RefundReason string  `json:"refund_reason" bson:"refund_reason"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata" bson:"metadata"`
	Notes    string                 `json:"notes" bson:"notes"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type TransactionBreakdown struct {
	BaseAmount      float64 `json:"base_amount" bson:"base_amount"`
	TipAmount       float64 `json:"tip_amount" bson:"tip_amount"`
	TaxAmount       float64 `json:"tax_amount" bson:"tax_amount"`
	DiscountAmount  float64 `json:"discount_amount" bson:"discount_amount"`
	SurchargeAmount float64 `json:"surcharge_amount" bson:"surcharge_amount"`
	ServiceFee      float64 `json:"service_fee" bson:"service_fee"`
	ProcessingFee   float64 `json:"processing_fee" bson:"processing_fee"`
	TotalAmount     float64 `json:"total_amount" bson:"total_amount"`
}

type Refund struct {
	ID            primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	TransactionID primitive.ObjectID  `json:"transaction_id" bson:"transaction_id"`
	RideID        *primitive.ObjectID `json:"ride_id,omitempty" bson:"ride_id,omitempty"`
	UserID        primitive.ObjectID  `json:"user_id" bson:"user_id"`

	// Refund Details
	Amount   float64      `json:"amount" bson:"amount"`
	Currency string       `json:"currency" bson:"currency"`
	Reason   string       `json:"reason" bson:"reason"`
	Status   RefundStatus `json:"status" bson:"status"`
	Type     string       `json:"type" bson:"type"` // full, partial

	// External References
	ExternalRefundID string `json:"external_refund_id" bson:"external_refund_id"`
	ProviderID       string `json:"provider_id" bson:"provider_id"`

	// Processing Information
	ProcessedAt   *time.Time `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	FailedAt      *time.Time `json:"failed_at,omitempty" bson:"failed_at,omitempty"`
	FailureReason string     `json:"failure_reason" bson:"failure_reason"`

	// Admin Information
	RequestedBy primitive.ObjectID  `json:"requested_by" bson:"requested_by"`
	ApprovedBy  *primitive.ObjectID `json:"approved_by,omitempty" bson:"approved_by,omitempty"`
	ProcessedBy *primitive.ObjectID `json:"processed_by,omitempty" bson:"processed_by,omitempty"`
	AdminNotes  string              `json:"admin_notes" bson:"admin_notes"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type PaymentDispute struct {
	ID            primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	TransactionID primitive.ObjectID  `json:"transaction_id" bson:"transaction_id"`
	RideID        *primitive.ObjectID `json:"ride_id,omitempty" bson:"ride_id,omitempty"`
	DisputerID    primitive.ObjectID  `json:"disputer_id" bson:"disputer_id"`

	// Dispute Details
	Type        string  `json:"type" bson:"type"` // chargeback, complaint, fraud
	Reason      string  `json:"reason" bson:"reason"`
	Description string  `json:"description" bson:"description"`
	Amount      float64 `json:"amount" bson:"amount"`
	Currency    string  `json:"currency" bson:"currency"`
	Status      string  `json:"status" bson:"status"`

	// Evidence
	Evidence []DisputeEvidence `json:"evidence" bson:"evidence"`

	// External Information
	ExternalDisputeID string `json:"external_dispute_id" bson:"external_dispute_id"`
	ProviderID        string `json:"provider_id" bson:"provider_id"`

	// Resolution
	ResolvedAt      *time.Time          `json:"resolved_at,omitempty" bson:"resolved_at,omitempty"`
	ResolvedBy      *primitive.ObjectID `json:"resolved_by,omitempty" bson:"resolved_by,omitempty"`
	Resolution      string              `json:"resolution" bson:"resolution"`
	ResolutionNotes string              `json:"resolution_notes" bson:"resolution_notes"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type DisputeEvidence struct {
	Type        string    `json:"type" bson:"type"` // document, screenshot, receipt
	URL         string    `json:"url" bson:"url"`
	Description string    `json:"description" bson:"description"`
	UploadedAt  time.Time `json:"uploaded_at" bson:"uploaded_at"`
}

type PromoCode struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Code        string             `json:"code" bson:"code"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`

	// Discount Details
	DiscountType      string  `json:"discount_type" bson:"discount_type"` // percentage, fixed, free_ride
	DiscountValue     float64 `json:"discount_value" bson:"discount_value"`
	MaxDiscountAmount float64 `json:"max_discount_amount" bson:"max_discount_amount"`
	MinPurchaseAmount float64 `json:"min_purchase_amount" bson:"min_purchase_amount"`

	// Usage Limits
	UsageLimit        int `json:"usage_limit" bson:"usage_limit"`
	UsageLimitPerUser int `json:"usage_limit_per_user" bson:"usage_limit_per_user"`
	CurrentUsage      int `json:"current_usage" bson:"current_usage"`

	// Validity
	ValidFrom time.Time `json:"valid_from" bson:"valid_from"`
	ValidTo   time.Time `json:"valid_to" bson:"valid_to"`

	// Restrictions
	ApplicableServices []ServiceType `json:"applicable_services" bson:"applicable_services"`
	ApplicableCities   []string      `json:"applicable_cities" bson:"applicable_cities"`
	FirstRideOnly      bool          `json:"first_ride_only" bson:"first_ride_only"`
	NewUsersOnly       bool          `json:"new_users_only" bson:"new_users_only"`

	// Status
	IsActive bool `json:"is_active" bson:"is_active"`
	IsPublic bool `json:"is_public" bson:"is_public"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
