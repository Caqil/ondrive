package routes

import (
	"indrive-backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupPaymentRoutes(rg *gin.RouterGroup, controller *controllers.PaymentController) {
	payments := rg.Group("/payments")
	{
		// Payment Methods
		payments.GET("/methods", controller.GetPaymentMethods)
		payments.POST("/methods", controller.AddPaymentMethod)
		payments.PUT("/methods/:id", controller.UpdatePaymentMethod)
		payments.DELETE("/methods/:id", controller.DeletePaymentMethod)
		payments.POST("/methods/:id/set-default", controller.SetDefaultPaymentMethod)

		// Wallet Management
		payments.GET("/wallet", controller.GetWallet)
		payments.POST("/wallet/add-money", controller.AddMoneyToWallet)
		payments.POST("/wallet/withdraw", controller.WithdrawFromWallet)
		payments.GET("/wallet/transactions", controller.GetWalletTransactions)

		// Payment Processing
		payments.POST("/process", controller.ProcessPayment)
		payments.POST("/rides/:ride_id/pay", controller.PayForRide)
		payments.GET("/rides/:ride_id/payment-status", controller.GetRidePaymentStatus)

		// Payment History
		payments.GET("/history", controller.GetPaymentHistory)
		payments.GET("/receipts/:payment_id", controller.GetPaymentReceipt)
		payments.POST("/receipts/:payment_id/send-email", controller.SendReceiptEmail)

		// Refunds & Disputes
		payments.POST("/refunds", controller.RequestRefund)
		payments.GET("/refunds", controller.GetRefunds)
		payments.GET("/refunds/:id", controller.GetRefundStatus)
		payments.POST("/disputes", controller.CreateDispute)
		payments.GET("/disputes", controller.GetDisputes)

		// Billing & Invoices
		payments.GET("/invoices", controller.GetInvoices)
		payments.GET("/invoices/:id", controller.GetInvoice)
		payments.POST("/invoices/:id/pay", controller.PayInvoice)

		// Cards & Banking
		payments.POST("/cards/verify", controller.VerifyCard)
		payments.GET("/banks", controller.GetSupportedBanks)
		payments.POST("/bank-accounts", controller.AddBankAccount)
		payments.GET("/bank-accounts", controller.GetBankAccounts)

		// Payment Analytics
		payments.GET("/spending-summary", controller.GetSpendingSummary)
		payments.GET("/monthly-report", controller.GetMonthlyReport)

		// Webhooks
		payments.POST("/webhooks/stripe", controller.StripeWebhook)
		payments.POST("/webhooks/paypal", controller.PayPalWebhook)

		// Promo Codes & Credits
		payments.GET("/promo-codes", controller.GetPromoCodes)
		payments.POST("/promo-codes/apply", controller.ApplyPromoCode)
		payments.GET("/credits", controller.GetCredits)
		payments.POST("/credits/apply", controller.ApplyCredits)
	}
}
