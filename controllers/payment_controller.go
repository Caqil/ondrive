package controllers

import (
	"net/http"
	"strconv"

	"ondrive/middleware"
	"ondrive/models"
	"ondrive/services"
	"ondrive/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentController handles payment-related HTTP requests
type PaymentController struct {
	paymentService services.PaymentService
	logger         utils.Logger
}

// NewPaymentController creates a new payment controller
func NewPaymentController(paymentService services.PaymentService) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
		logger:         utils.ControllerLogger("payment"),
	}
}

// Request structures
type AddPaymentMethodRequest struct {
	Type          models.PaymentMethodType `json:"type" binding:"required"`
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

type AddMoneyRequest struct {
	Amount          float64 `json:"amount" binding:"required,min=1"`
	PaymentMethodID string  `json:"payment_method_id" binding:"required"`
	Currency        string  `json:"currency,omitempty"`
}

type WithdrawRequest struct {
	Amount      float64 `json:"amount" binding:"required,min=1"`
	BankAccount string  `json:"bank_account" binding:"required"`
	Currency    string  `json:"currency,omitempty"`
	Reason      string  `json:"reason,omitempty"`
}

type ProcessPaymentRequest struct {
	Amount          float64                `json:"amount" binding:"required,min=0"`
	Currency        string                 `json:"currency,omitempty"`
	PaymentMethodID string                 `json:"payment_method_id" binding:"required"`
	Description     string                 `json:"description"`
	Type            models.TransactionType `json:"type" binding:"required"`
	RideID          *string                `json:"ride_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type RefundRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,min=0"`
	Reason        string  `json:"reason" binding:"required"`
	Type          string  `json:"type,omitempty"`
}

type DisputeRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	Reason        string  `json:"reason" binding:"required"`
	Description   string  `json:"description" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,min=0"`
}

type VerifyCardRequest struct {
	CardNumber  string `json:"card_number" binding:"required"`
	ExpiryMonth int    `json:"expiry_month" binding:"required,min=1,max=12"`
	ExpiryYear  int    `json:"expiry_year" binding:"required"`
	CVV         string `json:"cvv" binding:"required"`
}

type AddBankAccountRequest struct {
	BankName      string         `json:"bank_name" binding:"required"`
	AccountNumber string         `json:"account_number" binding:"required"`
	RoutingNumber string         `json:"routing_number" binding:"required"`
	AccountType   string         `json:"account_type" binding:"required"`
	HolderName    string         `json:"holder_name" binding:"required"`
	Address       models.Address `json:"address" binding:"required"`
}

type ApplyPromoCodeRequest struct {
	Code   string  `json:"code" binding:"required"`
	Amount float64 `json:"amount" binding:"required,min=0"`
}

type ApplyCreditsRequest struct {
	Amount float64 `json:"amount" binding:"required,min=0"`
	Type   string  `json:"type,omitempty"`
}

// Payment Methods

func (pc *PaymentController) GetPaymentMethods(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	methods, err := pc.paymentService.GetPaymentMethods(userID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get payment methods")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment methods retrieved successfully", methods)
}

func (pc *PaymentController) AddPaymentMethod(c *gin.Context) {
	var req AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	method, err := pc.paymentService.AddPaymentMethod(userID, &req)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add payment method")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Payment method added successfully", method)
}

func (pc *PaymentController) UpdatePaymentMethod(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment method ID")
		return
	}

	var req UpdatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	method, err := pc.paymentService.UpdatePaymentMethod(userID, objID, &req)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("method_id", id).Msg("Failed to update payment method")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment method updated successfully", method)
}

func (pc *PaymentController) DeletePaymentMethod(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment method ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err = pc.paymentService.DeletePaymentMethod(userID, objID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("method_id", id).Msg("Failed to delete payment method")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment method deleted successfully", nil)
}

func (pc *PaymentController) SetDefaultPaymentMethod(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment method ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err = pc.paymentService.SetDefaultPaymentMethod(userID, objID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("method_id", id).Msg("Failed to set default payment method")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Default payment method set successfully", nil)
}

// Wallet Management

func (pc *PaymentController) GetWallet(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	wallet, err := pc.paymentService.GetWallet(userID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get wallet")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wallet retrieved successfully", wallet)
}

func (pc *PaymentController) AddMoneyToWallet(c *gin.Context) {
	var req AddMoneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	transaction, err := pc.paymentService.AddMoneyToWallet(userID, req.Amount, req.PaymentMethodID, req.Currency)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Float64("amount", req.Amount).Msg("Failed to add money to wallet")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Money added to wallet successfully", transaction)
}

func (pc *PaymentController) WithdrawFromWallet(c *gin.Context) {
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	transaction, err := pc.paymentService.WithdrawFromWallet(userID, req.Amount, req.BankAccount, req.Currency, req.Reason)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Float64("amount", req.Amount).Msg("Failed to withdraw from wallet")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Withdrawal initiated successfully", transaction)
}

func (pc *PaymentController) GetWalletTransactions(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	transactionType := c.Query("type")

	transactions, total, err := pc.paymentService.GetWalletTransactions(userID, page, limit, transactionType)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get wallet transactions")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(page, limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Wallet transactions retrieved successfully", transactions, meta)
}

// Payment Processing

func (pc *PaymentController) ProcessPayment(c *gin.Context) {
	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	transaction, err := pc.paymentService.ProcessPayment(userID, &req)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Float64("amount", req.Amount).Msg("Failed to process payment")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment processed successfully", transaction)
}

func (pc *PaymentController) PayForRide(c *gin.Context) {
	rideID := c.Param("ride_id")
	rideObjID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid ride ID")
		return
	}

	var req struct {
		PaymentMethodID string  `json:"payment_method_id" binding:"required"`
		TipAmount       float64 `json:"tip_amount,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	transaction, err := pc.paymentService.PayForRide(userID, rideObjID, req.PaymentMethodID, req.TipAmount)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("ride_id", rideID).Msg("Failed to pay for ride")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride payment processed successfully", transaction)
}

func (pc *PaymentController) GetRidePaymentStatus(c *gin.Context) {
	rideID := c.Param("ride_id")
	rideObjID, err := primitive.ObjectIDFromHex(rideID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid ride ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	status, err := pc.paymentService.GetRidePaymentStatus(userID, rideObjID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("ride_id", rideID).Msg("Failed to get ride payment status")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Ride payment status retrieved successfully", status)
}

// Payment History

func (pc *PaymentController) GetPaymentHistory(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	transactionType := c.Query("type")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	transactions, total, err := pc.paymentService.GetPaymentHistory(userID, page, limit, transactionType, status, startDate, endDate)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get payment history")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(page, limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Payment history retrieved successfully", transactions, meta)
}

func (pc *PaymentController) GetPaymentReceipt(c *gin.Context) {
	paymentID := c.Param("payment_id")
	paymentObjID, err := primitive.ObjectIDFromHex(paymentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	receipt, err := pc.paymentService.GetPaymentReceipt(userID, paymentObjID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("payment_id", paymentID).Msg("Failed to get payment receipt")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment receipt retrieved successfully", receipt)
}

func (pc *PaymentController) SendReceiptEmail(c *gin.Context) {
	paymentID := c.Param("payment_id")
	paymentObjID, err := primitive.ObjectIDFromHex(paymentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment ID")
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	err = pc.paymentService.SendReceiptEmail(userID, paymentObjID, req.Email)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("payment_id", paymentID).Msg("Failed to send receipt email")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Receipt email sent successfully", nil)
}

// Refunds & Disputes

func (pc *PaymentController) RequestRefund(c *gin.Context) {
	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	refund, err := pc.paymentService.RequestRefund(userID, &req)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("transaction_id", req.TransactionID).Msg("Failed to request refund")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Refund request submitted successfully", refund)
}

func (pc *PaymentController) GetRefunds(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	refunds, total, err := pc.paymentService.GetRefunds(userID, page, limit, status)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get refunds")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(page, limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Refunds retrieved successfully", refunds, meta)
}

func (pc *PaymentController) GetRefundStatus(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid refund ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	refund, err := pc.paymentService.GetRefundStatus(userID, objID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("refund_id", id).Msg("Failed to get refund status")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Refund status retrieved successfully", refund)
}

func (pc *PaymentController) CreateDispute(c *gin.Context) {
	var req DisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	dispute, err := pc.paymentService.CreateDispute(userID, &req)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("transaction_id", req.TransactionID).Msg("Failed to create dispute")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Dispute created successfully", dispute)
}

func (pc *PaymentController) GetDisputes(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	disputes, total, err := pc.paymentService.GetDisputes(userID, page, limit, status)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get disputes")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(page, limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Disputes retrieved successfully", disputes, meta)
}

// Billing & Invoices

func (pc *PaymentController) GetInvoices(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	invoices, total, err := pc.paymentService.GetInvoices(userID, page, limit, status)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get invoices")
		utils.InternalServerErrorResponse(c)
		return
	}

	meta := utils.CalculatePagination(page, limit, total)
	utils.PaginatedResponse(c, http.StatusOK, "Invoices retrieved successfully", invoices, meta)
}

func (pc *PaymentController) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid invoice ID")
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	invoice, err := pc.paymentService.GetInvoice(userID, objID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("invoice_id", id).Msg("Failed to get invoice")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Invoice retrieved successfully", invoice)
}

func (pc *PaymentController) PayInvoice(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid invoice ID")
		return
	}

	var req struct {
		PaymentMethodID string `json:"payment_method_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	transaction, err := pc.paymentService.PayInvoice(userID, objID, req.PaymentMethodID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("invoice_id", id).Msg("Failed to pay invoice")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Invoice paid successfully", transaction)
}

// Cards & Banking

func (pc *PaymentController) VerifyCard(c *gin.Context) {
	var req VerifyCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	result, err := pc.paymentService.VerifyCard(userID, &req)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to verify card")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Card verified successfully", result)
}

func (pc *PaymentController) GetSupportedBanks(c *gin.Context) {
	country := c.Query("country")
	banks, err := pc.paymentService.GetSupportedBanks(country)
	if err != nil {
		pc.logger.Error().Err(err).Msg("Failed to get supported banks")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Supported banks retrieved successfully", banks)
}

func (pc *PaymentController) AddBankAccount(c *gin.Context) {
	var req AddBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	bankAccount, err := pc.paymentService.AddBankAccount(userID, &req)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to add bank account")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Bank account added successfully", bankAccount)
}

func (pc *PaymentController) GetBankAccounts(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	bankAccounts, err := pc.paymentService.GetBankAccounts(userID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get bank accounts")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Bank accounts retrieved successfully", bankAccounts)
}

// Payment Analytics

func (pc *PaymentController) GetSpendingSummary(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	period := c.DefaultQuery("period", "month")
	category := c.Query("category")

	summary, err := pc.paymentService.GetSpendingSummary(userID, period, category)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get spending summary")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Spending summary retrieved successfully", summary)
}

func (pc *PaymentController) GetMonthlyReport(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	month, _ := strconv.Atoi(c.DefaultQuery("month", "1"))

	report, err := pc.paymentService.GetMonthlyReport(userID, year, month)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get monthly report")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Monthly report retrieved successfully", report)
}

// Webhooks

func (pc *PaymentController) StripeWebhook(c *gin.Context) {
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		utils.BadRequestResponse(c, "Missing Stripe signature")
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		utils.BadRequestResponse(c, "Failed to read request body")
		return
	}

	err = pc.paymentService.HandleStripeWebhook(body, signature)
	if err != nil {
		pc.logger.Error().Err(err).Msg("Failed to handle Stripe webhook")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func (pc *PaymentController) PayPalWebhook(c *gin.Context) {
	authAlgo := c.GetHeader("PAYPAL-AUTH-ALGO")
	transmission := c.GetHeader("PAYPAL-TRANSMISSION-ID")
	certID := c.GetHeader("PAYPAL-CERT-ID")
	transmissionSig := c.GetHeader("PAYPAL-TRANSMISSION-SIG")
	transmissionTime := c.GetHeader("PAYPAL-TRANSMISSION-TIME")

	if authAlgo == "" || transmission == "" || certID == "" || transmissionSig == "" || transmissionTime == "" {
		utils.BadRequestResponse(c, "Missing PayPal headers")
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		utils.BadRequestResponse(c, "Failed to read request body")
		return
	}

	err = pc.paymentService.HandlePayPalWebhook(body, map[string]string{
		"auth_algo":         authAlgo,
		"transmission_id":   transmission,
		"cert_id":           certID,
		"transmission_sig":  transmissionSig,
		"transmission_time": transmissionTime,
	})
	if err != nil {
		pc.logger.Error().Err(err).Msg("Failed to handle PayPal webhook")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

// Promo Codes & Credits

func (pc *PaymentController) GetPromoCodes(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	promoCodes, err := pc.paymentService.GetPromoCodes(userID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get promo codes")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Promo codes retrieved successfully", promoCodes)
}

func (pc *PaymentController) ApplyPromoCode(c *gin.Context) {
	var req ApplyPromoCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	result, err := pc.paymentService.ApplyPromoCode(userID, req.Code, req.Amount)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Str("promo_code", req.Code).Msg("Failed to apply promo code")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Promo code applied successfully", result)
}

func (pc *PaymentController) GetCredits(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	credits, err := pc.paymentService.GetCredits(userID)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get credits")
		utils.InternalServerErrorResponse(c)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Credits retrieved successfully", credits)
}

func (pc *PaymentController) ApplyCredits(c *gin.Context) {
	var req ApplyCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, map[string]string{"error": "Invalid request data"})
		return
	}

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c)
		return
	}

	result, err := pc.paymentService.ApplyCredits(userID, req.Amount, req.Type)
	if err != nil {
		pc.logger.Error().Err(err).Str("user_id", userID).Float64("amount", req.Amount).Msg("Failed to apply credits")
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Credits applied successfully", result)
}
